---
name: pacto-status
description: Agent contract for the Pacto status workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=codex/skill/pacto-status workflow=status contract=v1 template_sha256=18c0cc251bfafabb5a0a4495621711ebf6fdb85ac4e71b0704464e6d92ea8b78 generated_by=pacto generated_at=2026-04-15T16:21:09Z -->
# Pacto Status Skill

Use this skill as an agent contract for the status workflow in Pacto projects.

## Objective

Report plan status, blockers, freshness, and optional path verification.

## When To Use

Use when you need a consolidated metadata-first status report for plans. Add `--verify` only when explicit file path claims need repo verification.

## Input Contract

### Required Inputs
- None when auto-discovery can resolve plans root from current directory or parents.

### Optional Inputs
- `--root <path>` to pin project root used for `.pacto/plans` discovery.
- `--repo-root <path>` to pin optional path verification root.
- `--format table|json`, `--verify`, `--fail-on`, `--state`, `--include-archive`.
- `--mode compat|strict`, `--config`, `--max-next-actions`, `--max-blockers`, `--verbose`.

## Execution Contract

- Tool target: codex
- Recommended command: pacto status --format json

## Output Contract
- Produces metadata-first `table` or `json` report with state summary, freshness, blockers, and next actions.
- When `--verify` is enabled, augments output with explicit file path claim verification results.
- Exit code follows `--fail-on` policy for CI automation.

## Validation Checklist
- Confirm resolved roots are correct for the user's intent.
- Confirm report includes expected plans/states.
- For agent or CI use cases, prefer `--format json`.
- Only enable `--verify` when path verification is intentionally required.

## Failure Modes and Handling
- Root resolution failure when no valid plans root is discoverable.
- Invalid config/flags or unsupported flag values.
- Path verification findings may be incomplete when `--verify` is omitted by design.

## Implementation Status

- Status: **Implemented**
- Fallback: Ask for explicit `--root` and `--repo-root` when auto-discovery fails.
<!-- pacto:managed:end -->
