package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitCreatesWorkspace(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		assertPathExists(t, filepath.Join(plansRoot, st))
	}
	for _, f := range []string{"README.md", "PACTO.md", "PLANTILLA_PACTO_PLAN.md", "SLASH_COMMANDS.md"} {
		assertPathExists(t, filepath.Join(plansRoot, f))
	}
	assertPathExists(t, filepath.Join(root, ".pacto", "config.yaml"))
	cfg, err := os.ReadFile(filepath.Join(root, ".pacto", "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cfg), "editor:") || strings.Contains(string(cfg), "custom_editor:") {
		t.Fatalf("config should not include editor fields: %s", string(cfg))
	}
	assertPathExists(t, filepath.Join(root, "prd.md"))
	if _, err := os.Stat(filepath.Join(root, "AGENTS.md")); !os.IsNotExist(err) {
		t.Fatalf("AGENTS.md should not be created by default")
	}
}

func TestRunInitIdempotentAndForce(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("first RunInit returned %d", code)
	}

	readmePath := filepath.Join(root, ".pacto", "plans", "README.md")
	if err := os.WriteFile(readmePath, []byte("custom readme"), 0o664); err != nil {
		t.Fatal(err)
	}

	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("second RunInit returned %d", code)
	}
	got, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(got)) != "custom readme" {
		t.Fatalf("README should not be overwritten without --force")
	}

	if code := RunInit([]string{"--root", root, "--force"}); code != 0 {
		t.Fatalf("forced RunInit returned %d", code)
	}
	got, err = os.ReadFile(readmePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "# Pacto Plans") {
		t.Fatalf("README was not overwritten with managed template")
	}
}

func TestRunInitWithAgentsManagedBlock(t *testing.T) {
	root := t.TempDir()
	agentsPath := filepath.Join(root, "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("Team notes\n"), 0o664); err != nil {
		t.Fatal(err)
	}

	if code := RunInit([]string{"--root", root, "--with-agents"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}
	first, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(first)
	if !strings.Contains(text, "Team notes") {
		t.Fatalf("existing AGENTS content should be preserved")
	}
	if strings.Count(text, agentsManagedStart) != 1 || strings.Count(text, agentsManagedEnd) != 1 {
		t.Fatalf("expected one managed block in AGENTS.md")
	}

	if code := RunInit([]string{"--root", root, "--with-agents"}); code != 0 {
		t.Fatalf("second RunInit returned %d", code)
	}
	second, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(second), agentsManagedStart) != 1 || strings.Count(string(second), agentsManagedEnd) != 1 {
		t.Fatalf("managed block should not be duplicated")
	}
}

func TestRunInitUpdatesManagedPRDBlockOnly(t *testing.T) {
	root := t.TempDir()
	prdPath := filepath.Join(root, "prd.md")
	initial := "Team intro\n\n<!-- pacto:prd:start -->\nold\n<!-- pacto:prd:end -->\n\nnotes"
	if err := os.WriteFile(prdPath, []byte(initial), 0o664); err != nil {
		t.Fatal(err)
	}

	if code := RunInit([]string{"--root", root, "--no-install"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}
	b, err := os.ReadFile(prdPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "Team intro") || !strings.Contains(text, "notes") {
		t.Fatalf("expected unmanaged prd content preserved, got %q", text)
	}
	if !strings.Contains(text, "<!-- pacto:prd:start -->") || !strings.Contains(text, "<!-- pacto:prd:end -->") {
		t.Fatalf("expected managed prd markers, got %q", text)
	}
}

func TestRunInitRejectsRemovedEditorFlag(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--editor", "cursor"}); code != 2 {
		t.Fatalf("RunInit returned %d, expected 2", code)
	}
}

func TestRunInitRejectsRemovedLanguageFlag(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root, "--language", "go"}); code != 2 {
		t.Fatalf("RunInit returned %d, expected 2", code)
	}
}

func assertPathExists(t *testing.T, p string) {
	t.Helper()
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("expected path to exist %q: %v", p, err)
	}
}
