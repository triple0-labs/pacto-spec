# Spec: Spec Baseline and Requirements Grammar

## Metadata

- Owner: pacto-core
- Created: 2026-04-29
- Last Modified: 2026-04-29
- State: to-implement
- Slug: spec-baseline-and-requirements

## Domains Affected

- specs
- new
- move
- status
- parser
- planfmt

## Problem Statement

Pacto today has no persistent answer to "what does the system already do?".
Every plan's `spec.md` re-states the world in free-form prose, and once a plan
is moved to `done/` that knowledge stays buried inside a historical artifact.
Two consequences follow:

1. New plans cannot cleanly express "I am changing this existing requirement"
   because there is no shared baseline to point at.
2. `pacto status` can only verify free-form claims (paths, symbols, endpoints,
   test refs) — it cannot tell you "Requirement R-003 has zero tasks" or
   "Scenario S-002 has no covering evidence", because Requirements and
   Scenarios are not first-class addressable units.

OpenSpec solves (1) with a baseline `openspec/specs/<capability>/` tree plus
delta blocks in change folders. spec-kit solves (2) with structured
`### Requirement` / acceptance scenario grammar that downstream commands
(`/analyze`, `/tasks`) consume. Pacto needs both.

## User Scenarios

### Scenario: greenfield workspace gains a baseline tree

- **GIVEN** a workspace freshly initialized with `pacto init`
- **WHEN** init completes
- **THEN** `.pacto/specs/` exists with a `README.md` explaining the baseline
- **AND** the directory is empty of capability folders until plans land

### Scenario: new plan declares structured requirements

- **GIVEN** the user runs `pacto new "Add OAuth login"`
- **WHEN** the generated `spec.md` is opened
- **THEN** it contains a `## Requirements` section with at least one
  `### Requirement: <name>` block and one nested `#### Scenario: <name>` block
  using `WHEN/THEN` placeholders
- **AND** it contains a `## Capabilities` section with `New Capabilities` and
  `Modified Capabilities` lists (kebab-case slugs) so the plan declares which
  baseline files it will create or change

### Scenario: planfmt validates the new grammar

- **GIVEN** an active plan whose `spec.md` defines a `### Requirement` with no
  nested `#### Scenario`
- **WHEN** `pacto status` runs in strict mode
- **THEN** `planfmt.Validate` reports the missing scenario as an error
- **AND** the plan is flagged as failing planfmt validation

### Scenario: status reports per-requirement coverage

- **GIVEN** a plan whose `spec.md` has Requirements R-001..R-003 and whose
  `tasks.md` references R-001 in two task descriptions and R-002 in one
- **WHEN** the user runs `pacto status`
- **THEN** the report lists each Requirement with its task count and evidence
  count
- **AND** R-003 is flagged as `uncovered` (zero tasks)

### Scenario: move done merges deltas into the baseline

- **GIVEN** a plan in `current/` that declares `Modified Capabilities: [auth]`
  and contains an `## ADDED Requirements` block under
  `## Capability: auth` in its `spec.md`
- **WHEN** the user runs `pacto move done <slug>`
- **THEN** pacto reads the delta blocks and merges ADDED requirements into
  `.pacto/specs/auth/spec.md`, creating that file if absent
- **AND** the plan moves to `done/` with its delta blocks intact (audit trail)
- **AND** a summary line lists each baseline file written

### Scenario: archive merge is skipped when no deltas are declared

- **GIVEN** a plan with no `## Capability:` sections in `spec.md`
- **WHEN** `pacto move done <slug>` runs
- **THEN** no baseline file is written and the move succeeds normally

### Scenario: existing plans without Requirements grammar still work

- **GIVEN** a plan created before this feature whose `spec.md` uses only
  `## Acceptance Criteria`
- **WHEN** any pacto command runs against it
- **THEN** no crash occurs and the plan is treated as a legacy plan with zero
  structured Requirements

## Acceptance Criteria

- AC-001: `pacto init` creates `.pacto/specs/` and a `.pacto/specs/README.md`
  describing the baseline tree (idempotent — preserves existing files).
- AC-002: `pacto new` generates a `spec.md` containing a `## Capabilities`
  section (with `New` and `Modified` lists) and a `## Requirements` section
  with one `### Requirement: <name>` and one `#### Scenario: <name>`
  placeholder using `WHEN/THEN` lines.
- AC-003: A new `internal/specsbaseline/` (or equivalent) package exposes
  `ParseRequirements(specPath) []Requirement` returning Requirements with
  stable IDs (`R-NNN`) and nested Scenarios (`S-NNN`).
- AC-004: `planfmt.Validate` enforces:
  (a) every `### Requirement` has at least one `#### Scenario`;
  (b) Scenarios use exactly four hashtags (no silent failure);
  (c) Requirement and Scenario names are unique within their parent block.
- AC-005: The grammar supports delta blocks `## Capability: <slug>` followed by
  `### ADDED Requirements`, `### MODIFIED Requirements`,
  `### REMOVED Requirements`, `### RENAMED Requirements` — parsed by the same
  package.
- AC-006: `pacto status` reports per-Requirement coverage: task count and
  evidence count, plus an `uncovered` flag when task count is zero.
  Available in both table and JSON output.
- AC-007: `pacto move done <slug>` merges declared deltas into
  `.pacto/specs/<capability>/spec.md`. ADDED appends, MODIFIED replaces the
  matching `### Requirement` block, REMOVED deletes it (with a
  `<!-- removed by <slug> on <date> -->` audit comment), RENAMED rewrites the
  header. The plan's own `spec.md` is left intact for history.
- AC-008: Merge is atomic per baseline file (write to temp, rename) so a
  failure mid-merge does not leave a partial baseline file.
- AC-009: Plans without the new grammar (legacy plans) continue to work for
  every command — no parse errors, no false coverage warnings, no merges
  attempted.
- AC-010: Spanish templates (`internal/assets/templates_es/`) gain matching
  `## Capacidades` and `## Requisitos` sections with equivalent placeholders.
- AC-011: All existing Go unit and integration tests continue to pass.
- AC-012: New unit tests cover: requirement parsing, scenario parsing, delta
  parsing, planfmt validation rules, baseline merge (each delta op), atomic
  write behaviour.
- AC-013: New integration test covers the full lifecycle: `init` →
  `new` → fill spec with one ADDED requirement under `## Capability: demo` →
  `move done` → assert `.pacto/specs/demo/spec.md` contains the merged
  Requirement → `status` reports R-001 covered.

## Out of Scope

- `pacto analyze` cross-artifact consistency command (separate Tier 1 plan).
- `pacto clarify` interactive `[NEEDS CLARIFICATION]` loop (separate plan).
- Promoting `PACTO.md` to a constitution with enforced gates (separate plan).
- Workflow YAML composition and presets (Tier 2/3).

## Non-Goals

- Replacing the existing claim/evidence verification — Requirements coexist
  with claims; a Requirement may carry one or more claim hints, but the
  ripgrep-based verification path is unchanged.
- Forcing legacy plans to be rewritten — adoption is opt-in via templates.
