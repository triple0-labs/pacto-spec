<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/command/pacto-update.md workflow=update contract=v1 template_sha256=7b03df4cc644c978d769f137c80fbe5ba711b1a34523138967258e55c21d88b2 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-update

Agent contract for update.

## Objective

Refresh previously installed managed Pacto artifacts.

## Input Contract

### Required Inputs
- Previously installed tool artifacts, or explicit `--tools` target.

### Optional Inputs
- `--tools <all|none|csv>` for explicit tool selection.
- `--force` to overwrite unmanaged files when needed.

## Execution Contract

- Tool target: cursor
- Recommended command: pacto update --artifacts [--tools <all|none|csv>] [--force]

## Output Contract
- Updates managed blocks in skill and command artifacts in place.
- Reports created/updated/skipped/failed counts.
- Preserves unmanaged files unless `--force` is set.

## Validation Checklist
- Confirm managed marker replacement happened for existing files.
- Review skipped unmanaged warnings and decide if force is appropriate.
- Spot-check one skill and one command artifact for expected template updates.

## Failure Modes and Handling
- Unsupported or invalid tool selection.
- Unmanaged files skipped without `--force`.
- Filesystem write errors.

## Implementation Status

- Status: **Implemented**
- Fallback: Use `--force` only when intentional overwrite of unmanaged files is acceptable.
<!-- pacto:managed:end -->
