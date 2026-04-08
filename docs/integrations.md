# Integrations

`pacto install` generates managed artifacts for supported AI tools.
`pacto update --artifacts` refreshes those managed artifacts.
`pacto update` without `--artifacts` updates the pacto binary and does not modify generated skill/command files.
`pacto doctor` audits artifact drift and legacy patterns.

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
- `doctor`
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
| Codex (`codex`) | `.agents/skills/pacto-*/SKILL.md` | `$CODEX_HOME/prompts/pacto-*.md` (or `~/.codex/prompts/pacto-*.md`) |
| Cursor (`cursor`) | `.cursor/skills/pacto-*/SKILL.md` | `.cursor/commands/pacto-*.md` |
| Claude (`claude`) | `.claude/skills/pacto-*/SKILL.md` | `.claude/commands/pacto-*.md` |
| OpenCode (`opencode`) | `.opencode/skills/pacto-*/SKILL.md` | `.opencode/commands/pacto-*.md` |

## Managed File Behavior

Generated files use managed markers:

```text
<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=... workflow=... contract=... template_sha256=... generated_by=... generated_at=... -->
...
<!-- pacto:managed:end -->
```

Update behavior:

- Managed block exists: block is updated in place.
- Unmanaged file exists: skipped unless `--force` is provided.
- Missing file: created.

## Drift Detection

Use `pacto doctor` to detect:

- missing managed artifacts
- unmanaged overrides
- legacy managed blocks without metadata
- metadata/checksum drift
- known legacy patterns in tool folders

Suggested fix flow:

```bash
pacto doctor --format table
pacto update --artifacts
pacto doctor --fail-on any
```

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
