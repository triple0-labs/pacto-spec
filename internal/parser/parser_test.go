package parser

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"pacto/internal/model"
)

func TestParsePlanStrictRequiresStatus(t *testing.T) {
	ref := writePlan(t, "# Plan\n\n## Phases\n| Phase | Desc | State | 10% |\n")
	_, err := ParsePlan(ref, "strict")
	if err == nil {
		t.Fatal("expected strict mode error for missing status")
	}
}

func TestParsePlanExtractsTotalProgressFallback(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\nProgress: 42%\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.Phases) != 1 {
		t.Fatalf("expected 1 derived phase, got %d", len(p.Phases))
	}
	if p.Phases[0].Progress != 42 {
		t.Fatalf("progress=%d, want 42", p.Phases[0].Progress)
	}
}

func TestParsePlanExtractsNextActions(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Next Steps\n1. Ship endpoint\n- [ ] Update docs\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.NextActions) != 2 {
		t.Fatalf("expected 2 next actions, got %d", len(p.NextActions))
	}
}

func TestParsePlanExtractsBlockersFromSection(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Blockers\n- waiting on API access\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.BlockerHints) != 1 {
		t.Fatalf("expected 1 blocker, got %d", len(p.BlockerHints))
	}
	if p.BlockerHints[0] != "waiting on API access" {
		t.Fatalf("blocker=%q", p.BlockerHints[0])
	}
}

func TestParsePlanIgnoresEmptyBlockersSectionPlaceholder(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Blockers\n- None currently.\n\n## Next Steps\n1. Ship endpoint\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.BlockerHints) != 0 {
		t.Fatalf("expected no blockers, got %v", p.BlockerHints)
	}
}

func TestParsePlanDoesNotTreatGoErrorReturnAsBlocked(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Phase 1: Setup\n- [ ] 1.1 Implement `InitContext(plansRoot string) error`\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(p.Tasks))
	}
	if p.Tasks[0].LikelyBlk {
		t.Fatalf("expected task not to be marked blocked: %+v", p.Tasks[0])
	}
}

func TestParsePlanExtractsPhaseTaskRefs(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Phase 1: Setup\n- [ ] 1.1 Define interfaces\n- [ ] 1.2 Add wiring\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(p.Tasks))
	}
	if p.Tasks[0].StepRef != "1.1" || p.Tasks[0].Phase != 1 || p.Tasks[0].Number != 1 {
		t.Fatalf("unexpected first task metadata: %+v", p.Tasks[0])
	}
	if p.Tasks[1].StepRef != "1.2" || p.Tasks[1].Phase != 1 || p.Tasks[1].Number != 2 {
		t.Fatalf("unexpected second task metadata: %+v", p.Tasks[1])
	}
}

func TestParsePlanExtractsPhaseTaskRefsSpanish(t *testing.T) {
	ref := writePlan(t, "Estado: En ejecución\n\n## Fase 1: Base\n- [ ] 1.1 Definir interfaces\n- [ ] 1.2 Conectar flujo\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.Tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(p.Tasks))
	}
	if p.Tasks[0].StepRef != "1.1" || p.Tasks[0].Phase != 1 || p.Tasks[0].Number != 1 {
		t.Fatalf("unexpected first task metadata: %+v", p.Tasks[0])
	}
	if p.Tasks[1].StepRef != "1.2" || p.Tasks[1].Phase != 1 || p.Tasks[1].Number != 2 {
		t.Fatalf("unexpected second task metadata: %+v", p.Tasks[1])
	}
}

func TestParsePlanStructuredDeltasEnglish(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Delta History\n### Delta D-2026-03-01-01\n- **Date:** 2026-03-01 10:30\n- **Type:** feat\n- **Status:** applied\n- **Changes:**\n  - `+ src/auth.go`\n  - `~ internal/parser/parser.go`\n- **Next Delta:** add tests\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if !p.HasStructuredDeltas {
		t.Fatal("expected HasStructuredDeltas=true")
	}
	if len(p.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(p.Deltas))
	}
	if got := p.Deltas[0].ID; got != "D-2026-03-01-01" {
		t.Fatalf("delta id=%q", got)
	}
	if p.Deltas[0].Date == nil {
		t.Fatal("expected parsed delta date")
	}
	if len(p.Deltas[0].Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(p.Deltas[0].Changes))
	}
	if p.LatestDeltaTime == nil {
		t.Fatal("expected LatestDeltaTime from structured delta")
	}
}

func TestParsePlanStructuredDeltasSpanishAliases(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Historial de deltas\n### Delta D-2026-03-02-01\n- **Fecha:** 2026-03-02 08:00\n- **Tipo:** fix\n- **Estado:** partial\n- **Cambios:**\n  - `+ src/api.go`\n- **Siguiente delta:** completar pruebas\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if !p.HasStructuredDeltas {
		t.Fatal("expected HasStructuredDeltas=true")
	}
	if len(p.Deltas) != 1 {
		t.Fatalf("expected 1 delta, got %d", len(p.Deltas))
	}
	if p.Deltas[0].Date == nil {
		t.Fatal("expected parsed Date from Fecha alias")
	}
	if got := p.Deltas[0].Status; got != "partial" {
		t.Fatalf("status=%q want partial", got)
	}
}

func TestParsePlanStructuredDeltaCompatWarning(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Delta History\n### Delta BAD-ID\n- **Date:** bad-date\n- **Status:** nonsense\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if len(p.ParseWarnings) == 0 {
		t.Fatal("expected parse warnings for malformed structured delta")
	}
}

func TestParsePlanStructuredDeltaStrictError(t *testing.T) {
	ref := writePlanInState(t, "done", "Status: Completed\n\n## Delta History\n### Delta BAD-ID\n- **Date:** bad-date\n")
	p, err := ParsePlan(ref, "strict")
	if err != nil {
		t.Fatalf("did not expect strict mode error for malformed structured delta: %v", err)
	}
	if len(p.ParseWarnings) == 0 {
		t.Fatal("expected warnings for malformed structured delta in strict mode")
	}
}

func TestParsePlanStrictDoneStateWarnOnlyForStructure(t *testing.T) {
	ref := writePlanInState(t, "done", "Status: Completed\n\n## Summary\n\nshort\n")
	_, err := ParsePlan(ref, "strict")
	if err != nil {
		t.Fatalf("expected done plans in strict mode to warn, not fail: %v", err)
	}
}

func TestParsePlanStrictCurrentStateFailsOnMissingSchema(t *testing.T) {
	ref := writePlanInState(t, "current", "Status: In Progress\n\n## Summary\n\nshort\n")
	_, err := ParsePlan(ref, "strict")
	if err == nil {
		t.Fatal("expected strict current plan to fail on missing required structure")
	}
}

func TestParsePlanLegacyDeltaFallbackWithoutSection(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\nRecent delta 2026-03-03 09:15\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if p.HasStructuredDeltas {
		t.Fatal("did not expect structured deltas")
	}
	if p.LatestDeltaTime == nil {
		t.Fatal("expected legacy fallback latest delta time")
	}
}

func TestParsePlanStructuredTakesPrecedenceOverLegacyHeuristic(t *testing.T) {
	ref := writePlan(t, "Status: In Progress\n\n## Delta History\n### Delta D-2026-03-01-01\n- **Date:** 2026-03-01 10:30\n- **Status:** applied\n\nSome random delta note 2026-03-04 18:00\n")
	p, err := ParsePlan(ref, "compat")
	if err != nil {
		t.Fatalf("ParsePlan returned error: %v", err)
	}
	if p.LatestDeltaTime == nil {
		t.Fatal("expected LatestDeltaTime")
	}
	want := time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC)
	if !p.LatestDeltaTime.Equal(want) {
		t.Fatalf("latest=%s want %s", p.LatestDeltaTime.Format("2006-01-02 15:04"), want.Format("2006-01-02 15:04"))
	}
}

func writePlan(t *testing.T, planText string) model.PlanRef {
	return writePlanInState(t, "current", planText)
}

func writePlanInState(t *testing.T, state, planText string) model.PlanRef {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, state, "sample")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(dir, "README.md")
	plan := filepath.Join(dir, "PLAN_SAMPLE.md")
	if err := os.WriteFile(readme, []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte(planText), 0o644); err != nil {
		t.Fatal(err)
	}
	return model.PlanRef{
		State:    state,
		Slug:     "sample",
		Dir:      dir,
		Readme:   readme,
		PlanDocs: []string{plan},
	}
}
