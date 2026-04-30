# Pacto

Created by: `pacto-cli`  
Last updated: `2026-02-27`

## Purpose

Pacto is a lightweight workflow for AI-assisted engineering plans:

- write plans before implementation
- keep plan state explicit
- validate implementation claims against repository evidence
- produce machine-readable status for CI/automation

## Workspace Model

Canonical plans root:

- `./.pacto/plans` (created by `pacto init`)

Also supported:

- any directory that already contains the 4 state folders (advanced/manual usage)

Required state folders:

- `current`
- `to-implement`
- `done`
- `outdated`

Core root files:

- `README.md` (workspace overview)
- `PACTO.md` (this contract)

## Plan Unit

Each plan lives at:

- `<plans-root>/<state>/<slug>/`

Minimum files per plan:

- `README.md` (human summary and links)
- `spec.md` (problem, scenarios, acceptance criteria, optional `## Capability:` deltas)
- `design.md` (technical context and decisions)
- `tasks.md` (phase tasks, evidence, blockers, next steps)

Slug rules:

- lowercase
- starts with `[a-z0-9]`
- contains only `[a-z0-9-]`

## Capability Baseline

Pacto maintains a persistent capability baseline at `<plans-root>/../specs/<slug>/spec.md`. Each baseline `spec.md` declares Requirements:

```markdown
## Requirements

### Requirement: <name>
The system SHALL <behaviour>.

#### Scenario: <name>
- WHEN <trigger>
- THEN <observable outcome>
```

A plan proposes changes to the baseline using delta blocks inside its own `spec.md`:

```markdown
## Capability: <slug>

### ADDED Requirements
#### Requirement: <name>
...
##### Scenario: <name>
- WHEN ...
- THEN ...

### MODIFIED Requirements
#### Requirement: <existing name>
...

### REMOVED Requirements
#### Requirement: <existing name>

### RENAMED Requirements
#### Requirement: <old name>
- to: <new name>
```

On `pacto move done`, deltas are pre-validated against the baseline and merged atomically. Spanish keywords (`Capacidad`, `Requisito`, `Escenario`, `Capacidades`, `Requisitos`) are accepted alongside English.

## Command Behavior

Canonical CLI commands:

- `pacto init`
- `pacto new`
- `pacto status`
- `pacto explore`
- `pacto install`
- `pacto update`
- `pacto exec`
- `pacto move`

Notes:

- `pacto exec` updates plan execution artifacts only (no source-code edits).
- `pacto move` performs explicit state transitions between plan folders.
- CLI supports `--lang en|es`; language defaults to workspace config when available.
- `status` and `new` auto-discover plans root from current directory and parents.

## How Status Works

`pacto status` performs five steps:

1. Resolve roots (`plans-root`, `repo-root`).
2. Discover plans by state/filter.
3. Parse plan documents (`compat` or `strict` mode).
4. Extract claims from plan text.
5. Verify claims against repository evidence.

Claim categories (configurable):

- `paths`
- `symbols`
- `endpoints`
- `test_refs`

Verification outcomes:

- `verified`
- `partial`
- `unverified`

Output formats:

- `table`
- `json` (stable interface for automation)

Fail policies:

- `none`
- `unverified`
- `partial`
- `blocked`

## How New Plan Creation Works

`pacto new <state> <slug>`:

1. resolves/validates plan root
2. creates `<state>/<slug>/`
3. generates `README.md`, `spec.md`, `design.md`, `tasks.md`

With `--allow-minimal-root`, Pacto can bootstrap missing root files with minimal defaults.

## Evidence Rules

- A plan claim is not considered reliable unless it can be verified in `repo-root`.
- State/progress should match current evidence, not only narrative text.
- Use absolute dates (`YYYY-MM-DD`) for state-relevant updates.

## Authoring Rules

- Keep plans concise and executable: scope, phases/tasks, blockers, next actions, evidence.
- Prefer one source of truth per plan; link out only when needed.
- Keep naming stable (`slug`, file names, section labels) so automation remains deterministic.

## Evolution

This file defines the operational contract for the current CLI behavior.  
If command behavior changes, update this file with the managed root templates.
