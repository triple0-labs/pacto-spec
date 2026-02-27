package integrations

import (
	"os"
	"path/filepath"
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
