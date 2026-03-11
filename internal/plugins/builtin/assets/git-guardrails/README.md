# git-guardrails plugin (built-in)

Made by Pacto, owned by 000geid.

This plugin enforces git preflight guardrails before key workflows (`status`, `new`, `explore`, `exec`, `move`).

## Behavior

- Requires execution inside a git repository.
- Runs `git fetch --prune <remote>` (upstream remote first, fallback `origin`).
- Blocks when unresolved merge conflicts exist.
- Blocks when branch is behind or diverged from tracking reference.
- Optionally runs `gh` PR diagnostics (non-blocking).

## Configuration

Copy/edit `.pacto/plugins/git-guardrails/config.env`:

- `STRICT_MODE=1` fail-closed (default), `0` warn-only.
- `REMOTE_OVERRIDE=<remote>` to force remote selection.
- `ENABLE_GH_DIAGNOSTICS=1|0` to toggle `gh` checks.
- `GH_REPO=owner/repo` to scope PR diagnostics.

## Bypass for one run

```bash
pacto --allow-guardrail git-guardrails/git-preflight <command ...>
```
