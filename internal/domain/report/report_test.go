package report

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"pacto/internal/domain/claim"
)

// TestStatusReportJSONShape pins the wire format of the status report.
// pacto status --format=json depends on these field names; this test is
// the canonical regression guard for that contract.
func TestStatusReportJSONShape(t *testing.T) {
	pct := 42
	updated := time.Date(2026, 4, 29, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 29, 13, 0, 0, 0, time.UTC)
	rep := StatusReport{
		GeneratedAt:         now,
		Root:                "/repo",
		PlansRoot:           "/repo/plans",
		RepoRoot:            "/repo",
		Mode:                "strict",
		VerificationEnabled: true, // intentionally not serialized (json:"-")
		Summary: Summary{
			TotalPlans:        2,
			ByState:           map[string]int{"current": 1, "to-implement": 1},
			ByVerification:    map[string]int{"ok": 1, "missing": 1},
			TotalPendingTasks: 3,
			TotalBlockedTasks: 1,
		},
		Plans: []PlanStatus{{
			StateFolder:    "current",
			Slug:           "alpha",
			Readme:         "README.md",
			DeclaredStatus: "in-progress",
			DerivedStatus:  "in-progress",
			ProgressPct:    &pct,
			LastUpdatedAt:  &updated,
			PendingTasks:   2,
			BlockedTasks:   1,
			Blockers:       []string{"db down"},
			NextActions:    []string{"ship it"},
			Verification:   "ok",
			Confidence:     "high",
			Claims: []claim.Result{{
				ClaimType: claim.Path, SourceText: "x.go", Result: "ok",
			}},
		}},
		Overlaps: []DomainOverlap{{
			Domain: "billing", Plans: []string{"current/alpha", "current/beta"},
		}},
	}

	b, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)

	for _, want := range []string{
		`"generated_at"`,
		`"root": "/repo"`,
		`"plans_root": "/repo/plans"`,
		`"repo_root": "/repo"`,
		`"mode": "strict"`,
		`"summary"`,
		`"total_plans": 2`,
		`"by_state"`,
		`"by_verification"`,
		`"total_pending_tasks": 3`,
		`"total_blocked_tasks": 1`,
		`"plans"`,
		`"state_folder": "current"`,
		`"slug": "alpha"`,
		`"readme": "README.md"`,
		`"declared_status": "in-progress"`,
		`"derived_status": "in-progress"`,
		`"progress_percent": 42`,
		`"last_updated_at"`,
		`"pending_tasks": 2`,
		`"blocked_tasks": 1`,
		`"blockers"`,
		`"next_actions"`,
		`"verification": "ok"`,
		`"confidence": "high"`,
		`"claims"`,
		`"overlaps"`,
		`"domain": "billing"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in JSON output:\n%s", want, got)
		}
	}

	// VerificationEnabled is presentation-only; must never be emitted.
	if strings.Contains(got, "VerificationEnabled") ||
		strings.Contains(got, `"verification_enabled"`) {
		t.Errorf("VerificationEnabled leaked into JSON: %s", got)
	}
}

func TestPlanStatusOmitEmpty(t *testing.T) {
	// Minimal PlanStatus; optional fields with omitempty must disappear.
	b, err := json.Marshal(PlanStatus{
		StateFolder:    "done",
		Slug:           "beta",
		Readme:         "README.md",
		DeclaredStatus: "done",
		DerivedStatus:  "done",
		Confidence:     "high",
		Blockers:       []string{},
		NextActions:    []string{},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	for _, omit := range []string{
		"progress_percent",
		"last_updated_at",
		"verification",
		"claims",
		"parse_warnings",
		"parse_error",
	} {
		if strings.Contains(got, omit) {
			t.Errorf("expected %q to be omitted, got %s", omit, got)
		}
	}
}

func TestSummaryOmitVerificationWhenNil(t *testing.T) {
	b, err := json.Marshal(Summary{TotalPlans: 1, ByState: map[string]int{"current": 1}})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "by_verification") {
		t.Errorf("expected by_verification to be omitted, got %s", b)
	}
}

func TestStatusReportRoundTrip(t *testing.T) {
	in := StatusReport{
		GeneratedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Mode:        "strict",
		Summary:     Summary{TotalPlans: 0, ByState: map[string]int{}},
		Plans:       []PlanStatus{},
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out StatusReport
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !out.GeneratedAt.Equal(in.GeneratedAt) || out.Mode != in.Mode {
		t.Errorf("round-trip mismatch: got %+v want %+v", out, in)
	}
}
