// Package status contains the application/use-case for building a project
// status report from the plans tree. It orchestrates discovery, parsing,
// claim extraction + verification, domain overlap detection, and analysis,
// returning a domain StatusReport. CLI wiring (flag parsing, config load,
// rendering, exit codes) lives in internal/command/status.
package status

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/analyze"
	"pacto/internal/claims"
	plancontext "pacto/internal/context"
	"pacto/internal/discovery"
	"pacto/internal/domain/claim"
	"pacto/internal/domain/report"
	"pacto/internal/parser"
	"pacto/internal/verify"
)

// Input captures everything BuildReport needs to produce a status report.
// It is intentionally a value type with no command-line/config concerns.
type Input struct {
	PlansRoot      string
	RepoRoot       string
	Mode           string
	State          string
	IncludeArchive bool
	Verify         bool
	ClaimsPaths    bool
	MaxNextActions int
	MaxBlockers    int

	// Warnings are propagated into every plan in the produced report so
	// callers can surface config/runtime warnings via the same channel.
	Warnings []string
}

// BuildReport orchestrates the status pipeline:
//  1. discover plans under PlansRoot,
//  2. parse each plan,
//  3. (optionally) extract + verify claims,
//  4. detect domain overlaps across active plans,
//  5. build the analyzed report.
func BuildReport(in Input) (report.StatusReport, error) {
	plans, err := discovery.FindPlans(in.PlansRoot, discovery.Options{
		StateFilter:    in.State,
		IncludeArchive: in.IncludeArchive,
	})
	if err != nil {
		return report.StatusReport{}, fmt.Errorf("discover plans: %w", err)
	}

	parsed := make([]parser.ParsedPlan, 0, len(plans))
	claimsByPlan := map[string][]claim.Result{}
	warningsByPlan := map[string][]string{}
	planDomains := map[string][]string{}
	verifier := verify.New(in.RepoRoot, in.PlansRoot)
	claimOpts := claims.Options{Paths: in.ClaimsPaths}

	for _, plan := range plans {
		pp, pErr := parser.ParsePlan(plan, in.Mode)
		if pErr != nil {
			pp.ParseError = pErr.Error()
		}
		parsed = append(parsed, pp)
		key := plan.State + "/" + plan.Slug

		var verified []claim.Result
		if in.Verify {
			raw := claims.Extract(pp, claimOpts)
			verified = make([]claim.Result, 0, len(raw))
			for _, c := range raw {
				verified = append(verified, verifier.VerifyClaim(plan, c))
			}
		}
		claimsByPlan[key] = verified

		if len(in.Warnings) > 0 {
			warningsByPlan[key] = append(warningsByPlan[key], in.Warnings...)
		}
		if plan.State == "current" || plan.State == "to-implement" {
			domains := plancontext.ExtractDomains(filepath.Join(plan.Dir, "spec.md"))
			if len(domains) > 0 {
				planDomains[key] = domains
			}
		}
	}

	overlaps := plancontext.DetectOverlaps(planDomains)
	for _, overlap := range overlaps {
		warning := fmt.Sprintf("domain overlap: %s shared by %s", overlap.Domain, strings.Join(overlap.Plans, ", "))
		for _, planKey := range overlap.Plans {
			warningsByPlan[planKey] = appendUnique(warningsByPlan[planKey], warning)
		}
	}

	rep := analyze.Build(analyze.Input{
		Root:                in.PlansRoot,
		PlansRoot:           in.PlansRoot,
		RepoRoot:            in.RepoRoot,
		Mode:                in.Mode,
		VerificationEnabled: in.Verify,
		Plans:               parsed,
		Claims:              claimsByPlan,
		Warnings:            warningsByPlan,
	}, analyze.Options{MaxNextActions: in.MaxNextActions, MaxBlockers: in.MaxBlockers})

	rep.Overlaps = make([]report.DomainOverlap, 0, len(overlaps))
	for _, overlap := range overlaps {
		plans := append([]string{}, overlap.Plans...)
		sort.Strings(plans)
		rep.Overlaps = append(rep.Overlaps, report.DomainOverlap{
			Domain: overlap.Domain,
			Plans:  plans,
		})
	}
	return rep, nil
}

func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}
