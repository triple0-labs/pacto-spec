package planfmt

import (
	"strings"
	"testing"
)

func TestValidateCompliantPlan(t *testing.T) {
	plan := `# Plan: Demo

## Metadata
- Status: Draft
- Last Modified: 2026-03-11

## Problem Statement
x

## User Scenarios
### Scenario: Happy path
- **GIVEN** a user
- **WHEN** they submit
- **THEN** it succeeds

## Acceptance Criteria
- AC-001: Request returns 200.

## Implementation Plan by Phases
## Phase 1: Setup
- [ ] 1.1 add endpoint

## Evidence
- 2026-03-11 go test ./...
`

	issues := Validate(plan)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %+v", issues)
	}
}

func TestNormalizeConvertsLegacyTaskWithoutAddingMissingSections(t *testing.T) {
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
	if strings.Contains(res.Content, "## Metadatos") || strings.Contains(res.Content, "## Metadata") {
		t.Fatalf("did not expect normalize to auto-add metadata, got: %s", res.Content)
	}
}

func TestSectionPlaceholderSpanishUsesLocalizedKeywords(t *testing.T) {
	scenarios := sectionPlaceholder("user_scenarios", "es")
	if !strings.Contains(scenarios, "**DADO**") || !strings.Contains(scenarios, "**CUANDO**") || !strings.Contains(scenarios, "**ENTONCES**") {
		t.Fatalf("expected localized spanish scenario placeholder, got %q", scenarios)
	}

	phases := sectionPlaceholder("implementation_phases", "es")
	if !strings.Contains(phases, "## Fase 1:") {
		t.Fatalf("expected spanish phase placeholder, got %q", phases)
	}
}
