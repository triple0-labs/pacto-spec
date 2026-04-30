package normalizeplans

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupWorkspace(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o775); err != nil {
			t.Fatal(err)
		}
	}
	return root, plansRoot
}

func writePlan(t *testing.T, plansRoot, state, slug, content string) string {
	t.Helper()
	dir := filepath.Join(plansRoot, state, slug)
	if err := os.MkdirAll(dir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# "+slug+"\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(dir, "PLAN_"+strings.ToUpper(slug)+".md")
	if err := os.WriteFile(planPath, []byte(content), 0o664); err != nil {
		t.Fatal(err)
	}
	return planPath
}

func TestApply_DryRunReportsChangesWithoutWriting(t *testing.T) {
	root, plansRoot := setupWorkspace(t)
	planPath := writePlan(t, plansRoot, "current", "demo", "# Plan: Demo\n\n## Fase 1: Setup\n- [ ] T2. arreglar\n")
	original, _ := os.ReadFile(planPath)

	report, err := Apply(Input{Root: root})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if report.TotalPlans != 1 || report.ChangedPlans != 1 || report.AppliedPlans != 0 {
		t.Fatalf("unexpected counters: %+v", report)
	}
	after, _ := os.ReadFile(planPath)
	if string(after) != string(original) {
		t.Fatal("dry-run must not modify the file")
	}
	if !report.Items[0].Changed || report.Items[0].Applied {
		t.Fatalf("expected Changed=true Applied=false, got %+v", report.Items[0])
	}
}

func TestApply_WriteAppliesChanges(t *testing.T) {
	root, plansRoot := setupWorkspace(t)
	planPath := writePlan(t, plansRoot, "current", "demo", "# Plan: Demo\n\n## Fase 1: Setup\n- [ ] T2. arreglar\n")
	original, _ := os.ReadFile(planPath)

	report, err := Apply(Input{Root: root, Write: true})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if report.AppliedPlans != 1 {
		t.Fatalf("expected AppliedPlans=1, got %+v", report)
	}
	after, _ := os.ReadFile(planPath)
	if string(after) == string(original) {
		t.Fatal("expected file content to change after write")
	}
	if !report.Items[0].Applied {
		t.Fatal("expected item Applied=true")
	}
}

func TestApply_NoChangesYieldsZeroChanged(t *testing.T) {
	clean := "# Plan: Clean\n\n## Phase 1: Setup\n\n- [ ] 1.1 do thing\n"
	root, plansRoot := setupWorkspace(t)
	writePlan(t, plansRoot, "current", "clean", clean)

	report, err := Apply(Input{Root: root})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if report.ChangedPlans != 0 {
		t.Fatalf("expected ChangedPlans=0, got %+v", report)
	}
}

func TestApply_RejectsUnresolvableRoot(t *testing.T) {
	_, err := Apply(Input{Root: t.TempDir()})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestApply_StateFilter(t *testing.T) {
	root, plansRoot := setupWorkspace(t)
	writePlan(t, plansRoot, "current", "alpha", "# Plan: Alpha\n\n## Phase 1\n\n- [ ] 1.1 x\n")
	writePlan(t, plansRoot, "to-implement", "beta", "# Plan: Beta\n\n## Phase 1\n\n- [ ] 1.1 x\n")

	report, err := Apply(Input{Root: root, State: "current"})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if report.TotalPlans != 1 || report.Items[0].Slug != "alpha" {
		t.Fatalf("expected only alpha, got %+v", report)
	}
}
