package analyze

import (
	"testing"
	"time"

	"pacto/internal/domain/claim"
	"pacto/internal/domain/plan"
	"pacto/internal/parser"
)

func TestBuildDerivesBlockedAndVerification(t *testing.T) {
	in := Input{
		Root:                ".",
		Mode:                "compat",
		VerificationEnabled: true,
		Plans: []parser.ParsedPlan{
			{
				Ref:          plan.PlanRef{State: "current", Slug: "a"},
				Tasks:        []plan.Task{{Text: "blocked by env", Completed: false, LikelyBlk: true}},
				BlockerHints: []string{"blocked by env"},
			},
		},
		Claims: map[string][]claim.Result{
			"current/a": {
				{ClaimType: claim.Path, SourceText: "x", Result: "unverified"},
			},
		},
	}

	rep := Build(in, Options{MaxNextActions: 3, MaxBlockers: 3})
	if len(rep.Plans) != 1 {
		t.Fatalf("plans=%d, want 1", len(rep.Plans))
	}
	p := rep.Plans[0]
	if p.DerivedStatus != "blocked" {
		t.Fatalf("DerivedStatus=%q, want blocked", p.DerivedStatus)
	}
	if p.Verification != "unverified" {
		t.Fatalf("Verification=%q, want unverified", p.Verification)
	}
}

func TestBuildUsesFallbackNextActions(t *testing.T) {
	in := Input{
		Root: ".",
		Mode: "compat",
		Plans: []parser.ParsedPlan{
			{
				Ref:   plan.PlanRef{State: "to-implement", Slug: "b"},
				Tasks: []plan.Task{{Text: "first task", Completed: false}},
			},
		},
		Claims: map[string][]claim.Result{},
	}

	rep := Build(in, Options{MaxNextActions: 3, MaxBlockers: 3})
	p := rep.Plans[0]
	if len(p.NextActions) == 0 {
		t.Fatal("expected fallback next action")
	}
	if p.NextActions[0] != "first task" {
		t.Fatalf("NextActions[0]=%q, want %q", p.NextActions[0], "first task")
	}
}

func TestBuildDeltaSignalDoesNotOverrideDerivedStatus(t *testing.T) {
	dt := time.Date(2026, 3, 1, 10, 30, 0, 0, time.UTC)
	in := Input{
		Root: ".",
		Mode: "compat",
		Plans: []parser.ParsedPlan{
			{
				Ref:                 plan.PlanRef{State: "to-implement", Slug: "c"},
				HasStructuredDeltas: true,
				LatestDeltaTime:     &dt,
			},
		},
		Claims: map[string][]claim.Result{},
	}

	rep := Build(in, Options{MaxNextActions: 3, MaxBlockers: 3})
	if len(rep.Plans) != 1 {
		t.Fatalf("plans=%d, want 1", len(rep.Plans))
	}
	p := rep.Plans[0]
	if p.DerivedStatus != "pending" {
		t.Fatalf("DerivedStatus=%q, want pending", p.DerivedStatus)
	}
}
