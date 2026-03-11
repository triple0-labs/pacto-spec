package builtin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListAvailableIncludesGitGuardrails(t *testing.T) {
	infos := ListAvailable()
	found := false
	for _, info := range infos {
		if info.ID == "git-guardrails" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected git-guardrails in available list: %#v", infos)
	}
}

func TestInstallGitGuardrailsWritesFilesAndConfig(t *testing.T) {
	root := t.TempDir()
	res, err := Install(root, "git-guardrails", InstallOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if res.PluginID != "git-guardrails" {
		t.Fatalf("unexpected plugin id %q", res.PluginID)
	}
	pluginDir := filepath.Join(root, ".pacto", "plugins", "git-guardrails")
	for _, p := range []string{
		filepath.Join(pluginDir, "plugin.yaml"),
		filepath.Join(pluginDir, "scripts", "preflight.sh"),
		filepath.Join(pluginDir, "guardrails", "git-guardrails.md"),
		filepath.Join(pluginDir, "README.md"),
		filepath.Join(pluginDir, "config.env.example"),
		filepath.Join(pluginDir, "config.env"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected path %s: %v", p, err)
		}
	}
	info, err := os.Stat(filepath.Join(pluginDir, "scripts", "preflight.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected executable script mode, got %o", info.Mode())
	}
}

func TestInstallUnknownPluginReturnsError(t *testing.T) {
	_, err := Install(t.TempDir(), "nope", InstallOptions{})
	if err == nil {
		t.Fatalf("expected error for unknown plugin")
	}
}
