package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/integrations"
	"pacto/internal/ui"
)

type doctorOptions struct {
	root    string
	tools   string
	format  string
	failOn  string
	verbose bool
}

func RunDoctor(args []string) int {
	lang := effectiveLanguage("")
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto doctor [--root <path>] [--tools <all|none|csv>] [--format table|json] [--fail-on none|drift|legacy|any] [--verbose]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	opts := doctorOptions{}
	fs.StringVar(&opts.root, "root", "", "Project root path")
	fs.StringVar(&opts.tools, "tools", "", "Tools to analyze: all, none, or comma-separated list")
	fs.StringVar(&opts.format, "format", "table", "Output format: table|json")
	fs.StringVar(&opts.failOn, "fail-on", "none", "Failure policy: none|drift|legacy|any")
	fs.BoolVar(&opts.verbose, "verbose", false, "Print debug details")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "%s: %v\n", tr(lang, "parse flags", "error parseando flags"), err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return 2
	}
	opts.format = strings.ToLower(strings.TrimSpace(opts.format))
	opts.failOn = strings.ToLower(strings.TrimSpace(opts.failOn))
	if opts.format != "table" && opts.format != "json" {
		fmt.Fprintln(os.Stderr, "invalid --format value (allowed: table|json)")
		return 2
	}
	switch opts.failOn {
	case "none", "drift", "legacy", "any":
	default:
		fmt.Fprintln(os.Stderr, "invalid --fail-on value (allowed: none|drift|legacy|any)")
		return 2
	}

	projectRoot := strings.TrimSpace(opts.root)
	if projectRoot == "" {
		abs, err := filepath.Abs(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
			return 2
		}
		projectRoot = abs
	}
	projectRoot = cleanAbs(projectRoot)

	tools, err := resolveDoctorTools(projectRoot, strings.TrimSpace(opts.tools))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 2
	}
	if len(tools) == 0 {
		fmt.Println(ui.Dim(tr(lang, "No tools selected; nothing to do.", "No se seleccionaron herramientas; nada por hacer.")))
		return 0
	}
	records, err := integrations.AnalyzeDrift(projectRoot, tools)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor analysis: %v\n", err)
		return 3
	}
	summary := summarizeDoctor(records)

	if opts.format == "json" {
		out := map[string]any{
			"root":    projectRoot,
			"tools":   tools,
			"summary": summary,
			"records": records,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "render doctor json: %v\n", err)
			return 3
		}
		fmt.Println(string(b))
	} else {
		fmt.Print(renderDoctorTable(projectRoot, tools, records, summary))
	}

	code := evaluateDoctorExit(opts.failOn, records)
	if code == 0 && opts.failOn == "none" && (summary["drift"] > 0 || summary["legacy"] > 0) {
		fmt.Fprintf(os.Stderr, "%s: %s\n", tr(lang, "warning", "advertencia"), tr(lang, "drift detected; run `pacto update --artifacts` to refresh managed artifacts", "drift detectado; ejecuta `pacto update --artifacts` para refrescar artefactos gestionados"))
	}
	if opts.verbose {
		fmt.Fprintf(os.Stderr, "doctor: root=%s tools=%s format=%s fail-on=%s records=%d\n", projectRoot, strings.Join(tools, ","), opts.format, opts.failOn, len(records))
	}
	return code
}

func resolveDoctorTools(projectRoot, raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		detected, err := integrations.DetectTools(projectRoot)
		if err != nil {
			return nil, err
		}
		if len(detected) == 0 {
			return nil, fmt.Errorf("no tools detected. Use --tools (%s)", strings.Join(integrations.SupportedTools(), ","))
		}
		return detected, nil
	}
	return integrations.ParseToolsArg(raw)
}

func summarizeDoctor(records []integrations.DriftRecord) map[string]int {
	summary := map[string]int{
		"ok":        0,
		"drift":     0,
		"legacy":    0,
		"missing":   0,
		"unmanaged": 0,
		"errors":    0,
	}
	for _, r := range records {
		switch r.Status {
		case integrations.DriftOK:
			summary["ok"]++
		case integrations.DriftLegacyPattern:
			summary["legacy"]++
		case integrations.DriftMissing:
			summary["drift"]++
			summary["missing"]++
		case integrations.DriftUnmanaged:
			summary["drift"]++
			summary["unmanaged"]++
		case integrations.DriftLegacyManaged, integrations.DriftStale, integrations.DriftMetaMismatch:
			summary["drift"]++
		default:
			summary["errors"]++
		}
	}
	return summary
}

func evaluateDoctorExit(policy string, records []integrations.DriftRecord) int {
	if policy == "none" {
		return 0
	}
	hasDrift := false
	hasLegacy := false
	for _, r := range records {
		switch r.Status {
		case integrations.DriftLegacyPattern:
			hasLegacy = true
		case integrations.DriftMissing, integrations.DriftUnmanaged, integrations.DriftLegacyManaged, integrations.DriftStale, integrations.DriftMetaMismatch:
			hasDrift = true
		}
	}
	switch policy {
	case "drift":
		if hasDrift {
			return 1
		}
	case "legacy":
		if hasLegacy {
			return 1
		}
	case "any":
		if hasDrift || hasLegacy {
			return 1
		}
	}
	return 0
}

func renderDoctorTable(projectRoot string, tools []string, records []integrations.DriftRecord, summary map[string]int) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("ROOT: %s\n", displayPath(projectRoot)))
	b.WriteString(fmt.Sprintf("TOOLS: %s\n", strings.Join(tools, ",")))
	b.WriteString(fmt.Sprintf("SUMMARY: ok=%d drift=%d legacy=%d missing=%d unmanaged=%d errors=%d\n\n",
		summary["ok"], summary["drift"], summary["legacy"], summary["missing"], summary["unmanaged"], summary["errors"]))

	rows := make([]integrations.DriftRecord, 0, len(records))
	rows = append(rows, records...)
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Status != rows[j].Status {
			return rows[i].Status < rows[j].Status
		}
		if rows[i].Tool != rows[j].Tool {
			return rows[i].Tool < rows[j].Tool
		}
		return rows[i].Path < rows[j].Path
	})
	for _, r := range rows {
		b.WriteString(fmt.Sprintf("- [%s] %s %s %s\n", r.Status, r.Tool, r.Kind, displayPath(r.Path)))
		if strings.TrimSpace(r.Reason) != "" {
			b.WriteString(fmt.Sprintf("  reason: %s\n", r.Reason))
		}
		if strings.TrimSpace(r.RecommendedFix) != "" {
			b.WriteString(fmt.Sprintf("  fix: %s\n", r.RecommendedFix))
		}
	}
	return b.String()
}
