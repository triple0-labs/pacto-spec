<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/command/pacto-status.md workflow=status contract=v1 template_sha256=5dab10e40d40e308a82ce34ca8b3659b21501300d6c62a0e3260c21325e6570c generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-status

Agent contract for status.

## Objective

Verify plan status, blockers, and evidence claims.

## Input Contract

### Required Inputs
- None when auto-discovery can resolve plans root from current directory or parents.

### Optional Inputs
- `--root <path>` to pin project root used for `.pacto/plans` discovery.
- `--repo-root <path>` to pin evidence verification root.
- `--format table|json`, `--fail-on`, `--state`, `--include-archive`.
- `--mode compat|strict`, `--config`, `--max-next-actions`, `--max-blockers`, `--verbose`.

## Execution Contract

- Tool target: cursor
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
- Fallback: Ask for explicit `--root` and `--repo-root` when auto-discovery fails.
<!-- pacto:managed:end -->
