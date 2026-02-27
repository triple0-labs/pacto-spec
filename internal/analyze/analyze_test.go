package analyze

import (
	"testing"

	"pacto/internal/model"
	"pacto/internal/parser"
)

func TestBuildDerivesBlockedAndVerification(t *testing.T) {
	in := Input{
		Root: ".",
		Mode: "compat",
		Plans: []parser.ParsedPlan{
			{
				Ref:          model.PlanRef{State: "current", Slug: "a"},
				Tasks:        []model.Task{{Text: "blocked by env", Completed: false, LikelyBlk: true}},
				BlockerHints: []string{"blocked by env"},
			},
		},
		Claims: map[string][]model.ClaimResult{
			"current/a": {
				{ClaimType: model.ClaimPath, SourceText: "x", Result: "unverified"},
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
				Ref:   model.PlanRef{State: "to-implement", Slug: "b"},
				Tasks: []model.Task{{Text: "first task", Completed: false}},
			},
		},
		Claims: map[string][]model.ClaimResult{},
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
