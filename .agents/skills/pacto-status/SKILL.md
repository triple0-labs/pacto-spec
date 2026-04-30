---
name: pacto-status
description: Agent contract for the Pacto status workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=agents/skill/pacto-status workflow=status contract=v1 template_sha256=b20794b0702f2520047460183ca867108b237713385215f0774dda82b6aadd69 generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
# Pacto Status Skill

Use this skill as an agent contract for the status workflow in Pacto projects.

## Objective

Report plan status, blockers, freshness, and optional path verification.

## When To Use

Use when you need a consolidated metadata-first status report for plans. Add `--verify` only when explicit file path claims need repo verification. Use `--mode strict` to surface Requirement/Scenario grammar issues before authoring fixes.

## Input Contract

### Required Inputs
- None when auto-discovery can resolve plans root from current directory or parents.

### Optional Inputs
- `--root <path>` to pin project root used for `.pacto/plans` discovery.
- `--repo-root <path>` to pin optional path verification root.
- `--format table|json`, `--verify`, `--fail-on`, `--state`, `--include-archive`.
- `--mode compat|strict`, `--config`, `--max-next-actions`, `--max-blockers`, `--verbose`.

## Execution Contract

- Tool target: Cursor Agent and Codex (project skills under .agents/skills/)
- Recommended command: pacto status --format json

## Output Contract
- Produces metadata-first `table` or `json` report with state summary, freshness, blockers, and next actions.
- When `--verify` is enabled, augments output with explicit file path claim verification results.
- Exit code follows `--fail-on` policy for CI automation.
- For plans that declare structured Requirements, includes per-Requirement coverage: each Requirement reports `tasks=N evidence=M`, and Requirements with zero `R-NNN` references in `tasks.md` are flagged `[uncovered]`. JSON output adds a `requirements[]` array per plan with `id`, `name`, `tasks`, `evidence`, `uncovered` fields.

## Validation Checklist
- Confirm resolved roots are correct for the user's intent.
- Confirm report includes expected plans/states.
- For agent or CI use cases, prefer `--format json`.
- Only enable `--verify` when path verification is intentionally required.
- When `--mode strict` reports `requirement_missing_scenario`, `requirements_grammar`, or `capability_grammar` issues, treat them as authoring bugs in the plan's `spec.md`. Open the file and add the missing `#### Scenario:` block, fix the heading level, or correct the delta op header rather than ignoring the warning.
- When a Requirement is flagged `[uncovered]`, add at least one task in `tasks.md` that references the Requirement ID (e.g. `- [ ] 2.3 implement R-014 sign-out flow`). Optionally add an Evidence line under `## Evidence` referencing the same ID once verified.

## Failure Modes and Handling
- Root resolution failure when no valid plans root is discoverable.
- Invalid config/flags or unsupported flag values.
- Path verification findings may be incomplete when `--verify` is omitted by design.

## Implementation Status

- Status: **Implemented**
- Fallback: Ask for explicit `--root` and `--repo-root` when auto-discovery fails. For Requirement coverage gaps, edit the plan's `tasks.md` and `spec.md` directly — never write to `.pacto/specs/` manually; that baseline is maintained exclusively by `pacto move ... done`.
<!-- pacto:managed:end -->
