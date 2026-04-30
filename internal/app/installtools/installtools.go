// Package installtools is the use-case layer for `pacto install` and the
// artifact-refresh half of `pacto update`. It validates inputs, resolves the
// target tool list (auto-detect or explicit), and runs the per-tool artifact
// generation pipeline. The presentation layer (CLI) is responsible for
// printing per-result lines and the final summary.
package installtools

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/integrations"
	"pacto/internal/workspace"
)

// ErrInvalid wraps validation problems (missing root, bad --tools value,
// unresolvable cwd, no tools detected). Callers map this to exit code 2.
var ErrInvalid = errors.New("invalid input")

// Input is the request for the install pipeline.
type Input struct {
	// Root is the project root. Empty means: resolve from CWD.
	Root string
	// Tools is the raw --tools value. Empty means: auto-detect.
	Tools string
	// Force overwrites unmanaged files when generating artifacts.
	Force bool
}

// Result is the aggregated outcome of an install run.
type Result struct {
	// ProjectRoot is the resolved absolute project root.
	ProjectRoot string
	// Tools is the final tool list that was processed (in order).
	Tools []string
	// Items is the per-file outcomes returned by integrations, in the order
	// they were generated.
	Items []integrations.ArtifactResult
	// Counts.
	Created, Updated, Skipped, Failed int
	// NoOp is true when the tool list resolved to an empty set (e.g.
	// `--tools none`). The caller should print a friendly message.
	NoOp bool
}

// Install runs the artifact generation pipeline. It does not print anything.
func Install(in Input) (Result, error) {
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

	var tools []string
	if strings.TrimSpace(in.Tools) != "" {
		parsed, err := integrations.ParseToolsArg(in.Tools)
		if err != nil {
			return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
		}
		tools = parsed
	} else {
		detected, err := integrations.DetectTools(projectRoot)
		if err != nil {
			return Result{}, fmt.Errorf("%w: detect tools: %v", ErrInvalid, err)
		}
		if len(detected) == 0 {
			return Result{}, fmt.Errorf(
				"%w: no tools detected (use --tools with one of: %s)",
				ErrInvalid,
				strings.Join(integrations.SupportedTools(), ","),
			)
		}
		tools = detected
	}

	res := Result{ProjectRoot: projectRoot, Tools: tools}
	if len(tools) == 0 {
		res.NoOp = true
		return res, nil
	}

	for _, toolID := range tools {
		results := integrations.GenerateForTool(projectRoot, toolID, in.Force)
		for _, r := range results {
			res.Items = append(res.Items, r)
			switch {
			case r.Err != nil:
				res.Failed++
			case r.Outcome == integrations.OutcomeCreated:
				res.Created++
			case r.Outcome == integrations.OutcomeUpdated:
				res.Updated++
			case r.Outcome == integrations.OutcomeSkipped:
				res.Skipped++
			}
		}
	}

	return res, nil
}
