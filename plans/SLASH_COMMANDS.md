# Pacto Assistant Commands

This file defines workflow command IDs used by Pacto-managed assistant integrations.

## Command IDs

- `pacto-status`
- `pacto-new`
- `pacto-explore`
- `pacto-init`
- `pacto-install`
- `pacto-update`
- `pacto-move`
- `pacto-exec`

## Command Contract

### `pacto-status`

- Goal: report consolidated plan state and verification outcomes.
- Uses detected or explicit plans root (`./.pacto/plans` or direct state-folder root).
- Verifies claims against `repo-root`.
- Outputs `table` or `json` with `verified|partial|unverified` classification.

### `pacto-new`

- Goal: create a new plan slice under a target state.
- Requires: `<current|to-implement|done|outdated> <slug>`.
- Generates `README.md` + `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.
- Updates root index counts/links in `<plans-root>/README.md`.

### `pacto-explore`

- Goal: capture and revisit ideas before formal planning.
- Stores ideas in `.pacto/ideas/<slug>/README.md`.
- Supports create, note append, list, and show flows.

### `pacto-init`

- Goal: initialize canonical workspace in `.pacto/plans`.
- Creates state folders and core templates (`README.md`, `PACTO.md`, `PLANTILLA_PACTO_PLAN.md`, `SLASH_COMMANDS.md`).
- Optional `--with-agents` updates managed Pacto block in `AGENTS.md`.

### `pacto-install`

- Goal: generate managed skill and command artifacts for supported AI tools.
- Supports explicit tool selection (`--tools`) and overwrite control (`--force`).

### `pacto-update`

- Goal: refresh managed blocks in previously generated artifacts.
- Preserves unmanaged files unless `--force` is used.

### `pacto-move`

- Goal: perform explicit state transitions for a plan slice.
- Usage: `pacto move <from-state> <slug> <to-state>`.
- Updates slice README status and root index counts/links.

### `pacto-exec`

- Goal: execute plan checklist progress without source-code edits.
- Usage: `pacto exec current <slug> [--step ...] [--note ...] [--blocker ...] [--evidence ...]`.
- Marks tasks complete and appends execution evidence in plan docs.

## Conventions

- Product docs are canonical in `docs/`.
- Workspace files under `.pacto/plans` are operational artifacts/templates.
