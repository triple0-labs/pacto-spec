package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindPlansByStateFilter(t *testing.T) {
	root := t.TempDir()
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	planDir := filepath.Join(root, "current", "sample")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_SAMPLE_2026-01-01.md"), []byte("# plan"), 0o644); err != nil {
		t.Fatal(err)
	}

	plans, err := FindPlans(root, Options{StateFilter: "current"})
	if err != nil {
		t.Fatal(err)
	}
	if len(plans) != 1 {
		t.Fatalf("expected 1 plan, got %d", len(plans))
	}
	if plans[0].State != "current" || plans[0].Slug != "sample" {
		t.Fatalf("unexpected plan ref: %#v", plans[0])
	}
}
