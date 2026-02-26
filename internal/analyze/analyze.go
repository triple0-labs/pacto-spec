package analyze

import (
	"sort"
	"strings"
	"time"

	"pacto/internal/model"
	"pacto/internal/parser"
)

type Options struct {
	MaxNextActions int
	MaxBlockers    int
}

type Input struct {
	Root     string
	Mode     string
	Plans    []parser.ParsedPlan
	Claims   map[string][]model.ClaimResult
	Warnings map[string][]string
}

func Build(in Input, opts Options) model.StatusReport {
	plans := make([]model.PlanStatus, 0, len(in.Plans))
	for _, p := range in.Plans {
		planKey := p.Ref.State + "/" + p.Ref.Slug
		claims := in.Claims[planKey]
		warn := append([]string{}, p.ParseWarnings...)
		warn = append(warn, in.Warnings[planKey]...)
		for _, c := range claims {
			if c.Evidence == "plan_doc_only" {
				warn = appendUniqueWarning(warn, "claim only found in plan docs; excluded from verification")
			}
		}

		pending := 0
		blocked := 0
		for _, t := range p.Tasks {
			if !t.Completed {
				pending++
			}
			if !t.Completed && t.LikelyBlk {
				blocked++
			}
		}
		if blocked == 0 {
			blocked = len(p.BlockerHints)
		}

		declared := p.DeclaredStatus
		if strings.TrimSpace(declared) == "" {
			declared = deriveStatusFromState(p.Ref.State)
		}
		derived := deriveFromSignals(p, blocked)
		if p.LatestDeltaTime != nil {
			derived = deriveFromDelta(*p.LatestDeltaTime, derived)
		}

		progress := deriveProgress(p)
		next := p.NextActions
		if len(next) == 0 {
			next = fallbackNextActions(p)
		}

		verification := classifyVerification(p, claims)
		confidence := classifyConfidence(p, claims)

		plans = append(plans, model.PlanStatus{
			StateFolder:    p.Ref.State,
			Slug:           p.Ref.Slug,
			Readme:         p.Ref.Readme,
			DeclaredStatus: declared,
			DerivedStatus:  derived,
			ProgressPct:    progress,
			PendingTasks:   pending,
			BlockedTasks:   blocked,
			Blockers:       truncateSlice(p.BlockerHints, opts.MaxBlockers),
			NextActions:    truncateSlice(next, opts.MaxNextActions),
			Verification:   verification,
			Confidence:     confidence,
			Claims:         claims,
			ParseWarnings:  warn,
			ParseError:     p.ParseError,
		})
	}

	sort.Slice(plans, func(i, j int) bool {
		if plans[i].StateFolder == plans[j].StateFolder {
			return plans[i].Slug < plans[j].Slug
		}
		return plans[i].StateFolder < plans[j].StateFolder
	})

	summary := summarize(plans)
	return model.StatusReport{GeneratedAt: time.Now().UTC(), Root: in.Root, Mode: in.Mode, Summary: summary, Plans: plans}
}

func summarize(plans []model.PlanStatus) model.Summary {
	byState := map[string]int{}
	byVerif := map[string]int{}
	pending := 0
	blocked := 0
	for _, p := range plans {
		byState[p.StateFolder]++
		byVerif[p.Verification]++
		pending += p.PendingTasks
		blocked += p.BlockedTasks
	}
	return model.Summary{TotalPlans: len(plans), ByState: byState, ByVerification: byVerif, TotalPendingTasks: pending, TotalBlockedTasks: blocked}
}

func deriveFromSignals(p parser.ParsedPlan, blocked int) string {
	if blocked > 0 {
		return "blocked"
	}
	if len(p.Tasks) > 0 {
		allDone := true
		for _, t := range p.Tasks {
			if !t.Completed {
				allDone = false
				break
			}
		}
		if allDone {
			return "completed"
		}
		return "in_progress"
	}
	if len(p.Phases) > 0 {
		for _, ph := range p.Phases {
			if ph.Progress > 0 && ph.Progress < 100 {
				return "in_progress"
			}
		}
	}
	return "pending"
}

func deriveFromDelta(_ time.Time, fallback string) string {
	return fallback
}

func deriveProgress(p parser.ParsedPlan) *int {
	if len(p.Phases) == 0 {
		return nil
	}
	sum := 0
	for _, ph := range p.Phases {
		sum += ph.Progress
	}
	avg := sum / len(p.Phases)
	return &avg
}

func classifyVerification(p parser.ParsedPlan, claims []model.ClaimResult) string {
	verified := 0
	unverified := 0
	for _, c := range claims {
		switch c.Result {
		case "verified":
			verified++
		case "unverified":
			unverified++
		}
	}
	if len(claims) == 0 {
		if p.HasEvidence {
			return "partial"
		}
		return "unverified"
	}
	if verified > 0 && unverified == 0 {
		if p.ParseError == "" {
			return "verified"
		}
		return "partial"
	}
	if verified > 0 {
		return "partial"
	}
	return "unverified"
}

func classifyConfidence(p parser.ParsedPlan, claims []model.ClaimResult) string {
	signals := 0
	if p.DeclaredStatus != "" {
		signals++
	}
	if len(p.Phases) > 0 {
		signals++
	}
	if len(p.Tasks) > 0 {
		signals++
	}
	if p.HasCheckpoint || p.LatestDeltaTime != nil {
		signals++
	}
	if len(claims) > 0 {
		signals++
	}
	if signals >= 4 {
		return "high"
	}
	if signals >= 2 {
		return "medium"
	}
	return "low"
}

func deriveStatusFromState(state string) string {
	switch state {
	case "current":
		return "In Progress"
	case "to-implement":
		return "Pending"
	case "done":
		return "Completed"
	case "outdated":
		return "Outdated"
	default:
		return "unknown"
	}
}

func fallbackNextActions(p parser.ParsedPlan) []string {
	actions := make([]string, 0, 2)
	for _, t := range p.Tasks {
		if !t.Completed {
			actions = append(actions, t.Text)
			if len(actions) == 2 {
				break
			}
		}
	}
	if len(actions) == 0 && len(p.BlockerHints) > 0 {
		actions = append(actions, "Resolve blockers detected in plan")
	}
	if len(actions) == 0 {
		actions = append(actions, "Update plan with concrete next steps")
	}
	return actions
}

func truncateSlice(items []string, n int) []string {
	if n <= 0 {
		return nil
	}
	if len(items) <= n {
		return items
	}
	return items[:n]
}

func appendUniqueWarning(items []string, msg string) []string {
	for _, it := range items {
		if it == msg {
			return items
		}
	}
	return append(items, msg)
}
