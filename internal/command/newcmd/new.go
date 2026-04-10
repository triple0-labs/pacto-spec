package newcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/ui"
	"pacto/internal/workspace"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type Options struct {
	Root         string
	Title        string
	Owner        string
	AllowMinimal bool
	Lang         string
	RootProvided bool
}

type newRequest struct {
	root         string
	state        string
	slug         string
	title        string
	owner        string
	allowMinimal bool
	date         string
	planDir      string
	readmePath   string
	specPath     string
	designPath   string
	tasksPath    string
}

func Run(opts Options, state, slug string) int {
	state = strings.ToLower(strings.TrimSpace(state))
	slug = strings.TrimSpace(slug)
	if !isValidState(state) {
		fmt.Fprintf(os.Stderr, "invalid state %q (allowed: current|to-implement|done|outdated)\n", state)
		return 2
	}
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return 2
	}
	if strings.TrimSpace(opts.Lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.Lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.Lang)
			return 2
		}
	}
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	if strings.TrimSpace(opts.Lang) != "" {
		cmdutil.SetGlobalLangOverride(opts.Lang)
		defer cmdutil.SetGlobalLangOverride("")
	}

	req, code, ok := buildNewRequest(opts, state, slug, opts.RootProvided)
	if !ok {
		return code
	}
	lang := cmdutil.EffectiveLanguage(req.root)
	created, code := createPlanScaffold(req)
	if code != 0 {
		return code
	}

	fmt.Println(ui.ActionHeader(cmdutil.Tr(lang, "Created Plan", "Plan creado"), req.state+"/"+req.slug))
	for _, p := range created {
		fmt.Println(cmdutil.PathLine("created", p))
	}
	return 0
}

func buildNewRequest(opts Options, state, slug string, rootProvided bool) (newRequest, int, bool) {
	absRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return newRequest{}, 2, false
	}
	if rootProvided {
		if resolved, ok := workspace.ResolvePlanRoot(absRoot); ok {
			absRoot = resolved
		}
	} else if resolved, _, ok := workspace.ResolvePlanRootFrom(absRoot); ok {
		absRoot = resolved
	}

	if opts.AllowMinimal {
		if err := workspace.EnsureMinimalRoot(absRoot, cmdutil.EffectiveLanguage(absRoot)); err != nil {
			fmt.Fprintf(os.Stderr, "prepare minimal root: %v\n", err)
			return newRequest{}, 2, false
		}
	} else {
		if err := workspace.ValidateRoot(absRoot); err != nil {
			fmt.Fprintf(os.Stderr, "invalid pacto root: %v\n", err)
			return newRequest{}, 2, false
		}
	}

	planTitle := strings.TrimSpace(opts.Title)
	if planTitle == "" {
		planTitle = slugToTitle(slug)
	}

	now := time.Now()
	date := now.Format("2006-01-02")
	planDir := filepath.Join(absRoot, state, slug)
	if _, err := os.Stat(planDir); err == nil {
		fmt.Fprintf(os.Stderr, "plan already exists: %s\n", planDir)
		return newRequest{}, 2, false
	}
	req := newRequest{
		root:         absRoot,
		state:        state,
		slug:         slug,
		title:        planTitle,
		owner:        opts.Owner,
		allowMinimal: opts.AllowMinimal,
		date:         date,
		planDir:      planDir,
		readmePath:   filepath.Join(planDir, "README.md"),
		specPath:     filepath.Join(planDir, "spec.md"),
		designPath:   filepath.Join(planDir, "design.md"),
		tasksPath:    filepath.Join(planDir, "tasks.md"),
	}
	return req, 0, true
}

func createPlanScaffold(req newRequest) ([]string, int) {
	lang := cmdutil.EffectiveLanguage(req.root)
	if err := os.MkdirAll(req.planDir, 0o775); err != nil {
		fmt.Fprintf(os.Stderr, "create plan dir: %v\n", err)
		return nil, 3
	}

	created := make([]string, 0, 4)
	created = append(created, req.readmePath)

	specText := defaultSpecTemplate(req.title, req.date, req.owner, req.state, req.slug, lang)
	designText := defaultDesignTemplate(req.title, req.date, req.owner, req.state, req.slug, lang)
	tasksText := defaultTasksTemplate(req.title, req.date, req.owner, req.state, req.slug, lang)

	if err := os.WriteFile(req.specPath, []byte(specText), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write spec file: %v\n", err)
		return nil, 3
	}
	if err := os.WriteFile(req.designPath, []byte(designText), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write design file: %v\n", err)
		return nil, 3
	}
	if err := os.WriteFile(req.tasksPath, []byte(tasksText), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write tasks file: %v\n", err)
		return nil, 3
	}
	if err := os.WriteFile(req.readmePath, []byte(buildPlanReadme(req.title, req.state, req.date, []string{"spec.md", "design.md", "tasks.md"}, lang)), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write readme: %v\n", err)
		return nil, 3
	}
	created = append(created, req.specPath, req.designPath, req.tasksPath)
	return created, 0
}

func normalizeNewArgs(args []string) ([]string, error) {
	withValue := map[string]bool{"--root": true, "-root": true, "--title": true, "-title": true, "--owner": true, "-owner": true, "--lang": true, "-lang": true}
	return cmdutil.NormalizeArgs(args, withValue)
}

func isValidState(state string) bool {
	switch state {
	case "current", "to-implement", "done", "outdated":
		return true
	default:
		return false
	}
}

func buildPlanReadme(title, state, date string, docs []string, lang i18n.Language) string {
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

func defaultSpecTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Spec: %s\n\n## Metadata\n\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Problem Statement\n\n<Describe the problem and scope.>\n\n## User Scenarios\n\n### Scenario: <name>\n\n- **GIVEN** <initial state>\n- **WHEN** <action>\n- **THEN** <outcome>\n\n## Acceptance Criteria\n\n- AC-001: <measurable outcome>.\n\n## Domains Affected\n\n- <domain>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Especificación: %s\n\n## Metadatos\n\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Planteamiento del Problema\n\n<Describe el problema y su alcance.>\n\n## Escenarios de Usuario\n\n### Escenario: <nombre>\n\n- **DADO** <estado inicial>\n- **CUANDO** <acción>\n- **ENTONCES** <resultado>\n\n## Criterios de Aceptación\n\n- AC-001: <resultado medible>.\n\n## Dominios Afectados\n\n- <dominio>\n", title, owner, date, date, state, slug),
	)
}

func defaultDesignTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Design: %s\n\n## Metadata\n\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Technical Context\n\n- Language/Version: <value>\n- Dependencies: <value>\n- Constraints: <value>\n\n## Architecture Decisions\n\n1. Decision: <text> | Rationale: <text>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Diseño: %s\n\n## Metadatos\n\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Contexto Técnico\n\n- Lenguaje/Versión: <valor>\n- Dependencias: <valor>\n- Restricciones: <valor>\n\n## Decisiones de Arquitectura\n\n1. Decisión: <texto> | Justificación: <texto>\n", title, owner, date, date, state, slug),
	)
}

func defaultTasksTemplate(title, date, owner, state, slug string, lang i18n.Language) string {
	return cmdutil.Tr(lang,
		fmt.Sprintf("# Tasks: %s\n\n## Execution Metadata\n\n- Status: Draft\n- Owner: %s\n- Created: %s\n- Last Modified: %s\n- State: %s\n- Slug: %s\n\n## Implementation Plan by Phases\n\n## Phase 1: <title>\n\n- [ ] 1.1 <task>\n\n## Evidence\n\n- <YYYY-MM-DD HH:MM> `<path|symbol|command>`\n\n## Blockers\n\n- <YYYY-MM-DD HH:MM> <blocker>\n\n## Next Steps\n\n1. <next step>\n", title, owner, date, date, state, slug),
		fmt.Sprintf("# Tareas: %s\n\n## Metadatos de Ejecución\n\n- Estado: Borrador\n- Responsable: %s\n- Creado: %s\n- Última Modificación: %s\n- Estado de Carpeta: %s\n- Slug: %s\n\n## Plan de Implementación por Fases\n\n## Fase 1: <título>\n\n- [ ] 1.1 <tarea>\n\n## Evidencia\n\n- <YYYY-MM-DD HH:MM> `<ruta|simbolo|comando>`\n\n## Bloqueadores\n\n- <YYYY-MM-DD HH:MM> <bloqueador>\n\n## Siguientes Pasos\n\n1. <siguiente paso>\n", title, owner, date, date, state, slug),
	)
}
