package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
)

func TestHasStateDirsTrueWhenAllPresent(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, root)
	if !HasStateDirs(root) {
		t.Error("expected true")
	}
}

func TestHasStateDirsFalseWhenMissing(t *testing.T) {
	root := t.TempDir()
	for _, st := range []string{"current", "to-implement", "done"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o775); err != nil {
			t.Fatal(err)
		}
	}
	if HasStateDirs(root) {
		t.Error("expected false: missing 'outdated'")
	}
}

func TestHasStateDirsFalseWhenStateIsAFile(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, root)
	// Replace one dir with a file.
	target := filepath.Join(root, "current")
	if err := os.RemoveAll(target); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	if HasStateDirs(root) {
		t.Error("expected false when 'current' is a file")
	}
}

func TestCleanAbsHandlesRelative(t *testing.T) {
	got := CleanAbs(".")
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute path, got %q", got)
	}
}

func TestPathExists(t *testing.T) {
	root := t.TempDir()
	if !PathExists(root) {
		t.Error("temp dir should exist")
	}
	if PathExists(filepath.Join(root, "missing")) {
		t.Error("missing path should not exist")
	}
}

func TestResolvePlansRootForActionWithExplicitRoot(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, root)
	got, err := ResolvePlansRootForAction(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != root {
		t.Errorf("got %q want %q", got, root)
	}
}

func TestResolvePlansRootForActionErrorWhenNoStateDirs(t *testing.T) {
	root := t.TempDir()
	_, err := ResolvePlansRootForAction(root)
	if err == nil {
		t.Fatal("expected error for missing state dirs")
	}
	if !strings.Contains(err.Error(), "could not resolve plans root") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestResolvePlansRootForActionWalksFromCwd(t *testing.T) {
	root := t.TempDir()
	mustCreateStateDirs(t, filepath.Join(root, ".pacto", "plans"))
	subdir := filepath.Join(root, "src", "pkg")
	if err := os.MkdirAll(subdir, 0o775); err != nil {
		t.Fatal(err)
	}

	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(subdir); err != nil {
		t.Fatal(err)
	}

	got, err := ResolvePlansRootForAction("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join(root, ".pacto", "plans")
	// Resolve symlinks to compare canonical paths (e.g. /tmp vs /private/tmp on macOS).
	gotResolved, _ := filepath.EvalSymlinks(got)
	wantResolved, _ := filepath.EvalSymlinks(want)
	if gotResolved != wantResolved {
		t.Errorf("got %q want %q", got, want)
	}
}

func TestValidateRootDetectsMissingFiles(t *testing.T) {
	root := t.TempDir()
	if err := ValidateRoot(root); err == nil {
		t.Fatal("expected error")
	}
	mustCreateStateDirs(t, root)
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "PACTO.md"), []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRoot(root); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateRootMissingStateFolder(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "PACTO.md"), []byte("x"), 0o664); err != nil {
		t.Fatal(err)
	}
	err := ValidateRoot(root)
	if err == nil || !strings.Contains(err.Error(), "missing state folder") {
		t.Errorf("expected missing state folder error, got %v", err)
	}
}

func TestEnsureMinimalRootCreatesScaffold(t *testing.T) {
	root := filepath.Join(t.TempDir(), "newroot")
	if err := EnsureMinimalRoot(root, i18n.English); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ValidateRoot(root); err != nil {
		t.Errorf("scaffolded root failed validation: %v", err)
	}
	// Idempotent.
	if err := EnsureMinimalRoot(root, i18n.English); err != nil {
		t.Errorf("second call failed: %v", err)
	}
}

func TestResolveExploreRootFallsBackToCwd(t *testing.T) {
	root := t.TempDir()
	prev, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(prev) })
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveExploreRoot("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	gotResolved, _ := filepath.EvalSymlinks(got)
	rootResolved, _ := filepath.EvalSymlinks(root)
	if gotResolved != rootResolved {
		t.Errorf("got %q want %q", got, root)
	}
}

func TestResolveExploreRootHonorsExplicitArg(t *testing.T) {
	root := t.TempDir()
	got, err := ResolveExploreRoot(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Errorf("expected absolute, got %q", got)
	}
}
