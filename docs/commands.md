# Commands

Global option:

- `--lang <en|es>`: override output language for current command. Workspace default is persisted by `pacto init`.

## `pacto status`

Verify plan status, blockers, and evidence claims.

```bash
pacto status [--root <path>] [--repo-root <path>] [--mode compat|strict] [--format table|json]
```

Behavior:

- TTY: launches interactive status UI.
- Non-TTY: renders `table|json` output.
- In TTY, `--format` is rejected; use non-TTY (pipe/redirection) for structured output.
- `--mode strict` enforces structured PRD plan contract for active plans (`current`, `to-implement`) and reports warnings for `done/outdated`.

Key options:

- `--root`, `--repo-root`
- `--mode`, `--format`, `--fail-on`
- `--state`, `--include-archive`
- `--config`
- `--max-next-actions`, `--max-blockers`
- `--verbose`

Examples:

```bash
pacto status
pacto status | cat
pacto status --format json --fail-on partial
pacto status --root . --repo-root .
```

## `pacto doctor`

Audit generated integration artifacts for drift and legacy patterns.

```bash
pacto doctor [--root <path>] [--tools <all|none|csv>] [--format table|json] [--fail-on none|drift|legacy|any]
```

Notes:

- Detects stale/missing/unmanaged/legacy-managed artifact files.
- Flags known legacy patterns (for example `pa-*` compatibility files).
- Recommended remediation path: `pacto update --artifacts`.

Examples:

```bash
pacto doctor
pacto doctor --tools cursor --format table
pacto doctor --format json --fail-on any
```

## `pacto new`

Create a plan scaffold and update root index.

```bash
pacto new <current|to-implement|done|outdated> <slug> [--title ...] [--owner ...]
```

Key options:

- `--root <path>`
- `--allow-minimal-root`

Examples:

```bash
pacto new to-implement polling-contactos-v2
pacto new current api-contract-refresh --title "API Contract Refresh" --owner "Backend Team"
```

## `pacto init`

Initialize local workspace in `.pacto/plans`.

```bash
pacto init [--root .] [--with-agents] [--force] [--tools <all|none|csv>] [--no-interactive] [--yes] [--no-install] [--dry-run]
```

Notes:

- Canonical workflow contract is `<plans-root>/PACTO.md`.
- `--with-agents` only adds/updates an optional managed hand-off block in root `AGENTS.md`.
- In agent-driven `pacto-init` workflows, run a short interview (problem, technologies, install targets) and create/update a basic project `prd.md`.
- `pacto init` writes `.pacto/config.yaml` with detected/selected technologies, tools, and problem statement.

## `pacto explore`

Capture and revisit ideas without implementation.

```bash
pacto explore <slug> [--title ...] [--note ...] [--root <path>]
pacto explore --list
pacto explore --show <slug>
```

## `pacto install`

Install managed Pacto skills and command prompts.

```bash
pacto install [--tools <all|none|csv>] [--force]
```

## `pacto update`

Update pacto binary by default. Use legacy artifact refresh with `--artifacts`.

```bash
pacto update [--check] [--yes] [--version <vX.Y.Z>] [--repo <owner/repo>]
pacto update --artifacts [--tools <all|none|csv>] [--force]
```

Notes:

- `pacto update` (without `--artifacts`) targets the pacto binary only.
- To refresh generated skills/commands, run `pacto update --artifacts`.

## `pacto exec`

Execute plan tasks and append execution evidence in plan docs.

```bash
pacto exec <current|to-implement|done|outdated> <slug> [--root <path>] [--step <phase.task>] [--note <text>] [--blocker <text>] [--evidence <claim>] [--dry-run]
```

`--step` uses phase task refs (`<phase>.<task>`), for example `1.2`.

## `pacto normalize`

Normalize plan documents to the structured PRD contract and report compliance drift.

```bash
pacto normalize [--root <path>] [--state <state|all>] [--include-archive] [--write] [--format table|json]
```

Notes:

- Dry-run by default; no files are written unless `--write` is set.
- Migrates legacy headings and task IDs (`T1` -> `N.M` when phase context is known).
- Emits per-plan issue counts before/after normalization.

## `pacto move`

Move plan slice between states explicitly.

```bash
pacto move <from-state> <slug> <to-state> [--root <path>] [--reason <text>] [--force]
```

## `pacto plugin`

Manage local plugins under `.pacto/plugins`.

```bash
pacto plugin list-available [--format table|json]
pacto plugin install <id> [--root <path>] [--force] [--no-enable]
pacto plugin list [--root <path>] [--format table|json]
pacto plugin validate [--root <path>] [--plugin <id>]
pacto plugin enable <id> [--root <path>]
pacto plugin disable <id> [--root <path>]
```

Notes:

- `list-available` shows built-in plugins shipped by pacto.
- Built-in example: `git-guardrails`.
- `install` copies a built-in plugin into `.pacto/plugins/<id>` and auto-enables it by default.
- Plugins are loaded from `.pacto/plugins/*/plugin.yaml`.
- Only plugins listed in `.pacto/config.yaml` under `plugins.enabled` are active.
- Supported commands enforce active plugin CLI guardrails by default (`status`, `new`, `move`, `exec`, `install`, `update`, `init`, and `explore` create/update paths).
- Use `--allow-guardrail <id[,id...]>` to bypass specific guardrails for a single run.
