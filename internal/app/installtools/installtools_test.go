package installtools

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestInstall_ExplicitToolCreatesArtifacts(t *testing.T) {
	root := t.TempDir()
	res, err := Install(Input{Root: root, Tools: "cursor"})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if res.NoOp {
		t.Fatal("did not expect NoOp")
	}
	if len(res.Tools) != 1 || res.Tools[0] != "cursor" {
		t.Fatalf("unexpected Tools: %v", res.Tools)
	}
	if res.Created == 0 {
		t.Fatalf("expected Created > 0, got %+v", res)
	}
	if res.Failed != 0 {
		t.Fatalf("unexpected failures: %d", res.Failed)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "skills", "pacto-status", "SKILL.md")); err != nil {
		t.Fatalf("expected cursor skill written: %v", err)
	}
	if len(res.Items) == 0 {
		t.Fatal("expected per-file items")
	}
}

func TestInstall_AutoDetectsOpenCode(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".opencode"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Install(Input{Root: root})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	found := false
	for _, tool := range res.Tools {
		if tool == "opencode" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected opencode in detected tools, got %v", res.Tools)
	}
	if _, err := os.Stat(filepath.Join(root, ".opencode", "skills", "pacto-status", "SKILL.md")); err != nil {
		t.Fatalf("expected opencode skill written: %v", err)
	}
}

func TestInstall_NoTools_NoOp(t *testing.T) {
	root := t.TempDir()
	res, err := Install(Input{Root: root, Tools: "none"})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if !res.NoOp {
		t.Fatalf("expected NoOp for tools=none, got %+v", res)
	}
	if len(res.Items) != 0 {
		t.Fatal("expected no items for NoOp run")
	}
}

func TestInstall_RejectsUnknownTool(t *testing.T) {
	_, err := Install(Input{Root: t.TempDir(), Tools: "not-a-tool"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestInstall_RejectsMissingRoot(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	_, err := Install(Input{Root: missing, Tools: "cursor"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestInstall_AllExpandsToSupportedTools(t *testing.T) {
	root := t.TempDir()
	res, err := Install(Input{Root: root, Tools: "all"})
	if err != nil {
		t.Fatalf("Install: %v", err)
	}
	if len(res.Tools) < 2 {
		t.Fatalf("expected multiple tools for --tools all, got %v", res.Tools)
	}
	if res.Failed != 0 {
		t.Fatalf("unexpected failures: %d", res.Failed)
	}
}
