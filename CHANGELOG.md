# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Changed

- Removed `bin/pacto-feature-matrix.sh` and the `make feature-matrix` target. Regression coverage for the CLI and command workflows is consolidated in `go test ./...`, with additional tests for `status` (YAML config load, `--fail-on`, `--mode strict`, `--state`), `new` validation, and `pacto new` usage when state/slug are missing.

## 0.1.22 - 2026-04-08

### Added
- Spanish (i18n) support for `pacto exec` workflow: phase headings ("Fase"), localized section headings (Notas de Ejecución, Evidencia, Bloqueadores), and automatic language detection from document content.
- `--root` flag to `pacto install` and `pacto update --artifacts` for specifying project root path explicitly.
- OpenCode skill frontmatter support for generated skills.
- Doctor integration check detects legacy `.codex/skills/pacto-*` entries and recommends cleanup.

### Changed
- Codex integration path migrated from `.codex/skills/` to `.agents/skills/` to align with Codex 2.0 structure.
- Tool detection now identifies Codex presence via `.agents/skills/` or `.codex/` directories.
- Parser and planfmt now accept both English ("Phase N") and Spanish ("Fase N") phase headings.
- Spanish templates updated: GIVEN/WHEN/THEN → DADO/CUANDO/ENTONCES, "Owner" → "Responsable", "Spec:" → "Especificación:", "Design:" → "Diseño:", "Tasks:" → "Tareas:".

### Fixed
- `pacto exec` correctly matches phase heading regex groups for both languages.
- Last Modified updates work with both bold and bullet metadata formats.

## 0.1.21 - 2026-03-27

### Changed
- `pacto init` and minimal-root bootstrapping no longer generate or reference the legacy `SLASH_COMMANDS.md` workspace file.
- Managed workspace templates, root contracts, help text, and getting-started docs were updated to remove legacy `SLASH_COMMANDS.md` guidance.
- Legacy single-file plan template artifacts (`PLANTILLA_PACTO_PLAN.md`) were removed from active templates and example workspace assets.
- The stale `samples/mock-pacto-repo` fixture was removed; examples and smoke flows now generate temporary workspaces instead of relying on checked-in sample plans.
- `bin/pacto-feature-matrix.sh` was updated to validate the current split-plan CLI behavior and current help/output surface.
- Release guidance and local publish skills were simplified to GitHub Release flow only, and the dormant npm publish workflow was removed.

## 0.1.20 - 2026-03-20

### Changed
- `pacto new` now uses split scaffolding only (`README.md`, `spec.md`, `design.md`, `tasks.md`); legacy single-file scaffold creation is deprecated and removed.
- Workspace bootstrap/root validation no longer requires or creates `PLANTILLA_PACTO_PLAN.md`.
- Contract/help/getting-started docs and managed templates were updated to reflect split-only plan creation.

## 0.1.19 - 2026-03-20

### Added
- `pacto new` now supports `--layout split|legacy` and defaults to split plan scaffolds (`spec.md`, `design.md`, `tasks.md`).
- New split-layout scaffolding templates for spec, design, and tasks artifacts.
- New tests covering split default scaffolding, legacy scaffold compatibility, and `pacto exec` preference for `tasks.md`.

### Changed
- `pacto exec` now targets `tasks.md` when present (split layout), while preserving legacy PLAN document behavior.
- Plan structure validation (`planfmt`) moved to a core contract (intent/problem, scenarios, acceptance, phase tasks, evidence, last-modified) with optional module checks only when present.
- `pacto normalize` keeps heading/task normalization but no longer auto-injects large missing section blocks.
- Strict parsing no longer hard-fails on missing declared status alone; structure enforcement follows the relaxed core contract.
- Help and command docs updated for layout-aware `pacto new` usage.

## 0.1.17 - 2026-03-04

## 0.1.18 - 2026-03-11

### Added
- Structured PRD-style plan contract support with dedicated normalization engine (`internal/planfmt`) and coverage tests.
- New `pacto normalize` command for dry-run and write-mode migration of plan docs to standardized headings/task IDs.
- New tests for plan normalization and strict-state parser behavior (`internal/app/normalize_test.go`, `internal/planfmt/planfmt_test.go`).

### Changed
- `pacto status --mode strict` now enforces structured plan rules for active states (`current`, `to-implement`) and reports warnings for non-active states.
- Plan templates were updated to a document-first standardized PRD layout (metadata, FR/AC, scenarios, phased tasks, evidence, risks, next steps).
- `pacto new` plan generation now fills state/slug placeholders in templates and fallback template output.
- `pacto exec` keeps plan metadata fresh by updating `Last Modified` whenever execution updates are written.
- Command/help/docs surface now includes plan normalization workflow and strict-structure guidance.

### Added
- New modular command builders in `internal/app/root_cmd.go`, with focused packages for explore/move/markdown helpers (`internal/explore`, `internal/move`, `internal/markdown`).
- New built-in plugin assets for `git-guardrails` under `internal/plugins/builtin/assets/git-guardrails/`.
- Project-local agent skill docs under `.agents/skills/pacto-local-dev/`.

### Changed
- CLI entrypoints (`cmd/pacto`, `cmd/pacto-engine`) now execute through the shared root command wiring in `internal/app`.
- `pacto plugin install` documentation/examples now use `git-guardrails` as the built-in plugin reference.
- `pacto update` docs now clarify binary update behavior vs artifacts refresh mode (`--artifacts`).
- Init/config/install/status/new/exec flows and tests were updated to align with the new command and plugin wiring.

### Removed
- Legacy `internal/app/cli.go` root command definition.
- Built-in/sample `git-sync` plugin assets and docs references.

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
