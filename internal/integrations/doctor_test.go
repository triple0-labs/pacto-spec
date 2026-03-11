package integrations

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeDriftReportsOKForFreshArtifacts(t *testing.T) {
	root := t.TempDir()
	results := GenerateForTool(root, "cursor", false)
	for _, r := range results {
		if r.Err != nil {
			t.Fatalf("generate error: %v", r.Err)
		}
	}
	drift, err := AnalyzeDrift(root, []string{"cursor"})
	if err != nil {
		t.Fatalf("AnalyzeDrift error: %v", err)
	}
	if len(drift) == 0 {
		t.Fatal("expected drift records")
	}
	for _, r := range drift {
		if r.Status != DriftOK {
			t.Fatalf("expected all fresh artifacts to be ok, got %s (%s)", r.Status, r.Path)
		}
	}
}

func TestAnalyzeDriftDetectsMissingArtifact(t *testing.T) {
	root := t.TempDir()
	_ = GenerateForTool(root, "cursor", false)
	target := filepath.Join(root, ".cursor", "commands", "pacto-status.md")
	if err := os.Remove(target); err != nil {
		t.Fatal(err)
	}
	drift, err := AnalyzeDrift(root, []string{"cursor"})
	if err != nil {
		t.Fatalf("AnalyzeDrift error: %v", err)
	}
	if !hasStatusForPath(drift, target, DriftMissing) {
		t.Fatalf("expected missing status for %s", target)
	}
}

func TestAnalyzeDriftDetectsUnmanagedFile(t *testing.T) {
	root := t.TempDir()
	_ = GenerateForTool(root, "cursor", false)
	target := filepath.Join(root, ".cursor", "commands", "pacto-status.md")
	if err := os.WriteFile(target, []byte("custom file"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift, err := AnalyzeDrift(root, []string{"cursor"})
	if err != nil {
		t.Fatalf("AnalyzeDrift error: %v", err)
	}
	if !hasStatusForPath(drift, target, DriftUnmanaged) {
		t.Fatalf("expected unmanaged status for %s", target)
	}
}

func TestAnalyzeDriftDetectsLegacyManagedWithoutMetadata(t *testing.T) {
	root := t.TempDir()
	_ = GenerateForTool(root, "cursor", false)
	target := filepath.Join(root, ".cursor", "commands", "pacto-status.md")
	legacy := ManagedStart + "\nlegacy body\n" + ManagedEnd + "\n"
	if err := os.WriteFile(target, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	drift, err := AnalyzeDrift(root, []string{"cursor"})
	if err != nil {
		t.Fatalf("AnalyzeDrift error: %v", err)
	}
	if !hasStatusForPath(drift, target, DriftLegacyManaged) {
		t.Fatalf("expected legacy managed status for %s", target)
	}
}

func TestAnalyzeDriftDetectsLegacyPatternFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".cursor", "commands"), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(root, ".cursor", "commands", "pa-status.md")
	if err := os.WriteFile(legacy, []byte("legacy alias"), 0o644); err != nil {
		t.Fatal(err)
	}
	drift, err := AnalyzeDrift(root, []string{"cursor"})
	if err != nil {
		t.Fatalf("AnalyzeDrift error: %v", err)
	}
	if !hasStatusForPath(drift, legacy, DriftLegacyPattern) {
		t.Fatalf("expected legacy pattern status for %s", legacy)
	}
}

func hasStatusForPath(records []DriftRecord, path string, status DriftStatus) bool {
	for _, r := range records {
		if r.Path == path && r.Status == status {
			return true
		}
	}
	return false
}

