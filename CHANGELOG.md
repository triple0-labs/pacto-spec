# Changelog

All notable changes to this project are documented in this file.

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
