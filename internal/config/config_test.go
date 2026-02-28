package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadResolvesRootRelativeToConfigDir(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".pacto-engine.yaml")
	writeFile(t, path, "root: plans\n")

	cfg, warnings, err := Load("", root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	want := filepath.Join(root, "plans")
	if cfg.Root != want {
		t.Fatalf("cfg.Root=%q, want %q", cfg.Root, want)
	}
}

func TestLoadResolvesSplitRootsRelativeToConfigDir(t *testing.T) {
	root := t.TempDir()
	cfgDir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cfgDir, ".pacto-engine.yaml")
	writeFile(t, path, "plans_root: ../plans\nrepo_root: ../repo\n")

	cfg, warnings, err := Load(path, root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(warnings) == 0 || !containsWarning(warnings, "plans_root") || !containsWarning(warnings, "deprecated") {
		t.Fatalf("expected plans_root deprecation warning, got %v", warnings)
	}
	if cfg.PlansRoot != filepath.Clean(filepath.Join(cfgDir, "../plans")) {
		t.Fatalf("cfg.PlansRoot=%q", cfg.PlansRoot)
	}
	if cfg.RepoRoot != filepath.Clean(filepath.Join(cfgDir, "../repo")) {
		t.Fatalf("cfg.RepoRoot=%q", cfg.RepoRoot)
	}
}

func TestLoadPreservesQuotedHashInValue(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".pacto-engine.yaml")
	writeFile(t, path, `plans_root: "folder#name"`+"\n")

	cfg, warnings, err := Load("", root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(warnings) == 0 || !containsWarning(warnings, "plans_root") || !containsWarning(warnings, "deprecated") {
		t.Fatalf("expected plans_root deprecation warning, got %v", warnings)
	}
	want := filepath.Join(root, "folder#name")
	if cfg.PlansRoot != want {
		t.Fatalf("cfg.PlansRoot=%q, want %q", cfg.PlansRoot, want)
	}
}

func TestLoadWarnsOnUnknownKey(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".pacto-engine.yaml")
	writeFile(t, path, "foo: bar\n")

	_, warnings, err := Load("", root)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown key")
	}
	if !containsWarning(warnings, "unknown config key") {
		t.Fatalf("unexpected warning: %v", warnings)
	}
}

func containsWarning(warnings []string, sub string) bool {
	for _, w := range warnings {
		if strings.Contains(strings.ToLower(w), strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
