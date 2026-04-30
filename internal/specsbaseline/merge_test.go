package specsbaseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPlanMerge_addedCreatesBaseline(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op: DeltaAdded,
			Requirements: []Requirement{{
				Name:      "Sign in",
				Body:      []string{"The system SHALL sign in users."},
				Scenarios: []Scenario{{Name: "Happy", Lines: []string{"- WHEN x", "- THEN y"}}},
			}},
		}},
	}}
	files, err := PlanMerge(specsDir, "demo-plan", caps)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if len(files) != 1 || files[0].Existed {
		t.Fatalf("unexpected: %+v", files)
	}
	if !strings.Contains(files[0].Content, "### Requirement: Sign in") {
		t.Fatalf("content missing requirement: %s", files[0].Content)
	}
	if !strings.Contains(files[0].Content, "#### Scenario: Happy") {
		t.Fatalf("content missing scenario: %s", files[0].Content)
	}
	if err := CommitMerge(files); err != nil {
		t.Fatalf("commit: %v", err)
	}
	got, err := os.ReadFile(files[0].Path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "Sign in") {
		t.Fatalf("file missing content: %s", got)
	}
}

func TestPlanMerge_addedDuplicateErrors(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "auth"), 0o775); err != nil {
		t.Fatal(err)
	}
	existing := "# Capability: auth\n\n## Requirements\n\n### Requirement: Sign in\nbody\n\n#### Scenario: A\n- WHEN x\n- THEN y\n"
	if err := os.WriteFile(filepath.Join(specsDir, "auth", "spec.md"), []byte(existing), 0o664); err != nil {
		t.Fatal(err)
	}
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op: DeltaAdded,
			Requirements: []Requirement{{
				Name:      "Sign in",
				Body:      []string{"dup"},
				Scenarios: []Scenario{{Name: "B", Lines: []string{"- WHEN x", "- THEN y"}}},
			}},
		}},
	}}
	if _, err := PlanMerge(specsDir, "p", caps); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestPlanMerge_modifiedReplaces(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "auth"), 0o775); err != nil {
		t.Fatal(err)
	}
	existing := "# Capability: auth\n\n## Requirements\n\n### Requirement: Sign in\nold body\n\n#### Scenario: A\n- WHEN x\n- THEN y\n"
	if err := os.WriteFile(filepath.Join(specsDir, "auth", "spec.md"), []byte(existing), 0o664); err != nil {
		t.Fatal(err)
	}
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op: DeltaModified,
			Requirements: []Requirement{{
				Name:      "Sign in",
				Body:      []string{"new body"},
				Scenarios: []Scenario{{Name: "Updated", Lines: []string{"- WHEN x", "- THEN y"}}},
			}},
		}},
	}}
	files, err := PlanMerge(specsDir, "p", caps)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if !strings.Contains(files[0].Content, "new body") {
		t.Fatalf("modified content not present: %s", files[0].Content)
	}
	if strings.Contains(files[0].Content, "old body") {
		t.Fatalf("old body still present: %s", files[0].Content)
	}
	if !strings.Contains(files[0].Content, "Updated") {
		t.Fatalf("updated scenario not present: %s", files[0].Content)
	}
}

func TestPlanMerge_modifiedMissingErrors(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "auth"), 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specsDir, "auth", "spec.md"), []byte("# Capability: auth\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op: DeltaModified,
			Requirements: []Requirement{{
				Name:      "Missing",
				Body:      []string{"x"},
				Scenarios: []Scenario{{Name: "S", Lines: []string{"- WHEN x", "- THEN y"}}},
			}},
		}},
	}}
	if _, err := PlanMerge(specsDir, "p", caps); err == nil {
		t.Fatal("expected error for missing target")
	}
}

func TestPlanMerge_removedAddsAuditComment(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "auth"), 0o775); err != nil {
		t.Fatal(err)
	}
	existing := "# Capability: auth\n\n## Requirements\n\n### Requirement: Old\nbody\n\n#### Scenario: A\n- WHEN x\n- THEN y\n"
	if err := os.WriteFile(filepath.Join(specsDir, "auth", "spec.md"), []byte(existing), 0o664); err != nil {
		t.Fatal(err)
	}
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op:           DeltaRemoved,
			Requirements: []Requirement{{Name: "Old"}},
		}},
	}}
	files, err := PlanMerge(specsDir, "plan-x", caps)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if strings.Contains(files[0].Content, "### Requirement: Old") {
		t.Fatalf("removed req still present: %s", files[0].Content)
	}
	if !strings.Contains(files[0].Content, "removed by plan-x") {
		t.Fatalf("audit comment missing: %s", files[0].Content)
	}
}

func TestPlanMerge_renamedRewritesHeader(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(filepath.Join(specsDir, "auth"), 0o775); err != nil {
		t.Fatal(err)
	}
	existing := "# Capability: auth\n\n## Requirements\n\n### Requirement: Old name\nbody\n\n#### Scenario: A\n- WHEN x\n- THEN y\n"
	if err := os.WriteFile(filepath.Join(specsDir, "auth", "spec.md"), []byte(existing), 0o664); err != nil {
		t.Fatal(err)
	}
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op:           DeltaRenamed,
			Requirements: []Requirement{{Name: "Old name", NewName: "New name"}},
		}},
	}}
	files, err := PlanMerge(specsDir, "p", caps)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	if !strings.Contains(files[0].Content, "### Requirement: New name") {
		t.Fatalf("rename not applied: %s", files[0].Content)
	}
	if !strings.Contains(files[0].Content, "body") {
		t.Fatalf("body lost on rename: %s", files[0].Content)
	}
}

func TestCommitMerge_atomicTempCleanup(t *testing.T) {
	dir := t.TempDir()
	specsDir := filepath.Join(dir, "specs")
	caps := []Capability{{
		Slug: "auth",
		Deltas: []Delta{{
			Op: DeltaAdded,
			Requirements: []Requirement{{
				Name:      "X",
				Scenarios: []Scenario{{Name: "S", Lines: []string{"- WHEN x", "- THEN y"}}},
			}},
		}},
	}}
	files, err := PlanMerge(specsDir, "p", caps)
	if err != nil {
		t.Fatal(err)
	}
	if err := CommitMerge(files); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(specsDir, "auth"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".spec-") && strings.HasSuffix(e.Name(), ".tmp") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}
