<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/skill/pacto-new workflow=new contract=v1 template_sha256=a3393d31f08be36c310c23b3f23029e9f8af98d5aa20626500ff7f05fb85694d generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# Pacto New Skill

Use this skill as an agent contract for the new workflow in Pacto projects.

## Objective

Create a new plan scaffold.

## When To Use

Use when a new plan slice must be created in one of the canonical states.

## Input Contract

### Required Inputs
- `<state>` in `current|to-implement|done|outdated`.
- `<slug>` matching `[a-z0-9][a-z0-9-]*`.

### Optional Inputs
- `--title`, `--owner` for richer metadata.
- `--root <path>` for explicit plan root.
- `--allow-minimal-root` to bootstrap missing root files.

## Execution Contract

- Tool target: cursor
- Recommended command: pacto new to-implement my-plan-slug

## Output Contract
- Creates `<state>/<slug>/README.md`, `spec.md`, `design.md`, and `tasks.md`.
- Prints created plan artifact paths.

## Validation Checklist
- Verify state and slug validity before execution.
- Confirm plan directory does not already exist.
- Confirm plan appears in `pacto status` output for the target state.

## Failure Modes and Handling
- Invalid state or invalid slug format.
- Invalid root (missing canonical files/folders) when minimal root is not allowed.
- Plan already exists for the same state/slug.

## Implementation Status

- Status: **Implemented**
- Fallback: If root validation fails, retry with explicit `--root` or `--allow-minimal-root` when appropriate.
<!-- pacto:managed:end -->
