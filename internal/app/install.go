package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/integrations"
)

func RunInstall(args []string) int {
	return runInstallLike("install", args)
}

func RunUpdate(args []string) int {
	return runInstallLike("update", args)
}

func runInstallLike(cmd string, args []string) int {
	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintf(os.Stderr, "  pacto %s [--tools <all|none|csv>] [--force]\n", cmd)
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	toolsArg := fs.String("tools", "", "Tools to configure: all, none, or comma-separated list (codex,cursor,claude,opencode)")
	force := fs.Bool("force", false, "Overwrite unmanaged existing files")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return 2
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve cwd: %v\n", err)
		return 2
	}

	tools := make([]string, 0)
	if strings.TrimSpace(*toolsArg) != "" {
		parsed, err := integrations.ParseToolsArg(*toolsArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 2
		}
		tools = parsed
	} else {
		detected, err := integrations.DetectTools(cwd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "detect tools: %v\n", err)
			return 2
		}
		if len(detected) == 0 {
			fmt.Fprintf(os.Stderr, "no tools detected. Use --tools (%s)\n", strings.Join(integrations.SupportedTools(), ","))
			return 2
		}
		tools = detected
	}

	if len(tools) == 0 {
		fmt.Println("No tools selected; nothing to do.")
		return 0
	}

	fmt.Printf("pacto %s: %s\n", cmd, strings.Join(tools, ", "))
	created := 0
	updated := 0
	skipped := 0
	failed := 0

	for _, toolID := range tools {
		results := integrations.GenerateForTool(cwd, toolID, *force)
		for _, r := range results {
			if r.Err != nil {
				failed++
				fmt.Fprintf(os.Stderr, "error: tool=%s kind=%s workflow=%s: %v\n", r.Tool, r.Kind, r.WorkflowID, r.Err)
				continue
			}
			switch r.Outcome {
			case integrations.OutcomeCreated:
				created++
				fmt.Printf("+ %s\n", r.Path)
			case integrations.OutcomeUpdated:
				updated++
				fmt.Printf("~ %s\n", r.Path)
			case integrations.OutcomeSkipped:
				skipped++
				if r.Reason == "unmanaged_exists" {
					fmt.Fprintf(os.Stderr, "warning: skipped unmanaged file (use --force): %s\n", r.Path)
				} else {
					fmt.Printf("= %s\n", r.Path)
				}
			}
		}
	}

	fmt.Printf("Created: %d  Updated: %d  Skipped: %d  Failed: %d\n", created, updated, skipped, failed)
	if failed > 0 {
		return 3
	}
	return 0
}
