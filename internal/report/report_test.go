package report

import (
	"strings"
	"testing"
	"time"

	"pacto/internal/model"
)

func TestRenderJSONAndTable(t *testing.T) {
	rep := model.StatusReport{
		GeneratedAt: time.Date(2026, 2, 28, 12, 0, 0, 0, time.UTC),
		PlansRoot:   "/tmp/plans",
		RepoRoot:    "/tmp/repo",
		Mode:        "compat",
		Summary:     model.Summary{TotalPlans: 1},
		Plans: []model.PlanStatus{{
			StateFolder:  "current",
			Slug:         "sample",
			Verification: "verified",
			Confidence:   "high",
		}},
	}
	j, err := Render(rep, "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(j, `"plans_root": "/tmp/plans"`) {
		t.Fatalf("unexpected json: %s", j)
	}

	table, err := Render(rep, "table")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(table, "sample") || !strings.Contains(table, "verified") {
		t.Fatalf("unexpected table: %s", table)
	}
}
