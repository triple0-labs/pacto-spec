<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=opencode/command/pacto-explore.md workflow=explore contract=v1 template_sha256=153846c9a3b5418eee3c864f527755852bce6e0b4c8cf61c08611f2723b29e12 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-explore

Agent contract for explore.

## Objective

Capture and revisit ideas without implementing them.

## Input Contract

### Required Inputs
- `<slug>` for create/update flows, or one of `--list` / `--show <slug>`.

### Optional Inputs
- `--title` for the initial idea heading.
- `--note` to append timestamped exploration notes.
- `--root <path>` to target a specific project root.

## Execution Contract

- Tool target: opencode
- Recommended command: pacto explore <slug> [--title <title>] [--note <note>]

## Output Contract
- Stores ideas in `.pacto/ideas/<slug>/README.md`.
- Tracks `Created At` and `Updated At` timestamps.
- Returns list/show output for discovery and review.

## Validation Checklist
- Confirm idea slug resolves to intended workspace.
- Confirm notes append with timestamp and preserve prior history.
- Use `--show` to verify resulting content when needed.

## Failure Modes and Handling
- Missing slug for create/show usage.
- Invalid flag combinations such as conflicting mode flags.
- Permission/path issues creating `.pacto/ideas` files.

## Implementation Status

- Status: **Implemented**
- Fallback: If idea lookup fails, run `pacto explore --list` to discover available slugs.
<!-- pacto:managed:end -->
