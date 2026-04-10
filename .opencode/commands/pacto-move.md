<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=opencode/command/pacto-move.md workflow=move contract=v1 template_sha256=e6ab424e004b938d5e585fdb971fc9697f364193219c2db5145afc7b462a39b5 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-move

Agent contract for move.

## Objective

Move a plan slice between canonical states.

## Input Contract

### Required Inputs
- `<from-state>` in `current|to-implement|done|outdated`.
- `<slug>` matching `[a-z0-9][a-z0-9-]*`.
- `<to-state>` in `current|to-implement|done|outdated`.

### Optional Inputs
- `--root <path>` to target a specific project root.
- `--reason <text>` to append transition context in plan README.
- `--force` to overwrite destination when it already exists.

## Execution Contract

- Tool target: opencode
- Recommended command: pacto move <from-state> <slug> <to-state> [--root <path>] [--reason <text>] [--force]

## Output Contract
- Moves plan directory from source state folder to destination state folder.
- Updates moved plan README `Status` line.

## Validation Checklist
- Confirm source plan exists before moving.
- Confirm destination does not exist unless force overwrite is intended.
- Confirm `pacto status` reflects the new state location.

## Failure Modes and Handling
- Invalid state values or invalid slug format.
- Source plan missing or destination conflict without `--force`.
- Filesystem move/write failure during transition.

## Implementation Status

- Status: **Implemented**
- Fallback: If transition fails, run `pacto status --root <path>` to verify current filesystem state before retrying.
<!-- pacto:managed:end -->
