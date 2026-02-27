# Commands

## `pacto status`

Verify plan status, blockers, and evidence claims.

```bash
pacto status [--plans-root <path>] [--repo-root <path>] [--mode compat|strict] [--format table|json]
```

Key options:

- `--plans-root`, `--repo-root`
- `--mode`, `--format`, `--fail-on`
- `--state`, `--include-archive`
- `--config`
- `--max-next-actions`, `--max-blockers`
- `--verbose`

Examples:

```bash
pacto status
pacto status --format json --fail-on partial
pacto status --plans-root ./.pacto/plans --repo-root .
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
pacto init [--root .] [--with-agents] [--force]
```

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

Refresh managed Pacto artifacts already installed.

```bash
pacto update [--tools <all|none|csv>] [--force]
```

## `pacto exec`

Planned command, not implemented yet.
