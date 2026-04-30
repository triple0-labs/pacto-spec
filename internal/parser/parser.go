package parser

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pacto/internal/domain/plan"
	"pacto/internal/planfmt"
)

type ParsedDeltaChange struct {
	Op   string
	Path string
}

type ParsedDelta struct {
	ID        string
	Date      *time.Time
	Type      string
	Status    string
	NextDelta string
	Changes   []ParsedDeltaChange
}

type ParsedPlan struct {
	Ref                 plan.PlanRef
	RawText             string
	DeclaredStatus      string
	Phases              []plan.Phase
	LastUpdatedAt       *time.Time
	Tasks               []plan.Task
	BlockerHints        []string
	NextActions         []string
	HasCheckpoint       bool
	HasEvidence         bool
	HasStructuredDeltas bool
	Deltas              []ParsedDelta
	LatestDeltaTime     *time.Time
	ParseWarnings       []string
	ParseError          string
}

var (
	reDeclaredStatus = regexp.MustCompile(`(?i)^[-*]?\s*(?:\*\*)?(estado|status)(?::)?(?:\*\*)?:\s*(.+)$`)
	reCheckbox       = regexp.MustCompile(`^\s*[-*]\s*\[( |x|X)\]\s*(.+)$`)
	reTaskNumbered   = regexp.MustCompile(`^\s*\d+\.\s+(.+)$`)
	rePhaseRow       = regexp.MustCompile(`(?i)^\|\s*((phase|fase)\s*[^|]+)\|([^|]+)\|([^|]+)\|\s*([0-9]{1,3})%\s*\|`)
	rePhaseHeading   = regexp.MustCompile(`(?i)^##\s*(phase|fase)\s+([1-9][0-9]*)(?::\s*(.*))?$`)
	reStepRef        = regexp.MustCompile(`^([1-9][0-9]*)\.([1-9][0-9]*)\b`)
	reAnyPercent     = regexp.MustCompile(`(?i)(progreso total|progress)[:\s*]*([0-9]{1,3})%`)
	reDateTime       = regexp.MustCompile(`(20[0-9]{2}-[0-9]{2}-[0-9]{2})(?:[ T]([0-9]{2}:[0-9]{2}))?`)
	reDeltaHeader    = regexp.MustCompile(`(?i)^###\s*delta\s+(.+)$`)
	reDeltaID        = regexp.MustCompile(`^D-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{2}$`)
)

func ParsePlan(ref plan.PlanRef, mode string) (ParsedPlan, error) {
	p := ParsedPlan{Ref: ref}
	text, err := readPlanText(ref)
	if err != nil {
		return p, err
	}
	p.RawText = text
	p.LastUpdatedAt = latestPlanUpdate(ref)
	lines := strings.Split(text, "\n")

	deltas, deltaWarnings, hasDeltaSection, deltaErr := parseStructuredDeltas(lines, "compat")
	if len(deltas) > 0 {
		p.HasStructuredDeltas = true
		p.Deltas = deltas
		for _, d := range deltas {
			if d.Date == nil {
				continue
			}
			if p.LatestDeltaTime == nil || d.Date.After(*p.LatestDeltaTime) {
				p.LatestDeltaTime = d.Date
			}
		}
	}
	p.ParseWarnings = append(p.ParseWarnings, deltaWarnings...)
	if deltaErr != nil {
		return p, deltaErr
	}

	currentPhase := 0
	var legacyLatestDelta *time.Time
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if m := rePhaseHeading.FindStringSubmatch(t); len(m) >= 3 {
			currentPhase, _ = strconv.Atoi(strings.TrimSpace(m[2]))
		}

		if m := reDeclaredStatus.FindStringSubmatch(t); len(m) == 3 && p.DeclaredStatus == "" {
			p.DeclaredStatus = cleanStatusValue(m[2])
		}
		if p.DeclaredStatus == "" {
			lt := strings.ToLower(t)
			if idx := strings.Index(lt, "estado:"); idx >= 0 {
				p.DeclaredStatus = cleanStatusValue(t[idx+len("estado:"):])
			} else if idx := strings.Index(lt, "status:"); idx >= 0 {
				p.DeclaredStatus = cleanStatusValue(t[idx+len("status:"):])
			}
		}

		if m := rePhaseRow.FindStringSubmatch(t); len(m) == 6 {
			prog, _ := strconv.Atoi(strings.TrimSpace(m[5]))
			p.Phases = append(p.Phases, plan.Phase{Name: strings.TrimSpace(m[1]), RawState: strings.TrimSpace(m[4]), Progress: prog})
		}

		if m := reCheckbox.FindStringSubmatch(line); len(m) == 3 {
			done := strings.EqualFold(strings.TrimSpace(m[1]), "x")
			text := strings.TrimSpace(m[2])
			task := plan.Task{Text: text, Completed: done, LikelyBlk: looksBlocked(text)}
			if currentPhase > 0 {
				if phase, number, ok := extractStepRef(text); ok && phase == currentPhase {
					task.StepRef = fmt.Sprintf("%d.%d", phase, number)
					task.Phase = phase
					task.Number = number
				}
			}
			p.Tasks = append(p.Tasks, task)
		}

		lt := strings.ToLower(t)
		if strings.HasPrefix(lt, "**checkpoint") || strings.HasPrefix(lt, "checkpoint") {
			p.HasCheckpoint = true
		}
		if strings.Contains(lt, "evidencia") || strings.Contains(lt, "evidence") || strings.Contains(lt, "smoke") || strings.Contains(lt, "validación") || strings.Contains(lt, "validacion") || strings.Contains(lt, "validation") {
			p.HasEvidence = true
		}
		if strings.Contains(lt, "delta") || strings.Contains(lt, "checkpoint") {
			if dt := parseDateTime(t); dt != nil {
				if legacyLatestDelta == nil || dt.After(*legacyLatestDelta) {
					legacyLatestDelta = dt
				}
			}
		}
	}

	if p.LatestDeltaTime == nil && !hasDeltaSection && legacyLatestDelta != nil {
		p.LatestDeltaTime = legacyLatestDelta
	}

	extractBlockers(lines, &p)
	extractNextActions(lines, &p)
	if len(p.Phases) == 0 {
		if pct := extractTotalProgress(text); pct >= 0 {
			p.Phases = append(p.Phases, plan.Phase{Name: "total", RawState: "derived", Progress: pct})
		}
	}

	activeStrict := mode == "strict" && (p.Ref.State == "current" || p.Ref.State == "to-implement")
	if len(p.Phases) == 0 {
		if activeStrict {
			p.ParseWarnings = append(p.ParseWarnings, "missing structured progress source")
		} else if mode == "strict" {
			p.ParseWarnings = append(p.ParseWarnings, "missing structured progress source")
		}
	}

	structureIssues := planfmt.Validate(text)
	for _, issue := range structureIssues {
		p.ParseWarnings = appendUnique(p.ParseWarnings, "plan_structure: "+issue.Code+" - "+issue.Message)
	}
	if activeStrict && len(structureIssues) > 0 {
		first := structureIssues[0]
		return p, fmt.Errorf("plan structure violation (%s): %s", first.Code, first.Message)
	}
	return p, nil
}

func parseStructuredDeltas(lines []string, mode string) ([]ParsedDelta, []string, bool, error) {
	start := findDeltaHistorySection(lines)
	if start < 0 {
		return nil, nil, false, nil
	}

	warnings := make([]string, 0)
	deltas := make([]ParsedDelta, 0)
	var curr *ParsedDelta
	insideChanges := false
	recordedIssues := map[string]struct{}{}

	addIssue := func(msg string) error {
		if _, ok := recordedIssues[msg]; ok {
			return nil
		}
		recordedIssues[msg] = struct{}{}
		if mode == "strict" {
			return fmt.Errorf("invalid structured delta section: %s", msg)
		}
		warnings = append(warnings, msg)
		return nil
	}

	flushCurrent := func() error {
		if curr == nil {
			return nil
		}
		for _, msg := range validateDelta(curr) {
			if err := addIssue(msg); err != nil {
				return err
			}
		}
		deltas = append(deltas, *curr)
		curr = nil
		insideChanges = false
		return nil
	}

	for i := start + 1; i < len(lines); i++ {
		t := strings.TrimSpace(lines[i])
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "## ") {
			break
		}
		if m := reDeltaHeader.FindStringSubmatch(t); len(m) == 2 {
			if err := flushCurrent(); err != nil {
				return nil, warnings, true, err
			}
			id := strings.TrimSpace(m[1])
			if !reDeltaID.MatchString(id) {
				if err := addIssue("invalid delta id: " + id); err != nil {
					return nil, warnings, true, err
				}
			}
			curr = &ParsedDelta{ID: id, Changes: make([]ParsedDeltaChange, 0)}
			insideChanges = false
			continue
		}
		if curr == nil {
			continue
		}

		if key, value, ok := parseDeltaField(t); ok {
			nowInsideChanges, issue := applyDeltaField(curr, key, value)
			insideChanges = nowInsideChanges
			if issue != "" {
				if err := addIssue(issue); err != nil {
					return nil, warnings, true, err
				}
			}
			continue
		}

		if insideChanges {
			if change, ok := parseDeltaChangeLine(t); ok {
				curr.Changes = append(curr.Changes, change)
				continue
			}
			if err := addIssue("delta " + curr.ID + " has invalid change line: " + t); err != nil {
				return nil, warnings, true, err
			}
		}
	}

	if err := flushCurrent(); err != nil {
		return nil, warnings, true, err
	}
	return deltas, warnings, true, nil
}

// findDeltaHistorySection returns the index of the "## Delta History" /
// "## Historial de Deltas" heading, or -1 if absent.
func findDeltaHistorySection(lines []string) int {
	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "##") {
			continue
		}
		h := normalizeHeading(strings.TrimSpace(line))
		if h == "delta history" || h == "historial de deltas" {
			return i
		}
	}
	return -1
}

// validateDelta returns the structural issues for a fully-collected delta
// (in the order they should be reported). An empty slice means valid.
// Status is normalized in place when recognized.
func validateDelta(d *ParsedDelta) []string {
	issues := make([]string, 0, 2)
	if d.ID == "" {
		issues = append(issues, "missing delta id in heading")
	}
	if d.Date == nil {
		issues = append(issues, "delta "+d.ID+" missing Date")
	}
	if d.Status != "" {
		s := strings.ToLower(strings.TrimSpace(d.Status))
		switch s {
		case "applied", "partial", "reverted":
			d.Status = s
		default:
			issues = append(issues, "delta "+d.ID+" has invalid Status: "+d.Status)
		}
	}
	return issues
}

// applyDeltaField writes a parsed (key,value) pair into the current delta
// and signals whether subsequent unstructured lines should be parsed as
// change entries. Returns an issue message when a value is rejected.
func applyDeltaField(curr *ParsedDelta, key, value string) (insideChanges bool, issue string) {
	switch key {
	case "date":
		dt, err := parseStructuredDate(value)
		if err != nil {
			return false, "delta " + curr.ID + " has invalid Date: " + value
		}
		curr.Date = dt
	case "type":
		curr.Type = strings.TrimSpace(value)
	case "status":
		curr.Status = strings.TrimSpace(value)
	case "next delta":
		curr.NextDelta = strings.TrimSpace(value)
	case "changes":
		return true, ""
	}
	return false, ""
}

func parseStructuredDate(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty date")
	}
	t, err := time.Parse("2006-01-02 15:04", value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func parseDeltaChangeLine(line string) (ParsedDeltaChange, bool) {
	v := strings.TrimSpace(line)
	for strings.HasPrefix(v, "- ") || strings.HasPrefix(v, "* ") {
		v = strings.TrimSpace(v[2:])
	}
	v = strings.Trim(v, "`")
	if v == "" {
		return ParsedDeltaChange{}, false
	}
	op := string(v[0])
	if op != "+" && op != "~" && op != "-" {
		return ParsedDeltaChange{}, false
	}
	path := strings.TrimSpace(v[1:])
	if path == "" {
		return ParsedDeltaChange{}, false
	}
	return ParsedDeltaChange{Op: op, Path: path}, true
}

func parseDeltaField(line string) (string, string, bool) {
	v := strings.TrimSpace(line)
	for strings.HasPrefix(v, "- ") || strings.HasPrefix(v, "* ") {
		v = strings.TrimSpace(v[2:])
	}
	if strings.HasPrefix(v, "**") {
		if idx := strings.Index(v[2:], "**"); idx >= 0 {
			end := 2 + idx
			label := strings.TrimSpace(v[2:end])
			label = strings.TrimSuffix(label, ":")
			rest := strings.TrimSpace(v[end+2:])
			rest = strings.TrimPrefix(rest, ":")
			rest = strings.TrimSpace(rest)
			if key, ok := canonicalDeltaField(label); ok {
				return key, rest, true
			}
		}
	}
	if idx := strings.Index(v, ":"); idx >= 0 {
		label := strings.TrimSpace(v[:idx])
		rest := strings.TrimSpace(v[idx+1:])
		if key, ok := canonicalDeltaField(label); ok {
			return key, rest, true
		}
	}
	return "", "", false
}

func canonicalDeltaField(label string) (string, bool) {
	norm := strings.ToLower(strings.TrimSpace(label))
	norm = strings.Trim(norm, "*")
	norm = strings.Join(strings.Fields(norm), " ")
	aliases := map[string]string{
		"date":            "date",
		"fecha":           "date",
		"author":          "author",
		"autor":           "author",
		"scope":           "scope",
		"alcance":         "scope",
		"type":            "type",
		"tipo":            "type",
		"status":          "status",
		"estado":          "status",
		"changes":         "changes",
		"cambios":         "changes",
		"validation":      "validation",
		"validacion":      "validation",
		"validación":      "validation",
		"risk":            "risk",
		"riesgo":          "risk",
		"rollback":        "rollback",
		"next delta":      "next delta",
		"siguiente delta": "next delta",
	}
	key, ok := aliases[norm]
	return key, ok
}

func normalizeHeading(line string) string {
	t := strings.TrimSpace(line)
	t = strings.TrimLeft(t, "#")
	t = strings.TrimSpace(t)
	t = strings.Trim(t, "*")
	t = strings.ToLower(t)
	t = strings.Join(strings.Fields(t), " ")
	return t
}

func extractStepRef(text string) (int, int, bool) {
	m := reStepRef.FindStringSubmatch(strings.TrimSpace(text))
	if len(m) != 3 {
		return 0, 0, false
	}
	phase, err := strconv.Atoi(strings.TrimSpace(m[1]))
	if err != nil || phase < 1 {
		return 0, 0, false
	}
	number, err := strconv.Atoi(strings.TrimSpace(m[2]))
	if err != nil || number < 1 {
		return 0, 0, false
	}
	return phase, number, true
}

func readPlanText(ref plan.PlanRef) (string, error) {
	parts := make([]string, 0, len(ref.PlanDocs)+1)
	readme, err := os.ReadFile(ref.Readme)
	if err != nil {
		return "", err
	}
	parts = append(parts, string(readme))
	for _, doc := range ref.PlanDocs {
		b, err := os.ReadFile(doc)
		if err != nil {
			continue
		}
		parts = append(parts, string(b))
	}
	return strings.Join(parts, "\n\n"), nil
}

func latestPlanUpdate(ref plan.PlanRef) *time.Time {
	paths := make([]string, 0, len(ref.PlanDocs)+1)
	paths = append(paths, ref.Readme)
	paths = append(paths, ref.PlanDocs...)

	var latest *time.Time
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		mod := info.ModTime().UTC()
		if latest == nil || mod.After(*latest) {
			ts := mod
			latest = &ts
		}
	}
	return latest
}

func looksBlocked(text string) bool {
	t := strings.ToLower(text)
	keys := []string{"blocked", "blocked by", "waiting on", "waiting for", "bloqueado", "bloqueado por", "pendiente crítico", "pendiente critico", "en espera"}
	for _, k := range keys {
		if strings.Contains(t, k) {
			return true
		}
	}
	return false
}

func extractBlockers(lines []string, p *ParsedPlan) {
	collect := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if isBlockersHeading(t) {
			collect = true
			continue
		}
		if collect {
			if strings.HasPrefix(t, "## ") || strings.HasPrefix(t, "### ") {
				collect = false
				continue
			}
			if blocker, ok := parseBlockerLine(line); ok {
				p.BlockerHints = appendUnique(p.BlockerHints, blocker)
			}
		}
	}
}

func isBlockersHeading(line string) bool {
	switch normalizeHeading(line) {
	case "blockers", "bloqueadores":
		return true
	default:
		return false
	}
}

func parseBlockerLine(line string) (string, bool) {
	t := strings.TrimSpace(line)
	if t == "" {
		return "", false
	}

	text := t
	switch {
	case strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* "):
		text = strings.TrimSpace(t[2:])
	case len(reCheckbox.FindStringSubmatch(line)) == 3:
		text = strings.TrimSpace(reCheckbox.FindStringSubmatch(line)[2])
	case len(reTaskNumbered.FindStringSubmatch(t)) == 2:
		text = strings.TrimSpace(reTaskNumbered.FindStringSubmatch(t)[1])
	}

	if isBlockerPlaceholder(text) {
		return "", false
	}
	return trimForReport(text), true
}

func isBlockerPlaceholder(text string) bool {
	t := strings.ToLower(strings.TrimSpace(text))
	t = strings.Trim(t, "*`")
	t = strings.TrimSpace(t)
	switch t {
	case "", "none", "none currently", "none currently.", "no blockers", "no blockers.", "ninguno", "ninguno actualmente", "ninguno actualmente.", "sin bloqueadores", "sin bloqueadores.":
		return true
	}
	if strings.HasPrefix(t, "<") && strings.HasSuffix(t, ">") {
		return true
	}
	return false
}

func extractNextActions(lines []string, p *ParsedPlan) {
	collect := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		lt := strings.ToLower(t)
		if strings.HasPrefix(lt, "## siguientes pasos") || strings.HasPrefix(lt, "### siguientes pasos") || strings.Contains(lt, "**siguientes pasos") || strings.Contains(lt, "**siguiente delta") || strings.HasPrefix(lt, "## next steps") || strings.Contains(lt, "**next delta") {
			collect = true
			continue
		}
		if collect {
			if strings.HasPrefix(t, "## ") || strings.HasPrefix(t, "### ") {
				collect = false
				continue
			}
			if m := reTaskNumbered.FindStringSubmatch(t); len(m) == 2 {
				p.NextActions = appendUnique(p.NextActions, trimForReport(m[1]))
			}
			if m := reCheckbox.FindStringSubmatch(line); len(m) == 3 && m[1] == " " {
				p.NextActions = appendUnique(p.NextActions, trimForReport(m[2]))
			}
		}
	}
}

func extractTotalProgress(text string) int {
	m := reAnyPercent.FindStringSubmatch(text)
	if len(m) == 3 {
		if n, err := strconv.Atoi(m[2]); err == nil {
			return n
		}
	}
	return -1
}

func parseDateTime(line string) *time.Time {
	m := reDateTime.FindStringSubmatch(line)
	if len(m) < 2 {
		return nil
	}
	ts := m[1] + " 00:00"
	if len(m) > 2 && m[2] != "" {
		ts = m[1] + " " + m[2]
	}
	t, err := time.Parse("2006-01-02 15:04", ts)
	if err != nil {
		return nil
	}
	return &t
}

func appendUnique(items []string, s string) []string {
	if s == "" {
		return items
	}
	for _, it := range items {
		if it == s {
			return items
		}
	}
	return append(items, s)
}

func trimForReport(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 160 {
		return s[:157] + "..."
	}
	return s
}

func cleanStatusValue(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "*")
	s = strings.TrimSpace(s)
	return s
}
