package movecmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMoveTransitionsStateAndUpdatesPlanReadme(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	if code := RunNew([]string{"to-implement", "move-sample", "--root", root}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	_, stderr := captureOutput(t, func() {
		code := RunMove([]string{"to-implement", "move-sample", "done", "--root", root, "--reason", "completed work"})
		if code != 0 {
			t.Fatalf("RunMove returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	dstReadme := filepath.Join(plansRoot, "done", "move-sample", "README.md")
	if _, err := os.Stat(dstReadme); err != nil {
		t.Fatalf("expected moved README: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plansRoot, "to-implement", "move-sample")); !os.IsNotExist(err) {
		t.Fatalf("expected source dir removed")
	}

	b, err := os.ReadFile(dstReadme)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "**Status:** Completed (Done)") {
		t.Fatalf("expected status updated to done, got %q", text)
	}
	if !strings.Contains(text, "## Move History") || !strings.Contains(text, "completed work") {
		t.Fatalf("expected move history note, got %q", text)
	}

}

func TestRunMoveDoneCreatesDomainDocsAndPrompt(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--no-interactive", "--no-install"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}
	if code := RunNew([]string{"to-implement", "domain-move", "--root", root}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	specPath := filepath.Join(root, ".pacto", "plans", "to-implement", "domain-move", "spec.md")
	spec := "# Spec: Domain Move\n\n## Acceptance Criteria\n\n- AC-001: done.\n\n## Domains Affected\n\n- auth\n- session\n"
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr := captureOutput(t, func() {
		code := RunMove([]string{"to-implement", "domain-move", "done", "--root", root, "--reason", "completed work"})
		if code != 0 {
			t.Fatalf("RunMove returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}
	if !strings.Contains(stdout, "auth.md") || !strings.Contains(stdout, "session.md") {
		t.Fatalf("expected affected domain docs in output, got %q", stdout)
	}
	if !strings.Contains(stdout, "Review design.md and update the affected domain docs") {
		t.Fatalf("expected enrichment prompt, got %q", stdout)
	}

	for _, name := range []string{"auth.md", "session.md"} {
		path := filepath.Join(root, ".pacto", "context", "domains", name)
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected domain doc %s: %v", name, err)
		}
		if !strings.Contains(string(b), "- done/domain-move") {
			t.Fatalf("expected related plan in %s, got %q", name, string(b))
		}
	}
}
