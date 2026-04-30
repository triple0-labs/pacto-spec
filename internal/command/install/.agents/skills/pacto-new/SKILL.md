---
name: pacto-new
description: Agent contract for the Pacto new workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=agents/skill/pacto-new workflow=new contract=v1 template_sha256=6e69345dc3e061db54c9f3b212b5c1f2d1a1f50df4cb96e851e4d25ba63da1cf generated_by=pacto generated_at=2026-04-30T00:22:28Z -->
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

- Tool target: Cursor Agent and Codex (project skills under .agents/skills/)
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
