// Package newplan contains the application/use-case for creating a new plan
// scaffold (README + spec/design/tasks). CLI flag parsing, language override
// state, and pretty output stay in internal/command/newcmd; this package
// returns a result struct or a typed error and writes only files.
package newplan

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Input is everything the use case needs to produce a plan scaffold.
// Root may be relative; it is resolved (and walked upward to .pacto/plans)
// internally. RootProvided records whether the caller explicitly supplied
// --root: if true we anchor at that directory; if false we walk upward to
// find an existing plans tree.
type Input struct {
	Root         string
	RootProvided bool
	State        string
	Slug         string
	Title        string
	Owner        string
	AllowMinimal bool
	Lang         i18n.Language
}

// Result captures what was created on disk.
type Result struct {
	PlanDir      string
	CreatedFiles []string // README + spec/design/tasks, in display order
}

// ErrInvalid signals validation problems (bad state/slug/root, plan already
// exists). The CLI maps these to exit code 2.
var ErrInvalid = errors.New("invalid input")

// Create validates Input and writes the plan scaffold under
// <root>/<state>/<slug>/. On success it returns the absolute paths of all
// files created.
func Create(in Input) (Result, error) {
	state := strings.ToLower(strings.TrimSpace(in.State))
	slug := strings.TrimSpace(in.Slug)
	if !isValidState(state) {
		return Result{}, fmt.Errorf("%w: invalid state %q (allowed: current|to-implement|done|outdated)", ErrInvalid, state)
	}
	if !slugRe.MatchString(slug) {
		return Result{}, fmt.Errorf("%w: invalid slug %q (use lowercase letters, numbers, dashes)", ErrInvalid, slug)
	}

	rootArg := in.Root
	if strings.TrimSpace(rootArg) == "" {
		rootArg = "."
	}
	absRoot, err := filepath.Abs(rootArg)
	if err != nil {
		return Result{}, fmt.Errorf("resolve root: %w", err)
	}
	if in.RootProvided {
		if resolved, ok := workspace.ResolvePlanRoot(absRoot); ok {
			absRoot = resolved
		}
	} else if resolved, _, ok := workspace.ResolvePlanRootFrom(absRoot); ok {
		absRoot = resolved
	}

	if in.AllowMinimal {
		if err := workspace.EnsureMinimalRoot(absRoot, in.Lang); err != nil {
			return Result{}, fmt.Errorf("prepare minimal root: %w", err)
		}
	} else {
		if err := workspace.ValidateRoot(absRoot); err != nil {
			return Result{}, fmt.Errorf("%w: %v", ErrInvalid, err)
		}
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = slugToTitle(slug)
	}
	date := time.Now().Format("2006-01-02")

	planDir := filepath.Join(absRoot, state, slug)
	if _, err := os.Stat(planDir); err == nil {
		return Result{}, fmt.Errorf("%w: plan already exists at %s", ErrInvalid, planDir)
	}

	readmePath := filepath.Join(planDir, "README.md")
	specPath := filepath.Join(planDir, "spec.md")
	designPath := filepath.Join(planDir, "design.md")
	tasksPath := filepath.Join(planDir, "tasks.md")

	if err := os.MkdirAll(planDir, 0o775); err != nil {
		return Result{}, fmt.Errorf("create plan dir: %w", err)
	}
	if err := os.WriteFile(specPath, []byte(specTemplate(title, date, in.Owner, state, slug, in.Lang)), 0o664); err != nil {
		return Result{}, fmt.Errorf("write spec file: %w", err)
	}
	if err := os.WriteFile(designPath, []byte(designTemplate(title, date, in.Owner, state, slug, in.Lang)), 0o664); err != nil {
		return Result{}, fmt.Errorf("write design file: %w", err)
	}
	if err := os.WriteFile(tasksPath, []byte(tasksTemplate(title, date, in.Owner, state, slug, in.Lang)), 0o664); err != nil {
		return Result{}, fmt.Errorf("write tasks file: %w", err)
	}
	if err := os.WriteFile(readmePath, []byte(readmeTemplate(title, state, date, []string{"spec.md", "design.md", "tasks.md"}, in.Lang)), 0o664); err != nil {
		return Result{}, fmt.Errorf("write readme: %w", err)
	}

	return Result{
		PlanDir:      planDir,
		CreatedFiles: []string{readmePath, specPath, designPath, tasksPath},
	}, nil
}

func isValidState(state string) bool {
	switch state {
	case "current", "to-implement", "done", "outdated":
		return true
	default:
		return false
	}
}

func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func readmeTemplate(title, state, date string, docs []string, lang i18n.Language) string {
	statusEN := map[string]string{
		"current":      "In Progress (Current)",
		"to-implement": "Pending (To Implement)",
		"done":         "Completed (Done)",
		"outdated":     "Outdated (Outdated)",
	}[state]
	statusES := map[string]string{
		"current":      "En ejecución",
		"to-implement": "Pendiente",
		"done":         "Completado",
		"outdated":     "Obsoleto",
	}[state]
	status := cmdutil.Tr(lang, statusEN, statusES)
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString(cmdutil.Tr(lang, "**Status:** ", "**Estado:** ") + status + "  \n")
	b.WriteString(cmdutil.Tr(lang, "**Date:** ", "**Fecha:** ") + date + "\n\n")
	b.WriteString(cmdutil.Tr(lang, "## Description\n\n", "## Descripción\n\n"))
	b.WriteString(cmdutil.Tr(lang, "Plan created with `pacto new`.\n\n", "Plan creado con `pacto new`.\n\n"))
	b.WriteString(cmdutil.Tr(lang, "## Documents\n\n", "## Documentos\n\n"))
	for _, doc := range docs {
		b.WriteString("- [" + doc + "](./" + doc + ")\n")
	}
	return b.String()
}

func specTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Spec: %s\n\n## Metadata\n\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Problem Statement\n\n<Describe the problem and scope.>\n\n## User Scenarios\n\n### Scenario: <name>\n\n- **GIVEN** <initial state>\n- **WHEN** <action>\n- **THEN** <outcome>\n\n## Acceptance Criteria\n\n- AC-001: <measurable outcome>.\n\n## Capabilities\n\n- New Capabilities: []\n- Modified Capabilities: []\n\n## Requirements\n\n### Requirement: <name>\n\nThe system SHALL <behaviour>.\n\n#### Scenario: <name>\n\n- WHEN <action>\n- THEN <outcome>\n\n## Domains Affected\n\n- <domain>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Especificación: %s\n\n## Metadatos\n\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Planteamiento del Problema\n\n<Describe el problema y su alcance.>\n\n## Escenarios de Usuario\n\n### Escenario: <nombre>\n\n- **DADO** <estado inicial>\n- **CUANDO** <acción>\n- **ENTONCES** <resultado>\n\n## Criterios de Aceptación\n\n- AC-001: <resultado medible>.\n\n## Capacidades\n\n- Capacidades Nuevas: []\n- Capacidades Modificadas: []\n\n## Requisitos\n\n### Requisito: <nombre>\n\nEl sistema DEBE <comportamiento>.\n\n#### Escenario: <nombre>\n\n- CUANDO <acción>\n- ENTONCES <resultado>\n\n## Dominios Afectados\n\n- <dominio>\n", title, owner, date, date, state, slug),
	)
}

func designTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Design: %s\n\n## Metadata\n\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Technical Context\n\n- Language/Version: <value>\n- Dependencies: <value>\n- Constraints: <value>\n\n## Architecture Decisions\n\n1. Decision: <text> | Rationale: <text>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Diseño: %s\n\n## Metadatos\n\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Contexto Técnico\n\n- Lenguaje/Versión: <valor>\n- Dependencias: <valor>\n- Restricciones: <valor>\n\n## Decisiones de Arquitectura\n\n1. Decisión: <texto> | Justificación: <texto>\n", title, owner, date, date, state, slug),
	)
}

func tasksTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Tasks: %s\n\n## Execution Metadata\n\n- Status: Draft\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Implementation Plan by Phases\n\n## Phase 1: <title>\n\n- [ ] 1.1 <task>\n\n## Evidence\n\n- <YYYY-MM-DD HH:MM> `<path|symbol|command>`\n\n## Blockers\n\n- <YYYY-MM-DD HH:MM> <blocker>\n\n## Next Steps\n\n1. <next step>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Tareas: %s\n\n## Metadatos de Ejecución\n\n- Estado: Borrador\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Plan de Implementación por Fases\n\n## Fase 1: <título>\n\n- [ ] 1.1 <tarea>\n\n## Evidencia\n\n- <YYYY-MM-DD HH:MM> `<ruta|simbolo|comando>`\n\n## Bloqueadores\n\n- <YYYY-MM-DD HH:MM> <bloqueador>\n\n## Siguientes Pasos\n\n1. <siguiente paso>\n", title, owner, date, date, state, slug),
	)
}
