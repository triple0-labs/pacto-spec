package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type execOptions struct {
	root     string
	step     string
	note     string
	blocker  string
	evidence string
	dryRun   bool
}

var (
	reExecCheckbox = regexp.MustCompile(`^\s*[-*]\s*\[( |x|X)\]\s*(.+)$`)
	reTaskID       = regexp.MustCompile(`(?i)\bT([0-9]+)\b`)
	reStrictStepID = regexp.MustCompile(`^T[0-9]+$`)
)

func RunExec(args []string) int {
	opts, pos, code, ok := parseExecArgs(args)
	if !ok {
		return code
	}
	if len(pos) != 2 {
		fmt.Fprintln(os.Stderr, "exec requires <state> <slug>")
		return 2
	}

	state := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	if !isValidState(state) {
		fmt.Fprintf(os.Stderr, "invalid state %q (allowed: current|to-implement|done|outdated)\n", state)
		return 2
	}
	if state != "current" {
		fmt.Fprintf(os.Stderr, "exec only supports state %q\n", "current")
		fmt.Fprintf(os.Stderr, "next action: move the plan to current, then retry exec\n")
		fmt.Fprintf(os.Stderr, "trigger: pacto move %s %s current\n", state, slug)
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

	ref, err := resolvePlanRef(plansRoot, state, slug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve plan: %v\n", err)
		return 2
	}

	planPath := ref.PlanDocs[0]
	orig, err := os.ReadFile(planPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read plan doc: %v\n", err)
		return 3
	}
	content := string(orig)

	actions := make([]string, 0, 4)
	updated, act, err := applyExecTaskUpdate(content, opts.step)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec task update: %v\n", err)
		return 2
	}
	if act != "" {
		actions = append(actions, act)
	}

	ts := time.Now().Format("2006-01-02 15:04")
	if note := strings.TrimSpace(opts.note); note != "" {
		updated = appendSectionBullet(updated, "## Execution Notes", fmt.Sprintf("- %s %s", ts, note))
		actions = append(actions, "appended execution note")
	}
	if blocker := strings.TrimSpace(opts.blocker); blocker != "" {
		updated = appendSectionBullet(updated, "## Blockers", fmt.Sprintf("- %s %s", ts, blocker))
		actions = append(actions, "appended blocker")
	}
	if evidence := strings.TrimSpace(opts.evidence); evidence != "" {
		e := evidence
		if !strings.Contains(e, "`") {
			e = "`" + e + "`"
		}
		updated = appendSectionBullet(updated, "## Evidence", fmt.Sprintf("- %s %s", ts, e))
		actions = append(actions, "appended evidence")
	}

	if updated == content {
		fmt.Println("No execution changes to apply.")
		return 0
	}

	if opts.dryRun {
		fmt.Printf("[dry-run] would update: %s\n", planPath)
		for _, a := range actions {
			fmt.Printf("- %s\n", a)
		}
		return 0
	}

	if err := os.WriteFile(planPath, []byte(updated), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write plan doc: %v\n", err)
		return 3
	}

	fmt.Printf("Executed plan: %s/%s\n", state, slug)
	fmt.Printf("~ %s\n", planPath)
	for _, a := range actions {
		fmt.Printf("- %s\n", a)
	}
	return 0
}

func parseExecArgs(args []string) (execOptions, []string, int, bool) {
	opts := execOptions{}
	fs := flag.NewFlagSet("exec", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto exec <current|to-implement|done|outdated> <slug> [--root <path>] [--step <task-id>] [--note <text>] [--blocker <text>] [--evidence <claim>] [--dry-run]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	fs.StringVar(&opts.root, "root", "", "Project root path (auto-discovers when omitted)")
	fs.StringVar(&opts.step, "step", "", "Target task id (e.g. T1)")
	fs.StringVar(&opts.note, "note", "", "Append execution note")
	fs.StringVar(&opts.blocker, "blocker", "", "Append blocker")
	fs.StringVar(&opts.evidence, "evidence", "", "Append evidence reference")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "Show intended changes without writing files")

	normalizedArgs, normErr := normalizeExecArgs(args)
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return execOptions{}, nil, 2, false
	}

	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return execOptions{}, nil, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return execOptions{}, nil, 2, false
	}

	return opts, fs.Args(), 0, true
}

func normalizeExecArgs(args []string) ([]string, error) {
	withValue := map[string]bool{
		"--root": true, "-root": true, "--step": true, "-step": true,
		"--note": true, "-note": true, "--blocker": true, "-blocker": true,
		"--evidence": true, "-evidence": true,
	}
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

func resolvePlansRootForAction(rawRoot string) (string, error) {
	if strings.TrimSpace(rawRoot) != "" {
		abs, err := filepath.Abs(rawRoot)
		if err != nil {
			return "", err
		}
		if resolved, ok := resolvePlanRoot(abs); ok {
			return resolved, nil
		}
		return "", fmt.Errorf("could not resolve plans root from %s (expected .pacto/plans)", abs)
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if resolved, _, ok := resolvePlanRootFrom(cwd); ok {
		return resolved, nil
	}
	return "", fmt.Errorf("could not resolve plans root from %s or parents (expected .pacto/plans)", cwd)
}

func resolvePlanRef(plansRoot, state, slug string) (planRef struct {
	Dir      string
	Readme   string
	PlanDocs []string
}, err error) {
	dir := filepath.Join(plansRoot, state, slug)
	readme := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readme); err != nil {
		return planRef, fmt.Errorf("plan not found: %s/%s", state, slug)
	}

	docs, _ := filepath.Glob(filepath.Join(dir, "PLAN*.md"))
	if len(docs) == 0 {
		docs, _ = filepath.Glob(filepath.Join(dir, "*.md"))
		filtered := make([]string, 0, len(docs))
		for _, d := range docs {
			if strings.EqualFold(filepath.Base(d), "README.md") {
				continue
			}
			filtered = append(filtered, d)
		}
		docs = filtered
	}
	sort.Strings(docs)
	if len(docs) == 0 {
		return planRef, fmt.Errorf("plan has no plan document: %s/%s", state, slug)
	}

	planRef.Dir = dir
	planRef.Readme = readme
	planRef.PlanDocs = docs
	return planRef, nil
}

func applyExecTaskUpdate(content, requestedStep string) (string, string, error) {
	step := normalizeTaskID(requestedStep)
	if requestedStep != "" && !reStrictStepID.MatchString(step) {
		return content, "", fmt.Errorf("invalid --step %q (use T<number>, e.g. T3)", requestedStep)
	}
	lines := strings.Split(content, "\n")
	target := -1
	targetID := ""

	for i, line := range lines {
		m := reExecCheckbox.FindStringSubmatch(line)
		if len(m) != 3 {
			continue
		}
		done := strings.EqualFold(strings.TrimSpace(m[1]), "x")
		text := strings.TrimSpace(m[2])
		tid := extractTaskID(text)

		if step == "" {
			if !done {
				target = i
				targetID = tid
				break
			}
			continue
		}

		if normalizeTaskID(tid) != step {
			continue
		}
		if done {
			return content, "", nil
		}
		target = i
		targetID = tid
		break
	}

	if target < 0 {
		if step != "" {
			return content, "", fmt.Errorf("task %s not found or already completed", step)
		}
		return content, "", nil
	}

	line := lines[target]
	if strings.Contains(line, "[ ]") {
		lines[target] = strings.Replace(line, "[ ]", "[x]", 1)
	} else {
		lines[target] = strings.Replace(line, "[  ]", "[x]", 1)
	}

	if targetID == "" {
		targetID = fmt.Sprintf("line %d", target+1)
	}
	return strings.Join(lines, "\n"), fmt.Sprintf("completed %s", targetID), nil
}

func extractTaskID(text string) string {
	m := reTaskID.FindStringSubmatch(text)
	if len(m) != 2 {
		return ""
	}
	return "T" + m[1]
}

func normalizeTaskID(s string) string {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return ""
	}
	return s
}

func appendSectionBullet(content, heading, bullet string) string {
	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != heading {
			continue
		}
		j := i + 1
		for ; j < len(lines); j++ {
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "## ") {
				break
			}
		}
		block := append([]string{}, lines[i+1:j]...)
		for _, ln := range block {
			if strings.TrimSpace(ln) == strings.TrimSpace(bullet) {
				return content
			}
		}
		if len(block) > 0 && strings.TrimSpace(block[len(block)-1]) != "" {
			block = append(block, "")
		}
		block = append(block, bullet)
		out := make([]string, 0, len(lines)+2)
		out = append(out, lines[:i+1]...)
		out = append(out, block...)
		out = append(out, lines[j:]...)
		return strings.Join(out, "\n")
	}

	trimmed := strings.TrimRight(content, "\n")
	if trimmed == "" {
		return heading + "\n\n" + bullet + "\n"
	}
	return trimmed + "\n\n" + heading + "\n\n" + bullet + "\n"
}
