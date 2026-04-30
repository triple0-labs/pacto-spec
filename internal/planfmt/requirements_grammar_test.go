package planfmt

import (
	"strings"
	"testing"
)

func TestValidateRequirementsGrammar_missingScenario(t *testing.T) {
	plan := `# Plan
## Metadata
- Last Modified: 2026-04-29
## Problem Statement
x
## User Scenarios
### Scenario: s
- WHEN x
- THEN y
## Acceptance Criteria
- AC-001: x
## Requirements
### Requirement: Lonely
body without scenario.
## Implementation Plan by Phases
## Phase 1: x
- [ ] 1.1 do
## Evidence
- 2026-04-29 cmd
`
	issues := Validate(plan)
	found := false
	for _, i := range issues {
		if i.Code == "requirement_missing_scenario" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected requirement_missing_scenario, got %+v", issues)
	}
}

func TestValidateRequirementsGrammar_capabilityDeltaMissingScenario(t *testing.T) {
	plan := `# Plan
## Metadata
- Last Modified: 2026-04-29
## Problem Statement
x
## User Scenarios
### Scenario: s
- WHEN x
- THEN y
## Acceptance Criteria
- AC-001: x
## Capability: auth
### ADDED Requirements
#### Requirement: New
body
## Implementation Plan by Phases
## Phase 1: x
- [ ] 1.1 do
## Evidence
- 2026-04-29 cmd
`
	issues := Validate(plan)
	got := ""
	for _, i := range issues {
		if i.Code == "requirement_missing_scenario" {
			got = i.Message
		}
	}
	if got == "" || !strings.Contains(got, "auth") {
		t.Fatalf("expected scenario error mentioning auth, got %+v", issues)
	}
}

func TestValidateRequirementsGrammar_unknownDeltaOp(t *testing.T) {
	plan := `## Capability: auth
### FROBNICATED Requirements
#### Requirement: x
##### Scenario: y
- WHEN x
- THEN y
`
	issues := Validate(plan)
	found := false
	for _, i := range issues {
		if i.Code == "capability_grammar" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected capability_grammar issue, got %+v", issues)
	}
}

func TestValidateRequirementsGrammar_legacyPlanNoIssues(t *testing.T) {
	// A legacy plan with no Requirements / Capability sections must not
	// produce any new grammar issues.
	plan := `# Plan
## Metadata
- Last Modified: 2026-04-29
## Problem Statement
x
## User Scenarios
### Scenario: s
- WHEN x
- THEN y
## Acceptance Criteria
- AC-001: x
## Implementation Plan by Phases
## Phase 1: x
- [ ] 1.1 do
## Evidence
- 2026-04-29 cmd
`
	for _, i := range Validate(plan) {
		if i.Code == "requirement_missing_scenario" || i.Code == "capability_grammar" || i.Code == "requirements_grammar" {
			t.Fatalf("legacy plan should have no grammar issues, got %+v", i)
		}
	}
}
