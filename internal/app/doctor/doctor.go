// Package doctor is the use-case layer for `pacto doctor`. It resolves the
// project root and target tools, runs the drift analysis, summarizes the
// results, and computes the exit code for a given fail-on policy. The CLI
// layer renders the report (table or JSON) and writes warnings.
package doctor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/integrations"
	"pacto/internal/workspace"
)

// ErrInvalid wraps validation problems (bad --format/--fail-on, bad --tools,
// unresolvable cwd, no tools detected). Callers map this to exit code 2.
var ErrInvalid = errors.New("invalid input")

// Input is the request for a doctor run.
type Input struct {
	// Root is the project root. Empty means: resolve from CWD.
	Root string
	// Tools is the raw --tools value. Empty means: auto-detect.
	Tools string
	// Format must be "table" or "json".
	Format string
	// FailOn must be "none", "drift", "legacy", or "any".
	FailOn string
}

// Result is the aggregated outcome of a doctor run.
type Result struct {
	ProjectRoot string
	Tools       []string
	Records     []integrations.DriftRecord
	Summary     map[string]int
	// ExitCode honors the FailOn policy.
	ExitCode int
	// NoOp is true when the resolved tool list is empty (e.g. --tools none).
	NoOp bool
	// HasNonFailingDrift is true when drift or legacy was detected but the
	// FailOn policy is "none" (so ExitCode is still 0). Lets the CLI emit a
	// soft warning.
	HasNonFailingDrift bool
}

// Analyze runs the doctor pipeline. It does not print anything.
func Analyze(in Input) (Result, error) {
	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format == "" {
		format = "table"
	}
	if format != "table" && format != "json" {
		return Result{}, fmt.Errorf("%w: invalid --format value (allowed: table|json)", ErrInvalid)
	}
	policy := strings.ToLower(strings.TrimSpace(in.FailOn))
	if policy == "" {
		policy = "none"
	}
	switch policy {
	case "none", "drift", "legacy", "any":
	default:
		return Result{}, fmt.Errorf("%w: invalid --fail-on value (allowed: none|drift|legacy|any)", ErrInvalid)
	}

	projectRoot := strings.TrimSpace(in.Root)
	if projectRoot == "" {
		abs, err := filepath.Abs(".")
		if err != nil {
			return Result{}, fmt.Errorf("%w: resolve cwd: %v", ErrInvalid, err)
		}
		projectRoot = abs
	}
	projectRoot = workspace.CleanAbs(projectRoot)
	if _, err := os.Stat(projectRoot); err != nil {
		return Result{}, fmt.Errorf("%w: resolve root: %v", ErrInvalid, err)
	}

	tools, err := resolveTools(projectRoot, strings.TrimSpace(in.Tools))
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	res := Result{ProjectRoot: projectRoot, Tools: tools}
	if len(tools) == 0 {
		res.NoOp = true
		return res, nil
	}

	records, err := integrations.AnalyzeDrift(projectRoot, tools)
	if err != nil {
		return Result{}, fmt.Errorf("doctor analysis: %w", err)
	}
	res.Records = records
	res.Summary = summarize(records)
	res.ExitCode = evaluateExit(policy, records)
	if res.ExitCode == 0 && policy == "none" && (res.Summary["drift"] > 0 || res.Summary["legacy"] > 0) {
		res.HasNonFailingDrift = true
	}
	return res, nil
}

func resolveTools(projectRoot, raw string) ([]string, error) {
	if raw == "" {
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

func summarize(records []integrations.DriftRecord) map[string]int {
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

func evaluateExit(policy string, records []integrations.DriftRecord) int {
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
