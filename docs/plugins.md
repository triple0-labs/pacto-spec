# Plugins

Pacto v1 plugins are project-local and loaded from:

- `.pacto/plugins/<plugin-id>/plugin.yaml`

Plugins are discovered automatically, but only plugins listed in `.pacto/config.yaml` under `plugins.enabled` are active.

## Manifest

```yaml
apiVersion: pacto/v1alpha1
kind: Plugin
metadata:
  id: acme-guardrails
  version: 0.1.0
  priority: 100
spec:
  cliGuardrails:
    - id: clean-worktree
      commands: [status, new, move, exec, install, update, init, explore]
      run:
        script: scripts/check-clean.sh
        timeoutMs: 5000
      onFail:
        message: Working tree must be clean.
  agentGuardrails:
    - id: status-first
      tools: [codex, cursor, claude, opencode]
      workflows: [new, exec, move]
      markdownFile: guardrails/status-first.md
```

## CLI Guardrails

- Run before command execution for supported commands (`status`, `new`, `move`, `exec`, `install`, `update`, `init`, and `explore` create/update paths).
- Non-zero exit blocks command.
- Timeout also blocks command.
- Use `--allow-guardrail <id[,id...]>` to bypass specific guardrails for one run.
- Guardrails are skipped for help requests (`--help`, `-h`, or `help`).

Linux/macOS are supported in v1 (`/bin/sh` runtime).

## Agent Guardrails

`agentGuardrails` markdown snippets are appended to generated skill/command artifacts during `pacto install` and `pacto update` in a managed plugin section.

## Commands

```bash
pacto plugin list-available
pacto plugin install git-sync
pacto plugin list
pacto plugin validate
pacto plugin enable <id>
pacto plugin disable <id>
```

Built-in plugins can be listed with `list-available` and installed directly with `plugin install <id>`.
`plugin install` writes files into `.pacto/plugins/<id>` and enables the plugin by default (unless `--no-enable` is set).

## Sample Plugin

See a complete example at:

- `samples/plugins/acme-guardrails/`
- `samples/plugins/git-sync/`
