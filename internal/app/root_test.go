package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePlanRootPrefersDirectStateDirs(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, root)

	got, ok := resolvePlanRoot(root)
	if !ok {
		t.Fatalf("expected root to resolve")
	}
	if got != root {
		t.Fatalf("expected %q, got %q", root, got)
	}
}

func TestResolvePlanRootPrefersDotPactoPlans(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, filepath.Join(root, ".pacto", "plans"))
	mustCreateStateDirs(t, filepath.Join(root, "plans"))

	got, ok := resolvePlanRoot(root)
	if !ok {
		t.Fatalf("expected root to resolve")
	}
	want := filepath.Join(root, ".pacto", "plans")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolvePlanRootFallsBackToPlans(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, filepath.Join(root, "plans"))

	got, ok := resolvePlanRoot(root)
	if !ok {
		t.Fatalf("expected root to resolve")
	}
	want := filepath.Join(root, "plans")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func mustCreateStateDirs(t *testing.T, root string) {
	t.Helper()
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o775); err != nil {
			t.Fatal(err)
		}
	}
}
