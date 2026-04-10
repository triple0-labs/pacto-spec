# Spec: Pacto Plan Artifact System v1

## Status

- Draft
- Date: 2026-03-18
- Owner: Pacto Core
- Related Idea: `pacto-plan-artifact-system`

## Problem

Pacto currently centers plan authoring and status derivation on a single strict `PLAN_*.md` document per slice.
This keeps automation deterministic, but it creates friction when teams need to separate intent, design, and execution details.
As a result, the plan document can become overloaded and difficult to maintain.

## Goals

1. Preserve Pacto's markdown-first and evidence-first model.
2. Keep one deterministic canonical source for status and execution tracking.
3. Allow optional supporting artifacts to improve clarity and maintainability.
4. Keep backward compatibility with existing single-document plans.

## Non-Goals

1. Replace markdown artifacts with a database or proprietary format.
2. Adopt full OpenSpec-style delta merging workflow.
3. Require multi-artifact plans for all repositories.

## Design Principles

1. Canonical execution artifact remains mandatory.
2. Optional artifacts should be additive, not breaking.
3. Strict mode should enforce reliability, not ceremony.
4. Existing plans should continue to pass without migration.

## Proposed Artifact Model

Each plan slice remains:

```text
<plans-root>/<state>/<slug>/
```

Canonical required files:

- `README.md`
- `PLAN_<TOPIC>_<YYYY-MM-DD>.md` (canonical execution doc)

Optional supporting artifacts:

- `PROPOSAL.md`
- `DESIGN.md`
- `TASKS.md`
- `DECISIONS.md`
- `EVIDENCE.md`

Optional artifact manifest:

- `ARTIFACTS.yaml`

### `ARTIFACTS.yaml` v1

```yaml
version: 1
mode: lite # lite|full
canonical: PLAN_AUTH_REFRESH_2026-03-18.md
artifacts:
  - id: proposal
    file: PROPOSAL.md
    required: false
    requires: []
  - id: design
    file: DESIGN.md
    required: false
    requires: [proposal]
  - id: tasks
    file: TASKS.md
    required: false
    requires: [design]
  - id: decisions
    file: DECISIONS.md
    required: false
    requires: [design]
```

Rules:

1. `canonical` must point to an existing markdown file in the same slice.
2. `artifacts[].id` must be unique kebab-case.
3. `requires` references must resolve to existing ids.
4. Dependency cycles are invalid.
5. If `ARTIFACTS.yaml` is absent, current behavior applies unchanged.

## Status and Verification Semantics

### Parsing

1. Status derivation (`declared_status`, phase tasks, blockers, next actions) remains sourced from canonical `PLAN_*.md`.
2. Optional artifacts are parsed only for evidence claims and context links.

### Claims

1. `pacto status` extracts claims from canonical plan plus optional artifacts.
2. Each claim stores source provenance (`file`, `line` when available).
3. Existing claim types remain unchanged in v1 (`path`, `symbol`, `endpoint`, `test_ref`, `delta`).

### Strictness

1. `strict` validates canonical plan structure exactly as today.
2. Optional artifacts may emit warnings for malformed structure, never hard-fail in v1.
3. In `full` mode (manifest), missing `required: true` optional artifacts become strict errors.

## CLI and UX Changes

### `pacto new`

New optional flags:

- `--artifact-mode lite|full` (default `lite`)
- `--with-artifacts proposal,design,tasks,decisions,evidence`

Behavior:

1. Still generates canonical `PLAN_*.md` + `README.md`.
2. If `--with-artifacts` is provided, generate selected files and `ARTIFACTS.yaml`.
3. Template expansion uses existing token replacement pipeline.

### `pacto status`

1. Discover optional artifacts from `ARTIFACTS.yaml` when present.
2. Fallback to filename conventions when manifest absent.
3. Add per-plan warnings:
   - `artifact_missing`
   - `artifact_cycle`
   - `artifact_unresolved_dependency`

JSON extension (additive):

```json
{
  "artifact_mode": "lite",
  "canonical_plan": ".../PLAN_...md",
  "artifacts": [
    {"id": "design", "file": ".../DESIGN.md", "present": true}
  ]
}
```

### `pacto normalize`

1. Continue normalizing canonical plan only by default.
2. Add `--include-artifacts` to lint/normalize optional artifacts with relaxed rules.

## Backward Compatibility

1. Existing repositories need no migration.
2. Existing templates remain valid.
3. `status --mode strict` on legacy single-file plans remains unchanged.
4. New fields in JSON output are additive and optional.

## Implementation Outline

## Phase 1: Discovery and Metadata

- [ ] 1.1 Add artifact manifest model and validation package (`internal/artifacts`).
- [ ] 1.2 Extend discovery to collect optional artifact references per plan.
- [ ] 1.3 Add artifact metadata fields to report model.

## Phase 2: Status Integration

- [ ] 2.1 Keep canonical parser unchanged for status derivation.
- [ ] 2.2 Extend claim extraction to additional artifact files.
- [ ] 2.3 Surface artifact warnings in table/json output.

## Phase 3: Authoring UX

- [ ] 3.1 Add `pacto new --with-artifacts` and `--artifact-mode`.
- [ ] 3.2 Add scaffold templates for supported optional artifacts.
- [ ] 3.3 Document workflow in `docs/concepts.md`, `docs/commands.md`, and plans contract docs.

## Phase 4: Optional Strict Full Mode

- [ ] 4.1 Enforce `required: true` artifact presence in strict mode for `mode: full`.
- [ ] 4.2 Keep `mode: lite` non-blocking for optional artifacts.

## Acceptance Criteria

1. Legacy single-plan slices produce identical status summaries as before.
2. Claim extraction includes entries from optional artifact files when present.
3. Invalid `ARTIFACTS.yaml` produces deterministic warnings/errors.
4. `pacto new` can scaffold canonical-only or multi-artifact slices without manual edits.
5. Strict mode remains reliable and deterministic across both legacy and v1 slices.

## Risks and Mitigations

1. Risk: Increased complexity in plan discovery.
   Mitigation: Isolate artifact logic in `internal/artifacts` with unit tests.

2. Risk: Confusion about which file drives status.
   Mitigation: Keep explicit `canonical` field and show it in status output.

3. Risk: Over-adoption of ceremony.
   Mitigation: Default mode remains `lite` and canonical-only scaffolding.

## Open Questions

1. Should `TASKS.md` eventually become a first-class execution source, or remain supplemental?
2. Should evidence claims from optional artifacts have lower confidence weight than canonical claims?
3. Should `pacto exec` append notes only to canonical plan in v1, or support target artifact selection?
