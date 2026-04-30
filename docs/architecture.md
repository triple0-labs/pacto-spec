# Pacto Architecture

This document defines the intended layering and extension points for the `pacto` CLI.

A light DDD layering is being adopted incrementally. Dependencies point inward:
`cli/tui` → `app` (use cases) → `domain` (pure types and rules); `infra`
(adapters: filesystem, markdown, yaml, exec, integrations, plugins) implements
ports declared by `app`.

## Layers

1. CLI orchestration (presentation)
- Packages: `cmd/pacto`, `internal/cli`, `internal/command/<cmd>/`.
- Responsibilities: Cobra root command, persistent flags (`--root`, `--lang`, guardrail bypasses), command registration, delegating to use cases, exit codes.
- Rule: thin wiring only — keep heavy parsing and domain logic out.

2. Domain (pure)
- Packages: `internal/domain/plan`, `internal/domain/claim`, `internal/domain/report`.
- Responsibilities: core types (PlanRef, Phase, Task, Claim/Result, StatusReport DTOs) and validation rules. Stdlib-only; no I/O.

3. Domain workflows
- Packages: `internal/discovery`, `internal/parser`, `internal/claims`, `internal/verify`, `internal/analyze`, `internal/render`.
- Responsibilities: discover plan artifacts, derive status signals, verify claims, compute report model, render report formats.

4. Persistence and config
- Packages: `internal/config`, `internal/onboarding`, `internal/yamlutil`.
- Responsibilities: load and normalize configuration, persist onboarding workspace state, merge managed sections while preserving unrelated user data.

5. Integrations and plugins
- Packages: `internal/integrations`, `internal/plugins`.
- Responsibilities: adapter-based generation of managed artifacts, plugin discovery/validation, guardrail enforcement.

6. UI
- Packages: `internal/ui`, `internal/tui/*`.
- Responsibilities: terminal styles and interactive displays only.

## Application Layer Policies

### i18n in use cases

Several `internal/app/*` use cases (`execplan`, `explore`, `move`, `newplan`,
`initws`) import `internal/i18n`. This is intentional: pacto's text-mutation
commands must match the language of the plan document they are editing
(autodetected from the document's headings), so the language selection is a
**domain rule** ("write the new bullet under the same `## Blockers` /
`## Bloqueadores` heading the document already uses"), not a presentation
concern.

The CLI layer still owns the `--lang` flag, but its only job is to set a
fallback that flows into use-case `Input` structs as `Lang i18n.Language`.
Use cases must:

- Detect the document's effective language when mutating existing files.
- Use the supplied fallback only when no markers are present.
- Never call `fmt.Print*` or write to stdout/stderr — text variants are
  returned via `Result` for the CLI to render.

This policy keeps `internal/i18n` confined to the rendering of strings; it
does not turn the application layer into a presentation layer.

## Design Constraints

1. Markdown-first plan model
- Source of truth remains plan markdown files under plans root.
- `pacto exec` mutates plan artifacts only.

2. Evidence over assumptions
- `pacto status` parses plan claims and verifies against `repo-root`.
- Verification outputs: `verified`, `partial`, `unverified`.

3. Safe managed writes
- Managed content should use explicit markers where applicable.
- Config updates must be merge-preserving and not clobber unrelated keys.

4. Extensibility by adapters/plugins
- Tool integrations are adapter-driven (`codex`, `cursor`, `claude`, `opencode`).
- Plugins extend behavior through validated manifests and guardrails.

## Binary Policy

`cmd/pacto` is the canonical CLI binary. The legacy `pacto-engine` alias has been retired.

## Non-goals

1. Introduce a full AST or database-backed plan store.
2. Replace markdown plans with proprietary formats.
3. Add remote/network coupling to core status/verification workflows.
