package execmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"pacto/internal/i18n"
	"pacto/internal/ui"
)

type Options struct {
	Root     string
	Step     string
	Note     string
	Blocker  string
	Evidence string
	DryRun   bool
}

var (
	reExecCheckbox = regexp.MustCompile(`^\s*[-*]\s*\[( |x|X)\]\s*(.+)$`)
	rePhaseHeading = regexp.MustCompile(`(?i)^##\s*(phase|fase)\s+([1-9][0-9]*)(?::\s*.*)?$`)
	reStepRef      = regexp.MustCompile(`^([1-9][0-9]*)\.([1-9][0-9]*)\b`)
	reStrictStepID = regexp.MustCompile(`^[1-9][0-9]*\.[1-9][0-9]*$`)
)

func Run(opts Options, pos []string) int {
	lang := effectiveLanguage(opts.Root)
	if len(pos) != 2 {
		fmt.Fprintln(os.Stderr, tr(lang, "exec requires <state> <slug>", "exec requiere <state> <slug>"))
		return 2
	}

	state := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	if !isValidState(state) {
		fmt.Fprintf(os.Stderr, tr(lang, "invalid state %q (allowed: current|to-implement|done|outdated)\n", "estado inválido %q (permitidos: current|to-implement|done|outdated)\n"), state)
		return 2
	}
	if state != "current" {
		fmt.Fprintf(os.Stderr, tr(lang, "exec only supports state %q\n", "exec solo soporta el estado %q\n"), "current")
		fmt.Fprintln(os.Stderr, tr(lang, "next action: move the plan to current, then retry exec", "siguiente acción: mueve el plan a current y vuelve a intentar exec"))
		fmt.Fprintf(os.Stderr, tr(lang, "trigger: pacto move %s %s current\n", "comando: pacto move %s %s current\n"), state, slug)
		return 2
	}
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return 2
	}

	plansRoot, err := resolvePlansRootForAction(opts.Root)
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
	if strings.EqualFold(filepath.Base(ref.ExecDoc), "tasks.md") {
		planPath = ref.ExecDoc
	}
	orig, err := os.ReadFile(planPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read plan doc: %v\n", err)
		return 3
	}
	content := string(orig)
	docLang := detectPlanDocLanguage(content, lang)

	actions := make([]string, 0, 4)
	updated, act, err := applyExecTaskUpdate(content, opts.Step)
	if err != nil {
		fmt.Fprintf(os.Stderr, "exec task update: %v\n", err)
		return 2
	}
	if act != "" {
		actions = append(actions, act)
	}

	ts := time.Now().Format("2006-01-02 15:04")
	if note := strings.TrimSpace(opts.Note); note != "" {
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("execution_notes", docLang), fmt.Sprintf("- %s %s", ts, note))
		actions = append(actions, "appended execution note")
	}
	if blocker := strings.TrimSpace(opts.Blocker); blocker != "" {
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("blockers", docLang), fmt.Sprintf("- %s %s", ts, blocker))
		actions = append(actions, "appended blocker")
	}
	if evidence := strings.TrimSpace(opts.Evidence); evidence != "" {
		e := evidence
		if !strings.Contains(e, "`") {
			e = "`" + e + "`"
		}
		updated = appendSectionBulletLocalized(updated, localizedSectionHeadings("evidence", docLang), fmt.Sprintf("- %s %s", ts, e))
		actions = append(actions, "appended evidence")
	}

	if updated == content {
		fmt.Println(ui.Dim(tr(lang, "No execution changes to apply.", "No hay cambios de ejecución para aplicar.")))
		return 0
	}
	updated = upsertPlanLastModified(updated, ts, docLang)

	if opts.DryRun {
		fmt.Println(ui.ActionHeader(tr(lang, "Dry Run", "Simulación"), tr(lang, "execution update", "actualización de ejecución")))
		fmt.Println(pathLine("updated", planPath))
		for _, a := range actions {
			fmt.Println(ui.Bullet(a))
		}
		return 0
	}

	if err := os.WriteFile(planPath, []byte(updated), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write plan doc: %v\n", err)
		return 3
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Executed Plan", "Plan ejecutado"), state+"/"+slug))
	fmt.Println(pathLine("updated", planPath))
	for _, a := range actions {
		fmt.Println(ui.Bullet(a))
	}
	return 0
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
	ExecDoc  string
	PlanDocs []string
}, err error) {
	dir := filepath.Join(plansRoot, state, slug)
	readme := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readme); err != nil {
		return planRef, fmt.Errorf("plan not found: %s/%s", state, slug)
	}

	tasksPath := filepath.Join(dir, "tasks.md")
	if _, err := os.Stat(tasksPath); err == nil {
		planRef.Dir = dir
		planRef.Readme = readme
		planRef.ExecDoc = tasksPath
		planRef.PlanDocs = []string{tasksPath}
		return planRef, nil
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
	planRef.ExecDoc = docs[0]
	planRef.PlanDocs = docs
	return planRef, nil
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
	boldLabel := tr(lang, "**Last Modified:** ", "**Última Modificación:** ")
	bulletLabel := tr(lang, "- Last Modified: ", "- Última Modificación: ")
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
