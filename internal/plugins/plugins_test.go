package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDiscoverAndLoadActive(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, pluginSpec{
		dir:      "acme",
		id:       "acme",
		priority: 10,
		script:   "#!/bin/sh\nexit 0\n",
		markdown: "Run status before exec.",
	})
	if err := WriteActiveConfig(root, []string{"acme"}); err != nil {
		t.Fatal(err)
	}
	active, errs := LoadActive(root)
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(active) != 1 {
		t.Fatalf("expected 1 active plugin, got %d", len(active))
	}
	contrib := CollectAgentContributions(active, "codex", "exec")
	if len(contrib) != 1 {
		t.Fatalf("expected 1 contribution, got %d", len(contrib))
	}
	if !strings.Contains(contrib[0].Markdown, "status") {
		t.Fatalf("unexpected markdown: %q", contrib[0].Markdown)
	}
}

func TestDiscoverReportsInvalidManifest(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, pluginSpec{
		dir:        "bad",
		id:         "bad",
		apiVersion: "pacto/v9",
		script:     "#!/bin/sh\nexit 0\n",
		markdown:   "x",
	})
	d := Discover(root)
	if len(d.Plugins) != 0 {
		t.Fatalf("expected no valid plugins, got %d", len(d.Plugins))
	}
	if len(d.Errors) == 0 {
		t.Fatalf("expected validation errors")
	}
}

func TestDiscoverReportsDuplicateIDs(t *testing.T) {
	root := t.TempDir()
	writePlugin(t, root, pluginSpec{dir: "one", id: "dup", script: "#!/bin/sh\nexit 0\n", markdown: "a"})
	writePlugin(t, root, pluginSpec{dir: "two", id: "dup", script: "#!/bin/sh\nexit 0\n", markdown: "b"})
	d := Discover(root)
	if len(d.Errors) == 0 {
		t.Fatalf("expected duplicate ID error")
	}
}

func TestLoadActiveMissingEnabledPluginReturnsError(t *testing.T) {
	root := t.TempDir()
	if err := WriteActiveConfig(root, []string{"missing"}); err != nil {
		t.Fatal(err)
	}
	active, errs := LoadActive(root)
	if len(active) != 0 {
		t.Fatalf("expected no active plugins, got %d", len(active))
	}
	if len(errs) == 0 {
		t.Fatalf("expected missing enabled plugin error")
	}
}

func TestEvaluateGuardrailsBlockedAndAllowed(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, ".pacto", "plugins", "acme")
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "fail.sh"), []byte("#!/bin/sh\necho nope 1>&2\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	m := Manifest{
		APIVersion: ManifestAPIVersion,
		Kind:       ManifestKind,
		Metadata:   Metadata{ID: "acme", Priority: 1},
		Spec: Spec{CLIGuardrails: []CLIGuardrail{{
			ID:       "no-fail",
			Commands: []string{"exec"},
			Run:      RunSpec{Script: "scripts/fail.sh", TimeoutMS: 1000},
			OnFail:   OnFail{Message: "no failing"},
		}}},
	}
	active := []Plugin{{Dir: pluginDir, Manifest: m}}
	violations := EvaluateGuardrails(active, HookRequest{Command: "exec", ProjectRoot: root, Allow: map[string]bool{}})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0].Allowed {
		t.Fatalf("expected blocked violation")
	}
	violations = EvaluateGuardrails(active, HookRequest{Command: "exec", ProjectRoot: root, Allow: map[string]bool{"acme/no-fail": true}})
	if len(violations) != 1 || !violations[0].Allowed {
		t.Fatalf("expected allowed violation, got %#v", violations)
	}
}

func TestEvaluateGuardrailsTimeout(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, ".pacto", "plugins", "acme")
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "slow.sh"), []byte("#!/bin/sh\nsleep 2\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	active := []Plugin{{Dir: pluginDir, Manifest: Manifest{
		APIVersion: ManifestAPIVersion,
		Kind:       ManifestKind,
		Metadata:   Metadata{ID: "acme"},
		Spec: Spec{CLIGuardrails: []CLIGuardrail{{
			ID:       "slow",
			Commands: []string{"new"},
			Run:      RunSpec{Script: "scripts/slow.sh", TimeoutMS: 500},
		}}},
	}}}
	violations := EvaluateGuardrails(active, HookRequest{Command: "new", ProjectRoot: root, Allow: map[string]bool{}})
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if !violations[0].TimedOut {
		t.Fatalf("expected timeout violation, got %#v", violations[0])
	}
}

func TestEvaluateGuardrailsSkipsNonMatchingCommand(t *testing.T) {
	root := t.TempDir()
	pluginDir := filepath.Join(root, ".pacto", "plugins", "acme")
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "fail.sh"), []byte("#!/bin/sh\nexit 2\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	active := []Plugin{{Dir: pluginDir, Manifest: Manifest{
		APIVersion: ManifestAPIVersion,
		Kind:       ManifestKind,
		Metadata:   Metadata{ID: "acme"},
		Spec:       Spec{CLIGuardrails: []CLIGuardrail{{ID: "x", Commands: []string{"exec"}, Run: RunSpec{Script: "scripts/fail.sh"}}}},
	}}}
	violations := EvaluateGuardrails(active, HookRequest{Command: "status", ProjectRoot: root, Allow: map[string]bool{}})
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %#v", violations)
	}
}

func TestActivationEnableDisableRoundTrip(t *testing.T) {
	root := t.TempDir()
	if err := Enable(root, "acme"); err != nil {
		t.Fatal(err)
	}
	cfg, err := ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Enabled) != 1 || cfg.Enabled[0] != "acme" {
		t.Fatalf("unexpected enabled list: %#v", cfg.Enabled)
	}
	if err := Disable(root, "acme"); err != nil {
		t.Fatal(err)
	}
	cfg, err = ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Enabled) != 0 {
		t.Fatalf("expected empty enabled list, got %#v", cfg.Enabled)
	}
}

func TestReadActiveConfigSupportsMultilineEnabledList(t *testing.T) {
	root := t.TempDir()
	cfgPath := filepath.Join(root, ".pacto", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "version: 1\nplugins:\n  enabled:\n    - Acme\n    - qa-guard\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := ReadActiveConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Enabled) != 2 || cfg.Enabled[0] != "acme" || cfg.Enabled[1] != "qa-guard" {
		t.Fatalf("unexpected enabled plugins: %#v", cfg.Enabled)
	}
}

type pluginSpec struct {
	dir        string
	id         string
	apiVersion string
	priority   int
	script     string
	markdown   string
}

func writePlugin(t *testing.T, root string, spec pluginSpec) {
	t.Helper()
	if spec.apiVersion == "" {
		spec.apiVersion = ManifestAPIVersion
	}
	if spec.id == "" {
		spec.id = spec.dir
	}
	pluginDir := filepath.Join(root, ".pacto", "plugins", spec.dir)
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "guardrails"), 0o755); err != nil {
		t.Fatal(err)
	}
	if spec.script == "" {
		spec.script = "#!/bin/sh\nexit 0\n"
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "check.sh"), []byte(spec.script), 0o755); err != nil {
		t.Fatal(err)
	}
	if spec.markdown == "" {
		spec.markdown = "Default guardrail"
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "guardrails", "status.md"), []byte(spec.markdown), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := "apiVersion: " + spec.apiVersion + "\n" +
		"kind: Plugin\n" +
		"metadata:\n" +
		"  id: " + spec.id + "\n" +
		"  version: 0.1.0\n" +
		"  priority: " + fmt.Sprintf("%d", spec.priority) + "\n" +
		"spec:\n" +
		"  cliGuardrails:\n" +
		"    - id: clean\n" +
		"      commands: [exec,new,init,explore]\n" +
		"      run:\n" +
		"        script: scripts/check.sh\n" +
		"        timeoutMs: 1000\n" +
		"      onFail:\n" +
		"        message: clean required\n" +
		"  agentGuardrails:\n" +
		"    - id: status-first\n" +
		"      tools: [codex]\n" +
		"      workflows: [exec]\n" +
		"      markdownFile: guardrails/status.md\n"
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
}
