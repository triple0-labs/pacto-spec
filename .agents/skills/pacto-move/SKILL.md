---
name: pacto-move
description: Agent contract for the Pacto move workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=agents/skill/pacto-move workflow=move contract=v1 template_sha256=a812b94907efa93c258b5a134030f8c72c0d20fd723b7ebbbdd3fa1b2330a413 generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
# Pacto Move Skill

Use this skill as an agent contract for the move workflow in Pacto projects.

## Objective

Move a plan slice between canonical states.

## When To Use

Use for explicit state transitions such as `to-implement -> current` or `current -> done`. When `<to-state>` is `done`, the move also folds the plan's `## Capability:` deltas into the persistent baseline at `.pacto/specs/<slug>/spec.md`.

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

- Tool target: Cursor Agent and Codex (project skills under .agents/skills/)
- Recommended command: pacto move <from-state> <slug> <to-state> [--root <path>] [--reason <text>] [--force]

## Output Contract
- Moves plan directory from source state folder to destination state folder.
- Updates moved plan README `Status` line.
- For `to-state=done` only: pre-validates Capability deltas against the baseline and, on success, atomically writes the merged `.pacto/specs/<slug>/spec.md` per affected capability. On any merge collision the move aborts BEFORE the plan folder is renamed.

## Validation Checklist
- Confirm source plan exists before moving.
- Confirm destination does not exist unless force overwrite is intended.
- Confirm `pacto status` reflects the new state location.
- Before `move ... done`: verify the plan's `spec.md` contains well-formed `## Capability: <slug>` blocks with `### ADDED|MODIFIED|REMOVED|RENAMED Requirements` subsections, and that every Requirement has at least one `##### Scenario:` (ADDED) or matches an existing baseline entry (MODIFIED/REMOVED/RENAMED). Run `pacto status --mode strict` first to catch grammar issues.
- After `move ... done`: confirm `.pacto/specs/<slug>/spec.md` was created or updated as expected and that the plan slug reference appears in the merged baseline (e.g. removal audit comments include the plan slug).

## Failure Modes and Handling
- Invalid state values or invalid slug format.
- Source plan missing or destination conflict without `--force`.
- Filesystem move/write failure during transition.
- Baseline merge would fail (ADDED Requirement already exists, MODIFIED/REMOVED target missing, RENAMED missing `- to:` line, unknown delta op). Plan folder is NOT renamed in this case — fix the deltas in `spec.md` and retry.

## Implementation Status

- Status: **Implemented**
- Fallback: If transition fails, run `pacto status --root <path>` to verify current filesystem state before retrying. For baseline merge failures, read the error message (it names the offending capability + Requirement), edit the plan's `spec.md` Capability block to align with the current baseline, and re-run `pacto move`.
<!-- pacto:managed:end -->
