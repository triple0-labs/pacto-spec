# Tasks: System Source of Truth

## Execution Metadata

- Status: Draft
- Owner: pacto-core
- Created: 2026-04-08
- Last Modified: 2026-04-09
- State: to-implement
- Slug: system-source-of-truth

## Implementation Plan by Phases

## Phase 1: Context Package — Domain Logic

- [ ] 1.1 Create `internal/context/` package with types and constants
- [ ] 1.2 Implement `ExtractDomains(specPath string) []string` — read `## Domains Affected` from spec.md
- [ ] 1.3 Implement `NormalizeDomainSlug(raw string) string` — convert freeform domain names into stable filenames
- [ ] 1.4 Implement `ReadContextDomains(contextDir string) []string` — list existing docs under `.pacto/context/domains/`
- [ ] 1.5 Implement `EnsureDomainDocs(contextDir string, domains []string, planRef string) error` — create/update `.pacto/context/domains/<domain>.md`
- [ ] 1.6 Implement `DetectOverlaps(planDomains map[string][]string) []Overlap` — overlap detection
- [ ] 1.7 Implement `InitContext(plansRoot string) error` — create `.pacto/context/README.md` plus `.pacto/context/domains/`
- [ ] 1.8 Write unit tests for all functions in `internal/context/context_test.go`

## Phase 2: Spec Template Update

- [ ] 2.1 Add `## Domains Affected` section to `defaultSpecTemplate` in `internal/command/newcmd/new.go`
- [ ] 2.2 Add `## Dominios Afectados` section to Spanish spec template
- [ ] 2.3 Update `internal/command/newcmd/new_test.go` — verify new section present in generated spec
- [ ] 2.4 Verify new section present in generated spec (`go test` / `new_test.go`)

## Phase 3: Init Integration

- [ ] 3.1 Call `context.InitContext` from `RunInit` after creating plans structure
- [ ] 3.2 Ensure idempotent init preserves existing `.pacto/context/README.md` and existing domain docs
- [ ] 3.3 Update `internal/command/initcmd/init_test.go` — verify context workspace created
- [ ] 3.4 Verify context workspace after init (`go test` / `init_test.go`)

## Phase 4: Move Done Integration

- [ ] 4.1 Hook domain extraction into `move.MovePlan` or the `RunMove` caller for `done` destination
- [ ] 4.2 Call `context.ExtractDomains` on moved plan's spec.md
- [ ] 4.3 Call `context.EnsureDomainDocs` to create/update per-domain docs under `.pacto/context/domains/`
- [ ] 4.4 Print Tier 2 enrichment prompt to stdout after successful move-to-done, including affected domain doc paths
- [ ] 4.5 Handle missing/malformed domains section gracefully
- [ ] 4.6 Write integration tests: `TestMoveDoneCreatesDomainDocs`, `TestMoveDonePrintsEnrichmentPrompt`
- [ ] 4.7 Verify domain extraction after move done (`go test` / `move_test.go`)

## Phase 5: Status Overlap Detection

- [ ] 5.1 In `internal/command/status/status.go`, read domains from all active plans after plan discovery
- [ ] 5.2 Run `DetectOverlaps` and collect warnings
- [ ] 5.3 Add overlap warnings to table format output
- [ ] 5.4 Add `"overlaps"` field to JSON format output
- [ ] 5.5 Write integration tests: `TestStatusShowsOverlapWarning`, `TestStatusOverlapsInJSON`
- [ ] 5.6 Verify overlap warning in status output (`go test` / `status_test.go`)

## Phase 6: Regression and Polish

- [ ] 6.1 Run `go test ./...` — all existing tests pass
- [ ] 6.2 Run `go test ./...` — all tests pass
- [ ] 6.3 Write `TestFullLifecycleIntegration` end-to-end test
- [ ] 6.4 Verify Spanish template works end-to-end (init --lang es → new → move done → domain docs created)
- [ ] 6.5 Verify existing plans without domains don't break status or move

## Evidence

- <YYYY-MM-DD HH:MM> `<path|symbol|command>`

## Blockers

- None currently.

## Next Steps

1. Start with Phase 1 (context package) — all logic is pure and testable in isolation.
2. Phase 2 (template) and Phase 3 (init) are independent and can proceed in parallel after Phase 1.
3. Phase 4 (move) and Phase 5 (status) depend on Phase 1 but are independent of each other.
