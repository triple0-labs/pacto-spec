package normalizecmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/app/normalizeplans"
	"pacto/internal/cmdutil"
	"pacto/internal/ui"
)

type Options struct {
	Root           string
	State          string
	IncludeArchive bool
	Write          bool
	Format         string
}

func Run(opts Options) int {
	if strings.TrimSpace(opts.Format) == "" {
		opts.Format = "table"
	}
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "table", "json":
	default:
		fmt.Fprintf(os.Stderr, "invalid --format value %q (allowed: table|json)\n", opts.Format)
		return 2
	}

	report, err := normalizeplans.Apply(normalizeplans.Input{
		Root:           opts.Root,
		State:          opts.State,
		IncludeArchive: opts.IncludeArchive,
		Write:          opts.Write,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if errors.Is(err, normalizeplans.ErrInvalid) {
			return 2
		}
		return 3
	}

	if strings.EqualFold(opts.Format, "json") {
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

func printNormalizeReport(report normalizeplans.Report) {
	action := "Dry Run"
	if report.Write {
		action = "Applied"
	}
	fmt.Println(ui.ActionHeader("Normalized Plans", action))
	fmt.Println(cmdutil.PathLine("root", report.Root))
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
