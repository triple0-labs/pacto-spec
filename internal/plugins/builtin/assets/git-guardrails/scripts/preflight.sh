#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PLUGIN_DIR=$(dirname "$SCRIPT_DIR")
CONFIG_FILE="$PLUGIN_DIR/config.env"

STRICT_MODE=1
REMOTE_OVERRIDE=
ENABLE_GH_DIAGNOSTICS=1
GH_REPO=

if [ -f "$CONFIG_FILE" ]; then
  # shellcheck disable=SC1090
  . "$CONFIG_FILE"
fi

warn() {
  echo "git-guardrails: $1" >&2
}

fail_or_continue() {
  if [ "${STRICT_MODE:-1}" = "1" ]; then
    warn "$1"
    exit 2
  fi
  warn "$1 (STRICT_MODE=0, continuing)"
  exit 0
}

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  fail_or_continue "not a git repository; run this workflow inside a git worktree"
fi

BRANCH=$(git symbolic-ref --quiet --short HEAD 2>/dev/null || true)
UPSTREAM=$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)
REMOTE=""
COMPARE_REF=""

if [ -n "$UPSTREAM" ]; then
  REMOTE=${UPSTREAM%%/*}
  COMPARE_REF=$UPSTREAM
fi

if [ -n "${REMOTE_OVERRIDE:-}" ]; then
  REMOTE=$REMOTE_OVERRIDE
fi

if [ -z "$REMOTE" ]; then
  REMOTE=origin
fi

if ! git fetch --prune "$REMOTE" >/dev/null 2>&1; then
  fail_or_continue "git fetch --prune $REMOTE failed"
fi

if [ -z "$COMPARE_REF" ] && [ -n "$BRANCH" ]; then
  if git show-ref --verify --quiet "refs/remotes/$REMOTE/$BRANCH"; then
    COMPARE_REF="$REMOTE/$BRANCH"
  fi
fi

if [ -z "$COMPARE_REF" ]; then
  fail_or_continue "no upstream/tracking branch found for comparison; set upstream with: git branch --set-upstream-to=$REMOTE/<branch> <branch>"
fi

CONFLICTS=$(git diff --name-only --diff-filter=U || true)
if [ -n "$CONFLICTS" ]; then
  warn "unresolved merge conflicts detected:"
  echo "$CONFLICTS" >&2
  fail_or_continue "resolve conflicts and stage files before continuing"
fi

COUNTS=$(git rev-list --left-right --count "HEAD...$COMPARE_REF" 2>/dev/null || true)
if [ -z "$COUNTS" ]; then
  fail_or_continue "unable to compare HEAD against $COMPARE_REF"
fi

AHEAD=$(echo "$COUNTS" | awk '{print $1}')
BEHIND=$(echo "$COUNTS" | awk '{print $2}')

if [ "${BEHIND:-0}" -gt 0 ]; then
  if [ "${AHEAD:-0}" -gt 0 ]; then
    fail_or_continue "branch diverged from $COMPARE_REF (ahead=$AHEAD, behind=$BEHIND); rebase/merge before running pacto workflow"
  fi
  fail_or_continue "branch is behind $COMPARE_REF by $BEHIND commit(s); pull/rebase before running pacto workflow"
fi

if [ "${ENABLE_GH_DIAGNOSTICS:-1}" = "1" ]; then
  if ! command -v gh >/dev/null 2>&1; then
    warn "gh CLI not found; skipping PR diagnostics"
  elif ! gh auth status >/dev/null 2>&1; then
    warn "gh CLI not authenticated; run 'gh auth login' to enable PR diagnostics"
  elif [ -n "${GH_REPO:-}" ]; then
    if ! gh pr status --repo "$GH_REPO" >/dev/null 2>&1; then
      warn "gh pr status check failed for repo '$GH_REPO'"
    fi
  elif ! gh pr status >/dev/null 2>&1; then
    warn "gh pr status check failed"
  fi
fi

exit 0
