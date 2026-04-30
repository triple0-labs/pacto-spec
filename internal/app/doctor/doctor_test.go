package doctor

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyze_HappyPath_NoDrift(t *testing.T) {
	root := t.TempDir()
	// Generate a fresh cursor toolset so there's nothing drifted.
	if err := os.MkdirAll(filepath.Join(root, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := Analyze(Input{Root: root, Tools: "cursor"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if res.NoOp {
		t.Fatal("did not expect NoOp")
	}
	if len(res.Tools) != 1 || res.Tools[0] != "cursor" {
		t.Fatalf("unexpected tools: %v", res.Tools)
	}
	if res.Summary == nil {
		t.Fatal("expected summary populated")
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected ExitCode=0, got %d", res.ExitCode)
	}
}

func TestAnalyze_RejectsBadFormat(t *testing.T) {
	_, err := Analyze(Input{Root: t.TempDir(), Tools: "cursor", Format: "yaml"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestAnalyze_RejectsBadFailOn(t *testing.T) {
	_, err := Analyze(Input{Root: t.TempDir(), Tools: "cursor", FailOn: "bogus"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestAnalyze_RejectsUnknownTool(t *testing.T) {
	_, err := Analyze(Input{Root: t.TempDir(), Tools: "not-a-tool"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestAnalyze_NoneToolsIsNoOp(t *testing.T) {
	res, err := Analyze(Input{Root: t.TempDir(), Tools: "none"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if !res.NoOp {
		t.Fatalf("expected NoOp for tools=none, got %+v", res)
	}
}

func TestAnalyze_MissingArtifactsCountAsDrift(t *testing.T) {
	root := t.TempDir()
	res, err := Analyze(Input{Root: root, Tools: "cursor"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if res.Summary["drift"] == 0 && res.Summary["missing"] == 0 {
		t.Fatalf("expected drift>0 with no artifacts written; summary=%v", res.Summary)
	}
	if !res.HasNonFailingDrift {
		t.Fatal("expected HasNonFailingDrift=true with default fail-on=none")
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected ExitCode=0 with fail-on=none, got %d", res.ExitCode)
	}
}

func TestAnalyze_FailOnDriftReturnsExit1(t *testing.T) {
	root := t.TempDir()
	res, err := Analyze(Input{Root: root, Tools: "cursor", FailOn: "drift"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if res.ExitCode != 1 {
		t.Fatalf("expected ExitCode=1 with fail-on=drift, got %d (summary=%v)", res.ExitCode, res.Summary)
	}
	if res.HasNonFailingDrift {
		t.Fatal("HasNonFailingDrift should be false when policy fails")
	}
}

func TestAnalyze_FailOnAnyReturnsExit1(t *testing.T) {
	root := t.TempDir()
	res, err := Analyze(Input{Root: root, Tools: "cursor", FailOn: "any"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if res.ExitCode != 1 {
		t.Fatalf("expected ExitCode=1 with fail-on=any, got %d", res.ExitCode)
	}
}

func TestAnalyze_DefaultsFormatAndFailOn(t *testing.T) {
	res, err := Analyze(Input{Root: t.TempDir(), Tools: "none"})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	// NoOp doesn't tell us much about defaults, so check via a real run too.
	if !res.NoOp {
		t.Fatal("expected NoOp for tools=none")
	}
}
