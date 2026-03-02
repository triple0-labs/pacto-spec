# Changelog

All notable changes to this project are documented in this file.

## 0.1.16 - 2026-03-02

### Added
- New i18n foundation for UI output (`en`/`es`) plus workspace-level UI language persistence in `.pacto/config.yaml`.
- Language selection step at the beginning of `pacto init` onboarding; selected language now drives generated managed docs/templates.
- Spanish managed template set for workspace artifacts (`README.md`, `PACTO.md`, `PLANTILLA_PACTO_PLAN.md`, `SLASH_COMMANDS.md`, `AGENTS.md`).
- New plugin subcommands:
  - `pacto plugin list-available [--format table|json]`
  - `pacto plugin install <id> [--root <path>] [--force] [--no-enable]`
- Embedded built-in plugin catalog (first built-in: `git-sync`) with installer support and default `config.env` generation.
- New sample plugin package at `samples/plugins/git-sync` for manual/project-local usage.

### Changed
- `pacto init` output summary redesigned to be more user-friendly and structured, including localized labels and next-step guidance.
- Path output in action-oriented commands now prefers relative paths to current working directory (with absolute fallback outside base).
- `pacto status` now supports plugin CLI guardrails; guardrails are skipped for help invocations (`--help`, `-h`, `help`).
- Help and command docs updated to reflect new plugin install/list-available workflows and active language behavior.
- Onboarding copy updated for clearer, more conversational guidance in problem and technologies steps.

## 0.1.15 - 2026-02-28

### Added
- Interactive TUI onboarding flow for `pacto init` (problem, technologies, install targets).
- New onboarding persistence for `.pacto/config.yaml` and managed `prd.md` block generation/update.
- New `--tools`, `--no-interactive`, `--yes`, `--no-install`, and `--dry-run` options for `pacto init`.
- Interactive TUI for `pacto status` in TTY mode (search/filter/details panel).
- Shared terminal UI styling helpers and global `--no-color` support.
- Phase-task metadata parsing (`<phase>.<task>`) in parser/model structures.
- Project-local Codex skills under `.codex/skills/*` for managed workflow contracts.

### Changed
- `pacto exec --step` now uses `<phase>.<task>` format (for example, `1.2`) and auto-completion targets first incomplete phase-task.
- Plan templates now use phase-oriented English headings and numbered phase-task checklist examples (`1.1`, `1.2`, ...).
- `pacto status` now rejects `--format` in TTY mode and keeps table/json rendering for non-TTY output.
- `pacto init` now validates profile completeness and reports created/updated/skipped outputs with improved CLI formatting.
- Docs and workflow metadata updated for onboarding + PRD generation and phase-task execution semantics.

### Removed
- `pacto init --editor` flag support.
- `pacto init --language` flag support.
- Legacy `pacto exec --step T<number>` task reference support.
