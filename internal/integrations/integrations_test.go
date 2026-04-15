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
	if err := os.MkdirAll(filepath.Join(root, ".agents", "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	got, err := DetectTools(root)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(got, "codex") || !contains(got, "opencode") || !contains(got, "cursor") {
		t.Fatalf("expected detected tools to include codex, opencode and cursor, got %v", got)
	}
}

func TestWriteManagedCreatesAndUpdates(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "x.md")
	meta := ManagedMetadata{
		Artifact:    "cursor/command/pacto-status.md",
		Workflow:    "status",
		Contract:    ContractVersion,
		TemplateSHA: TemplateChecksum("hello"),
		GeneratedBy: "pacto",
		GeneratedAt: "2026-03-06T00:00:00Z",
	}

	wr, err := WriteManaged(p, "hello", meta, false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeCreated {
		t.Fatalf("expected created, got %s", wr.Outcome)
	}

	wr, err = WriteManaged(p, "hello", meta, false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeSkipped {
		t.Fatalf("expected skipped unchanged, got %s", wr.Outcome)
	}

	if err := os.WriteFile(p, []byte("custom no markers\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	wr, err = WriteManaged(p, "hello", meta, false)
	if err != nil {
		t.Fatal(err)
	}
	if wr.Outcome != OutcomeSkipped || wr.Reason != "unmanaged_exists" {
		t.Fatalf("expected unmanaged skip, got %#v", wr)
	}

	wr, err = WriteManaged(p, "hello", meta, true)
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
}

func TestGetAdapterCursorSkillsUseAgentsPath(t *testing.T) {
	root := t.TempDir()
	a, ok := GetAdapter("cursor")
	if !ok {
		t.Fatal("expected cursor adapter")
	}
	skill, err := a.SkillFilePath(root, "status")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(root, ".agents", "skills", "pacto-status", "SKILL.md"); skill != want {
		t.Fatalf("cursor skill path mismatch: got %q want %q", skill, want)
	}
}

func TestRenderTemplatesIncludeContractSections(t *testing.T) {
	for _, wf := range Workflows() {
		skill := RenderSkill(SkillToolTargetForIntegration("codex"), wf)

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
		}
	}
}

func TestStatusWorkflowPrefersJSONAndOptionalVerify(t *testing.T) {
	var status WorkflowSpec
	for _, wf := range Workflows() {
		if wf.WorkflowID == "status" {
			status = wf
			break
		}
	}
	if status.WorkflowID == "" {
		t.Fatal("expected status workflow")
	}
	if status.Command != "pacto status --format json" {
		t.Fatalf("unexpected status command: %q", status.Command)
	}
	skill := RenderSkill(SkillToolTargetForIntegration("codex"), status)
	if !strings.Contains(skill, "optional path verification") {
		t.Fatalf("expected metadata-first status summary, got %q", skill)
	}
	if !strings.Contains(skill, "`--verify`") {
		t.Fatalf("expected verify guidance in rendered skill, got %q", skill)
	}
}

func TestGenerateForToolWritesSkillsOnly(t *testing.T) {
	root := t.TempDir()
	results := GenerateForTool(root, "opencode", false)
	if len(results) == 0 {
		t.Fatal("expected generation results")
	}
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
		if r.Kind != "skill" {
			t.Fatalf("expected only skill results, got kind %q", r.Kind)
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
		"## Implementation Status",
	} {
		if !strings.Contains(statusContent, section) {
			t.Fatalf("status skill missing %q", section)
		}
	}
	if !strings.Contains(statusContent, "Status: **Implemented**") {
		t.Fatalf("status skill should be marked implemented, got: %q", statusContent)
	}

	execWorkflow := filepath.Join(root, ".opencode", "skills", "pacto-exec", "SKILL.md")
	b, err = os.ReadFile(execWorkflow)
	if err != nil {
		t.Fatalf("read exec skill: %v", err)
	}
	execContent := string(b)
	if !strings.Contains(execContent, "## Implementation Status") {
		t.Fatalf("exec skill should include implementation status section, got: %q", execContent)
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

func TestGenerateForToolClaudeSkillsHaveFrontmatter(t *testing.T) {
	root := t.TempDir()

	results := GenerateForTool(root, "claude", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}

	skillPath := filepath.Join(root, ".claude", "skills", "pacto-doctor", "SKILL.md")
	b, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read claude skill: %v", err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "---\nname: pacto-doctor\n") {
		t.Fatalf("expected claude skill frontmatter prefix, got: %q", content)
	}
	if !strings.Contains(content, "<!-- pacto:managed:start -->") {
		t.Fatalf("expected managed block in claude skill, got: %q", content)
	}
}

func TestGenerateForToolCodexSkillsHaveFrontmatter(t *testing.T) {
	root := t.TempDir()
	t.Setenv("CODEX_HOME", filepath.Join(root, "_codex_home"))

	results := GenerateForTool(root, "codex", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}

	skillPath := filepath.Join(root, ".agents", "skills", "pacto-doctor", "SKILL.md")
	b, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read codex skill: %v", err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "---\nname: pacto-doctor\n") {
		t.Fatalf("expected codex skill frontmatter prefix, got: %q", content)
	}
	if !strings.Contains(content, "<!-- pacto:managed:start -->") {
		t.Fatalf("expected managed block in codex skill, got: %q", content)
	}
}

func TestGenerateForToolOpenCodeSkillsHaveFrontmatter(t *testing.T) {
	root := t.TempDir()

	results := GenerateForTool(root, "opencode", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("unexpected generation error for %s/%s: %v", r.Kind, r.WorkflowID, r.Err)
		}
	}

	skillPath := filepath.Join(root, ".opencode", "skills", "pacto-doctor", "SKILL.md")
	b, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read opencode skill: %v", err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "---\nname: pacto-doctor\n") {
		t.Fatalf("expected opencode skill frontmatter prefix, got: %q", content)
	}
	if !strings.Contains(content, "compatibility: opencode") {
		t.Fatalf("expected opencode skill compatibility field, got: %q", content)
	}
	if !strings.Contains(content, "<!-- pacto:managed:start -->") {
		t.Fatalf("expected managed block in opencode skill, got: %q", content)
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
