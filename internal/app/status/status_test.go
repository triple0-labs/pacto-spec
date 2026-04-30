package status

import (
	"os"
	"path/filepath"
	"testing"
)

// writePlan creates a minimal plan layout under root/<state>/<slug>/ with a
// PLAN_*.md file containing the given body, plus an empty README.md and
// optional spec.md.
func writePlan(t *testing.T, root, state, slug, body, spec string) {
	t.Helper()
	dir := filepath.Join(root, state, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# "+slug), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "PLAN_"+slug+"_2026-01-01.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if spec != "" {
		if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(spec), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func setupPlansRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestBuildReport_EmptyPlansRoot(t *testing.T) {
	root := setupPlansRoot(t)

	rep, err := BuildReport(Input{
		PlansRoot: root,
		RepoRoot:  root,
		Mode:      "lenient",
	})
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if len(rep.Plans) != 0 {
		t.Fatalf("expected 0 plans, got %d", len(rep.Plans))
	}
	if len(rep.Overlaps) != 0 {
		t.Fatalf("expected 0 overlaps, got %d", len(rep.Overlaps))
	}
}

func TestBuildReport_DiscoversPlans(t *testing.T) {
	root := setupPlansRoot(t)
	writePlan(t, root, "current", "alpha", "# Plan: alpha\n", "")
	writePlan(t, root, "to-implement", "beta", "# Plan: beta\n", "")

	rep, err := BuildReport(Input{
		PlansRoot: root,
		RepoRoot:  root,
		Mode:      "lenient",
	})
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if len(rep.Plans) != 2 {
		t.Fatalf("expected 2 plans, got %d", len(rep.Plans))
	}
}

func TestBuildReport_StateFilterHonoured(t *testing.T) {
	root := setupPlansRoot(t)
	writePlan(t, root, "current", "alpha", "# Plan: alpha\n", "")
	writePlan(t, root, "to-implement", "beta", "# Plan: beta\n", "")

	rep, err := BuildReport(Input{
		PlansRoot: root,
		RepoRoot:  root,
		Mode:      "lenient",
		State:     "current",
	})
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if len(rep.Plans) != 1 {
		t.Fatalf("expected 1 plan after state filter, got %d", len(rep.Plans))
	}
	if rep.Plans[0].StateFolder != "current" {
		t.Fatalf("expected plan state=current, got %s", rep.Plans[0].StateFolder)
	}
}

func TestBuildReport_PropagatesWarnings(t *testing.T) {
	root := setupPlansRoot(t)
	writePlan(t, root, "current", "alpha", "# Plan: alpha\n", "")

	rep, err := BuildReport(Input{
		PlansRoot: root,
		RepoRoot:  root,
		Mode:      "lenient",
		Warnings:  []string{"config: deprecated foo"},
	})
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if len(rep.Plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(rep.Plans))
	}
	found := false
	for _, w := range rep.Plans[0].ParseWarnings {
		if w == "config: deprecated foo" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected warning to propagate to plan; got %v", rep.Plans[0].ParseWarnings)
	}
}

func TestBuildReport_DetectsDomainOverlaps(t *testing.T) {
	root := setupPlansRoot(t)
	spec := "# Spec\n\n## Domains Affected\n\n- payments\n- billing\n"
	writePlan(t, root, "current", "alpha", "# Plan: alpha\n", spec)
	writePlan(t, root, "to-implement", "beta", "# Plan: beta\n", spec)

	rep, err := BuildReport(Input{
		PlansRoot: root,
		RepoRoot:  root,
		Mode:      "lenient",
	})
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if len(rep.Overlaps) == 0 {
		t.Fatalf("expected at least one domain overlap; got none")
	}
}

func TestBuildReport_MissingPlansRootIsEmpty(t *testing.T) {
	// Discovery treats a missing plans root as "no plans" rather than an
	// error; the use case mirrors that behavior.
	rep, err := BuildReport(Input{
		PlansRoot: "/nonexistent/path/that/should/not/exist",
		RepoRoot:  "/tmp",
		Mode:      "lenient",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rep.Plans) != 0 {
		t.Fatalf("expected 0 plans, got %d", len(rep.Plans))
	}
}
