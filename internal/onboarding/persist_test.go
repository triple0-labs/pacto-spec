package onboarding

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteConfigPreservesEnabledPlugins(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, ".pacto", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := "version: 1\nplugins:\n  enabled:\n    - acme\n    - qa-guard\n"
	if err := os.WriteFile(cfgPath, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := WriteConfig(root, Profile{
		Languages: []string{"go"},
		Tools:     []string{"codex"},
		Intents:   Intents{Problem: "reduce toil"},
		Sources:   Sources{Languages: "user", Tools: "user"},
	})
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "- acme") || !strings.Contains(text, "- qa-guard") {
		t.Fatalf("expected enabled plugins preserved, got %q", text)
	}
	if !strings.Contains(text, "problem: reduce toil") {
		t.Fatalf("expected merged profile values, got %q", text)
	}
}
