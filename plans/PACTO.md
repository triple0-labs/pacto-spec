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

- `README.md` (index + counts)
- `PACTO.md` (this contract)
- `PLANTILLA_PACTO_PLAN.md` (plan template)
- `SLASH_COMMANDS.md` (assistant command conventions)

## Plan Unit

Each plan lives at:

- `<plans-root>/<state>/<slug>/`

Minimum files per plan:

- `README.md` (human summary and links)
- `PLAN_<TOPIC>_<YYYY-MM-DD>.md` (detailed spec/progress)

Slug rules:

- lowercase
- starts with `[a-z0-9]`
- contains only `[a-z0-9-]`

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
- CLI output is English-only; `--lang` is deprecated/ignored.
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
3. generates `README.md` + `PLAN_*.md`
4. updates root `README.md` counts and section links

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
If command behavior changes, update this file and `PLANTILLA_PACTO_PLAN.md` together.
