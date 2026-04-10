package statuscmd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
	"pacto/internal/model"
)

func TestRunStatusSplitRootsVerifiesRepoArtifact(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "sample")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{"to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_SAMPLE.md"), []byte("Status: In Progress\n- `src/auth.go`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workspace, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "src", "auth.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		code := RunStatus([]string{"--root", workspace, "--repo-root", workspace, "--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, `"source_text": "src/auth.go"`) {
		t.Fatalf("expected src/auth.go claim in output, got %q", stdout)
	}
	if !strings.Contains(stdout, `"result": "verified"`) {
		t.Fatalf("expected verified claim in output, got %q", stdout)
	}
}

func TestRunStatusDeprecatedPlansRootWarns(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "sample")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_SAMPLE.md"), []byte("Status: In Progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		code := RunStatus([]string{"--plans-root", plansRoot, "--repo-root", workspace, "--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stderr, "deprecated") {
		t.Fatalf("expected deprecated warning, got %q", stderr)
	}
}

func TestRunStatusAutoDetectsRootsFromNestedDir(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "sample")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, st := range []string{"to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_SAMPLE.md"), []byte("Status: In Progress\n- `src/auth.go`\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(workspace, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(workspace, "src", "auth.go"), []byte("package src\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nestedDir := filepath.Join(workspace, "src", "module", "nested")
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

	stdout, _ := captureOutput(t, func() {
		code := RunStatus([]string{"--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, `"source_text": "src/auth.go"`) {
		t.Fatalf("expected src/auth.go claim in output, got %q", stdout)
	}
	if !strings.Contains(stdout, `"result": "verified"`) {
		t.Fatalf("expected verified claim in output, got %q", stdout)
	}
}

func TestRunStatusIgnoresEmptyBlockersSection(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "to-implement", "sample")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	tasks := "# Tasks: Sample\n\n## Phase 1: Setup\n- [ ] 1.1 Add parser fix\n\n## Blockers\n- None currently.\n\n## Next Steps\n1. Add regression test\n"
	if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		code := RunStatus([]string{"--root", workspace, "--repo-root", workspace, "--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, `"blocked_tasks": 0`) {
		t.Fatalf("expected blocked_tasks 0, got %q", stdout)
	}
	if !strings.Contains(stdout, `"derived_status": "in_progress"`) {
		t.Fatalf("expected in_progress derived status, got %q", stdout)
	}
}

func TestRunStatusShowsOverlapWarningAndJSONField(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, ref := range []struct {
		state string
		slug  string
	}{
		{state: "current", slug: "plan-a"},
		{state: "to-implement", slug: "plan-b"},
	} {
		planDir := filepath.Join(plansRoot, ref.state, ref.slug)
		if err := os.MkdirAll(planDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(planDir, "spec.md"), []byte("# Spec\n\n## Domains Affected\n\n- auth\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(planDir, "tasks.md"), []byte("# Tasks\n\n## Phase 1: Setup\n\n- [ ] 1.1 Do work\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stdout, _ := captureOutput(t, func() {
		code := RunStatus([]string{"--root", workspace, "--repo-root", workspace, "--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, `"overlaps":`) {
		t.Fatalf("expected overlaps field in output, got %q", stdout)
	}
	if !strings.Contains(stdout, `"domain": "auth"`) {
		t.Fatalf("expected auth overlap in output, got %q", stdout)
	}
	if !strings.Contains(stdout, `domain overlap: auth shared by current/plan-a, to-implement/plan-b`) {
		t.Fatalf("expected overlap warning in plan warnings, got %q", stdout)
	}
}

func TestRunStatusExplicitFormatInTTYRendersTable(t *testing.T) {
	workspace := t.TempDir()
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	planDir := filepath.Join(plansRoot, "current", "sample")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(plansRoot, st), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "README.md"), []byte("# sample\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "PLAN_SAMPLE.md"), []byte("Status: In Progress\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	origIsTerminal := isTerminalFn
	origRunStatusUI := runStatusUI
	t.Cleanup(func() {
		isTerminalFn = origIsTerminal
		runStatusUI = origRunStatusUI
	})

	isTerminalFn = func(io.Writer) bool { return true }
	tuiCalled := false
	runStatusUI = func(model.StatusReport, i18n.Language) error {
		tuiCalled = true
		return nil
	}

	stdout, stderr := captureOutput(t, func() {
		code := RunStatus([]string{"--root", workspace, "--repo-root", workspace, "--format", "table"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if tuiCalled {
		t.Fatal("expected explicit --format to bypass TUI in terminal mode")
	}
	if stderr != "" {
		t.Fatalf("expected no stderr output, got %q", stderr)
	}
	if !strings.Contains(stdout, "PLANS_ROOT:") {
		t.Fatalf("expected table output, got %q", stdout)
	}
}
