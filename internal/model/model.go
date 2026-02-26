package model

import "time"

type PlanRef struct {
	State    string
	Slug     string
	Dir      string
	Readme   string
	PlanDocs []string
}

type Phase struct {
	Name     string
	RawState string
	Progress int
}

type Task struct {
	Text      string
	Completed bool
	LikelyBlk bool
}

type ClaimType string

const (
	ClaimPath     ClaimType = "path"
	ClaimSymbol   ClaimType = "symbol"
	ClaimEndpoint ClaimType = "endpoint"
	ClaimTestRef  ClaimType = "test_ref"
	ClaimDelta    ClaimType = "delta"
)

type ClaimResult struct {
	ClaimType  ClaimType `json:"claim_type"`
	SourceText string    `json:"source_text"`
	Evidence   string    `json:"evidence"`
	Result     string    `json:"result"`
	References []string  `json:"references,omitempty"`
}

type PlanStatus struct {
	StateFolder    string        `json:"state_folder"`
	Slug           string        `json:"slug"`
	Readme         string        `json:"readme"`
	DeclaredStatus string        `json:"declared_status"`
	DerivedStatus  string        `json:"derived_status"`
	ProgressPct    *int          `json:"progress_percent,omitempty"`
	PendingTasks   int           `json:"pending_tasks"`
	BlockedTasks   int           `json:"blocked_tasks"`
	Blockers       []string      `json:"blockers"`
	NextActions    []string      `json:"next_actions"`
	Verification   string        `json:"verification"`
	Confidence     string        `json:"confidence"`
	Claims         []ClaimResult `json:"claims,omitempty"`
	ParseWarnings  []string      `json:"parse_warnings,omitempty"`
	ParseError     string        `json:"parse_error,omitempty"`
}

type Summary struct {
	TotalPlans        int            `json:"total_plans"`
	ByState           map[string]int `json:"by_state"`
	ByVerification    map[string]int `json:"by_verification"`
	TotalPendingTasks int            `json:"total_pending_tasks"`
	TotalBlockedTasks int            `json:"total_blocked_tasks"`
}

type StatusReport struct {
	GeneratedAt time.Time    `json:"generated_at"`
	Root        string       `json:"root"`
	Mode        string       `json:"mode"`
	Summary     Summary      `json:"summary"`
	Plans       []PlanStatus `json:"plans"`
}
