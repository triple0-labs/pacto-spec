package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/plugins"
)

func TestRunBlocksGuardrailOnMutatingCommand(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-new", []string{"new"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"new", "to-implement", "guarded", "--allow-minimal-root"})
		if code != 3 {
			t.Fatalf("Run returned %d, want 3", code)
		}
	})
	if !strings.Contains(stderr, "guardrail blocked") {
		t.Fatalf("expected guardrail blocked message, got %q", stderr)
	}
}

func TestRunAllowsSpecificGuardrailBypass(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-explore", []string{"explore"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"--allow-guardrail", "acme/block-explore", "explore", "idea-1"})
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	if _, err := os.Stat(filepath.Join(root, ".pacto", "ideas", "idea-1", "README.md")); err != nil {
		t.Fatalf("expected idea file created: %v", err)
	}
}

func TestRunSkipsGuardrailsForExploreList(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-explore", []string{"explore"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"explore", "--list"})
	if code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
}

func TestRunSkipsGuardrailsForInitDryRun(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-init", []string{"init"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureOutput(t, func() {
		code := Run([]string{"init", "--dry-run", "--no-interactive"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, "Init Dry Run") {
		t.Fatalf("expected dry run output, got %q", stdout)
	}
}

func TestRunBlocksGuardrailOnStatus(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-status", []string{"status"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"status"})
		if code != 3 {
			t.Fatalf("Run returned %d, want 3", code)
		}
	})
	if !strings.Contains(stderr, "guardrail blocked") {
		t.Fatalf("expected guardrail blocked message, got %q", stderr)
	}
}

func TestRunSkipsGuardrailsForStatusHelp(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-status", []string{"status"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	stdout, stderr := captureOutput(t, func() {
		code := Run([]string{"status", "--help"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if strings.Contains(stderr, "guardrail blocked") {
		t.Fatalf("did not expect guardrail blocked message on help, got %q", stderr)
	}
	if !strings.Contains(stdout, "Command: status") {
		t.Fatalf("expected status help output, got %q", stdout)
	}
}

func TestRunAllowsGuardrailBypassOnStatus(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-status", []string{"status"}, "#!/bin/sh\nexit 2\n")
	if err := plugins.WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	if code := RunInit([]string{"--root", root, "--no-interactive", "--no-install"}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"--allow-guardrail", "acme/block-status", "status"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if strings.Contains(stderr, "guardrail blocked") {
		t.Fatalf("did not expect guardrail blocked message with bypass, got %q", stderr)
	}
}

func TestRunErrorsWhenEnabledPluginMissing(t *testing.T) {
	root := t.TempDir()
	if err := plugins.WriteActiveConfig(root, []string{"missing-plugin"}); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"new", "to-implement", "x", "--allow-minimal-root"})
		if code != 3 {
			t.Fatalf("Run returned %d, want 3", code)
		}
	})
	if !strings.Contains(stderr, "enabled plugin not found") {
		t.Fatalf("expected missing plugin error, got %q", stderr)
	}
}

func TestRunPluginCommands(t *testing.T) {
	root := t.TempDir()
	writeTestPlugin(t, root, "acme", "block-new", []string{"new"}, "#!/bin/sh\nexit 0\n")
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	stdout, _ := captureOutput(t, func() {
		code := Run([]string{"plugin", "list", "--format", "json"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, "\"id\": \"acme\"") {
		t.Fatalf("expected plugin in list output, got %q", stdout)
	}

	if code := Run([]string{"plugin", "enable", "acme"}); code != 0 {
		t.Fatalf("enable returned %d", code)
	}
	cfg, err := plugins.ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Enabled) != 1 || cfg.Enabled[0] != "acme" {
		t.Fatalf("unexpected enabled list: %#v", cfg.Enabled)
	}

	if code := Run([]string{"plugin", "disable", "acme"}); code != 0 {
		t.Fatalf("disable returned %d", code)
	}
	cfg, err = plugins.ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Enabled) != 0 {
		t.Fatalf("expected empty enabled list, got %#v", cfg.Enabled)
	}
}

func TestRunPluginListAvailableShowsGitSync(t *testing.T) {
	root := t.TempDir()
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureOutput(t, func() {
		code := Run([]string{"plugin", "list-available", "--format", "json"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, "\"id\": \"git-sync\"") {
		t.Fatalf("expected git-sync in available plugins, got %q", stdout)
	}
}

func TestRunPluginInstallAutoEnablesByDefault(t *testing.T) {
	root := t.TempDir()
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	stdout, _ := captureOutput(t, func() {
		code := Run([]string{"plugin", "install", "git-sync"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if !strings.Contains(stdout, "Installed plugin git-sync") || !strings.Contains(stdout, "Enabled plugin git-sync") {
		t.Fatalf("expected install+enable output, got %q", stdout)
	}
	pluginDir := filepath.Join(root, ".pacto", "plugins", "git-sync")
	for _, p := range []string{
		filepath.Join(pluginDir, "plugin.yaml"),
		filepath.Join(pluginDir, "scripts", "sync-status.sh"),
		filepath.Join(pluginDir, "config.env"),
	} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected installed path %s: %v", p, err)
		}
	}
	cfg, err := plugins.ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, id := range cfg.Enabled {
		if id == "git-sync" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected git-sync enabled, got %#v", cfg.Enabled)
	}
	info, err := os.Stat(filepath.Join(pluginDir, "scripts", "sync-status.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&0o111 == 0 {
		t.Fatalf("expected sync-status.sh to be executable, mode=%o", info.Mode())
	}
}

func TestRunPluginInstallNoEnable(t *testing.T) {
	root := t.TempDir()
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if code := Run([]string{"plugin", "install", "git-sync", "--no-enable"}); code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	cfg, err := plugins.ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range cfg.Enabled {
		if id == "git-sync" {
			t.Fatalf("expected git-sync not enabled, got %#v", cfg.Enabled)
		}
	}
}

func TestRunPluginInstallUnknownBuiltin(t *testing.T) {
	root := t.TempDir()
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"plugin", "install", "not-real"})
		if code != 2 {
			t.Fatalf("Run returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "unknown built-in plugin") {
		t.Fatalf("expected unknown builtin error, got %q", stderr)
	}
}

func TestRunPluginValidateFailsForInvalidManifest(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, ".pacto", "plugins", "bad")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte("apiVersion: bad\nkind: Plugin\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	oldWD, _ := os.Getwd()
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"plugin", "validate"})
	if code != 3 {
		t.Fatalf("Run returned %d, want 3", code)
	}
}

func writeTestPlugin(t *testing.T, root, pluginID, guardrailID string, commands []string, script string) {
	t.Helper()
	pluginDir := filepath.Join(root, ".pacto", "plugins", pluginID)
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "guardrails"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "check.sh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "guardrails", "status.md"), []byte("Always run status first."), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := "apiVersion: pacto/v1alpha1\n" +
		"kind: Plugin\n" +
		"metadata:\n" +
		"  id: " + pluginID + "\n" +
		"  version: 0.1.0\n" +
		"  priority: 1\n" +
		"spec:\n" +
		"  cliGuardrails:\n" +
		"    - id: " + guardrailID + "\n" +
		"      commands: [" + strings.Join(commands, ",") + "]\n" +
		"      run:\n" +
		"        script: scripts/check.sh\n" +
		"        timeoutMs: 1000\n" +
		"      onFail:\n" +
		"        message: blocked\n" +
		"  agentGuardrails:\n" +
		"    - id: status-first\n" +
		"      tools: [codex, cursor, claude, opencode]\n" +
		"      workflows: [new, exec, move, status]\n" +
		"      markdownFile: guardrails/status.md\n"
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}
