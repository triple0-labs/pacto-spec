---
name: pacto-new
description: Agent contract for the Pacto new workflow to create plan scaffolds and update index metadata.
---

<!-- pacto:managed:start -->
# Pacto New Skill

Use this skill as an agent contract for the new workflow in Pacto projects.

## Objective

Create a new plan scaffold and update plan index metadata.

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

- Tool target: codex
- Recommended command: pacto new to-implement my-plan-slug

## Output Contract
- Creates `<state>/<slug>/README.md` and `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.
- Updates root `README.md` counts, section links, and last update date.
- Prints created paths and updated index path.

## Validation Checklist
- Verify state and slug validity before execution.
- Confirm plan directory does not already exist.
- Confirm index update succeeded in root `README.md`.

## Failure Modes and Handling
- Invalid state or invalid slug format.
- Invalid root (missing canonical files/folders) when minimal root is not allowed.
- Plan already exists for the same state/slug.

## Implementation Status

- Status: **Implemented**
- Fallback: If root validation fails, retry with explicit `--root` or `--allow-minimal-root` when appropriate.
<!-- pacto:managed:end -->
