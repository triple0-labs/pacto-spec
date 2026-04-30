---
name: pacto-new
description: Agent contract for the Pacto new workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=claude/skill/pacto-new workflow=new contract=v1 template_sha256=43dcd897a1db239ff4396b94dc107f6c3d595eabbb0ab2f74ad89c54705e52c1 generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
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

- Tool target: claude
- Recommended command: pacto new to-implement my-plan-slug

## Output Contract
- Creates `<state>/<slug>/README.md`, `spec.md`, `design.md`, and `tasks.md`.
- Prints created plan artifact paths.
- `spec.md` ships with placeholder `## Capabilities` and `## Requirements` sections — replace with the real Capabilities/Requirements before implementation begins.

## Validation Checklist
- Verify state and slug validity before execution.
- Confirm plan directory does not already exist.
- Confirm plan appears in `pacto status` output for the target state.
- After scaffolding, expand any prose into structured `### Requirement: <name>` blocks (each with at least one nested `#### Scenario: <name>` containing `- WHEN ...` / `- THEN ...`). Use `The system SHALL <behaviour>.` phrasing (or `El sistema DEBE ...` in Spanish).
- List the capabilities this plan touches under `## Capabilities` (`- New Capabilities: [slug, ...]`, `- Modified Capabilities: [slug, ...]`). For new capabilities, add a `## Capability: <slug>` block with `### ADDED Requirements`. For modifications, use `### MODIFIED|REMOVED|RENAMED Requirements` (RENAMED needs `- to: <new name>` in the body).
- Run `pacto status --mode strict` after editing to surface `requirement_missing_scenario`, `requirements_grammar`, or `capability_grammar` issues; fix them before moving the plan to `current`.

## Failure Modes and Handling
- Invalid state or invalid slug format.
- Invalid root (missing canonical files/folders) when minimal root is not allowed.
- Plan already exists for the same state/slug.

## Implementation Status

- Status: **Implemented**
- Fallback: If root validation fails, retry with explicit `--root` or `--allow-minimal-root` when appropriate. If Requirement/Scenario authoring is unclear, leave the placeholders in place and capture the open question under `## Open Questions` rather than inventing structure.
<!-- pacto:managed:end -->
