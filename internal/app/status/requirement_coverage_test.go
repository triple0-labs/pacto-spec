package status

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComputeRequirementCoverage(t *testing.T) {
	dir := t.TempDir()
	spec := `# Spec
## Requirements
### Requirement: Sign in
#### Scenario: Happy
- WHEN x
- THEN y
### Requirement: Sign out
#### Scenario: Happy
- WHEN x
- THEN y
### Requirement: Forgotten
#### Scenario: x
- WHEN x
- THEN y
`
	tasks := `# Tasks
## Implementation Plan by Phases
## Phase 1: x
- [ ] 1.1 implement R-001 sign-in
- [ ] 1.2 polish R-001
- [ ] 1.3 implement R-002 logout
## Evidence
- 2026-04-29 R-001 verified via cmd
`
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tasks.md"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}
	got := computeRequirementCoverage(dir)
	if len(got) != 3 {
		t.Fatalf("want 3 requirements, got %d (%+v)", len(got), got)
	}
	if got[0].ID != "R-001" || got[0].Tasks != 3 || got[0].Evidence != 1 || got[0].Uncovered {
		t.Fatalf("R-001 unexpected: %+v", got[0])
	}
	if got[1].ID != "R-002" || got[1].Tasks != 1 || got[1].Uncovered {
		t.Fatalf("R-002 unexpected: %+v", got[1])
	}
	if got[2].ID != "R-003" || got[2].Tasks != 0 || !got[2].Uncovered {
		t.Fatalf("R-003 should be uncovered: %+v", got[2])
	}
}

func TestComputeRequirementCoverage_legacyPlanReturnsNil(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "spec.md"), []byte("# Old plan\n## Acceptance Criteria\n- AC-001: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := computeRequirementCoverage(dir); got != nil {
		t.Fatalf("legacy plan should return nil, got %+v", got)
	}
}
