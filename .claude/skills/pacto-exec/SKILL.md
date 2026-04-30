---
name: pacto-exec
description: Agent contract for the Pacto exec workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=claude/skill/pacto-exec workflow=exec contract=v1 template_sha256=d2ad1dc1ba6b7fa9b9ebfb071a853ccce2eb509e65cc9208e846b63aaf9dff77 generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
# Pacto Exec Skill

Use this skill as an agent contract for the exec workflow in Pacto projects.

## Objective

Execute plan tasks and register execution evidence in plan artifacts.

## When To Use

Use after moving a plan to `current` to advance tasks while keeping execution evidence in plan documents.

## Input Contract

### Required Inputs
- `<state>` and `<slug>` identifying an existing plan slice (`state` must be `current`).

### Optional Inputs
- `--root <path>` to target a specific project root.
- `--step <phase.task>` to complete a specific task (for example, `1.2`).
- `--note`, `--blocker`, `--evidence` to append execution context.
- `--dry-run` to preview updates without writing files.

## Execution Contract

- Tool target: claude
- Recommended command: pacto exec <state> <slug> [--root <path>] [--step <phase.task>] [--note <text>] [--blocker <text>] [--evidence <claim>] [--dry-run]

## Output Contract
- Marks pending tasks as completed in plan markdown checklists.
- Appends execution notes, blockers, and evidence references.
- Writes only plan artifact files (no source-code edits).

## Validation Checklist
- Confirm target plan exists before execution updates.
- Confirm intended task selection (`next pending` vs explicit `--step`).
- When dry-run is used, confirm no files are modified.

## Failure Modes and Handling
- Missing/invalid plan target or invalid step id.
- No matching pending task for selected step.
- Filesystem write errors updating plan markdown.

## Implementation Status

- Status: **Implemented**
- Fallback: Use `pacto status` to inspect blockers or task state if execution updates cannot be applied.
<!-- pacto:managed:end -->
