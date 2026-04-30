// Package plan defines core plan-domain types: references, phases, and tasks.
//
// This package is pure: it depends only on the standard library and does not
// perform any I/O. It is the innermost layer of the pacto domain model.
package plan

// PlanRef identifies a plan slice on disk and the documents that compose it.
type PlanRef struct {
	State    string
	Slug     string
	Dir      string
	Readme   string
	PlanDocs []string
}

// Phase represents a milestone or grouping of tasks within a plan.
type Phase struct {
	Name     string
	RawState string
	Progress int
}

// Task represents a single actionable item declared in a plan document.
type Task struct {
	StepRef   string
	Phase     int
	Number    int
	Text      string
	Completed bool
	LikelyBlk bool
}
