package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNormalizeDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "demo")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_DEMO.md")
	orig := "# Plan: Demo\n\n## Objetivos\n\n1. x\n"
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte(orig), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureOutput(t, func() {
		code := RunNormalize([]string{"--root", root})
		if code != 0 {
			t.Fatalf("RunNormalize returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "would-change") {
		t.Fatalf("expected dry-run status in output, got %q", stdout)
	}
	b, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != orig {
		t.Fatalf("expected dry-run to avoid writing files")
	}
}

func TestRunNormalizeWriteAppliesChanges(t *testing.T) {
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "demo")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_DEMO.md")
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(planPath, []byte("# Plan: Demo\n\n## Fase 1: Setup\n- [ ] T2. arreglar\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunNormalize([]string{"--root", root, "--write"})
		if code != 0 {
			t.Fatalf("RunNormalize returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	b, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, "- [ ] 1.2 arreglar") {
		t.Fatalf("expected converted task ID, got %q", got)
	}
	if strings.Contains(got, "## Metadatos") || strings.Contains(got, "## Metadata") {
		t.Fatalf("did not expect normalize to auto-add metadata section, got %q", got)
	}
}
