package newcmd

import (
	"os"
	"path/filepath"
	"strings"
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
	for _, name := range []string{"spec.md", "design.md", "tasks.md"} {
		if _, err := os.Stat(filepath.Join(planDir, name)); err != nil {
			t.Fatalf("expected %s in split layout: %v", name, err)
		}
	}
	planDocs, err := filepath.Glob(filepath.Join(planDir, "PLAN_*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(planDocs) != 0 {
		t.Fatalf("expected no PLAN_*.md files for split default, got %d", len(planDocs))
	}
}

func TestRunNewPrintsRelativePathsFromCWD(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--no-interactive", "--no-install"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		if code := RunNew([]string{"to-implement", "relative-output"}); code != 0 {
			t.Fatalf("RunNew returned %d", code)
		}
	})
	out := filepath.ToSlash(stdout)

	if !strings.Contains(out, ".pacto/plans/to-implement/relative-output/README.md") {
		t.Fatalf("expected relative README path in output, got %q", stdout)
	}
}

func TestRunNewMinimalRootDoesNotReferenceLegacySlashCommands(t *testing.T) {
	root := t.TempDir()

	if code := RunNew([]string{"to-implement", "minimal-root", "--root", root, "--allow-minimal-root"}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	plansReadme, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(plansReadme), "SLASH_COMMANDS.md") {
		t.Fatalf("expected minimal root README to avoid legacy slash commands reference, got %q", string(plansReadme))
	}
	if _, err := os.Stat(filepath.Join(root, "SLASH_COMMANDS.md")); !os.IsNotExist(err) {
		t.Fatalf("expected minimal root to avoid creating SLASH_COMMANDS.md, got err=%v", err)
	}
}

func TestRunNewSpecIncludesDomainsAffected(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--no-interactive", "--no-install"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	if code := RunNew([]string{"to-implement", "domain-scaffold", "--root", root}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	specPath := filepath.Join(root, ".pacto", "plans", "to-implement", "domain-scaffold", "spec.md")
	b, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "## Domains Affected") || !strings.Contains(text, "- <domain>") {
		t.Fatalf("expected domains affected scaffold, got %q", text)
	}
}
