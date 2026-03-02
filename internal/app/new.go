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

	if err := updateRootIndex(req.root, req.state, req.slug, req.title, req.date, lang); err != nil {
		fmt.Fprintf(os.Stderr, "update root README: %v\n", err)
		return 3
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Created Plan", "Plan creado"), req.state+"/"+req.slug))
	fmt.Println(pathLine("created", req.readmePath))
	fmt.Println(pathLine("created", req.planPath))
	fmt.Println(pathLine("updated", filepath.Join(req.root, "README.md")))
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

	planText, err := buildPlanFromTemplate(req.root, req.title, req.date, req.owner, req.allowMinimal, lang)
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

func buildPlanFromTemplate(root, title, date, owner string, allowMinimal bool, lang i18n.Language) (string, error) {
	tplPath := filepath.Join(root, "PLANTILLA_PACTO_PLAN.md")
	b, err := os.ReadFile(tplPath)
	if err != nil {
		if !allowMinimal {
			return "", err
		}
		return defaultPlanTemplate(title, date, owner, lang), nil
	}
	t := string(b)
	t = strings.ReplaceAll(t, "<Title>", title)
	t = strings.ReplaceAll(t, "<Título del plan>", title)
	t = strings.ReplaceAll(t, "<YYYY-MM-DD>", date)
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

func updateRootIndex(root, state, slug, title, date string, lang i18n.Language) error {
	readmePath := filepath.Join(root, "README.md")
	b, err := os.ReadFile(readmePath)
	if err != nil {
		return err
	}
	text := string(b)

	counts, err := countPlans(root)
	if err != nil {
		return err
	}
	text = updateCountsTable(text, counts)
	text, err = upsertLinkInSection(text, state, title, fmt.Sprintf("./%s/%s/", state, slug))
	if err != nil {
		return err
	}
	text = updateLastUpdate(text, date, lang)

	return os.WriteFile(readmePath, []byte(text), 0o664)
}

func countPlans(root string) (map[string]int, error) {
	out := map[string]int{}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		dir := filepath.Join(root, st)
		ents, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		n := 0
		for _, e := range ents {
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(dir, e.Name(), "README.md")); err == nil {
				n++
			}
		}
		out[st] = n
	}
	return out, nil
}

func updateCountsTable(text string, counts map[string]int) string {
	repls := map[string]string{
		"| 🟢 **Current** |":      fmt.Sprintf("| 🟢 **Current** | %d |", counts["current"]),
		"| 🟡 **To Implement** |": fmt.Sprintf("| 🟡 **To Implement** | %d |", counts["to-implement"]),
		"| ✅ **Done** |":         fmt.Sprintf("| ✅ **Done** | %d |", counts["done"]),
		"| ⚠️ **Outdated** |":    fmt.Sprintf("| ⚠️ **Outdated** | %d |", counts["outdated"]),
	}
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		for prefix, rep := range repls {
			if strings.HasPrefix(strings.TrimSpace(ln), prefix) {
				lines[i] = rep
			}
		}
	}
	return strings.Join(lines, "\n")
}

func upsertLinkInSection(text, state, title, relPath string) (string, error) {
	entry := fmt.Sprintf("- [%s](%s)", title, relPath)
	lines := strings.Split(text, "\n")

	start, end, sep := findCanonicalSection(lines, state)
	if start < 0 {
		start, end, sep = findSectionByStateLink(lines, state)
	}
	if start < 0 {
		return addFallbackSection(lines, state, entry), nil
	}

	sec := append([]string{}, lines[start+1:sep]...)
	for _, ln := range sec {
		if strings.TrimSpace(ln) == strings.TrimSpace(entry) {
			return text, nil
		}
	}

	bullets := make([]string, 0)
	for _, ln := range sec {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "- [") {
			bullets = append(bullets, t)
		}
	}
	bullets = append(bullets, entry)
	sort.Strings(bullets)

	newSec := make([]string, 0, len(sec)+2)
	for _, ln := range sec {
		t := strings.TrimSpace(ln)
		if strings.HasPrefix(t, "- [") || strings.HasPrefix(t, "_No plans") || strings.HasPrefix(t, "_No hay planes") || strings.HasPrefix(t, "<!-- Add:") || strings.HasPrefix(t, "<!-- Añadir:") {
			continue
		}
		newSec = append(newSec, ln)
	}
	for len(newSec) > 0 && strings.TrimSpace(newSec[len(newSec)-1]) == "" {
		newSec = newSec[:len(newSec)-1]
	}
	if len(newSec) > 0 {
		newSec = append(newSec, "")
	}
	newSec = append(newSec, bullets...)

	out := make([]string, 0, len(lines)+4)
	out = append(out, lines[:start+1]...)
	out = append(out, newSec...)
	out = append(out, lines[sep:]...)
	if end > sep {
		_ = end // keep shape explicit; sep drives splice point
	}
	return strings.Join(out, "\n"), nil
}

func findCanonicalSection(lines []string, state string) (start, end, sep int) {
	candidates := map[string][]string{
		"current":      {"## 🟢 Current (En Ejecución)", "## 🟢 Current (In Progress)", "## 🟢 Current"},
		"to-implement": {"## 🟡 To Implement (Pendientes)", "## 🟡 To Implement (Pending)", "## 🟡 To Implement"},
		"done":         {"## ✅ Done (Completados)", "## ✅ Done (Completed)", "## ✅ Done"},
		"outdated":     {"## ⚠️ Outdated (Obsoletos)", "## ⚠️ Outdated (Outdated)", "## ⚠️ Outdated"},
	}[state]
	if len(candidates) == 0 {
		return -1, -1, -1
	}
	start = -1
	for i, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		for _, heading := range candidates {
			if trimmed == heading {
				start = i
				break
			}
		}
		if start >= 0 {
			break
		}
	}
	if start < 0 {
		return -1, -1, -1
	}
	end = len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "## ") {
			end = i
			break
		}
	}
	sep = end
	for i := start + 1; i < end; i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			sep = i
			break
		}
	}
	return start, end, sep
}

func findSectionByStateLink(lines []string, state string) (start, end, sep int) {
	needle := "./" + state + "/"
	for i := 0; i < len(lines); i++ {
		if !strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			continue
		}
		end = len(lines)
		for j := i + 1; j < len(lines); j++ {
			if strings.HasPrefix(strings.TrimSpace(lines[j]), "## ") {
				end = j
				break
			}
		}
		for j := i + 1; j < end; j++ {
			if strings.Contains(lines[j], needle) {
				sep = end
				for k := i + 1; k < end; k++ {
					if strings.TrimSpace(lines[k]) == "---" {
						sep = k
						break
					}
				}
				return i, end, sep
			}
		}
	}
	return -1, -1, -1
}

func addFallbackSection(lines []string, state, entry string) string {
	heading := "## Plans (" + state + ")"
	insertAt := len(lines)
	for i, ln := range lines {
		if strings.TrimSpace(ln) == "## 📜 Pacto" {
			insertAt = i
			break
		}
	}
	block := []string{"", heading, entry, "---"}
	out := make([]string, 0, len(lines)+len(block))
	out = append(out, lines[:insertAt]...)
	out = append(out, block...)
	out = append(out, lines[insertAt:]...)
	return strings.Join(out, "\n")
}

func updateLastUpdate(text, date string, lang i18n.Language) string {
	lines := strings.Split(text, "\n")
	for i, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "**Last Updated:**") || strings.HasPrefix(trimmed, "**Última Actualización:**") {
			lines[i] = tr(lang, "**Last Updated:** ", "**Última Actualización:** ") + date
			return strings.Join(lines, "\n")
		}
	}
	return text + "\n\n" + tr(lang, "**Last Updated:** ", "**Última Actualización:** ") + date + "\n"
}

func defaultPlanTemplate(title, date, owner string, lang i18n.Language) string {
	var b strings.Builder
	b.WriteString("# Plan: " + title + "\n\n")
	b.WriteString(tr(lang, "**Version:** 1.0  \n", "**Versión:** 1.0  \n"))
	b.WriteString(tr(lang, "**Date:** ", "**Fecha:** ") + date + "  \n")
	b.WriteString(tr(lang, "**Status:** Draft  \n", "**Estado:** Borrador  \n"))
	b.WriteString("**Owner:** " + owner + "\n\n")
	b.WriteString(tr(lang, "## Summary\n\n", "## Resumen\n\n"))
	b.WriteString(tr(lang, "Plan scaffold generated by pacto CLI.\n", "Plantilla de plan generada por la CLI de pacto.\n"))
	return b.String()
}

func defaultRootReadme(lang i18n.Language) string {
	if lang == i18n.Spanish {
		return "# Planes de Pacto\n\n" +
			"## Resumen\n\n" +
			"| Estado | Cantidad |\n" +
			"|-------|-------|\n" +
			"| 🟢 **Current** | 0 |\n" +
			"| 🟡 **To Implement** | 0 |\n" +
			"| ✅ **Done** | 0 |\n" +
			"| ⚠️ **Outdated** | 0 |\n\n" +
			"---\n\n" +
			"## 🟢 Current (En Ejecución)\n_No hay planes._\n\n---\n\n" +
			"## 🟡 To Implement (Pendientes)\n_No hay planes._\n\n---\n\n" +
			"## ✅ Done (Completados)\n_No hay planes._\n\n---\n\n" +
			"## ⚠️ Outdated (Obsoletos)\n_No hay planes._\n\n---\n\n" +
			"## 📜 Pacto\n\n" +
			"- [PACTO.md](./PACTO.md)\n" +
			"- [PLANTILLA_PACTO_PLAN.md](./PLANTILLA_PACTO_PLAN.md)\n\n" +
			"---\n\n" +
			"**Última Actualización:** 1970-01-01\n"
	}
	return "# Pacto Plans\n\n" +
		"## Summary\n\n" +
		"| State | Count |\n" +
		"|-------|-------|\n" +
		"| 🟢 **Current** | 0 |\n" +
		"| 🟡 **To Implement** | 0 |\n" +
		"| ✅ **Done** | 0 |\n" +
		"| ⚠️ **Outdated** | 0 |\n\n" +
		"---\n\n" +
		"## 🟢 Current (In Progress)\n_No plans._\n\n---\n\n" +
		"## 🟡 To Implement (Pending)\n_No plans._\n\n---\n\n" +
		"## ✅ Done (Completed)\n_No plans._\n\n---\n\n" +
		"## ⚠️ Outdated (Outdated)\n_No plans._\n\n---\n\n" +
		"## 📜 Pacto\n\n" +
		"- [PACTO.md](./PACTO.md)\n" +
		"- [PLANTILLA_PACTO_PLAN.md](./PLANTILLA_PACTO_PLAN.md)\n\n" +
		"---\n\n" +
		"**Last Updated:** 1970-01-01\n"
}

func defaultMinimalTemplate(lang i18n.Language) string {
	return tr(lang,
		"# Plan: <Title>\n\n**Version:** 1.0  \n**Date:** <YYYY-MM-DD>  \n**Status:** <Draft | In Progress | Completed | Blocked>  \n**Owner:** <team>\n",
		"# Plan: <Título del plan>\n\n**Versión:** 1.0  \n**Fecha:** <YYYY-MM-DD>  \n**Estado:** <Draft | En ejecución | Completado | Bloqueado>  \n**Owner:** <nombre o equipo>\n",
	)
}
