package integrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/plugins"
)

func TestParseToolsArg(t *testing.T) {
	got, err := ParseToolsArg("codex,cursor,codex")
	if err != nil {
		t.Fatalf("ParseToolsArg error: %v", err)
	}
	if len(got) != 2 || got[0] != "codex" || got[1] != "cursor" {
		t.Fatalf("unexpected parse result: %#v", got)
	}

	if _, err := ParseToolsArg("badtool"); err == nil {
		t.Fatal("expected error for invalid tool")
	}
}

func TestDetectToolsIncludesOpenCode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := DetectTools(root)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(got, "opencode") || !contains(got, "cursor") {
		t.Fatalf("expected detected tools to include opencode and cursor, got %v", got)
	}
}

func TestWriteManagedCreatesAndUpdates(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "x.md")

	wr, err := WriteManaged(p, "hello", false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeCreated {
		t.Fatalf("expected created, got %s", wr.Outcome)
	}

	wr, err = WriteManaged(p, "hello", false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeSkipped {
		t.Fatalf("expected skipped unchanged, got %s", wr.Outcome)
	}

	if err := os.WriteFile(p, []byte("custom no markers\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wr, err = WriteManaged(p, "hello", false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeSkipped || wr.Reason != "unmanaged_exists" {
		t.Fatalf("expected unmanaged skip, got %#v", wr)
	}

	wr, err = WriteManaged(p, "hello", true)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeUpdated {
		t.Fatalf("expected force updated, got %#v", wr)
	}
}

func TestGetAdapterPaths(t *testing.T) {
	root := t.TempDir()
	a, ok := GetAdapter("opencode")
	if !ok {
		t.Fatal("expected opencode adapter")
	}
	skill, err := a.SkillFilePath(root, "status")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md"); skill != want {
		t.Fatalf("skill path mismatch: got %q want %q", skill, want)
	}
	cmd, err := a.CommandFilePath(root, "pacto-status")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, ".opencode", "commands", "pacto-status.md"); cmd != want {
		t.Fatalf("command path mismatch: got %q want %q", cmd, want)
	}
}

func TestRenderTemplatesIncludeContractSections(t *testing.T) {
	for _, wf := range Workflows() {
		skill := RenderSkill("codex", wf)
		command := RenderCommand("codex", wf)

		for _, section := range []string{
			"## Input Contract",
			"## Output Contract",
			"## Validation Checklist",
			"## Failure Modes and Handling",
			"## Implementation Status",
		} {
			if !strings.Contains(skill, section) {
				t.Fatalf("skill template for %s missing section %q", wf.WorkflowID, section)
			}
			if !strings.Contains(command, section) {
				t.Fatalf("command template for %s missing section %q", wf.WorkflowID, section)
			}
		}
	}
}

func TestGenerateForToolWritesContractAndExecCommand(t *testing.T) {
	root := t.TempDir()
	results := GenerateForTool(root, "opencode", false)
	if len(results) == 0 {
		t.Fatal("expected generation results")
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}

	statusSkill := filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md")
	b, err := os.ReadFile(statusSkill)
	if err != nil {
		t.Fatalf("read status skill: %v", err)
	}
	statusContent := string(b)
	for _, section := range []string{
		"## Input Contract",
		"## Output Contract",
		"## Validation Checklist",
	} {
		if !strings.Contains(statusContent, section) {
			t.Fatalf("status skill missing %q", section)
		}
	}

	execCommand := filepath.Join(root, ".opencode", "commands", "pacto-exec.md")
	b, err = os.ReadFile(execCommand)
	if err != nil {
		t.Fatalf("read exec command: %v", err)
	}
	execContent := string(b)
	if !strings.Contains(execContent, "## Implementation Status") {
		t.Fatalf("exec command should include implementation status section, got: %q", execContent)
	}
	if !strings.Contains(execContent, "Status: **Implemented**") {
		t.Fatalf("exec command should be marked implemented, got: %q", execContent)
	}
}

func TestGenerateForToolIncludesPluginGuardrailsWhenActive(t *testing.T) {
	root := t.TempDir()
	writeIntegrationPlugin(t, root, "acme", true)

	results := GenerateForTool(root, "opencode", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}
	statusSkill := filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md")
	b, err := os.ReadFile(statusSkill)
	if err != nil {
		t.Fatalf("read status skill: %v", err)
	}
	content := string(b)
	if !strings.Contains(content, "## Plugin Guardrails") {
		t.Fatalf("expected plugin guardrails section, got: %q", content)
	}
	if !strings.Contains(content, "acme/status-first") {
		t.Fatalf("expected plugin guardrail id, got: %q", content)
	}
	if !strings.Contains(content, "Always run pacto status first") {
		t.Fatalf("expected plugin markdown content, got: %q", content)
	}
}

func TestGenerateForToolSkipsPluginGuardrailsWhenDisabled(t *testing.T) {
	root := t.TempDir()
	writeIntegrationPlugin(t, root, "acme", false)

	results := GenerateForTool(root, "opencode", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}
	statusSkill := filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md")
	b, err := os.ReadFile(statusSkill)
	if err != nil {
		t.Fatalf("read status skill: %v", err)
	}
	content := string(b)
	if strings.Contains(content, "## Plugin Guardrails") {
		t.Fatalf("did not expect plugin guardrails section when disabled, got: %q", content)
	}
}

func writeIntegrationPlugin(t *testing.T, root, pluginID string, enable bool) {
	t.Helper()
	pluginDir := filepath.Join(root, ".pacto", "plugins", pluginID)
	if err := os.MkdirAll(filepath.Join(pluginDir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(pluginDir, "guardrails"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "scripts", "check.sh"), []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "guardrails", "status.md"), []byte("Always run pacto status first."), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := `apiVersion: pacto/v1alpha1
kind: Plugin
metadata:
  id: ` + pluginID + `
  version: 0.1.0
  priority: 1
spec:
  cliGuardrails:
    - id: clean
      commands: [new]
      run:
        script: scripts/check.sh
        timeoutMs: 1000
  agentGuardrails:
    - id: status-first
      tools: [opencode]
      workflows: [status]
      markdownFile: guardrails/status.md
`
	if err := os.WriteFile(filepath.Join(pluginDir, "plugin.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	if enable {
		if err := plugins.WriteActiveConfig(root, []string{pluginID}); err != nil {
			t.Fatal(err)
		}
	}
}
