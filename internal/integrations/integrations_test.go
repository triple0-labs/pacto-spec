package integrations

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func TestGenerateForToolWritesContractAndExecPlanned(t *testing.T) {
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
	if !strings.Contains(execContent, "Planned (Not Implemented)") {
		t.Fatalf("exec command should be marked planned, got: %q", execContent)
	}
	if !strings.Contains(execContent, "Use `pacto status` for verification") {
		t.Fatalf("exec command should include fallback guidance, got: %q", execContent)
	}
}
