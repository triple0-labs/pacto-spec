package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"pacto/internal/i18n"
	"pacto/internal/model"
)

func Render(r model.StatusReport, format string) (string, error) {
	return RenderWithLanguage(r, format, i18n.English)
}

func RenderWithLanguage(r model.StatusReport, format string, lang i18n.Language) (string, error) {
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
		return renderTable(r, lang), nil
	default:
		return "", fmt.Errorf("unsupported format %q", format)
	}
}

func renderTable(r model.StatusReport, lang i18n.Language) string {
	var b strings.Builder
	plansRoot := r.PlansRoot
	if plansRoot == "" {
		plansRoot = r.Root
	}
	repoRoot := r.RepoRoot
	if repoRoot == "" {
		repoRoot = r.Root
	}
	fmt.Fprintf(&b, "%s: %s | %s: %s | MODE: %s | %s: %s\n", i18n.T(lang, "PLANS_ROOT", "RAIZ_PLANES"), plansRoot, i18n.T(lang, "REPO_ROOT", "RAIZ_REPO"), repoRoot, r.Mode, i18n.T(lang, "GENERATED", "GENERADO"), r.GeneratedAt.Format(time.RFC3339))
	fmt.Fprintf(&b, "%s: plans=%d pending=%d blocked=%d\n", i18n.T(lang, "SUMMARY", "RESUMEN"), r.Summary.TotalPlans, r.Summary.TotalPendingTasks, r.Summary.TotalBlockedTasks)
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 130))
	fmt.Fprintf(&b, "%-14s %-36s %-11s %-8s %-8s %-10s %-12s\n", i18n.T(lang, "STATE", "ESTADO"), i18n.T(lang, "PLAN", "PLAN"), i18n.T(lang, "VERIF", "VERIF"), i18n.T(lang, "PENDING", "PEND"), i18n.T(lang, "BLOCKED", "BLOQ"), i18n.T(lang, "CONF", "CONF"), i18n.T(lang, "DERIVED", "DERIVADO"))
	fmt.Fprintf(&b, "%s\n", strings.Repeat("-", 130))
	for _, p := range r.Plans {
		fmt.Fprintf(&b, "%-14s %-36s %-11s %-8d %-8d %-10s %-12s\n", p.StateFolder, shorten(p.Slug, 36), p.Verification, p.PendingTasks, p.BlockedTasks, p.Confidence, shorten(p.DerivedStatus, 12))
		if len(p.Blockers) > 0 {
			fmt.Fprintf(&b, "  %s: %s\n", i18n.T(lang, "blockers", "bloqueadores"), strings.Join(p.Blockers, " | "))
		}
		if len(p.NextActions) > 0 {
			fmt.Fprintf(&b, "  %s: %s\n", i18n.T(lang, "next", "siguiente"), strings.Join(p.NextActions, " | "))
		}
		if len(p.Claims) > 0 {
			fmt.Fprintf(&b, "  %s: %d\n", i18n.T(lang, "claims", "afirmaciones"), len(p.Claims))
		}
		if len(p.ParseWarnings) > 0 {
			fmt.Fprintf(&b, "  %s: %s\n", i18n.T(lang, "warnings", "advertencias"), strings.Join(p.ParseWarnings, " | "))
		}
		if p.ParseError != "" {
			fmt.Fprintf(&b, "  %s: %s\n", i18n.T(lang, "parse_error", "error_parseo"), p.ParseError)
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
