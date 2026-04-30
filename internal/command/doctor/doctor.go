package doctorcmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"pacto/internal/app/doctor"
	"pacto/internal/cmdutil"
	"pacto/internal/integrations"
	"pacto/internal/ui"
)

type Options struct {
	Root    string
	Tools   string
	Format  string
	FailOn  string
	Verbose bool
}

func Run(opts Options) int {
	res, err := doctor.Analyze(doctor.Input{
		Root:   opts.Root,
		Tools:  opts.Tools,
		Format: opts.Format,
		FailOn: opts.FailOn,
	})
	lang := cmdutil.EffectiveLanguage(res.ProjectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if errors.Is(err, doctor.ErrInvalid) {
			return 2
		}
		return 3
	}
	if res.NoOp {
		fmt.Println(ui.Dim(cmdutil.Tr(lang, "No tools selected; nothing to do.", "No se seleccionaron herramientas; nada por hacer.")))
		return 0
	}

	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = "table"
	}
	if format == "json" {
		out := map[string]any{
			"root":    res.ProjectRoot,
			"tools":   res.Tools,
			"summary": res.Summary,
			"records": res.Records,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "render doctor json: %v\n", err)
			return 3
		}
		fmt.Println(string(b))
	} else {
		fmt.Print(renderDoctorTable(res.ProjectRoot, res.Tools, res.Records, res.Summary))
	}

	if res.HasNonFailingDrift {
		fmt.Fprintf(os.Stderr, "%s: %s\n", cmdutil.Tr(lang, "warning", "advertencia"), cmdutil.Tr(lang, "drift detected; run `pacto update --artifacts` to refresh managed artifacts", "drift detectado; ejecuta `pacto update --artifacts` para refrescar artefactos gestionados"))
	}
	if opts.Verbose {
		policy := strings.ToLower(strings.TrimSpace(opts.FailOn))
		if policy == "" {
			policy = "none"
		}
		fmt.Fprintf(os.Stderr, "doctor: root=%s tools=%s format=%s fail-on=%s records=%d\n", res.ProjectRoot, strings.Join(res.Tools, ","), format, policy, len(res.Records))
	}
	return res.ExitCode
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
