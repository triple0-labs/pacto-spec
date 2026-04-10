<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=cursor/command/pacto-exec.md workflow=exec contract=v1 template_sha256=88252d7693a6b6a22ea6c338faaa8bfdc66a4ce1103a3c34163ff224aa474c41 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-exec

Agent contract for exec.

## Objective

Execute plan tasks and register execution evidence in plan artifacts.

## Input Contract

### Required Inputs
- `<state>` and `<slug>` identifying an existing plan slice (`state` must be `current`).

### Optional Inputs
- `--root <path>` to target a specific project root.
- `--step <phase.task>` to complete a specific task (for example, `1.2`).
- `--note`, `--blocker`, `--evidence` to append execution context.
- `--dry-run` to preview updates without writing files.

## Execution Contract

- Tool target: cursor
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
