package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pacto/internal/model"
)

func Render(r model.StatusReport, format string) (string, error) {
	switch strings.ToLower(format) {
	case "json":
		out := struct {
			GeneratedAt string             `json:"generated_at"`
			Root        string             `json:"root"`
			PlansRoot   string             `json:"plans_root,omitempty"`
			RepoRoot    string             `json:"repo_root,omitempty"`
			Mode        string             `json:"mode"`
			Summary     model.Summary      `json:"summary"`
			Plans       []model.PlanStatus `json:"plans"`
		}{
			GeneratedAt: r.GeneratedAt.Format(time.RFC3339),
			Root:        r.Root,
			PlansRoot:   r.PlansRoot,
			RepoRoot:    r.RepoRoot,
			Mode:        r.Mode,
			Summary:     r.Summary,
			Plans:       r.Plans,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	case "table":
		return renderTable(r), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func renderTable(r model.StatusReport) string {
	var b strings.Builder
	plansRoot := r.PlansRoot
	if plansRoot == "" {
		plansRoot = r.Root
	}
	repoRoot := r.RepoRoot
	if repoRoot == "" {
		repoRoot = r.Root
	}
	fmt.Fprintf(&b, "PLANS_ROOT: %s | REPO_ROOT: %s | MODE: %s | GENERATED: %s\n", plansRoot, repoRoot, r.Mode, r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "SUMMARY: plans=%d pending=%d blocked=%d\n", r.Summary.TotalPlans, r.Summary.TotalPendingTasks, r.Summary.TotalBlockedTasks)
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 130))
	fmt.Fprintf(&b, "%-14s %-36s %-11s %-8s %-8s %-10s %-12s\n", "STATE", "PLAN", "VERIF", "PENDING", "BLOCKED", "CONF", "DERIVED")
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 130))
	for _, p := range r.Plans {
		fmt.Fprintf(&b, "%-14s %-36s %-11s %-8d %-8d %-10s %-12s\n", p.StateFolder, shorten(p.Slug, 36), p.Verification, p.PendingTasks, p.BlockedTasks, p.Confidence, shorten(p.DerivedStatus, 12))
		if len(p.Blockers) > 0 {
			fmt.Fprintf(&b, "  blockers: %s\n", strings.Join(p.Blockers, " | "))
		}
		if len(p.NextActions) > 0 {
			fmt.Fprintf(&b, "  next: %s\n", strings.Join(p.NextActions, " | "))
		}
		if len(p.Claims) > 0 {
			fmt.Fprintf(&b, "  claims: %d\n", len(p.Claims))
		}
		if len(p.ParseWarnings) > 0 {
			fmt.Fprintf(&b, "  warnings: %s\n", strings.Join(p.ParseWarnings, " | "))
		}
		if p.ParseError != "" {
			fmt.Fprintf(&b, "  parse_error: %s\n", p.ParseError)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

func shorten(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}
