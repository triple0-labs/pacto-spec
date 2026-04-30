// Package normalizeplans is the use-case layer for `pacto normalize`. It
// discovers plans, validates and normalizes each plan document, and
// optionally writes the normalized content back. The CLI layer renders the
// resulting Report as a table or JSON.
package normalizeplans

import (
	"errors"
	"fmt"
	"os"

	"pacto/internal/discovery"
	"pacto/internal/planfmt"
	"pacto/internal/workspace"
)

// ErrInvalid wraps validation problems (bad root, etc.). The CLI maps this
// to exit code 2.
var ErrInvalid = errors.New("invalid input")

// Input is the request for a normalize run.
type Input struct {
	Root           string
	State          string
	IncludeArchive bool
	Write          bool
}

// Item is the per-plan outcome.
type Item struct {
	State        string          `json:"state"`
	Slug         string          `json:"slug"`
	PlanPath     string          `json:"plan_path"`
	Changed      bool            `json:"changed"`
	Applied      bool            `json:"applied"`
	Changes      []string        `json:"changes,omitempty"`
	Warnings     []string        `json:"warnings,omitempty"`
	IssuesBefore []planfmt.Issue `json:"issues_before,omitempty"`
	IssuesAfter  []planfmt.Issue `json:"issues_after,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// Report is the aggregate result.
type Report struct {
	Root         string `json:"root"`
	Write        bool   `json:"write"`
	TotalPlans   int    `json:"total_plans"`
	ChangedPlans int    `json:"changed_plans"`
	AppliedPlans int    `json:"applied_plans"`
	ErroredPlans int    `json:"errored_plans"`
	Items        []Item `json:"items"`
}

// Apply runs the normalize pipeline. It does not print anything.
func Apply(in Input) (Report, error) {
	plansRoot, err := workspace.ResolvePlansRootForAction(in.Root)
	if err != nil {
		return Report{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	plans, err := discovery.FindPlans(plansRoot, discovery.Options{
		StateFilter:    in.State,
		IncludeArchive: in.IncludeArchive,
	})
	if err != nil {
		return Report{}, fmt.Errorf("discover plans: %w", err)
	}

	report := Report{
		Root:  plansRoot,
		Write: in.Write,
		Items: make([]Item, 0, len(plans)),
	}

	for _, ref := range plans {
		if len(ref.PlanDocs) == 0 {
			continue
		}
		planPath := ref.PlanDocs[0]
		item := Item{
			State:    ref.State,
			Slug:     ref.Slug,
			PlanPath: planPath,
		}

		b, err := os.ReadFile(planPath)
		if err != nil {
			item.Error = err.Error()
			report.Items = append(report.Items, item)
			report.ErroredPlans++
			continue
		}
		content := string(b)
		item.IssuesBefore = planfmt.Validate(content)
		norm := planfmt.Normalize(content)
		item.Changed = norm.Changed
		item.Changes = norm.Changes
		item.Warnings = norm.Warnings
		item.IssuesAfter = planfmt.Validate(norm.Content)

		if norm.Changed {
			report.ChangedPlans++
		}
		if in.Write && norm.Changed {
			if err := os.WriteFile(planPath, []byte(norm.Content), 0o664); err != nil {
				item.Error = err.Error()
				report.ErroredPlans++
			} else {
				item.Applied = true
				report.AppliedPlans++
			}
		}
		report.Items = append(report.Items, item)
	}
	report.TotalPlans = len(report.Items)
	return report, nil
}
