// Package specsbaseline parses and merges capability baseline specs and the
// Requirement / Scenario grammar used inside plan spec.md files.
//
// Two grammar forms are recognised:
//
//  1. Plan-local Requirements (greenfield, not tied to baseline merge):
//
//     ## Requirements
//     ### Requirement: <name>
//     #### Scenario: <name>
//     - WHEN ...
//     - THEN ...
//
//  2. Capability deltas (consumed at `pacto move done`):
//
//     ## Capability: <slug>
//     ### ADDED Requirements
//     #### Requirement: <name>
//     ##### Scenario: <name>
//     - WHEN ...
//     - THEN ...
//
// English and Spanish header keywords are both recognised.
package specsbaseline

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// DeltaOp identifies a delta operation against a capability baseline.
type DeltaOp string

const (
	DeltaAdded    DeltaOp = "ADDED"
	DeltaModified DeltaOp = "MODIFIED"
	DeltaRemoved  DeltaOp = "REMOVED"
	DeltaRenamed  DeltaOp = "RENAMED"
)

// Scenario is a single behavioural example nested under a Requirement.
type Scenario struct {
	ID    string // S-NNN, scoped per Requirement
	Name  string
	Lines []string // raw body lines (WHEN/THEN bullets etc.) without the heading
}

// Requirement is the addressable unit of a capability spec.
type Requirement struct {
	ID         string // R-NNN
	Name       string
	NewName    string // populated only for RENAMED deltas
	Body       []string
	Scenarios  []Scenario
	HeaderLine int // 1-based source line of the Requirement heading
}

// Delta groups requirements modified by a single op inside a capability block.
type Delta struct {
	Op           DeltaOp
	Requirements []Requirement
}

// Capability is a delta block targeting one baseline file.
type Capability struct {
	Slug   string
	Deltas []Delta
}

var (
	reIDComment = regexp.MustCompile(`<!--\s*id:\s*([A-Za-z0-9_-]+)\s*-->`)
	reHeading   = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*$`)
)

// ParseRequirements reads the file at specPath and returns Requirements
// declared inside the plan-local `## Requirements` section, if any.
//
// Unknown files or files without the section yield (nil, nil).
func ParseRequirements(specPath string) ([]Requirement, error) {
	text, err := readFile(specPath)
	if err != nil {
		return nil, err
	}
	return ParseRequirementsText(text)
}

// ParseDeltas reads the file at specPath and returns Capability blocks
// (`## Capability: <slug>`) along with their delta operations.
//
// Files without any Capability block yield (nil, nil).
func ParseDeltas(specPath string) ([]Capability, error) {
	text, err := readFile(specPath)
	if err != nil {
		return nil, err
	}
	return ParseDeltasText(text)
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(b), nil
}

// --- helpers ---------------------------------------------------------------

func headingLevel(line string) (int, string, bool) {
	m := reHeading.FindStringSubmatch(line)
	if m == nil {
		return 0, "", false
	}
	return len(m[1]), strings.TrimSpace(m[2]), true
}

// matchKeyword returns the value following the first occurrence of any keyword
// in `keys` followed by ":". Case-insensitive. Returns ("", false) if no match.
//
// Example: matchKeyword("Requirement: Foo", []string{"Requirement","Requisito"})
// → ("Foo", true).
func matchKeyword(text string, keys []string) (string, bool) {
	lower := strings.ToLower(text)
	for _, k := range keys {
		kk := strings.ToLower(k) + ":"
		if strings.HasPrefix(lower, kk) {
			return strings.TrimSpace(text[len(kk):]), true
		}
	}
	return "", false
}

// matchExact returns true if text equals (case-insensitive, trimmed) any
// of `keys`.
func matchExact(text string, keys []string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	for _, k := range keys {
		if t == strings.ToLower(k) {
			return true
		}
	}
	return false
}

// extractIDOverride scans body lines for a `<!-- id: X -->` comment and
// returns the override (or "" if absent).
func extractIDOverride(body []string) string {
	for _, l := range body {
		if m := reIDComment.FindStringSubmatch(l); m != nil {
			return m[1]
		}
	}
	return ""
}

func assignReqID(idx int, override string) string {
	if override != "" {
		return override
	}
	return fmt.Sprintf("R-%03d", idx+1)
}

func assignScnID(idx int, override string) string {
	if override != "" {
		return override
	}
	return fmt.Sprintf("S-%03d", idx+1)
}

// keyword sets ----------------------------------------------------------------

var (
	kwRequirements = []string{"Requirements", "Requisitos"}
	kwRequirement  = []string{"Requirement", "Requisito"}
	kwScenario     = []string{"Scenario", "Escenario"}
	kwCapability   = []string{"Capability", "Capacidad"}
)

// reqHeading classifies a heading line as a Requirement and returns the
// declared name. Accepts EN+ES.
func reqHeading(text string) (string, bool) {
	return matchKeyword(text, kwRequirement)
}

func scnHeading(text string) (string, bool) {
	return matchKeyword(text, kwScenario)
}

func capHeading(text string) (string, bool) {
	return matchKeyword(text, kwCapability)
}

func isRequirementsSection(text string) bool {
	return matchExact(text, kwRequirements)
}

// deltaOpHeading recognises `### ADDED Requirements` (and ES) and returns
// the op. Accepts the English op keywords ADDED/MODIFIED/REMOVED/RENAMED
// followed by `Requirements` (EN) or `Requisitos` (ES).
func deltaOpHeading(text string) (DeltaOp, bool) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return "", false
	}
	op := strings.ToUpper(parts[0])
	switch DeltaOp(op) {
	case DeltaAdded, DeltaModified, DeltaRemoved, DeltaRenamed:
	default:
		return "", false
	}
	tail := strings.Join(parts[1:], " ")
	if !matchExact(tail, kwRequirements) {
		return "", false
	}
	return DeltaOp(op), true
}
