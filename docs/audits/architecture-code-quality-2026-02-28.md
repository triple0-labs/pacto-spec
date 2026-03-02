# Architecture and Code Quality Audit (2026-02-28)

## Scope

Audit target: current working tree of `pacto` repository.  
Reference benchmark: OpenSpec clone in `./OpenSpec` (light, directional comparison only).

## Repository Baseline

1. Tooling: Go 1.23, Bubble Tea/Lipgloss for TUI.
2. Entry points:
- `cmd/pacto/main.go`
- `cmd/pacto-engine/main.go` (compatibility alias shape)
3. Core package map:
- Orchestration: `internal/app`
- Domain: `internal/{discovery,parser,claims,verify,analyze,report}`
- Integrations/plugins: `internal/{integrations,plugins}`
- Persistence/config/onboarding: `internal/{config,onboarding}`
- Presentation: `internal/{ui,tui}`

## Test and Quality Snapshot

Commands executed:
1. `go test ./...`
2. `go test ./... -cover`
3. `go vet ./...` (previous baseline run)

Current outcome:
1. All tests passing.
2. Coverage improved in previously untested risk areas:
- `internal/claims`: 82.0%
- `internal/discovery`: 57.1%
- `internal/report`: 68.4%
- `internal/onboarding`: 17.1% (still low but no longer zero)

## Findings and Remediation Status

### 1. Config clobber risk during init

- Risk: high
- Previous behavior: `WriteConfig` overwrote `.pacto/config.yaml`, dropping user/plugin state.
- Status: fixed
- Change:
  - `WriteConfig` now performs merge-preserving updates.
  - `plugins.enabled` is preserved when present.
- Files:
  - `internal/onboarding/persist.go`
  - `internal/app/init_test.go`
  - `internal/onboarding/persist_test.go`

### 2. Guardrail command execution hardening

- Risk: high
- Previous behavior: executed scripts via `sh -c <path>`.
- Status: fixed
- Change:
  - execution switched to `sh <path>`, removing command-string interpolation.
- Files:
  - `internal/plugins/hooks.go`

### 3. YAML parsing fragility and duplication

- Risk: high-medium
- Previous behavior: multiple custom parsers for config/manifests/activation.
- Status: fixed
- Change:
  - introduced shared `internal/yamlutil`.
  - migrated `config`, plugin manifest parse, and plugin activation to typed/map YAML handling.
- Files:
  - `internal/yamlutil/yamlutil.go`
  - `internal/config/config.go`
  - `internal/plugins/parse.go`
  - `internal/plugins/activation.go`

### 4. CLI argument normalization duplication

- Risk: medium
- Previous behavior: duplicate normalization functions in multiple commands.
- Status: fixed
- Change:
  - added shared `normalizeArgs` helper and wired command handlers to it.
- Files:
  - `internal/app/argnorm.go`
  - `internal/app/{new,exec,move,explore}.go`

### 5. Architecture intent documentation gap

- Risk: medium
- Status: fixed
- Change:
  - added architecture document with layers, constraints, and binary policy.
- File:
  - `docs/architecture.md`

## Added Regression Tests

1. `internal/app/init_test.go`
- preserves existing `plugins.enabled` after `RunInit`.

2. `internal/plugins/plugins_test.go`
- reads multiline YAML plugin enabled lists.

3. New package tests:
- `internal/claims/claims_test.go`
- `internal/discovery/discovery_test.go`
- `internal/report/report_test.go`
- `internal/onboarding/persist_test.go`

## Light Benchmark vs OpenSpec

Directional observations:
1. OpenSpec uses stronger modular boundaries around schemas/artifact graph.
2. Pacto is intentionally markdown/workspace-centric and lighter-weight.
3. This remediation aligns Pacto closer to robust config/schema handling while preserving markdown-first workflow.

## Remaining Risks

1. `internal/onboarding` and TUI packages still have lower coverage relative to core status/parser paths.
2. `cmd/pacto-engine` remains an alias; future work can decide deprecation or explicit differentiation.

## Acceptance Check

1. `go test ./...` passes.
2. YAML handling is consolidated across config and plugin paths.
3. Init config writes preserve existing plugin activation entries.
4. Guardrail invocation no longer uses `sh -c`.
