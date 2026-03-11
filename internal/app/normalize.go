package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/discovery"
	"pacto/internal/planfmt"
	"pacto/internal/ui"
)

type normalizeOptions struct {
	root           string
	state          string
	includeArchive bool
	write          bool
	format         string
}

type normalizeItem struct {
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

type normalizeReport struct {
	Root         string          `json:"root"`
	Write        bool            `json:"write"`
	TotalPlans   int             `json:"total_plans"`
	ChangedPlans int             `json:"changed_plans"`
	AppliedPlans int             `json:"applied_plans"`
	ErroredPlans int             `json:"errored_plans"`
	Items        []normalizeItem `json:"items"`
}

func RunNormalize(args []string) int {
	opts, code, ok := parseNormalizeArgs(args)
	if !ok {
		return code
	}
	plansRoot, err := resolvePlansRootForAction(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	plans, err := discovery.FindPlans(plansRoot, discovery.Options{
		StateFilter:    opts.state,
		IncludeArchive: opts.includeArchive,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "discover plans: %v\n", err)
		return 3
	}

	report := normalizeReport{
		Root:  plansRoot,
		Write: opts.write,
		Items: make([]normalizeItem, 0, len(plans)),
	}

	for _, ref := range plans {
		if len(ref.PlanDocs) == 0 {
			continue
		}
		planPath := ref.PlanDocs[0]
		item := normalizeItem{
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
		if opts.write && norm.Changed {
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

	if strings.EqualFold(opts.format, "json") {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "encode report: %v\n", err)
			return 3
		}
	} else {
		printNormalizeReport(report)
	}

	if report.ErroredPlans > 0 {
		return 3
	}
	return 0
}

func parseNormalizeArgs(args []string) (normalizeOptions, int, bool) {
	opts := normalizeOptions{}
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto normalize [--root <path>] [--state <state|all>] [--include-archive] [--write] [--format table|json]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}
	fs.StringVar(&opts.root, "root", "", "Project root path (auto-discovers when omitted)")
	fs.StringVar(&opts.state, "state", "all", "State filter: current|to-implement|done|outdated|all")
	fs.BoolVar(&opts.includeArchive, "include-archive", false, "Include archive plans")
	fs.BoolVar(&opts.write, "write", false, "Apply normalized content to files (default is dry-run)")
	fs.StringVar(&opts.format, "format", "table", "Output format: table|json")

	normalizedArgs, normErr := normalizeNormalizeArgs(args)
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return normalizeOptions{}, 2, false
	}
	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return normalizeOptions{}, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return normalizeOptions{}, 2, false
	}
	if strings.TrimSpace(opts.format) == "" {
		opts.format = "table"
	}
	switch strings.ToLower(strings.TrimSpace(opts.format)) {
	case "table", "json":
	default:
		fmt.Fprintf(os.Stderr, "invalid --format value %q (allowed: table|json)\n", opts.format)
		return normalizeOptions{}, 2, false
	}
	return opts, 0, true
}

func normalizeNormalizeArgs(args []string) ([]string, error) {
	withValue := map[string]bool{"--root": true, "-root": true, "--state": true, "-state": true, "--format": true, "-format": true}
	return normalizeArgs(args, withValue)
}

func printNormalizeReport(report normalizeReport) {
	action := "Dry Run"
	if report.Write {
		action = "Applied"
	}
	fmt.Println(ui.ActionHeader("Normalized Plans", action))
	fmt.Println(pathLine("root", report.Root))
	fmt.Printf("plans: %d  changed: %d  applied: %d  errors: %d\n", report.TotalPlans, report.ChangedPlans, report.AppliedPlans, report.ErroredPlans)

	for _, item := range report.Items {
		rel := item.State + "/" + item.Slug
		status := "ok"
		if item.Error != "" {
			status = "error"
		} else if item.Changed {
			if item.Applied {
				status = "applied"
			} else {
				status = "would-change"
			}
		}
		fmt.Printf("- [%s] %s (%s)\n", status, rel, filepath.Base(item.PlanPath))
		if item.Error != "" {
			fmt.Printf("  error: %s\n", item.Error)
			continue
		}
		if len(item.Changes) > 0 {
			fmt.Printf("  changes: %d\n", len(item.Changes))
		}
		if len(item.IssuesBefore) > 0 || len(item.IssuesAfter) > 0 {
			fmt.Printf("  issues: %d -> %d\n", len(item.IssuesBefore), len(item.IssuesAfter))
		}
	}
	if !report.Write {
		fmt.Println("no files were written (use --write to apply)")
	}
}
