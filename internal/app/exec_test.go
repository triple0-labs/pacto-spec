package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExecCompletesNextTaskAndAppendsEvidence(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "sample-exec")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	readme := "# Sample Exec\n\n**Status:** Pending (To Implement)  \n**Date:** 2026-02-28\n"
	plan := "# Plan: Sample Exec\n\n## Tasks\n\n- [ ] T1. first task\n- [ ] T2. second task\n"
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte(readme), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_SAMPLE_EXEC.md")
	if err := os.WriteFile(planPath, []byte(plan), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"current", "sample-exec", "--root", root, "--note", "ran checklist", "--evidence", "src/auth.go"})
		if code != 0 {
			t.Fatalf("RunExec returned %d", code)
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
	if !strings.Contains(got, "- [x] T1. first task") {
		t.Fatalf("expected T1 completed, got %q", got)
	}
	if !strings.Contains(got, "## Execution Notes") || !strings.Contains(got, "ran checklist") {
		t.Fatalf("expected execution note section, got %q", got)
	}
	if !strings.Contains(got, "## Evidence") || !strings.Contains(got, "`src/auth.go`") {
		t.Fatalf("expected evidence section, got %q", got)
	}
}

func TestRunExecDryRunDoesNotWrite(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "dry-run-exec")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# Dry Run\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_DRY_RUN_EXEC.md")
	orig := "# Plan: Dry Run\n\n- [ ] T1. first task\n"
	if err := os.WriteFile(planPath, []byte(orig), 0o664); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		code := RunExec([]string{"current", "dry-run-exec", "--root", root, "--dry-run"})
		if code != 0 {
			t.Fatalf("RunExec returned %d", code)
		}
	})
	if !strings.Contains(stdout, "[dry-run] would update") {
		t.Fatalf("expected dry-run output, got %q", stdout)
	}

	b, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != orig {
		t.Fatalf("dry-run mutated file")
	}
}

func TestRunExecRejectsNonCurrentState(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "to-implement", "needs-move")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# Needs Move\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_NEEDS_MOVE.md"), []byte("# Plan\n\n- [ ] T1. a task\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"to-implement", "needs-move", "--root", root})
		if code != 2 {
			t.Fatalf("RunExec returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "only supports state") {
		t.Fatalf("expected state restriction message, got %q", stderr)
	}
	if !strings.Contains(stderr, "trigger: pacto move to-implement needs-move current") {
		t.Fatalf("expected explicit trigger command, got %q", stderr)
	}
}

func TestRunExecRejectsInvalidStepFormat(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "invalid-step")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# Invalid Step\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_INVALID_STEP.md")
	if err := os.WriteFile(planPath, []byte("# Plan: Invalid Step\n\n- [ ] T1. a task\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"current", "invalid-step", "--root", root, "--step", "1"})
		if code != 2 {
			t.Fatalf("RunExec returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "invalid --step") {
		t.Fatalf("expected invalid --step error, got %q", stderr)
	}
}
