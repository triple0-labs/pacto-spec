#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${PACTO_BIN:-/tmp/pacto-feature-matrix-bin}"
WORKDIR="${1:-$(mktemp -d /tmp/pacto-feature-matrix-XXXXXX)}"
PROJECT="$WORKDIR/project"
PLANS="$PROJECT/.pacto/plans"

PASS=0
FAIL=0

print_header() {
  printf "\n== %s ==\n" "$1"
}

record_pass() {
  PASS=$((PASS + 1))
  printf "PASS  %s\n" "$1"
}

record_fail() {
  FAIL=$((FAIL + 1))
  printf "FAIL  %s\n" "$1"
}

contains() {
  local haystack="$1"
  local needle="$2"
  grep -Fq -- "$needle" <<<"$haystack"
}

run_expect() {
  local label="$1"
  local expect_code="$2"
  local needle="${3:-}"
  shift 3
  local out code
  set +e
  out="$("$@" 2>&1)"
  code=$?
  set -e
  if [[ "$code" -ne "$expect_code" ]]; then
    record_fail "$label (exit=$code expected=$expect_code)"
    printf "      output: %s\n" "$(head -n 2 <<<"$out" | tr '\n' ' ')"
    return
  fi
  if [[ -n "$needle" ]] && ! contains "$out" "$needle"; then
    record_fail "$label (missing: $needle)"
    printf "      output: %s\n" "$(head -n 2 <<<"$out" | tr '\n' ' ')"
    return
  fi
  record_pass "$label"
}

assert_file() {
  local label="$1"
  local path="$2"
  if [[ -f "$path" ]]; then
    record_pass "$label"
  else
    record_fail "$label (missing $path)"
  fi
}

assert_dir() {
  local label="$1"
  local path="$2"
  if [[ -d "$path" ]]; then
    record_pass "$label"
  else
    record_fail "$label (missing $path)"
  fi
}

print_header "Build CLI"
(
  cd "$ROOT_DIR"
  go build -o "$BIN" ./cmd/pacto
)
record_pass "Built binary at $BIN"

mkdir -p "$PROJECT"

print_header "CLI Basics"
run_expect "version command" 0 "pacto version" "$BIN" version
run_expect "root help no args" 0 "Pacto is a specialized CLI tool" "$BIN"
run_expect "help command" 0 "Commands:" "$BIN" help
run_expect "help status" 0 "Show current workspace status." "$BIN" help status
run_expect "help unknown topic" 0 "Unknown help topic" "$BIN" help unknown-topic
run_expect "unknown command" 1 "unknown command" "$BIN" does-not-exist
run_expect "version with lang override" 0 "pacto version" "$BIN" --lang es version

print_header "init"
run_expect "init basic" 0 "Workspace Ready" "$BIN" init --root "$PROJECT"
assert_dir "plans root exists" "$PLANS"
assert_dir "state current exists" "$PLANS/current"
assert_dir "state to-implement exists" "$PLANS/to-implement"
assert_dir "state done exists" "$PLANS/done"
assert_dir "state outdated exists" "$PLANS/outdated"
assert_file "plans README exists" "$PLANS/README.md"
assert_file "plans PACTO exists" "$PLANS/PACTO.md"
run_expect "init idempotent" 0 "Workspace Ready" "$BIN" init --root "$PROJECT"
run_expect "init with agents" 0 "AGENTS.md" "$BIN" init --root "$PROJECT" --with-agents
assert_file "agents created" "$PROJECT/AGENTS.md"

print_header "new"
run_expect "new invalid state" 2 "invalid state" "$BIN" new invalid slug --root "$PLANS"
run_expect "new invalid slug" 2 "invalid slug" "$BIN" new current BadSlug --root "$PLANS"
run_expect "new missing args" 2 "Usage:" "$BIN" new --root "$PLANS"
run_expect "new help" 0 "Create a new plan." "$BIN" new --help
run_expect "new creates current plan" 0 "Created Plan" "$BIN" new current api-core --root "$PLANS" --title "API Core" --owner "Platform"
run_expect "new duplicate rejected" 2 "plan already exists" "$BIN" new current api-core --root "$PLANS"

print_header "status setup data"
mkdir -p "$PROJECT/src"
cat >"$PROJECT/src/auth.go" <<'EOF'
package src

func ValidateToken() bool { return true }
EOF
cat >"$PLANS/current/api-core/spec.md" <<'EOF'
# Spec: API Core

## Metadata

- Owner: Platform
- Created: 2026-03-27
- Last Modified: 2026-03-27
- State: current
- Slug: api-core

## Problem Statement

Expose auth health and token validation status for verification flows.

## User Scenarios

### Scenario: health check

- **GIVEN** token middleware is wired
- **WHEN** auth health is requested
- **THEN** the service exposes token validation state

## Acceptance Criteria

- AC-001: Status verification can trace auth health implementation to repository evidence.
EOF
cat >"$PLANS/current/api-core/design.md" <<'EOF'
# Design: API Core

## Metadata

- Owner: Platform
- Created: 2026-03-27
- Last Modified: 2026-03-27
- State: current
- Slug: api-core

## Technical Context

- Language/Version: Go 1.24
- Dependencies: standard library
- Constraints: keep the flow local and testable

## Architecture Decisions

1. Decision: Validate token state in-process | Rationale: keep evidence easy to verify in smoke coverage
EOF
cat >"$PLANS/current/api-core/tasks.md" <<'EOF'
# Tasks: API Core

## Execution Metadata

- Status: Draft
- Owner: Platform
- Created: 2026-03-27
- Last Modified: 2026-03-27
- State: current
- Slug: api-core

## Implementation Plan by Phases

## Phase 1: Build

- [ ] 1.1 Wire token validator

## Evidence

- 2026-03-27 10:00 `src/auth.go`
- 2026-03-27 10:01 `ValidateToken`
- 2026-03-27 10:02 `GET /api/auth/health`

## Blockers

- 2026-03-27 10:03 Fix blocked deploy in QA

## Next Steps

1. Ship QA deploy
EOF

run_expect "new creates to-implement plan" 0 "Created Plan" "$BIN" new to-implement docs-cleanup --root "$PLANS" --title "Docs cleanup"

print_header "status split-root behavior"
run_expect "status json split roots" 0 "\"plans_root\"" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --format json
run_expect "status root mode" 0 "PLANS_ROOT:" "$BIN" status --root "$PROJECT" --format table
run_expect "status state filter current" 0 "api-core" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --state current --format table
run_expect "status strict mode" 0 "MODE: strict" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --mode strict --format table
run_expect "status fail-on none" 0 "" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --fail-on none --format table
run_expect "status fail-on blocked" 1 "" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --fail-on blocked --format table
run_expect "status fail-on unverified" 1 "" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --fail-on unverified --format table
run_expect "status fail-on partial" 1 "" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --fail-on partial --format table

print_header "config split roots"
CFG_DIR="$WORKDIR/cfg"
mkdir -p "$CFG_DIR"
cat >"$CFG_DIR/engine.yaml" <<EOF
plans_root: ../project/.pacto/plans
repo_root: ../project
format: json
EOF
run_expect "status with config split roots" 0 "\"repo_root\"" "$BIN" status --config "$CFG_DIR/engine.yaml"

cat >"$CFG_DIR/legacy.yaml" <<EOF
root: ../project
format: json
EOF
run_expect "status with config root" 0 "\"plans_root\"" "$BIN" status --config "$CFG_DIR/legacy.yaml"

print_header "exec planned"
run_expect "exec missing args" 2 "exec requires <state> <slug>" "$BIN" exec

print_header "Summary"
printf "Workdir: %s\n" "$WORKDIR"
printf "Passed:  %d\n" "$PASS"
printf "Failed:  %d\n" "$FAIL"

if [[ "$FAIL" -ne 0 ]]; then
  exit 1
fi
