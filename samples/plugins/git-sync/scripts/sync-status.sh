#!/bin/sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PLUGIN_DIR=$(dirname "$SCRIPT_DIR")
CONFIG_FILE="$PLUGIN_DIR/config.env"

PULL_ON_STATUS=0
STRICT_MODE=0

if [ -f "$CONFIG_FILE" ]; then
  # shellcheck disable=SC1090
  . "$CONFIG_FILE"
fi

warn() {
  echo "git-sync: $1" >&2
}

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  warn "not a git repository; skipping sync"
  exit 0
fi

UPSTREAM=$(git rev-parse --abbrev-ref --symbolic-full-name '@{upstream}' 2>/dev/null || true)
if [ -z "$UPSTREAM" ]; then
  warn "no upstream configured for current branch; set it with git branch --set-upstream-to=origin/<branch> <branch>"
  exit 0
fi

REMOTE=${UPSTREAM%%/*}
BRANCH=${UPSTREAM#*/}

if ! git fetch --prune "$REMOTE" >/dev/null 2>&1; then
  if [ "${STRICT_MODE:-0}" = "1" ]; then
    warn "fetch failed for remote '$REMOTE' (strict mode)"
    exit 2
  fi
  warn "fetch failed for remote '$REMOTE'; continuing (STRICT_MODE=0)"
  exit 0
fi

if [ "${PULL_ON_STATUS:-0}" = "1" ]; then
  if ! git pull --ff-only "$REMOTE" "$BRANCH" >/dev/null 2>&1; then
    if [ "${STRICT_MODE:-0}" = "1" ]; then
      warn "pull --ff-only failed for '$REMOTE/$BRANCH' (strict mode)"
      exit 2
    fi
    warn "pull --ff-only failed for '$REMOTE/$BRANCH'; continuing (STRICT_MODE=0)"
    exit 0
  fi
fi

exit 0
