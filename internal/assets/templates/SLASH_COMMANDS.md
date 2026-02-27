# Pacto Assistant Commands

Command IDs generated/managed by `pacto install` and `pacto update`:

- `pacto-status`
- `pacto-new`
- `pacto-explore`
- `pacto-init`
- `pacto-install`
- `pacto-update`
- `pacto-exec` (planned)

## Minimal Contract

### `pacto-status`

- Consolidated state + blockers + next actions.
- Verification outcome per plan: `verified|partial|unverified`.

### `pacto-new`

- Creates plan folder in target state.
- Generates `README.md` and `PLAN_<TOPIC>_<YYYY-MM-DD>.md`.
- Updates root index metadata.

### `pacto-explore`

- Stores and updates idea notes in `.pacto/ideas/<slug>/README.md`.

### `pacto-init`

- Bootstraps `.pacto/plans` and template docs.

### `pacto-install` / `pacto-update`

- Manages generated skill/command artifacts for supported tools.

### `pacto-exec`

- Planned command, not implemented.
