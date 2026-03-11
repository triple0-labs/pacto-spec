package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/move"
	"pacto/internal/ui"
)

type moveOptions struct {
	root   string
	reason string
	force  bool
}

func RunMove(args []string) int {
	opts, pos, code, ok := parseMoveArgs(args)
	if !ok {
		return code
	}
	return runMove(opts, pos)
}

func runMove(opts moveOptions, pos []string) int {
	if len(pos) != 3 {
		fmt.Fprintln(os.Stderr, "move requires <from-state> <slug> <to-state>")
		return 2
	}

	fromState := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	toState := strings.ToLower(strings.TrimSpace(pos[2]))

	plansRoot, err := resolvePlansRootForAction(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	lang := effectiveLanguage(filepath.Dir(plansRoot))

	readmePath, rootReadme, err := move.MovePlan(plansRoot, fromState, slug, toState, opts.force, opts.reason, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "move error: %v\n", err)
		return 3
	}

	fmt.Println(ui.ActionHeader(move.Tr(lang, "Moved Plan", "Plan movido"), fmt.Sprintf("%s/%s -> %s/%s", fromState, slug, toState, slug)))
	fmt.Println(pathLine("updated", readmePath))
	fmt.Println(pathLine("updated", rootReadme))
	return 0
}

func parseMoveArgs(args []string) (moveOptions, []string, int, bool) {
	opts := moveOptions{}
	fs := flag.NewFlagSet("move", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto move <from-state> <slug> <to-state> [--root <path>] [--reason <text>] [--force]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	fs.StringVar(&opts.root, "root", "", "Project root path (auto-discovers when omitted)")
	fs.StringVar(&opts.reason, "reason", "", "Optional reason to record in plan README")
	fs.BoolVar(&opts.force, "force", false, "Overwrite destination if it exists")

	normalizedArgs, normErr := normalizeMoveArgs(args)
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return moveOptions{}, nil, 2, false
	}

	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return moveOptions{}, nil, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return moveOptions{}, nil, 2, false
	}
	return opts, fs.Args(), 0, true
}

func normalizeMoveArgs(args []string) ([]string, error) {
	withValue := map[string]bool{"--root": true, "-root": true, "--reason": true, "-reason": true}
	return normalizeArgs(args, withValue)
}
