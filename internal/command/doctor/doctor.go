package doctorcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/cmdutil"
	"pacto/internal/integrations"
	"pacto/internal/ui"
	"pacto/internal/workspace"
)

type Options struct {
	Root    string
	Tools   string
	Format  string
	FailOn  string
	Verbose bool
}

func Run(opts Options) int {
	opts.Format = strings.ToLower(strings.TrimSpace(opts.Format))
	if opts.Format == "" {
		opts.Format = "table"
	}
	opts.FailOn = strings.ToLower(strings.TrimSpace(opts.FailOn))
	if opts.FailOn == "" {
		opts.FailOn = "none"
	}
	if opts.Format != "table" && opts.Format != "json" {
		fmt.Fprintln(os.Stderr, "invalid --format value (allowed: table|json)")
		return 2
	}
	switch opts.FailOn {
	case "none", "drift", "legacy", "any":
	default:
		fmt.Fprintln(os.Stderr, "invalid --fail-on value (allowed: none|drift|legacy|any)")
		return 2
	}
	projectRoot := strings.TrimSpace(opts.Root)
	if projectRoot == "" {
		abs, err := filepath.Abs(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
			return 2
		}
		projectRoot = abs
	}
	projectRoot = workspace.CleanAbs(projectRoot)
	lang := cmdutil.EffectiveLanguage(projectRoot)

	tools, err := resolveDoctorTools(projectRoot, strings.TrimSpace(opts.Tools))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 2
	}
	if len(tools) == 0 {
		fmt.Println(ui.Dim(cmdutil.Tr(lang, "No tools selected; nothing to do.", "No se seleccionaron herramientas; nada por hacer.")))
		return 0
	}
	records, err := integrations.AnalyzeDrift(projectRoot, tools)
	if err != nil {
		fmt.Fprintf(os.Stderr, "doctor analysis: %v\n", err)
		return 3
	}
	summary := summarizeDoctor(records)

	if opts.Format == "json" {
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

	code := evaluateDoctorExit(opts.FailOn, records)
	if code == 0 && opts.FailOn == "none" && (summary["drift"] > 0 || summary["legacy"] > 0) {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdutil.Tr(lang, "warning", "advertencia"), cmdutil.Tr(lang, "drift detected; run `pacto update --artifacts` to refresh managed artifacts", "drift detectado; ejecuta `pacto update --artifacts` para refrescar artefactos gestionados"))
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "doctor: root=%s tools=%s format=%s fail-on=%s records=%d\n", projectRoot, strings.Join(tools, ","), opts.Format, opts.FailOn, len(records))
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
	b.WriteString(fmt.Sprintf("ROOT: %s\n", cmdutil.DisplayPath(projectRoot)))
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
		b.WriteString(fmt.Sprintf("- [%s] %s %s %s\n", r.Status, r.Tool, r.Kind, cmdutil.DisplayPath(r.Path)))
		if strings.TrimSpace(r.Reason) != "" {
			b.WriteString(fmt.Sprintf("  reason: %s\n", r.Reason))
		}
		if strings.TrimSpace(r.RecommendedFix) != "" {
			b.WriteString(fmt.Sprintf("  fix: %s\n", r.RecommendedFix))
		}
	}
	return b.String()
}
