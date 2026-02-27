package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunNewAutoDetectsRootFromNestedDir(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	mustCreateStateDirs(t, plansRoot)
	if err := os.WriteFile(filepath.Join(plansRoot, "README.md"), []byte("# Plans\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plansRoot, "PACTO.md"), []byte("# Pacto\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(plansRoot, "PLANTILLA_PACTO_PLAN.md"), []byte("# Plan: <Título del plan>\n\n**Date:** <YYYY-MM-DD>\n**Owner:** <nombre o equipo>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nestedDir := filepath.Join(workspace, "src", "pkg", "nested")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatal(err)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(nestedDir); err != nil {
		t.Fatal(err)
	}

	_, _ = captureOutput(t, func() {
		code := RunNew([]string{"to-implement", "auto-root-plan"})
		if code != 0 {
			t.Fatalf("RunNew returned %d, want 0", code)
		}
	})

	planDir := filepath.Join(plansRoot, "to-implement", "auto-root-plan")
	if _, err := os.Stat(filepath.Join(planDir, "README.md")); err != nil {
		t.Fatalf("expected README.md in plan dir: %v", err)
	}
	planDocs, err := filepath.Glob(filepath.Join(planDir, "PLAN_*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(planDocs) != 1 {
		t.Fatalf("expected exactly one PLAN_*.md file, got %d", len(planDocs))
	}
}
