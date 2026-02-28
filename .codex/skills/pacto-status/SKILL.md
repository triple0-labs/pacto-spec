---
name: pacto-status
description: Agent contract for the Pacto status workflow to report plan state and evidence verification.
---

<!-- pacto:managed:start -->
# Pacto Status Skill

Use this skill as an agent contract for the status workflow in Pacto projects.

## Objective

Verify plan status, blockers, and evidence claims.

## When To Use

Use when you need a consolidated status report for plans and claim verification against repo evidence.

## Input Contract

### Required Inputs
- None when auto-discovery can resolve plans root from current directory or parents.

### Optional Inputs
- `--plans-root <path>` to pin plan discovery root.
- `--repo-root <path>` to pin evidence verification root.
- `--format table|json`, `--fail-on`, `--state`, `--include-archive`.
- `--mode compat|strict`, `--config`, `--max-next-actions`, `--max-blockers`, `--verbose`.

## Execution Contract

- Tool target: codex
- Recommended command: pacto status --format table

## Output Contract
- Produces `table` or `json` report with state summary, blockers, next actions, and verification outcomes.
- Verification classifications are `verified`, `partial`, or `unverified`.
- Exit code follows `--fail-on` policy for CI automation.

## Validation Checklist
- Confirm resolved roots are correct for the user's intent.
- Confirm report includes expected plans/states.
- If CI use case, ensure `--format json` and explicit `--fail-on` are set.

## Failure Modes and Handling
- Root resolution failure when no valid plans root is discoverable.
- Invalid config/flags or unsupported flag values.
- Partial verification due to missing or stale repository evidence.

## Implementation Status

- Status: **Implemented**
- Fallback: Ask for explicit `--plans-root` and `--repo-root` when auto-discovery fails.
<!-- pacto:managed:end -->
