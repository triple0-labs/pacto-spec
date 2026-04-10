package execmd

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
	plan := "# Plan: Sample Exec\n\n## Phase 1: Setup\n\n- [ ] 1.1 first task\n- [ ] 1.2 second task\n"
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
	if !strings.Contains(got, "- [x] 1.1 first task") {
		t.Fatalf("expected 1.1 completed, got %q", got)
	}
	if !strings.Contains(got, "## Execution Notes") || !strings.Contains(got, "ran checklist") {
		t.Fatalf("expected execution note section, got %q", got)
	}
	if !strings.Contains(got, "## Evidence") || !strings.Contains(got, "`src/auth.go`") {
		t.Fatalf("expected evidence section, got %q", got)
	}
	if !strings.Contains(got, "**Last Modified:** ") {
		t.Fatalf("expected Last Modified metadata update, got %q", got)
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
	orig := "# Plan: Dry Run\n\n## Phase 1\n\n- [ ] 1.1 first task\n"
	if err := os.WriteFile(planPath, []byte(orig), 0o664); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		code := RunExec([]string{"current", "dry-run-exec", "--root", root, "--dry-run"})
		if code != 0 {
			t.Fatalf("RunExec returned %d", code)
		}
	})
	if !strings.Contains(stdout, "Dry Run") && !strings.Contains(stdout, "Simulación") {
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
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_NEEDS_MOVE.md"), []byte("# Plan\n\n## Phase 1\n\n- [ ] 1.1 a task\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"to-implement", "needs-move", "--root", root})
		if code != 2 {
			t.Fatalf("RunExec returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "only supports state") && !strings.Contains(stderr, "solo soporta el estado") {
		t.Fatalf("expected state restriction message, got %q", stderr)
	}
	if !strings.Contains(stderr, "pacto move to-implement needs-move current") {
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
	if err := os.WriteFile(planPath, []byte("# Plan: Invalid Step\n\n## Phase 1\n\n- [ ] 1.1 a task\n"), 0o664); err != nil {
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

func TestRunExecRejectsLegacyStepFormat(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "legacy-step")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# Legacy Step\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_LEGACY_STEP.md")
	if err := os.WriteFile(planPath, []byte("# Plan: Legacy Step\n\n## Phase 1\n\n- [ ] 1.1 a task\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"current", "legacy-step", "--root", root, "--step", "T1"})
		if code != 2 {
			t.Fatalf("RunExec returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "legacy --step") {
		t.Fatalf("expected legacy format error, got %q", stderr)
	}
}

func TestRunExecRejectsPlanWithoutPhaseRefs(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "no-phase-refs")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# No Phase Refs\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	planPath := filepath.Join(planDir, "PLAN_NO_PHASE_REFS.md")
	if err := os.WriteFile(planPath, []byte("# Plan\n\n## Tasks\n\n- [ ] plain task\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{"current", "no-phase-refs", "--root", root})
		if code != 2 {
			t.Fatalf("RunExec returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "no phase tasks found") {
		t.Fatalf("expected phase task contract error, got %q", stderr)
	}
}

func TestRunExecPrefersTasksDocumentForSplitLayout(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "split-pref")
	if err := os.MkdirAll(planDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# Split Pref\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	tasksPath := filepath.Join(planDir, "tasks.md")
	if err := os.WriteFile(tasksPath, []byte("# Tasks\n\n## Phase 1\n\n- [ ] 1.1 do it\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	legacyPlanPath := filepath.Join(planDir, "PLAN_SPLIT_PREF.md")
	legacyOriginal := "# Plan\n\n## Phase 1\n\n- [ ] 1.1 should stay untouched\n"
	if err := os.WriteFile(legacyPlanPath, []byte(legacyOriginal), 0o664); err != nil {
		t.Fatal(err)
	}

	if code := RunExec([]string{"current", "split-pref", "--root", root}); code != 0 {
		t.Fatalf("RunExec returned %d", code)
	}

	tasksBody, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(tasksBody), "- [x] 1.1 do it") {
		t.Fatalf("expected tasks.md to be updated, got %q", string(tasksBody))
	}
	legacyBody, err := os.ReadFile(legacyPlanPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(legacyBody) != legacyOriginal {
		t.Fatalf("expected legacy PLAN file to remain unchanged")
	}
}

func TestRunExecHandlesSpanishTasksTemplate(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--no-interactive", "--no-install", "--lang", "es"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}
	if code := RunNew([]string{"current", "spanish-exec", "--root", filepath.Join(root, ".pacto", "plans")}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	_, stderr := captureOutput(t, func() {
		code := RunExec([]string{
			"current", "spanish-exec", "--root", root,
			"--note", "se validó el flujo", "--evidence", "src/api.go", "--blocker", "pendiente QA",
		})
		if code != 0 {
			t.Fatalf("RunExec returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	tasksPath := filepath.Join(root, ".pacto", "plans", "current", "spanish-exec", "tasks.md")
	body, err := os.ReadFile(tasksPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	if !strings.Contains(got, "- [x] 1.1 <tarea>") {
		t.Fatalf("expected first spanish task completion, got %q", got)
	}
	if !strings.Contains(got, "## Notas de Ejecución") || !strings.Contains(got, "se validó el flujo") {
		t.Fatalf("expected spanish execution notes section, got %q", got)
	}
	if !strings.Contains(got, "## Evidencia") || !strings.Contains(got, "`src/api.go`") {
		t.Fatalf("expected spanish evidence section, got %q", got)
	}
	if !strings.Contains(got, "## Bloqueadores") || !strings.Contains(got, "pendiente QA") {
		t.Fatalf("expected spanish blockers section, got %q", got)
	}
	if !strings.Contains(got, "- Última Modificación: ") {
		t.Fatalf("expected spanish metadata last-modified update, got %q", got)
	}
	if strings.Contains(got, "**Last Modified:**") {
		t.Fatalf("did not expect english readme-style metadata insertion, got %q", got)
	}
}
