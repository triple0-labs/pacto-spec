package parser

import (
	"os"
	"path/filepath"
	"testing"

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

func writePlan(t *testing.T, planText string) model.PlanRef {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "current", "sample")
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
		State:    "current",
		Slug:     "sample",
		Dir:      dir,
		Readme:   readme,
		PlanDocs: []string{plan},
	}
}
