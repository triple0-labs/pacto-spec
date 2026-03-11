package planfmt

import (
	"strings"
	"testing"
)

func TestValidateCompliantPlan(t *testing.T) {
	plan := `# Plan: Demo

## Metadata
- Status: Draft

## Problem Statement
x

## Goals
x

## Non-Goals
x

## User Scenarios
### Scenario: Happy path
- **GIVEN** a user
- **WHEN** they submit
- **THEN** it succeeds

## Functional Requirements
- FR-001: The system MUST persist data.

## Non-Functional Requirements
- NFR-001: p95 < 200ms.

## Acceptance Criteria
- AC-001: Request returns 200.

## Technical Context
x

## Implementation Plan by Phases
## Phase 1: Setup
- [ ] 1.1 add endpoint

## Evidence
- 2026-03-11 go test ./...

## Risks and Mitigations
x

## Next Steps
1. ship`

	issues := Validate(plan)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %+v", issues)
	}
}

func TestNormalizeConvertsLegacyTaskAndAddsMissingSections(t *testing.T) {
	plan := `# Plan: Demo

## Objetivos
1. x

## Plan de implementación por fases
## Fase 2: Trabajo
- [ ] T3. completar flujo`

	res := Normalize(plan)
	if !res.Changed {
		t.Fatal("expected normalized content to change")
	}
	if !strings.Contains(res.Content, "- [ ] 2.3 completar flujo") {
		t.Fatalf("expected legacy task conversion, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "## Metadatos") {
		t.Fatalf("expected metadata section added, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "## Requerimientos Funcionales") {
		t.Fatalf("expected missing required section added, got: %s", res.Content)
	}
}
