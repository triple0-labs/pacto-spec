<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/skill/pacto-explore workflow=explore contract=v1 template_sha256=3d14dfe155986aef8f1e100972803d1a2cf1c54f9d4ee13726c2ab40dec74a8b generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# Pacto Explore Skill

Use this skill as an agent contract for the explore workflow in Pacto projects.

## Objective

Capture and revisit ideas without implementing them.

## When To Use

Use for ideation and notes when work is not ready for a formal plan slice.

## Input Contract

### Required Inputs
- `<slug>` for create/update flows, or one of `--list` / `--show <slug>`.

### Optional Inputs
- `--title` for the initial idea heading.
- `--note` to append timestamped exploration notes.
- `--root <path>` to target a specific project root.

## Execution Contract

- Tool target: cursor
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
