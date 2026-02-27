package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInstallExplicitToolCreatesArtifacts(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		if code := RunInstall([]string{"--tools", "cursor"}); code != 0 {
			t.Fatalf("RunInstall returned %d", code)
		}
	})
	if !strings.Contains(stdout, "Created:") {
		t.Fatalf("expected summary in stdout, got %q", stdout)
	}

	assertExists(t, filepath.Join(root, ".cursor", "skills", "pacto-status", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".cursor", "commands", "pacto-status.md"))
	assertExists(t, filepath.Join(root, ".cursor", "skills", "pacto-new", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".cursor", "commands", "pacto-new.md"))
}

func TestRunInstallAutoDetectsOpenCode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	if code := RunInstall([]string{}); code != 0 {
		t.Fatalf("RunInstall returned %d", code)
	}
	assertExists(t, filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md"))
	assertExists(t, filepath.Join(root, ".opencode", "commands", "pacto-status.md"))
}

func TestRunUpdateSkipsUnmanagedWithoutForce(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cursor", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, ".cursor", "commands", "pacto-status.md")
	if err := os.WriteFile(path, []byte("custom\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	_, stderr := captureOutput(t, func() {
		if code := RunUpdate([]string{"--tools", "cursor"}); code != 0 {
			t.Fatalf("RunUpdate returned %d", code)
		}
	})
	if !strings.Contains(stderr, "skipped unmanaged file") {
		t.Fatalf("expected unmanaged warning, got %q", stderr)
	}
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected path to exist: %s (%v)", path, err)
	}
}
