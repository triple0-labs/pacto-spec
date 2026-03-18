package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"pacto/internal/i18n"
	"pacto/internal/ui"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

type newOptions struct {
	root         string
	title        string
	owner        string
	allowMinimal bool
	lang         string
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
	planFileName string
	planPath     string
	readmePath   string
}

func RunNew(args []string) int {
	opts, state, slug, rootProvided, code, ok := parseAndValidateNewArgs(args)
	if !ok {
		return code
	}
	return runNewParsed(opts, state, slug, rootProvided)
}

func runNewParsed(opts newOptions, state, slug string, rootProvided bool) int {
	if strings.TrimSpace(opts.lang) != "" {
		setGlobalLangOverride(opts.lang)
		defer setGlobalLangOverride("")
	}

	req, code, ok := buildNewRequest(opts, state, slug, rootProvided)
	if !ok {
		return code
	}
	lang := effectiveLanguage(req.root)
	if code = createPlanScaffold(req); code != 0 {
		return code
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Created Plan", "Plan creado"), req.state+"/"+req.slug))
	fmt.Println(pathLine("created", req.readmePath))
	fmt.Println(pathLine("created", req.planPath))
	return 0
}

func parseAndValidateNewArgs(args []string) (newOptions, string, string, bool, int, bool) {
	opts := newOptions{}
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto new <current|to-implement|done|outdated> <slug> [--title ...] [--owner ...] [--root <path>] [--allow-minimal-root]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}
	fs.StringVar(&opts.root, "root", ".", "Path to pacto root")
	fs.StringVar(&opts.title, "title", "", "Optional plan title")
	fs.StringVar(&opts.owner, "owner", "Platform Team", "Owner for generated plan")
	fs.BoolVar(&opts.allowMinimal, "allow-minimal-root", false, "Allow creating plans in lightweight/non-canonical roots")
	fs.StringVar(&opts.lang, "lang", "", "Output language override: en|es")

	normalizedArgs, normErr := normalizeNewArgs(args)
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return newOptions{}, "", "", false, 2, false
	}
	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return newOptions{}, "", "", false, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return newOptions{}, "", "", false, 2, false
	}
	if strings.TrimSpace(opts.lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.lang)
			return newOptions{}, "", "", false, 2, false
		}
	}
	rootProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "root" {
			rootProvided = true
		}
	})

	pos := fs.Args()
	if len(pos) != 2 {
		fs.Usage()
		return newOptions{}, "", "", rootProvided, 2, false
	}

	state := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	if !isValidState(state) {
		fmt.Fprintf(os.Stderr, "invalid state %q (allowed: current|to-implement|done|outdated)\n", state)
		return newOptions{}, "", "", rootProvided, 2, false
	}
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return newOptions{}, "", "", rootProvided, 2, false
	}
	return opts, state, slug, rootProvided, 0, true
}

func buildNewRequest(opts newOptions, state, slug string, rootProvided bool) (newRequest, int, bool) {
	absRoot, err := filepath.Abs(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return newRequest{}, 2, false
	}
	if rootProvided {
		if resolved, ok := resolvePlanRoot(absRoot); ok {
			absRoot = resolved
		}
	} else if resolved, _, ok := resolvePlanRootFrom(absRoot); ok {
		absRoot = resolved
	}

	if opts.allowMinimal {
		if err := ensureMinimalRoot(absRoot); err != nil {
			fmt.Fprintf(os.Stderr, "prepare minimal root: %v\n", err)
			return newRequest{}, 2, false
		}
	} else {
		if err := validateRoot(absRoot); err != nil {
			fmt.Fprintf(os.Stderr, "invalid pacto root: %v\n", err)
			return newRequest{}, 2, false
		}
	}

	planTitle := strings.TrimSpace(opts.title)
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
	planFileName := fmt.Sprintf("PLAN_%s_%s.md", slugToTopic(slug), date)
	req := newRequest{
		root:         absRoot,
		state:        state,
		slug:         slug,
		title:        planTitle,
		owner:        opts.owner,
		allowMinimal: opts.allowMinimal,
		date:         date,
		planDir:      planDir,
		planFileName: planFileName,
		planPath:     filepath.Join(planDir, planFileName),
		readmePath:   filepath.Join(planDir, "README.md"),
	}
	return req, 0, true
}

func createPlanScaffold(req newRequest) int {
	lang := effectiveLanguage(req.root)
	if err := os.MkdirAll(req.planDir, 0o775); err != nil {
		fmt.Fprintf(os.Stderr, "create plan dir: %v\n", err)
		return 3
	}

	planText, err := buildPlanFromTemplate(req.root, req.state, req.slug, req.title, req.date, req.owner, req.allowMinimal, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "build plan from template: %v\n", err)
		return 3
	}
	if err := os.WriteFile(req.planPath, []byte(planText), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write plan file: %v\n", err)
		return 3
	}
	if err := os.WriteFile(req.readmePath, []byte(buildPlanReadme(req.title, req.state, req.date, req.planFileName, lang)), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "write readme: %v\n", err)
		return 3
	}
	return 0
}

func normalizeNewArgs(args []string) ([]string, error) {
	withValue := map[string]bool{"--root": true, "-root": true, "--title": true, "-title": true, "--owner": true, "-owner": true, "--lang": true, "-lang": true}
	return normalizeArgs(args, withValue)
}

func isValidState(state string) bool {
	switch state {
	case "current", "to-implement", "done", "outdated":
		return true
	default:
		return false
	}
}

func validateRoot(root string) error {
	for _, p := range []string{"README.md", "PLANTILLA_PACTO_PLAN.md", "PACTO.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			return fmt.Errorf("missing %s", p)
		}
	}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if _, err := os.Stat(filepath.Join(root, st)); err != nil {
			return fmt.Errorf("missing state folder %s", st)
		}
	}
	return nil
}

func ensureMinimalRoot(root string) error {
	if err := os.MkdirAll(root, 0o775); err != nil {
		return err
	}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o775); err != nil {
			return err
		}
	}
	readmePath := filepath.Join(root, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		lang := effectiveLanguage(root)
		if err := os.WriteFile(readmePath, []byte(defaultRootReadme(lang)), 0o664); err != nil {
			return err
		}
	}
	pactoPath := filepath.Join(root, "PACTO.md")
	if _, err := os.Stat(pactoPath); err != nil {
		lang := effectiveLanguage(root)
		if err := os.WriteFile(pactoPath, []byte(tr(lang, "# Pacto\n\nMinimal root created by pacto CLI.\n", "# Pacto\n\nRaíz mínima creada por la CLI de pacto.\n")), 0o664); err != nil {
			return err
		}
	}
	templatePath := filepath.Join(root, "PLANTILLA_PACTO_PLAN.md")
	if _, err := os.Stat(templatePath); err != nil {
		lang := effectiveLanguage(root)
		if err := os.WriteFile(templatePath, []byte(defaultMinimalTemplate(lang)), 0o664); err != nil {
			return err
		}
	}
	return nil
}

func buildPlanFromTemplate(root, state, slug, title, date, owner string, allowMinimal bool, lang i18n.Language) (string, error) {
	tplPath := filepath.Join(root, "PLANTILLA_PACTO_PLAN.md")
	b, err := os.ReadFile(tplPath)
	if err != nil {
		if !allowMinimal {
			return "", err
		}
		return defaultPlanTemplate(state, slug, title, date, owner, lang), nil
	}
	t := string(b)
	t = strings.ReplaceAll(t, "<Title>", title)
	t = strings.ReplaceAll(t, "<Título del plan>", title)
	t = strings.ReplaceAll(t, "<YYYY-MM-DD>", date)
	t = strings.ReplaceAll(t, "<slug>", slug)
	t = strings.ReplaceAll(t, "<current|to-implement|done|outdated>", state)
	t = strings.ReplaceAll(t, "<Draft | In Progress | Completed | Blocked>", tr(lang, "Draft", "Borrador"))
	t = strings.ReplaceAll(t, "<Draft | En ejecución | Completado | Bloqueado>", "Draft")
	t = strings.ReplaceAll(t, "<nombre o equipo>", owner)
	t = strings.ReplaceAll(t, "<owner>", owner)
	t = strings.ReplaceAll(t, "<team>", owner)
	return t, nil
}

func buildPlanReadme(title, state, date, planFileName string, lang i18n.Language) string {
	statusEN := map[string]string{
		"current":      "In Progress (Current)",
		"to-implement": "Pending (To Implement)",
		"done":         "Completed (Done)",
		"outdated":     "Outdated (Outdated)",
	}[state]
	statusES := map[string]string{
		"current":      "En ejecución (Current)",
		"to-implement": "Pendiente (To Implement)",
		"done":         "Completado (Done)",
		"outdated":     "Obsoleto (Outdated)",
	}[state]
	status := tr(lang, statusEN, statusES)
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString(tr(lang, "**Status:** ", "**Estado:** ") + status + "  \n")
	b.WriteString(tr(lang, "**Date:** ", "**Fecha:** ") + date + "\n\n")
	b.WriteString(tr(lang, "## Description\n\n", "## Descripción\n\n"))
	b.WriteString(tr(lang, "Plan created with `pacto new`.\n\n", "Plan creado con `pacto new`.\n\n"))
	b.WriteString(tr(lang, "## Documents\n\n", "## Documentos\n\n"))
	b.WriteString("- [" + planFileName + "](./" + planFileName + ")\n")
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

func slugToTopic(slug string) string {
	up := strings.ToUpper(strings.ReplaceAll(slug, "-", "_"))
	up = strings.Trim(up, "_")
	if up == "" {
		return "PLAN"
	}
	return up
}

func defaultPlanTemplate(state, slug, title, date, owner string, lang i18n.Language) string {
	t := defaultMinimalTemplate(lang)
	t = strings.ReplaceAll(t, "<Title>", title)
	t = strings.ReplaceAll(t, "<Título del plan>", title)
	t = strings.ReplaceAll(t, "<YYYY-MM-DD>", date)
	t = strings.ReplaceAll(t, "<team>", owner)
	t = strings.ReplaceAll(t, "<nombre o equipo>", owner)
	t = strings.ReplaceAll(t, "<current|to-implement|done|outdated>", state)
	t = strings.ReplaceAll(t, "<slug>", slug)
	return t
}

func defaultRootReadme(lang i18n.Language) string {
	if lang == i18n.Spanish {
		return "# Planes de Pacto\n\n" +
			"## Resumen\n\n" +
			"El estado de planes se deriva de carpetas y documentos mediante `pacto status`.\n" +
			"Este README es solo una vista general, no la fuente de verdad del estado.\n\n" +
			"---\n\n" +
			"## 📜 Pacto\n\n" +
			"- [PACTO.md](./PACTO.md)\n" +
			"- [PLANTILLA_PACTO_PLAN.md](./PLANTILLA_PACTO_PLAN.md)\n" +
			"- [SLASH_COMMANDS.md](./SLASH_COMMANDS.md)\n"
	}
	return "# Pacto Plans\n\n" +
		"## Summary\n\n" +
		"Plan status is derived from plan folders and documents via `pacto status`.\n" +
		"This README is an overview, not a status source of truth.\n\n" +
		"---\n\n" +
		"## 📜 Pacto\n\n" +
		"- [PACTO.md](./PACTO.md)\n" +
		"- [PLANTILLA_PACTO_PLAN.md](./PLANTILLA_PACTO_PLAN.md)\n" +
		"- [SLASH_COMMANDS.md](./SLASH_COMMANDS.md)\n"
}

func defaultMinimalTemplate(lang i18n.Language) string {
	return tr(lang,
		"# Plan: <Title>\n\n## Metadata\n\n- Status: Draft\n- Owner: <team>\n- Created: <YYYY-MM-DD>\n- Last Modified: <YYYY-MM-DD>\n- State: <current|to-implement|done|outdated>\n- Slug: <slug>\n\n## Problem Statement\n\n<Describe the problem and scope.>\n\n## Goals\n\n1. <Goal 1>\n2. <Goal 2>\n\n## Non-Goals\n\n1. <Non-goal 1>\n\n## User Scenarios\n\n### Scenario: <name>\n\n- **GIVEN** <initial state>\n- **WHEN** <action>\n- **THEN** <outcome>\n\n## Functional Requirements\n\n- FR-001: The system MUST <verifiable capability>.\n\n## Non-Functional Requirements\n\n- NFR-001: <constraint or quality requirement>.\n\n## Acceptance Criteria\n\n- AC-001: <measurable outcome>.\n\n## Technical Context\n\n- Language/Version: <value>\n- Dependencies: <value>\n\n## Implementation Plan by Phases\n\n## Phase 1: <title>\n\n- [ ] 1.1 <task>\n\n## Evidence\n\n- <YYYY-MM-DD HH:MM> `<path|symbol|command>`\n\n## Risks and Mitigations\n\n1. Risk: <description> | Mitigation: <description>\n\n## Next Steps\n\n1. <next step>\n",
		"# Plan: <Título del plan>\n\n## Metadatos\n\n- Estado: Borrador\n- Owner: <nombre o equipo>\n- Creado: <YYYY-MM-DD>\n- Última Modificación: <YYYY-MM-DD>\n- Estado de Carpeta: <current|to-implement|done|outdated>\n- Slug: <slug>\n\n## Planteamiento del Problema\n\n<Describe el problema y su alcance.>\n\n## Objetivos\n\n1. <Objetivo 1>\n2. <Objetivo 2>\n\n## No Objetivos\n\n1. <No objetivo 1>\n\n## Escenarios de Usuario\n\n### Escenario: <nombre>\n\n- **GIVEN** <estado inicial>\n- **WHEN** <acción>\n- **THEN** <resultado>\n\n## Requerimientos Funcionales\n\n- FR-001: El sistema MUST <capacidad verificable>.\n\n## Requerimientos No Funcionales\n\n- NFR-001: <restricción o requisito de calidad>.\n\n## Criterios de Aceptación\n\n- AC-001: <resultado medible>.\n\n## Contexto Técnico\n\n- Lenguaje/Versión: <valor>\n- Dependencias: <valor>\n\n## Plan de Implementación por Fases\n\n## Phase 1: <título>\n\n- [ ] 1.1 <tarea>\n\n## Evidencia\n\n- <YYYY-MM-DD HH:MM> `<ruta|símbolo|comando>`\n\n## Riesgos y Mitigaciones\n\n1. Riesgo: <descripción> | Mitigación: <descripción>\n\n## Siguientes Pasos\n\n1. <siguiente paso>\n",
	)
}
