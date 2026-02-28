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

	got, ok := resolvePlanRoot(root)
	if !ok {
		t.Fatalf("expected root to resolve")
	}
	want := filepath.Join(root, ".pacto", "plans")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolvePlanRootFromWalksParents(t *testing.T) {
	root := t.TempDir()
	wantPlansRoot := filepath.Join(root, ".pacto", "plans")
	mustCreateStateDirs(t, wantPlansRoot)
	start := filepath.Join(root, "src", "pkg", "subpkg")
	if err := os.MkdirAll(start, 0o775); err != nil {
		t.Fatal(err)
	}

	gotPlansRoot, gotProjectRoot, ok := resolvePlanRootFrom(start)
	if !ok {
		t.Fatalf("expected root to resolve")
	}
	if gotPlansRoot != wantPlansRoot {
		t.Fatalf("expected plans root %q, got %q", wantPlansRoot, gotPlansRoot)
	}
	if gotProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, gotProjectRoot)
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
