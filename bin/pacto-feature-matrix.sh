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
run_expect "root help no args" 0 "Pacto CLI" "$BIN"
run_expect "help command" 0 "Commands:" "$BIN" help
run_expect "help status" 0 "Command: status" "$BIN" help status
run_expect "help unknown topic" 2 "unknown help topic" "$BIN" help unknown-topic
run_expect "unknown command" 2 "unknown command" "$BIN" does-not-exist
run_expect "deprecated --lang warning" 0 "--lang is deprecated" "$BIN" --lang es version

print_header "init"
run_expect "init basic" 0 "Initialized Pacto workspace" "$BIN" init --root "$PROJECT"
assert_dir "plans root exists" "$PLANS"
assert_dir "state current exists" "$PLANS/current"
assert_dir "state to-implement exists" "$PLANS/to-implement"
assert_dir "state done exists" "$PLANS/done"
assert_dir "state outdated exists" "$PLANS/outdated"
assert_file "plans README exists" "$PLANS/README.md"
assert_file "plans template exists" "$PLANS/PLANTILLA_PACTO_PLAN.md"
run_expect "init idempotent" 0 "Skipped:" "$BIN" init --root "$PROJECT"
run_expect "init with agents" 0 "AGENTS.md" "$BIN" init --root "$PROJECT" --with-agents
assert_file "agents created" "$PROJECT/AGENTS.md"

print_header "new"
run_expect "new invalid state" 2 "invalid state" "$BIN" new invalid slug --root "$PLANS"
run_expect "new invalid slug" 2 "invalid slug" "$BIN" new current BadSlug --root "$PLANS"
run_expect "new missing args" 2 "Usage:" "$BIN" new --root "$PLANS"
run_expect "new help" 0 "Command: new" "$BIN" new --help
run_expect "new creates current plan" 0 "Created plan: current/api-core" "$BIN" new current api-core --root "$PLANS" --title "API Core" --owner "Platform"
run_expect "new duplicate rejected" 2 "plan already exists" "$BIN" new current api-core --root "$PLANS"

print_header "status setup data"
mkdir -p "$PROJECT/src"
cat >"$PROJECT/src/auth.go" <<'EOF'
package src

func ValidateToken() bool { return true }
EOF
cat >"$PLANS/current/api-core/README.md" <<'EOF'
# API Core

**Status:** In Progress

## Next Steps
1. Wire token validator
2. Ship QA deploy

- [ ] Integrate `src/auth.go` in middleware
- [ ] Fix blocked deploy in QA
EOF

PLAN_FILE="$(find "$PLANS/current/api-core" -maxdepth 1 -type f -name 'PLAN_*.md' | head -n 1)"
cat >"$PLAN_FILE" <<'EOF'
# Plan: API Core

**Status:** In Progress

| Fase 1 | Build | In Progress | 50% |

## Evidence
- `src/auth.go`
- `ValidateToken`
- `GET /api/auth/health`
EOF

run_expect "new creates to-implement plan" 0 "Created plan: to-implement/docs-cleanup" "$BIN" new to-implement docs-cleanup --root "$PLANS" --title "Docs cleanup"
PLAN2_FILE="$(find "$PLANS/to-implement/docs-cleanup" -maxdepth 1 -type f -name 'PLAN_*.md' | head -n 1)"
cat >"$PLANS/to-implement/docs-cleanup/README.md" <<'EOF'
# Docs cleanup

**Status:** Pending

## Next Steps
1. Update onboarding docs

- [ ] Refresh README examples
EOF
cat >"$PLAN2_FILE" <<'EOF'
# Plan: Docs cleanup

**Status:** Pending

Progress: 15%
EOF

print_header "status split-root behavior"
run_expect "status json split roots" 0 "\"plans_root\"" "$BIN" status --plans-root "$PLANS" --repo-root "$PROJECT" --format json
run_expect "status deprecated root warning" 0 "deprecated for status" "$BIN" status --root "$PROJECT" --format table
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
run_expect "status with deprecated config root" 0 "deprecated for status" "$BIN" status --config "$CFG_DIR/legacy.yaml"

print_header "exec planned"
run_expect "exec not implemented guidance" 2 "planned but not implemented" "$BIN" exec

print_header "Summary"
printf "Workdir: %s\n" "$WORKDIR"
printf "Passed:  %d\n" "$PASS"
printf "Failed:  %d\n" "$FAIL"

if [[ "$FAIL" -ne 0 ]]; then
  exit 1
fi
