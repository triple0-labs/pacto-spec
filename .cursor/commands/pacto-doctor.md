<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/command/pacto-doctor.md workflow=doctor contract=v1 template_sha256=8fa558c0cd3b06f5185abb50a2ed215f96c3b5ffec7c6e0ffc2c7c3b78f563a9 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-doctor

Agent contract for doctor.

## Objective

Audit managed integration artifacts for drift and legacy patterns.

## Input Contract

### Required Inputs
- None when tool auto-detection succeeds.

### Optional Inputs
- `--root <path>` to target a specific project root.
- `--tools <all|none|csv>` for explicit tool selection.
- `--format table|json` for human vs automation output.
- `--fail-on none|drift|legacy|any` for CI enforcement.

## Execution Contract

- Tool target: cursor
- Recommended command: pacto doctor [--root <path>] [--tools <all|none|csv>] [--format table|json] [--fail-on none|drift|legacy|any]

## Output Contract
- Reports managed artifact status (`ok|missing|unmanaged|legacy_managed|stale|meta_mismatch|legacy_pattern`).
- Provides recommended remediation action (typically `pacto update --artifacts`).
- Emits summary counters for drift and legacy findings.

## Validation Checklist
- Confirm analyzed root and tool selection match user intent.
- Review drift/legacy findings before remediation.
- If CI use case, set explicit `--fail-on` policy.

## Failure Modes and Handling
- No tools detected when `--tools` is omitted.
- Invalid tool list or fail-on/format flag values.
- Filesystem read errors while auditing artifacts.

## Implementation Status

- Status: **Implemented**
- Fallback: If findings are expected but unresolved, run `pacto update --artifacts` and re-run `pacto doctor`.
<!-- pacto:managed:end -->
