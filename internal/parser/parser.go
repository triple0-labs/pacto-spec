package parser

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"pacto/internal/model"
)

type ParsedPlan struct {
	Ref             model.PlanRef
	RawText         string
	DeclaredStatus  string
	Phases          []model.Phase
	Tasks           []model.Task
	BlockerHints    []string
	NextActions     []string
	HasCheckpoint   bool
	HasEvidence     bool
	LatestDeltaTime *time.Time
	ParseWarnings   []string
	ParseError      string
}

var (
	reDeclaredStatus = regexp.MustCompile(`(?i)^[-*]?\s*(?:\*\*)?(estado|status)(?::)?(?:\*\*)?:\s*(.+)$`)
	reCheckbox       = regexp.MustCompile(`^\s*[-*]\s*\[( |x|X)\]\s*(.+)$`)
	reTaskNumbered   = regexp.MustCompile(`^\s*\d+\.\s+(.+)$`)
	rePhaseRow       = regexp.MustCompile(`(?i)^\|\s*(fase\s*[^|]+)\|([^|]+)\|([^|]+)\|\s*([0-9]{1,3})%\s*\|`)
	reAnyPercent     = regexp.MustCompile(`(?i)(progreso total|progress)[:\s*]*([0-9]{1,3})%`)
	reDateTime       = regexp.MustCompile(`(20[0-9]{2}-[0-9]{2}-[0-9]{2})(?:[ T]([0-9]{2}:[0-9]{2}))?`)
)

func ParsePlan(ref model.PlanRef, mode string) (ParsedPlan, error) {
	p := ParsedPlan{Ref: ref}
	text, err := readPlanText(ref)
	if err != nil {
		return p, err
	}
	p.RawText = text
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
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

		if m := rePhaseRow.FindStringSubmatch(t); len(m) == 5 {
			prog, _ := strconv.Atoi(strings.TrimSpace(m[4]))
			p.Phases = append(p.Phases, model.Phase{Name: strings.TrimSpace(m[1]), RawState: strings.TrimSpace(m[3]), Progress: prog})
		}

		if m := reCheckbox.FindStringSubmatch(line); len(m) == 3 {
			done := strings.EqualFold(strings.TrimSpace(m[1]), "x")
			text := strings.TrimSpace(m[2])
			p.Tasks = append(p.Tasks, model.Task{Text: text, Completed: done, LikelyBlk: looksBlocked(text)})
		}

		lt := strings.ToLower(t)
		if strings.HasPrefix(lt, "**checkpoint") || strings.HasPrefix(lt, "checkpoint") {
			p.HasCheckpoint = true
		}
		if strings.Contains(lt, "evidencia") || strings.Contains(lt, "smoke") || strings.Contains(lt, "validación") || strings.Contains(lt, "validacion") {
			p.HasEvidence = true
		}
		if looksBlockerLine(t) {
			p.BlockerHints = appendUnique(p.BlockerHints, trimForReport(t))
		}

		if strings.Contains(strings.ToLower(t), "delta") || strings.Contains(strings.ToLower(t), "checkpoint") {
			if dt := parseDateTime(t); dt != nil {
				if p.LatestDeltaTime == nil || dt.After(*p.LatestDeltaTime) {
					p.LatestDeltaTime = dt
				}
			}
		}
	}

	extractNextActions(lines, &p)
	if len(p.Phases) == 0 {
		if pct := extractTotalProgress(text); pct >= 0 {
			p.Phases = append(p.Phases, model.Phase{Name: "total", RawState: "derived", Progress: pct})
		}
	}

	if p.DeclaredStatus == "" && mode == "strict" {
		return p, fmt.Errorf("missing declared status")
	}
	if len(p.Phases) == 0 && mode == "strict" {
		p.ParseWarnings = append(p.ParseWarnings, "missing structured progress source")
	}
	return p, nil
}

func readPlanText(ref model.PlanRef) (string, error) {
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

func looksBlocked(text string) bool {
	t := strings.ToLower(text)
	keys := []string{"blocked", "bloqueado", "pendiente crítico", "pendiente critico", "falla", "error", "pendiente", "en espera", "bloqueado por"}
	for _, k := range keys {
		if strings.Contains(t, k) {
			return true
		}
	}
	return false
}

func looksBlockerLine(text string) bool {
	t := strings.ToLower(text)
	if strings.Contains(t, "bloqueador") || strings.Contains(t, "blocker") {
		return true
	}
	if strings.Contains(t, "pendiente crítico") || strings.Contains(t, "pendiente critico") {
		return true
	}
	if strings.Contains(t, "bloqueado") || strings.Contains(t, "blocked") {
		return true
	}
	return false
}

func extractNextActions(lines []string, p *ParsedPlan) {
	collect := false
	for _, line := range lines {
		t := strings.TrimSpace(line)
		lt := strings.ToLower(t)
		if strings.HasPrefix(lt, "## siguientes pasos") || strings.HasPrefix(lt, "### siguientes pasos") || strings.Contains(lt, "**siguientes pasos") || strings.Contains(lt, "**siguiente delta") || strings.HasPrefix(lt, "## next steps") {
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
