# Integrations

`pacto install` generates managed artifacts for supported AI tools.
`pacto update --artifacts` refreshes those managed artifacts.

Primary guidance surface is Pacto workspace contracts (`<plans-root>/PACTO.md`).
Tool artifacts and optional root `AGENTS.md` content are integration hand-offs.

Generated outputs:

- Skills: `.../skills/pacto-<workflow>/SKILL.md`
- Commands/prompts: `pacto-<workflow>.md`

## Generated Agent Contract Layer

Generated skills and command prompts include:

- `Input Contract`
- `Execution Contract`
- `Output Contract`
- `Validation Checklist`
- `Failure Modes and Handling`
- `Implementation Status`

## Workflows Generated

- `status`
- `new`
- `explore`
- `init`
- `install`
- `update`
- `move`
- `exec`

## Supported Tools and Paths

| Tool | Skills path | Command path |
|------|-------------|--------------|
| Codex (`codex`) | `.codex/skills/pacto-*/SKILL.md` | `$CODEX_HOME/prompts/pacto-*.md` (or `~/.codex/prompts/pacto-*.md`) |
| Cursor (`cursor`) | `.cursor/skills/pacto-*/SKILL.md` | `.cursor/commands/pacto-*.md` |
| Claude (`claude`) | `.claude/skills/pacto-*/SKILL.md` | `.claude/commands/pacto-*.md` |
| OpenCode (`opencode`) | `.opencode/skills/pacto-*/SKILL.md` | `.opencode/commands/pacto-*.md` |

## Managed File Behavior

Generated files use managed markers:

```text
<!-- pacto:managed:start -->
...
<!-- pacto:managed:end -->
```

Update behavior:

- Managed block exists: block is updated in place.
- Unmanaged file exists: skipped unless `--force` is provided.
- Missing file: created.

## Plugin Guardrail Injection

If active plugins define `agentGuardrails`, `pacto install` and `pacto update --artifacts` append a managed plugin section to generated artifacts:

```text
<!-- pacto:plugin-guardrails:start -->
...
<!-- pacto:plugin-guardrails:end -->
```

This section is regenerated on update and remains under the main managed block for each generated file.

## Examples

```bash
# auto-detect tools from project
pacto install

# explicit tools
pacto install --tools codex,cursor

# refresh existing managed artifacts
pacto update --artifacts
```
