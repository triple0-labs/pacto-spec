package planfmt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Issue struct {
	Code    string
	Message string
}

type NormalizeResult struct {
	Content  string
	Changed  bool
	Changes  []string
	Warnings []string
}

var (
	reHeading      = regexp.MustCompile(`^##\s+(.+?)\s*$`)
	rePhaseHeading = regexp.MustCompile(`(?i)^##\s*(phase|fase)\s+([1-9][0-9]*)`)
	reLegacyTask   = regexp.MustCompile(`^(\s*[-*]\s*\[[ xX]\]\s*)T([1-9][0-9]*)\.\s*(.+)$`)
	reFRLine       = regexp.MustCompile(`(?im)^\s*[-*]?\s*FR-[0-9]{3}\s*:\s*(.+)$`)
	reACLine       = regexp.MustCompile(`(?im)^\s*[-*]?\s*AC-[0-9]{3}\s*:\s*(.+)$`)
	reScenario     = regexp.MustCompile(`(?im)^###\s*(scenario|escenario)\s*:`)
	reNumericTask  = regexp.MustCompile(`(?im)^\s*[-*]\s*\[[ xX]\]\s*[1-9][0-9]*\.[1-9][0-9]*\s+`)
	reLegacyTaskID = regexp.MustCompile(`(?im)^\s*[-*]\s*\[[ xX]\]\s*T[0-9]+\.`)
	reLastModified = regexp.MustCompile(`(?im)^\s*[-*]?\s*(last modified|última modificación|ultima modificacion)\s*:\s*(.+)$`)
)

var requiredCanon = []string{
	"metadata",
	"problem",
	"goals",
	"non_goals",
	"user_scenarios",
	"functional_requirements",
	"non_functional_requirements",
	"acceptance_criteria",
	"technical_context",
	"implementation_phases",
	"evidence",
	"risks",
	"next_steps",
}

var canonicalAliases = map[string]string{
	"metadata":                         "metadata",
	"metadatos":                        "metadata",
	"intent":                           "intent",
	"intención":                        "intent",
	"intencion":                        "intent",
	"problem statement":                "problem",
	"planteamiento del problema":       "problem",
	"declaracion del problema":         "problem",
	"declaración del problema":         "problem",
	"goals":                            "goals",
	"objetivos":                        "goals",
	"non-goals":                        "non_goals",
	"non goals":                        "non_goals",
	"no objetivos":                     "non_goals",
	"user scenarios":                   "user_scenarios",
	"escenarios de usuario":            "user_scenarios",
	"functional requirements":          "functional_requirements",
	"requerimientos funcionales":       "functional_requirements",
	"requisitos funcionales":           "functional_requirements",
	"non-functional requirements":      "non_functional_requirements",
	"non functional requirements":      "non_functional_requirements",
	"requerimientos no funcionales":    "non_functional_requirements",
	"requisitos no funcionales":        "non_functional_requirements",
	"acceptance criteria":              "acceptance_criteria",
	"criterios de aceptación":          "acceptance_criteria",
	"criterios de aceptacion":          "acceptance_criteria",
	"success criteria":                 "acceptance_criteria",
	"criterios de éxito":               "acceptance_criteria",
	"criterios de exito":               "acceptance_criteria",
	"technical context":                "technical_context",
	"contexto técnico":                 "technical_context",
	"contexto tecnico":                 "technical_context",
	"context":                          "technical_context",
	"contexto":                         "technical_context",
	"implementation plan by phases":    "implementation_phases",
	"plan de implementación por fases": "implementation_phases",
	"plan de implementacion por fases": "implementation_phases",
	"evidence":                         "evidence",
	"evidencia":                        "evidence",
	"risks and mitigations":            "risks",
	"riesgos y mitigaciones":           "risks",
	"next steps":                       "next_steps",
	"siguientes pasos":                 "next_steps",
	"próximos pasos":                   "next_steps",
	"proximos pasos":                   "next_steps",
}

func Validate(content string) []Issue {
	issues := make([]Issue, 0, 16)
	headings := collectCanonicalHeadings(content)
	if _, ok := headings["problem"]; !ok {
		if _, hasIntent := headings["intent"]; !hasIntent {
			issues = append(issues, Issue{
				Code:    "missing_core_intent",
				Message: "missing core section: problem statement or intent",
			})
		}
	}
	if _, ok := headings["acceptance_criteria"]; !ok && len(reACLine.FindAllStringSubmatch(content, -1)) == 0 {
		issues = append(issues, Issue{Code: "missing_core_acceptance", Message: "missing core acceptance criteria"})
	}
	if len(reScenario.FindAllStringSubmatch(content, -1)) == 0 {
		issues = append(issues, Issue{Code: "missing_core_scenarios", Message: "missing core scenarios"})
	}
	if len(reNumericTask.FindAllStringSubmatch(content, -1)) == 0 {
		issues = append(issues, Issue{Code: "missing_core_tasks", Message: "missing core numeric phase tasks (N.M format)"})
	}
	if _, ok := headings["evidence"]; !ok {
		issues = append(issues, Issue{Code: "missing_core_evidence", Message: "missing core evidence section"})
	}
	if len(reLastModified.FindAllStringSubmatch(content, -1)) == 0 {
		issues = append(issues, Issue{Code: "missing_core_last_modified", Message: "missing core last modified metadata"})
	}

	fr := reFRLine.FindAllStringSubmatch(content, -1)
	if _, hasFRModule := headings["functional_requirements"]; hasFRModule && len(fr) == 0 {
		issues = append(issues, Issue{Code: "module_missing_fr", Message: "functional requirements section present but no FR-### entries"})
	}
	if len(fr) > 0 {
		for _, m := range fr {
			if len(m) < 2 {
				continue
			}
			text := strings.ToUpper(m[1])
			if !strings.Contains(text, "MUST") && !strings.Contains(text, "SHALL") {
				issues = append(issues, Issue{Code: "fr_must_shall", Message: "each FR must include MUST or SHALL"})
				break
			}
		}
	}
	if _, hasACModule := headings["acceptance_criteria"]; hasACModule && len(reACLine.FindAllStringSubmatch(content, -1)) == 0 {
		issues = append(issues, Issue{Code: "module_missing_ac", Message: "acceptance criteria section present but no AC-### entries"})
	}
	if reLegacyTaskID.MatchString(content) {
		issues = append(issues, Issue{Code: "legacy_task_id", Message: "legacy task IDs (T1/T2) are not allowed in strict mode"})
	}
	return issues
}

func Normalize(content string) NormalizeResult {
	lines := strings.Split(content, "\n")
	lang := detectLanguage(content)
	out := make([]string, 0, len(lines)+32)
	changes := make([]string, 0, 16)
	warnings := make([]string, 0, 8)

	currentPhase := 0
	for _, line := range lines {
		normalizedLine := line
		if m := reHeading.FindStringSubmatch(line); len(m) == 2 {
			if canon, ok := canonicalAliases[normalizeHeadingText(m[1])]; ok {
				target := "## " + sectionTitle(canon, lang)
				if strings.TrimSpace(line) != target {
					normalizedLine = target
					changes = append(changes, "normalized heading: "+strings.TrimSpace(line)+" -> "+target)
				}
			}
		}
		if m := rePhaseHeading.FindStringSubmatch(strings.TrimSpace(normalizedLine)); len(m) == 3 {
			currentPhase, _ = strconv.Atoi(m[2])
		}
		if m := reLegacyTask.FindStringSubmatch(normalizedLine); len(m) == 4 {
			if currentPhase > 0 {
				step, _ := strconv.Atoi(m[2])
				normalizedLine = m[1] + fmt.Sprintf("%d.%d %s", currentPhase, step, strings.TrimSpace(m[3]))
				changes = append(changes, "normalized legacy task ID to numeric format")
			} else {
				warnings = append(warnings, "found legacy task ID outside phase context; left unchanged")
			}
		}
		out = append(out, normalizedLine)
	}

	resultText := strings.Join(out, "\n")
	return NormalizeResult{
		Content:  resultText,
		Changed:  resultText != content,
		Changes:  dedupe(changes),
		Warnings: dedupe(warnings),
	}
}

func ensureMetadataSection(lines []string, lang string) ([]string, bool) {
	for _, line := range lines {
		if m := reHeading.FindStringSubmatch(line); len(m) == 2 {
			if canon, ok := canonicalAliases[normalizeHeadingText(m[1])]; ok && canon == "metadata" {
				return lines, false
			}
		}
	}

	insertAt := 0
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			insertAt = i + 1
			break
		}
	}

	metadata := []string{
		"",
		"## " + sectionTitle("metadata", lang),
		"",
		"- " + label("status", lang) + ": <pending>",
		"- " + label("owner", lang) + ": <owner>",
		"- " + label("created", lang) + ": <YYYY-MM-DD>",
		"- " + label("last_modified", lang) + ": <YYYY-MM-DD>",
		"- " + label("state", lang) + ": <current|to-implement|done|outdated>",
		"- " + label("slug", lang) + ": <slug>",
	}
	out := make([]string, 0, len(lines)+len(metadata))
	out = append(out, lines[:insertAt]...)
	out = append(out, metadata...)
	out = append(out, lines[insertAt:]...)
	return out, true
}

func collectCanonicalHeadings(content string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, line := range strings.Split(content, "\n") {
		m := reHeading.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		if canon, ok := canonicalAliases[normalizeHeadingText(m[1])]; ok {
			out[canon] = struct{}{}
		}
	}
	return out
}

func sectionTitle(canon, lang string) string {
	if lang == "es" {
		switch canon {
		case "metadata":
			return "Metadatos"
		case "problem":
			return "Planteamiento del Problema"
		case "goals":
			return "Objetivos"
		case "non_goals":
			return "No Objetivos"
		case "user_scenarios":
			return "Escenarios de Usuario"
		case "functional_requirements":
			return "Requerimientos Funcionales"
		case "non_functional_requirements":
			return "Requerimientos No Funcionales"
		case "acceptance_criteria":
			return "Criterios de Aceptación"
		case "technical_context":
			return "Contexto Técnico"
		case "implementation_phases":
			return "Plan de Implementación por Fases"
		case "evidence":
			return "Evidencia"
		case "risks":
			return "Riesgos y Mitigaciones"
		case "next_steps":
			return "Siguientes Pasos"
		}
	}
	switch canon {
	case "metadata":
		return "Metadata"
	case "problem":
		return "Problem Statement"
	case "goals":
		return "Goals"
	case "non_goals":
		return "Non-Goals"
	case "user_scenarios":
		return "User Scenarios"
	case "functional_requirements":
		return "Functional Requirements"
	case "non_functional_requirements":
		return "Non-Functional Requirements"
	case "acceptance_criteria":
		return "Acceptance Criteria"
	case "technical_context":
		return "Technical Context"
	case "implementation_phases":
		return "Implementation Plan by Phases"
	case "evidence":
		return "Evidence"
	case "risks":
		return "Risks and Mitigations"
	case "next_steps":
		return "Next Steps"
	default:
		return canon
	}
}

func sectionPlaceholder(canon, lang string) string {
	if lang == "es" {
		switch canon {
		case "problem":
			return "<Describe el problema y su alcance.>"
		case "functional_requirements":
			return "- FR-001: El sistema MUST <capacidad verificable>"
		case "acceptance_criteria":
			return "- AC-001: <resultado medible>"
		case "user_scenarios":
			return "### Escenario: <nombre>\n- **GIVEN** <estado inicial>\n- **WHEN** <acción>\n- **THEN** <resultado>"
		case "implementation_phases":
			return "## Phase 1: <título>\n- [ ] 1.1 <tarea>"
		default:
			return "<Completar>"
		}
	}
	switch canon {
	case "problem":
		return "<Describe the problem and scope.>"
	case "functional_requirements":
		return "- FR-001: The system MUST <verifiable capability>"
	case "acceptance_criteria":
		return "- AC-001: <measurable outcome>"
	case "user_scenarios":
		return "### Scenario: <name>\n- **GIVEN** <initial state>\n- **WHEN** <action>\n- **THEN** <outcome>"
	case "implementation_phases":
		return "## Phase 1: <title>\n- [ ] 1.1 <task>"
	default:
		return "<Fill in>"
	}
}

func normalizeHeadingText(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	s = strings.TrimPrefix(s, "🟢 ")
	s = strings.TrimPrefix(s, "🟡 ")
	s = strings.TrimPrefix(s, "✅ ")
	s = strings.TrimPrefix(s, "⚠️ ")
	s = strings.TrimPrefix(s, "⚠ ")
	s = strings.TrimPrefix(s, "📋 ")
	s = strings.TrimPrefix(s, "📝 ")
	s = strings.TrimPrefix(s, "🎯 ")
	s = strings.TrimPrefix(s, "🔄 ")
	s = strings.TrimPrefix(s, "🚨 ")
	s = strings.TrimPrefix(s, "📊 ")
	s = strings.TrimPrefix(s, "🧪 ")
	s = strings.TrimPrefix(s, "🔍 ")
	s = strings.TrimPrefix(s, "🔧 ")
	s = strings.TrimPrefix(s, "🏗️ ")
	s = strings.TrimPrefix(s, "🏗 ")
	s = strings.ReplaceAll(s, "  ", " ")
	return strings.TrimSpace(s)
}

func detectLanguage(content string) string {
	l := strings.ToLower(content)
	scoreES := 0
	for _, tok := range []string{"objetivos", "siguientes pasos", "riesgos", "criterios", "evidencia", "fase"} {
		if strings.Contains(l, tok) {
			scoreES++
		}
	}
	if scoreES >= 2 {
		return "es"
	}
	return "en"
}

func label(name, lang string) string {
	if lang == "es" {
		switch name {
		case "status":
			return "Estado"
		case "owner":
			return "Owner"
		case "created":
			return "Creado"
		case "last_modified":
			return "Última Modificación"
		case "state":
			return "Estado de Carpeta"
		case "slug":
			return "Slug"
		}
	}
	switch name {
	case "status":
		return "Status"
	case "owner":
		return "Owner"
	case "created":
		return "Created"
	case "last_modified":
		return "Last Modified"
	case "state":
		return "State"
	case "slug":
		return "Slug"
	default:
		return name
	}
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
