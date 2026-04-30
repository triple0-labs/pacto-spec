package specsbaseline

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "spec.md")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestParseRequirements_empty(t *testing.T) {
	got, err := ParseRequirements(writeTemp(t, "# Title\n\nNothing here.\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 requirements, got %d", len(got))
	}
}

func TestParseRequirements_basic(t *testing.T) {
	src := `# Spec

## Requirements

### Requirement: User can sign in
The system SHALL allow sign-in.

#### Scenario: Happy path
- WHEN the user submits credentials
- THEN a session is created

#### Scenario: Bad credentials
- WHEN the credentials are wrong
- THEN an error is shown

### Requirement: User can sign out
- detail

#### Scenario: Logout
- WHEN the user clicks logout
- THEN the session ends

## Other Section
ignored
`
	got, err := ParseRequirements(writeTemp(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 reqs, got %d", len(got))
	}
	if got[0].ID != "R-001" || got[0].Name != "User can sign in" {
		t.Fatalf("req0: %+v", got[0])
	}
	if len(got[0].Scenarios) != 2 {
		t.Fatalf("req0 scenarios: %d", len(got[0].Scenarios))
	}
	if got[0].Scenarios[0].ID != "S-001" || got[0].Scenarios[1].ID != "S-002" {
		t.Fatalf("scn ids: %+v", got[0].Scenarios)
	}
	if got[1].ID != "R-002" || got[1].Name != "User can sign out" {
		t.Fatalf("req1: %+v", got[1])
	}
}

func TestParseRequirements_idOverride(t *testing.T) {
	src := `## Requirements

### Requirement: Foo
<!-- id: R-042 -->
body

#### Scenario: S
<!-- id: S-099 -->
- WHEN x
- THEN y
`
	got, err := ParseRequirements(writeTemp(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got[0].ID != "R-042" {
		t.Fatalf("override id: %s", got[0].ID)
	}
	if got[0].Scenarios[0].ID != "S-099" {
		t.Fatalf("scn override: %s", got[0].Scenarios[0].ID)
	}
}

func TestParseRequirements_spanish(t *testing.T) {
	src := `## Requisitos

### Requisito: Usuario puede entrar
- detalle

#### Escenario: Camino feliz
- WHEN x
- THEN y
`
	got, err := ParseRequirements(writeTemp(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 1 || got[0].Name != "Usuario puede entrar" {
		t.Fatalf("got: %+v", got)
	}
	if got[0].Scenarios[0].Name != "Camino feliz" {
		t.Fatalf("scn: %+v", got[0].Scenarios)
	}
}

func TestParseRequirements_duplicateNameError(t *testing.T) {
	src := `## Requirements

### Requirement: Foo
#### Scenario: A
- WHEN x
- THEN y

### Requirement: Foo
#### Scenario: B
- WHEN x
- THEN y
`
	_, err := ParseRequirements(writeTemp(t, src))
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestParseRequirements_scenarioTooDeep(t *testing.T) {
	src := `## Requirements

### Requirement: Foo
##### Scenario: Wrong level
- WHEN x
- THEN y
`
	_, err := ParseRequirements(writeTemp(t, src))
	if err == nil {
		t.Fatal("expected error for scenario too deep")
	}
}

func TestParseDeltas_empty(t *testing.T) {
	got, err := ParseDeltas(writeTemp(t, "# Spec\n\nNo capabilities.\n"))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

func TestParseDeltas_addedAndModified(t *testing.T) {
	src := `# Spec

## Capability: auth

### ADDED Requirements

#### Requirement: New thing
The system SHALL do new thing.

##### Scenario: Happy
- WHEN x
- THEN y

### MODIFIED Requirements

#### Requirement: Existing thing
Updated body.

##### Scenario: Updated
- WHEN a
- THEN b

## Capability: billing

### REMOVED Requirements

#### Requirement: Old thing
Reason: no longer needed.
`
	caps, err := ParseDeltas(writeTemp(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(caps) != 2 {
		t.Fatalf("want 2 caps, got %d", len(caps))
	}
	if caps[0].Slug != "auth" || len(caps[0].Deltas) != 2 {
		t.Fatalf("auth: %+v", caps[0])
	}
	if caps[0].Deltas[0].Op != DeltaAdded || caps[0].Deltas[1].Op != DeltaModified {
		t.Fatalf("ops: %+v", caps[0].Deltas)
	}
	if caps[1].Slug != "billing" || caps[1].Deltas[0].Op != DeltaRemoved {
		t.Fatalf("billing: %+v", caps[1])
	}
	// REMOVED requirements need not have scenarios; verify parser accepts.
	if len(caps[1].Deltas[0].Requirements) != 1 {
		t.Fatalf("removed reqs: %+v", caps[1].Deltas[0].Requirements)
	}
}

func TestParseDeltas_renamed(t *testing.T) {
	src := `## Capability: auth

### RENAMED Requirements

#### Requirement: Old name
- to: New shiny name
`
	caps, err := ParseDeltas(writeTemp(t, src))
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	r := caps[0].Deltas[0].Requirements[0]
	if r.Name != "Old name" || r.NewName != "New shiny name" {
		t.Fatalf("renamed: %+v", r)
	}
}

func TestParseDeltas_renamedMissingTo(t *testing.T) {
	src := `## Capability: auth

### RENAMED Requirements

#### Requirement: Old name
some text without a to: line
`
	_, err := ParseDeltas(writeTemp(t, src))
	if err == nil {
		t.Fatal("expected error for missing `to:`")
	}
}

func TestParseDeltas_unknownOp(t *testing.T) {
	src := `## Capability: auth

### FROBNICATED Requirements

#### Requirement: x
#### Scenario: y
- WHEN x
- THEN y
`
	_, err := ParseDeltas(writeTemp(t, src))
	if err == nil {
		t.Fatal("expected unknown op error")
	}
}

func TestParseDeltas_duplicateCapability(t *testing.T) {
	src := `## Capability: auth

### ADDED Requirements

#### Requirement: A
##### Scenario: s
- WHEN x
- THEN y

## Capability: auth

### ADDED Requirements

#### Requirement: B
##### Scenario: s
- WHEN x
- THEN y
`
	_, err := ParseDeltas(writeTemp(t, src))
	if err == nil {
		t.Fatal("expected duplicate cap error")
	}
}
