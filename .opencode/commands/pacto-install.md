<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=opencode/command/pacto-install.md workflow=install contract=v1 template_sha256=8bc5a311d44570ed3fa1b18bc75706e893ec7c47c30b441f2810a343addc3bb3 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-install

Agent contract for install.

## Objective

Install managed Pacto skills and command prompts for supported tools.

## Input Contract

### Required Inputs
- Either detectable tool directories or explicit `--tools` selection.

### Optional Inputs
- `--tools <all|none|csv>` for explicit selection.
- `--force` to overwrite unmanaged existing files.

## Execution Contract

- Tool target: opencode
- Recommended command: pacto install [--tools <all|none|csv>] [--force]

## Output Contract
- Generates managed skill and command files per workflow and selected tool.
- Returns per-file outcome summary: created, updated, skipped, failed.

## Validation Checklist
- Confirm selected/detected tools match user intent.
- Check warnings for unmanaged file skips.
- Confirm generated artifacts are wrapped with managed markers.

## Failure Modes and Handling
- No tools detected when `--tools` is omitted.
- Invalid `--tools` argument values.
- Filesystem write failures for target tool paths.

## Implementation Status

- Status: **Implemented**
- Fallback: If detection fails, rerun with explicit `--tools` list.
<!-- pacto:managed:end -->
