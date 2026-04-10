<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/skill/pacto-install workflow=install contract=v1 template_sha256=6929358e74acc24a3aca335c3131bbe57ea9f744425edc6b6231a62d7d375065 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# Pacto Install Skill

Use this skill as an agent contract for the install workflow in Pacto projects.

## Objective

Install managed Pacto skills and command prompts for supported tools.

## When To Use

Use to bootstrap Pacto-generated skills/prompts for compatible AI tools.

## Input Contract

### Required Inputs
- Either detectable tool directories or explicit `--tools` selection.

### Optional Inputs
- `--tools <all|none|csv>` for explicit selection.
- `--force` to overwrite unmanaged existing files.

## Execution Contract

- Tool target: cursor
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
