# Integrations

`pacto install` generates managed Agent Skill artifacts for supported AI tools.
`pacto update --artifacts` refreshes those managed artifacts (use the global `--root` flag to point at the target repository if you are not running from its root).
`pacto update` without `--artifacts` updates the pacto binary and does not modify generated skill files.
`pacto doctor` audits artifact drift and legacy patterns.

Primary guidance surface is Pacto workspace contracts (`<plans-root>/PACTO.md`).
Tool artifacts and optional root `AGENTS.md` content are integration hand-offs.

Generated outputs:

- Skills only: `.../skills/pacto-<workflow>/SKILL.md` (no command or prompt files).

## Generated Agent Contract Layer

Generated skills include:

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

| Tool                  | Skills path                                                                 |
| --------------------- | --------------------------------------------------------------------------- |
| Codex (`codex`)       | `.agents/skills/pacto-*/SKILL.md`                                            |
| Cursor (`cursor`)     | `.agents/skills/pacto-*/SKILL.md` (same tree as Codex; shared contract)  |
| Claude (`claude`)     | `.claude/skills/pacto-*/SKILL.md`                                          |
| OpenCode (`opencode`) | `.opencode/skills/pacto-*/SKILL.md`                                        |

**Cursor and Codex** both discover project skills under **`.agents/skills/`** (see each product’s Agent Skills docs). **Claude does not use that directory today:** [Anthropic’s Agent Skills overview](https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview) describes **Claude Code** using project skills under **`.claude/skills/`** (and user-level `~/.claude/skills/`). Pacto writes Claude targets there, not under `.agents/`.

Legacy paths (`.cursor/skills/`, `~/.codex/prompts/`, and `~/.codex/skills/`) are reported by `pacto doctor` for cleanup.

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
- known legacy patterns in tool folders (including deprecated command/prompt files from earlier releases)

Suggested fix flow:

```bash
pacto doctor --format table
pacto update --artifacts
pacto doctor --fail-on any
```

## Plugin Guardrail Injection

If active plugins define `agentGuardrails`, `pacto install` and `pacto update --artifacts` append a managed plugin section to each generated skill:

```text
<!-- pacto:plugin-guardrails:start -->
...
<!-- pacto:plugin-guardrails:end -->
```

This section is regenerated on update and remains under the main managed block for each generated skill.

## Examples

```bash
# auto-detect tools from project
pacto install

# explicit tools
pacto install --tools codex,cursor

# refresh existing managed artifacts
pacto update --artifacts
```
