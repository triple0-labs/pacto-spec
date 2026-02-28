package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	if len(pos) != 3 {
		fmt.Fprintln(os.Stderr, "move requires <from-state> <slug> <to-state>")
		return 2
	}

	fromState := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	toState := strings.ToLower(strings.TrimSpace(pos[2]))
	if !isValidState(fromState) || !isValidState(toState) {
		fmt.Fprintf(os.Stderr, "invalid state transition %q -> %q (allowed: current|to-implement|done|outdated)\n", fromState, toState)
		return 2
	}
	if fromState == toState {
		fmt.Fprintln(os.Stderr, "source and destination states are the same")
		return 2
	}
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return 2
	}

	plansRoot, err := resolvePlansRootForAction(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	srcDir := filepath.Join(plansRoot, fromState, slug)
	dstDir := filepath.Join(plansRoot, toState, slug)
	if _, err := os.Stat(filepath.Join(srcDir, "README.md")); err != nil {
		fmt.Fprintf(os.Stderr, "source plan not found: %s/%s\n", fromState, slug)
		return 2
	}
	if _, err := os.Stat(dstDir); err == nil {
		if !opts.force {
			fmt.Fprintf(os.Stderr, "destination already exists: %s (use --force to overwrite)\n", dstDir)
			return 2
		}
		if err := os.RemoveAll(dstDir); err != nil {
			fmt.Fprintf(os.Stderr, "remove destination: %v\n", err)
			return 3
		}
	}

	if err := os.Rename(srcDir, dstDir); err != nil {
		fmt.Fprintf(os.Stderr, "move plan directory: %v\n", err)
		return 3
	}

	readmePath := filepath.Join(dstDir, "README.md")
	if err := rewritePlanReadmeStatus(readmePath, toState, fromState, opts.reason); err != nil {
		fmt.Fprintf(os.Stderr, "update moved README: %v\n", err)
		return 3
	}

	rootReadme := filepath.Join(plansRoot, "README.md")
	b, err := os.ReadFile(rootReadme)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read root README: %v\n", err)
		return 3
	}
	text := string(b)
	counts, err := countPlans(plansRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "count plans: %v\n", err)
		return 3
	}
	text = updateCountsTable(text, counts)
	text = removePlanLinkFromSection(text, fromState, slug)
	title := readPlanTitle(readmePath)
	if title == "" {
		title = slugToTitle(slug)
	}
	text2, err := upsertLinkInSection(text, toState, title, fmt.Sprintf("./%s/%s/", toState, slug))
	if err != nil {
		fmt.Fprintf(os.Stderr, "update root section: %v\n", err)
		return 3
	}
	text = updateLastUpdate(text2, time.Now().Format("2006-01-02"))
	if err := os.WriteFile(rootReadme, []byte(text), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write root README: %v\n", err)
		return 3
	}

	fmt.Printf("Moved plan: %s/%s -> %s/%s\n", fromState, slug, toState, slug)
	fmt.Printf("~ %s\n", readmePath)
	fmt.Printf("~ %s\n", rootReadme)
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
	flags := make([]string, 0, len(args))
	pos := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") || strings.HasPrefix(a, "-") {
			if withValue[a] {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag %s expects a value", a)
				}
				flags = append(flags, a, args[i+1])
				i++
				continue
			}
			flags = append(flags, a)
			continue
		}
		pos = append(pos, a)
	}
	return append(flags, pos...), nil
}

func rewritePlanReadmeStatus(path, toState, fromState, reason string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(b)
	lines := strings.Split(text, "\n")
	newStatus := stateStatusLabel(toState)
	updated := false
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "**Status:**") {
			lines[i] = "**Status:** " + newStatus + "  "
			updated = true
			break
		}
	}
	if !updated {
		lines = append([]string{"**Status:** " + newStatus + "  "}, lines...)
	}
	text = strings.Join(lines, "\n")
	if strings.TrimSpace(reason) != "" {
		note := fmt.Sprintf("- %s moved from `%s` to `%s`: %s", time.Now().Format("2006-01-02 15:04"), fromState, toState, strings.TrimSpace(reason))
		text = appendSectionBullet(text, "## Move History", note)
	}
	return os.WriteFile(path, []byte(text), 0o664)
}

func stateStatusLabel(state string) string {
	return map[string]string{
		"current":      "In Progress (Current)",
		"to-implement": "Pending (To Implement)",
		"done":         "Completed (Done)",
		"outdated":     "Outdated (Outdated)",
	}[state]
}

func readPlanTitle(readmePath string) string {
	b, err := os.ReadFile(readmePath)
	if err != nil {
		return ""
	}
	for _, ln := range strings.Split(string(b), "\n") {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(t, "# "))
		}
	}
	return ""
}

func removePlanLinkFromSection(text, state, slug string) string {
	lines := strings.Split(text, "\n")
	start, end, sep := findCanonicalSection(lines, state)
	if start < 0 {
		start, end, sep = findSectionByStateLink(lines, state)
	}
	if start < 0 {
		return text
	}

	needle := fmt.Sprintf("./%s/%s/", state, slug)
	sec := append([]string{}, lines[start+1:sep]...)
	newSec := make([]string, 0, len(sec))
	bulletCount := 0
	for _, ln := range sec {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "- [") {
			if strings.Contains(t, needle) {
				continue
			}
			bulletCount++
		}
		newSec = append(newSec, ln)
	}
	if bulletCount == 0 {
		clean := make([]string, 0, len(newSec)+1)
		for _, ln := range newSec {
			t := strings.TrimSpace(ln)
			if strings.HasPrefix(t, "- [") || strings.HasPrefix(strings.ToLower(t), "_no plans") || strings.HasPrefix(strings.ToLower(t), "_no hay planes") {
				continue
			}
			clean = append(clean, ln)
		}
		if len(clean) > 0 && strings.TrimSpace(clean[len(clean)-1]) != "" {
			clean = append(clean, "")
		}
		clean = append(clean, "_No plans._")
		newSec = clean
	}

	out := make([]string, 0, len(lines))
	out = append(out, lines[:start+1]...)
	out = append(out, newSec...)
	out = append(out, lines[sep:]...)
	if end > sep {
		_ = end
	}
	return strings.Join(out, "\n")
}
