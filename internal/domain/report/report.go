// Package report defines the status-report domain DTOs assembled by the
// status use case and consumed by renderers and TUI.
//
// This package is pure: it depends only on the standard library and the
// claim domain package.
package report

import (
	"time"

	"pacto/internal/domain/claim"
)

// PlanStatus is the per-plan view aggregated from parsing + verification.
type PlanStatus struct {
	StateFolder    string                `json:"state_folder"`
	Slug           string                `json:"slug"`
	Readme         string                `json:"readme"`
	DeclaredStatus string                `json:"declared_status"`
	DerivedStatus  string                `json:"derived_status"`
	ProgressPct    *int                  `json:"progress_percent,omitempty"`
	LastUpdatedAt  *time.Time            `json:"last_updated_at,omitempty"`
	PendingTasks   int                   `json:"pending_tasks"`
	BlockedTasks   int                   `json:"blocked_tasks"`
	Blockers       []string              `json:"blockers"`
	NextActions    []string              `json:"next_actions"`
	Verification   string                `json:"verification,omitempty"`
	Confidence     string                `json:"confidence"`
	Claims         []claim.Result        `json:"claims,omitempty"`
	Requirements   []RequirementCoverage `json:"requirements,omitempty"`
	ParseWarnings  []string              `json:"parse_warnings,omitempty"`
	ParseError     string                `json:"parse_error,omitempty"`
}

// RequirementCoverage is the per-Requirement coverage view computed from
// the plan's spec.md and tasks.md.
type RequirementCoverage struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Tasks     int    `json:"tasks"`
	Evidence  int    `json:"evidence"`
	Uncovered bool   `json:"uncovered"`
}

// Summary aggregates counts across all discovered plans.
type Summary struct {
	TotalPlans        int            `json:"total_plans"`
	ByState           map[string]int `json:"by_state"`
	ByVerification    map[string]int `json:"by_verification,omitempty"`
	TotalPendingTasks int            `json:"total_pending_tasks"`
	TotalBlockedTasks int            `json:"total_blocked_tasks"`
}

// DomainOverlap groups plans that touch the same logical domain.
type DomainOverlap struct {
	Domain string   `json:"domain"`
	Plans  []string `json:"plans"`
}

// StatusReport is the full status-command output model.
type StatusReport struct {
	GeneratedAt         time.Time       `json:"generated_at"`
	Root                string          `json:"root"`
	PlansRoot           string          `json:"plans_root,omitempty"`
	RepoRoot            string          `json:"repo_root,omitempty"`
	Mode                string          `json:"mode"`
	VerificationEnabled bool            `json:"-"`
	Summary             Summary         `json:"summary"`
	Plans               []PlanStatus    `json:"plans"`
	Overlaps            []DomainOverlap `json:"overlaps,omitempty"`
}
