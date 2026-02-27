package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		code := RunStatus([]string{"--plans-root", plansRoot, "--repo-root", workspace, "--format", "json"})
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

func TestRunStatusDeprecatedRootWarns(t *testing.T) {
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
		code := RunStatus([]string{"--root", workspace, "--format", "json"})
		if code != 0 {
			t.Fatalf("RunStatus returned %d, want 0", code)
		}
	})
	if !strings.Contains(stderr, "deprecated") {
		t.Fatalf("expected deprecated warning, got %q", stderr)
	}
}
