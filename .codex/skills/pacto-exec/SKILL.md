---
name: pacto-exec
description: Agent contract for the Pacto exec workflow (currently planned/not implemented).
---

<!-- pacto:managed:start -->
# Pacto Exec Skill

Use this skill as an agent contract for the exec workflow in Pacto projects.

## Objective

Execute plan slices and capture deltas/evidence.

## When To Use

Reserved workflow for guided plan execution when command implementation becomes available.

## Input Contract

### Required Inputs
- `<path-to-plan-md>` of the plan to execute (future behavior).

### Optional Inputs
- None currently, command is planned.

## Execution Contract

- Tool target: codex
- Recommended command: pacto exec <path-to-plan-md>

## Output Contract
- Current expected output is a planned/not-implemented message.
- Do not represent this workflow as executing plan slices today.

## Validation Checklist
- Explicitly communicate that `pacto exec` is planned and not implemented.
- Offer a fallback workflow (`status`, `new`, or `explore`) aligned with user intent.

## Failure Modes and Handling
- Users may expect execution; command currently cannot execute slices.

## Implementation Status

- Status: **Planned (Not Implemented)**
- Fallback: Use `pacto status` for verification, `pacto new` for scaffolding, and `pacto explore` for ideation until exec ships.
<!-- pacto:managed:end -->
