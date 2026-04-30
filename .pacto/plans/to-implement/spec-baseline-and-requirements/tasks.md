# Tasks: Spec Baseline and Requirements Grammar

## Execution Metadata

- Status: Draft
- Owner: pacto-core
- Created: 2026-04-29
- Last Modified: 2026-04-29
- State: to-implement
- Slug: spec-baseline-and-requirements

## Implementation Plan by Phases

## Phase 1: specsbaseline package — parsing

- [ ] 1.1 Create `internal/specsbaseline/` package with public types: `Capability`, `Requirement`, `Scenario`, `Delta`, `DeltaOp` (ADDED/MODIFIED/REMOVED/RENAMED).
- [ ] 1.2 Implement `ParseRequirements(specPath string) ([]Requirement, error)` for the `## Requirements` form.
- [ ] 1.3 Implement `ParseDeltas(specPath string) ([]Capability, error)` for the `## Capability: <slug>` form with delta op blocks.
- [ ] 1.4 Implement stable ID assignment (`R-NNN`, `S-NNN`) with optional `<!-- id: ... -->` override.
- [ ] 1.5 Add EN + ES header recognition (`Requirement|Requisito`, `Scenario|Escenario`, `Capability|Capacidad`, `Requirements|Requisitos`).
- [ ] 1.6 Unit tests: empty spec, requirements only, deltas only, mixed, malformed (unknown delta op, scenario at wrong level, duplicate names), ES headers.

## Phase 2: planfmt validation

- [ ] 2.1 Add validation rule: every `### Requirement` (or `#### Requirement` under a Capability) has at least one nested Scenario.
- [ ] 2.2 Add validation rule: Scenario header level matches its Requirement (one deeper).
- [ ] 2.3 Add validation rule: unique Requirement names within a section; unique Scenario names within a Requirement.
- [ ] 2.4 Add validation rule: known delta op headers only.
- [ ] 2.5 Wire rules into `planfmt.Validate` strict-mode path.
- [ ] 2.6 Unit tests for each validation rule (positive + negative).

## Phase 3: Templates

- [ ] 3.1 Update `internal/app/newplan/newplan.go` `specTemplate` to include `## Capabilities` and `## Requirements` sections with one Requirement and one Scenario placeholder.
- [ ] 3.2 Mirror in Spanish template (`internal/assets/templates_es/` if applicable; otherwise inline ES branch in `specTemplate`).
- [ ] 3.3 Update `internal/app/newplan/newplan_test.go` to assert new sections present in EN and ES outputs.

## Phase 4: Init integration

- [ ] 4.1 Add `specsbaseline.InitBaseline(plansRoot string) error` — creates `.pacto/specs/` and `README.md` if missing.
- [ ] 4.2 Call from `internal/app/initws/` after existing init steps.
- [ ] 4.3 Verify idempotency: re-running init does not overwrite `README.md` or any capability files.
- [ ] 4.4 Unit + integration tests.

## Phase 5: Move done — baseline merge

- [ ] 5.1 In `internal/app/move/move.go`, when destination state is `done`, call `specsbaseline.ParseDeltas` on the moved plan's `spec.md`.
- [ ] 5.2 Implement `specsbaseline.MergeDeltas(plansRoot string, slug string, caps []Capability) (written []string, err error)` — handles ADDED / MODIFIED / REMOVED / RENAMED with atomic temp-file + rename.
- [ ] 5.3 On error, abort the move (do not rename the plan folder) and return a clear message.
- [ ] 5.4 Print a summary listing each baseline file written.
- [ ] 5.5 Cross-plan conflict detection: warn (not block in v1) if another active plan declares the same Capability with conflicting MODIFIED targets.
- [ ] 5.6 Integration tests: ADDED-only, MODIFIED-only, REMOVED-only, RENAMED-only, mixed, missing baseline file (ADDED creates it), MODIFIED missing target (errors), atomicity (simulate write failure on second capability — first must roll back).

## Phase 6: Status — per-Requirement coverage

- [ ] 6.1 In `internal/app/status/status.go`, parse Requirements per active plan.
- [ ] 6.2 Scan `tasks.md` for `R-NNN` references; count matches.
- [ ] 6.3 Reuse claim/evidence pipeline to count evidence per Requirement.
- [ ] 6.4 Add `requirements` block to JSON output (`{id, name, tasks, evidence, uncovered}`).
- [ ] 6.5 Add a Requirements coverage section to table output (collapsed when empty).
- [ ] 6.6 Integration tests: plan with full coverage, plan with one uncovered Requirement, legacy plan with no Requirements.

## Phase 7: Documentation

- [ ] 7.1 Update [docs/concepts.md](../../../../docs/concepts.md) with the baseline tree and Requirement grammar.
- [ ] 7.2 Update [docs/commands.md](../../../../docs/commands.md) for `move done` merge behaviour and new `status` columns.
- [ ] 7.3 Update [plans/PACTO.md](../../../../plans/PACTO.md) to describe baseline + delta semantics.

## Phase 8: Regression and polish

- [ ] 8.1 `go test ./...` passes.
- [ ] 8.2 Run end-to-end manually: init → new → edit spec with one ADDED Requirement under `## Capability: demo` → move done → confirm `.pacto/specs/demo/spec.md` exists with the Requirement.
- [ ] 8.3 Verify Spanish lifecycle: `pacto init --lang es` → `pacto new --lang es` → manual delta → move done.
- [ ] 8.4 Confirm legacy plans (no Capabilities/Requirements sections) work for `status`, `move done`, `exec`.

## Evidence

- <YYYY-MM-DD HH:MM> `<path|symbol|command>`

## Blockers

- None currently.

## Next Steps

1. Phase 1 (parsing) is pure-function work — start there; everything else depends on it.
2. Phase 2 (planfmt) and Phase 3 (templates) can run in parallel after Phase 1.
3. Phase 4, 5, 6 are independent of each other once Phases 1–3 land.
4. Phase 7 (docs) and Phase 8 (regression) close the plan.
