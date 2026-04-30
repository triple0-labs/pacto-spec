// Package execplan is the use-case layer for `pacto exec`. It validates
// inputs, locates the plan document for a given <state>/<slug>, applies the
// requested execution updates (mark a task done, append note/blocker/
// evidence, refresh "Last Modified"), and either previews or persists the
// change. The CLI layer prints headers and bullet lines.
package execplan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

// ErrInvalid wraps validation problems (bad state/slug, unresolvable root,
// missing plan, bad step ref, plan without phase tasks). The CLI maps this
// to exit code 2.
var ErrInvalid = errors.New("invalid input")

var (
	slugRe         = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
	reExecCheckbox = regexp.MustCompile(`^\s*[-*]\s*\[( |x|X)\]\s*(.+)$`)
	rePhaseHeading = regexp.MustCompile(`(?i)^##\s*(phase|fase)\s+([1-9][0-9]*)(?::\s*.*)?$`)
	reStepRef      = regexp.MustCompile(`^([1-9][0-9]*)\.([1-9][0-9]*)\b`)
	reStrictStepID = regexp.MustCompile(`^[1-9][0-9]*\.[1-9][0-9]*$`)
)

// Input is the request for an exec apply.
type Input struct {
	Root     string
	State    string
	Slug     string
	Step     string
	Note     string
	Blocker  string
	Evidence string
	DryRun   bool
	// Lang is the preferred UI language (used as fallback when the plan
	// document has no obvious Spanish markers).
	Lang i18n.Language
	// Now lets tests inject a deterministic timestamp. Zero falls back to
	// time.Now().
	Now time.Time
}

// Result is the aggregated outcome of an exec apply.
type Result struct {
	// PlanPath is the absolute path of the plan document targeted.
	PlanPath string
	// Actions is the human-readable list of mutations that were applied
	// (e.g. "completed 1.2", "appended evidence"), in the order applied.
	Actions []string
	// NoChange is true when the requested updates produced no diff (the
	// caller should print a friendly "no changes" message).
	NoChange bool
	// DryRun mirrors the input flag: true means nothing was written.
	DryRun bool
}

// Apply runs the exec pipeline. It does not print anything.
func Apply(in Input) (Result, error) {
	state := strings.ToLower(strings.TrimSpace(in.State))
	slug := strings.TrimSpace(in.Slug)
	if !isValidState(state) {
		return Result{}, fmt.Errorf("%w: invalid state %q (allowed: current|to-implement|done|outdated)", ErrInvalid, state)
	}
	if state != "current" {
		return Result{}, fmt.Errorf("%w: exec only supports state %q (move the plan to current first: pacto move %s %s current)", ErrInvalid, "current", state, slug)
	}
	if !slugRe.MatchString(slug) {
		return Result{}, fmt.Errorf("%w: invalid slug %q (use lowercase letters, numbers, dashes)", ErrInvalid, slug)
	}

	plansRoot, err := resolvePlansRootForAction(in.Root)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	ref, err := resolvePlanRef(plansRoot, state, slug)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}

	planPath := ref.PlanDocs[0]
	if strings.EqualFold(filepath.Base(ref.ExecDoc), "tasks.md") {
		planPath = ref.ExecDoc
	}
	orig, err := os.ReadFile(planPath)
	if err != nil {
		return Result{}, fmt.Errorf("read plan doc: %w", err)
	}
	content := string(orig)
	docLang := detectPlanDocLanguage(content, in.Lang)

	res := Result{PlanPath: planPath, DryRun: in.DryRun}

	updated, act, err := applyExecTaskUpdate(content, in.Step)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	if act != "" {
		res.Actions = append(res.Actions, act)
	}

	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}
	ts := now.Format("2006-01-02 15:04")
	if note := strings.TrimSpace(in.Note); note != "" {
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("execution_notes", docLang), fmt.Sprintf("- %s %s", ts, note))
		res.Actions = append(res.Actions, "appended execution note")
	}
	if blocker := strings.TrimSpace(in.Blocker); blocker != "" {
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("blockers", docLang), fmt.Sprintf("- %s %s", ts, blocker))
		res.Actions = append(res.Actions, "appended blocker")
	}
	if evidence := strings.TrimSpace(in.Evidence); evidence != "" {
		e := evidence
		if !strings.Contains(e, "`") {
			e = "`" + e + "`"
		}
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("evidence", docLang), fmt.Sprintf("- %s %s", ts, e))
		res.Actions = append(res.Actions, "appended evidence")
	}

	if updated == content {
		res.NoChange = true
		return res, nil
	}
	updated = upsertPlanLastModified(updated, ts, docLang)

	if in.DryRun {
		return res, nil
	}
	if err := os.WriteFile(planPath, []byte(updated), 0o664); err != nil {
		return Result{}, fmt.Errorf("write plan doc: %w", err)
	}
	return res, nil
}

func isValidState(state string) bool {
	switch state {
	case "current", "to-implement", "done", "outdated":
		return true
	}
	return false
}

func resolvePlansRootForAction(rawRoot string) (string, error) {
	if strings.TrimSpace(rawRoot) != "" {
		abs, err := filepath.Abs(rawRoot)
		if err != nil {
			return "", err
		}
		if resolved, ok := workspace.ResolvePlanRoot(abs); ok {
			return resolved, nil
		}
		return "", fmt.Errorf("could not resolve plans root from %s (expected .pacto/plans)", abs)
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if resolved, _, ok := workspace.ResolvePlanRootFrom(cwd); ok {
		return resolved, nil
	}
	return "", fmt.Errorf("could not resolve plans root from %s or parents (expected .pacto/plans)", cwd)
}

type planRef struct {
	Dir      string
	Readme   string
	ExecDoc  string
	PlanDocs []string
}

func resolvePlanRef(plansRoot, state, slug string) (planRef, error) {
	var ref planRef
	dir := filepath.Join(plansRoot, state, slug)
	readme := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readme); err != nil {
		return ref, fmt.Errorf("plan not found: %s/%s", state, slug)
	}

	tasksPath := filepath.Join(dir, "tasks.md")
	if _, err := os.Stat(tasksPath); err == nil {
		ref.Dir = dir
		ref.Readme = readme
		ref.ExecDoc = tasksPath
		ref.PlanDocs = []string{tasksPath}
		return ref, nil
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
		return ref, fmt.Errorf("plan has no plan document: %s/%s", state, slug)
	}

	ref.Dir = dir
	ref.Readme = readme
	ref.ExecDoc = docs[0]
	ref.PlanDocs = docs
	return ref, nil
}

func applyExecTaskUpdate(content, requestedStep string) (string, string, error) {
	step := strings.TrimSpace(requestedStep)
	if strings.HasPrefix(strings.ToUpper(step), "T") {
		return content, "", fmt.Errorf("legacy --step %q is no longer supported (use <phase>.<task>, e.g. 1.2)", requestedStep)
	}
	if requestedStep != "" && !reStrictStepID.MatchString(step) {
		return content, "", fmt.Errorf("invalid --step %q (use <phase>.<task>, e.g. 1.2)", requestedStep)
	}
	lines := strings.Split(content, "\n")
	type candidate struct {
		line  int
		ref   string
		phase int
		task  int
		done  bool
	}
	candidates := make([]candidate, 0, 16)
	currentPhase := 0
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if m := rePhaseHeading.FindStringSubmatch(t); len(m) == 3 {
			currentPhase = parsePosInt(m[2])
			continue
		}
		m := reExecCheckbox.FindStringSubmatch(line)
		if len(m) != 3 || currentPhase == 0 {
			continue
		}
		done := strings.EqualFold(strings.TrimSpace(m[1]), "x")
		taskText := strings.TrimSpace(m[2])
		phase, task, ok := extractStepRef(taskText)
		if !ok || phase != currentPhase {
			continue
		}
		candidates = append(candidates, candidate{
			line:  i,
			ref:   fmt.Sprintf("%d.%d", phase, task),
			phase: phase,
			task:  task,
			done:  done,
		})
	}

	if len(candidates) == 0 {
		return content, "", fmt.Errorf("no phase tasks found (expected '- [ ] 1.1 ...' under '## Phase N' or '## Fase N' headings)")
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].phase == candidates[j].phase {
			if candidates[i].task == candidates[j].task {
				return candidates[i].line < candidates[j].line
			}
			return candidates[i].task < candidates[j].task
		}
		return candidates[i].phase < candidates[j].phase
	})

	target := -1
	targetID := ""
	if step != "" {
		for _, c := range candidates {
			if c.ref != step {
				continue
			}
			if c.done {
				return content, "", nil
			}
			target = c.line
			targetID = c.ref
			break
		}
	} else {
		for _, c := range candidates {
			if !c.done {
				target = c.line
				targetID = c.ref
				break
			}
		}
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

func extractStepRef(text string) (int, int, bool) {
	m := reStepRef.FindStringSubmatch(strings.TrimSpace(text))
	if len(m) != 3 {
		return 0, 0, false
	}
	phase := parsePosInt(m[1])
	task := parsePosInt(m[2])
	if phase <= 0 || task <= 0 {
		return 0, 0, false
	}
	return phase, task, true
}

func parsePosInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil || n < 1 {
		return 0
	}
	return n
}

func appendSectionBulletLocalized(content string, headings []string, bullet string) string {
	if len(headings) == 0 {
		return content
	}
	primary := strings.TrimSpace(headings[0])
	allowed := map[string]struct{}{}
	for _, h := range headings {
		h = strings.ToLower(strings.TrimSpace(h))
		if h == "" {
			continue
		}
		allowed[h] = struct{}{}
	}

	lines := strings.Split(content, "\n")
	for i := 0; i < len(lines); i++ {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(lines[i]))]; !ok {
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
		return primary + "\n\n" + bullet + "\n"
	}
	return trimmed + "\n\n" + primary + "\n\n" + bullet + "\n"
}

func upsertPlanLastModified(content, timestamp string, lang i18n.Language) string {
	lines := strings.Split(content, "\n")
	boldLabel := trLang(lang, "**Last Modified:** ", "**Última Modificación:** ")
	bulletLabel := trLang(lang, "- Last Modified: ", "- Última Modificación: ")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "**Last Modified:**") || strings.HasPrefix(trimmed, "**Última Modificación:**") {
			lines[i] = boldLabel + timestamp + "  "
			return strings.Join(lines, "\n")
		}
		if strings.HasPrefix(trimmed, "- Last Modified:") || strings.HasPrefix(trimmed, "- Última Modificación:") || strings.HasPrefix(trimmed, "- Ultima Modificacion:") {
			lines[i] = bulletLabel + timestamp
			return strings.Join(lines, "\n")
		}
	}

	insertAt := -1
	preferBullet := true
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "**Date:**") || strings.HasPrefix(trimmed, "**Fecha:**") {
			insertAt = i + 1
			preferBullet = false
			break
		}
		if strings.HasPrefix(trimmed, "- Created:") || strings.HasPrefix(trimmed, "- Creado:") {
			insertAt = i + 1
			preferBullet = true
			break
		}
	}
	if insertAt < 0 {
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "# ") {
				insertAt = i + 1
				preferBullet = false
				break
			}
		}
	}
	if insertAt < 0 {
		insertAt = 0
	}

	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:insertAt]...)
	if preferBullet {
		out = append(out, bulletLabel+timestamp)
	} else {
		out = append(out, boldLabel+timestamp+"  ")
	}
	out = append(out, lines[insertAt:]...)
	return strings.Join(out, "\n")
}

func trLang(lang i18n.Language, en, es string) string {
	if lang == i18n.Spanish {
		return es
	}
	return en
}

func detectPlanDocLanguage(content string, fallback i18n.Language) i18n.Language {
	l := strings.ToLower(content)
	scoreES := 0
	for _, tok := range []string{
		"## fase ", "## plan de implementación por fases", "## plan de implementacion por fases",
		"## metadatos de ejecución", "## metadatos de ejecucion", "## evidencia",
		"## bloqueadores", "## siguientes pasos", "última modificación", "ultima modificacion",
	} {
		if strings.Contains(l, tok) {
			scoreES++
		}
	}
	if scoreES >= 2 {
		return i18n.Spanish
	}
	return fallback
}

func localizedSectionHeadings(section string, lang i18n.Language) []string {
	switch section {
	case "execution_notes":
		if lang == i18n.Spanish {
			return []string{"## Notas de Ejecución", "## Notas de Ejecucion", "## Execution Notes"}
		}
		return []string{"## Execution Notes", "## Notas de Ejecución", "## Notas de Ejecucion"}
	case "evidence":
		if lang == i18n.Spanish {
			return []string{"## Evidencia", "## Evidence"}
		}
		return []string{"## Evidence", "## Evidencia"}
	case "blockers":
		if lang == i18n.Spanish {
			return []string{"## Bloqueadores", "## Blockers"}
		}
		return []string{"## Blockers", "## Bloqueadores"}
	default:
		return nil
	}
}
