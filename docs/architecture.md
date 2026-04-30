# Pacto Architecture

This document defines the intended layering and extension points for the `pacto` CLI.

A light DDD layering is being adopted incrementally. Dependencies point inward:
`cli/tui` → `app` (use cases) → `domain` (pure types and rules); `infra`
(adapters: filesystem, markdown, yaml, exec, integrations, plugins) implements
ports declared by `app`.

## Layers

1. CLI orchestration (presentation)
- Packages: `cmd/pacto`, `cmd/pacto-engine`, `internal/cli`, `internal/command/<cmd>/`.
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

`cmd/pacto` is the primary CLI binary.  
`cmd/pacto-engine` currently acts as a compatibility alias and intentionally shares the same runtime entrypoint.

## Non-goals

1. Introduce a full AST or database-backed plan store.
2. Replace markdown plans with proprietary formats.
3. Add remote/network coupling to core status/verification workflows.
