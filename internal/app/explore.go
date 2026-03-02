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

var (
	reCreatedAt = regexp.MustCompile(`(?m)^\*\*(?:Created At|Creado):\*\*\s*(.+)$`)
	reUpdatedAt = regexp.MustCompile(`(?m)^\*\*(?:Updated At|Actualizado):\*\*\s*(.+)$`)
)

type exploreOptions struct {
	root  string
	title string
	note  string
	list  bool
	show  string
}

func RunExplore(args []string) int {
	opts, pos, code, ok := parseExploreArgs(args)
	if !ok {
		return code
	}

	root, err := resolveExploreRoot(opts.root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	lang := effectiveLanguage(root)

	switch {
	case opts.list:
		return runExploreList(root, lang)
	case strings.TrimSpace(opts.show) != "":
		return runExploreShow(root, strings.TrimSpace(opts.show), lang)
	default:
		if len(pos) != 1 {
			fmt.Fprintln(os.Stderr, tr(lang, "explore requires a slug, or use --list/--show", "explore requiere un slug, o usar --list/--show"))
			return 2
		}
		return runExploreCreateOrUpdate(root, pos[0], opts.title, opts.note, lang)
	}
}

func parseExploreArgs(args []string) (exploreOptions, []string, int, bool) {
	opts := exploreOptions{}
	fs := flag.NewFlagSet("explore", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto explore <slug> [--title <title>] [--note <note>] [--root <path>]")
		fmt.Fprintln(os.Stderr, "  pacto explore --list [--root <path>]")
		fmt.Fprintln(os.Stderr, "  pacto explore --show <slug> [--root <path>]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	fs.StringVar(&opts.root, "root", "", "Project root path (defaults to auto-detected project root or current directory)")
	fs.StringVar(&opts.title, "title", "", "Optional idea title")
	fs.StringVar(&opts.note, "note", "", "Append exploration note and refresh update timestamp")
	fs.BoolVar(&opts.list, "list", false, "List saved ideas")
	fs.StringVar(&opts.show, "show", "", "Show a saved idea by slug")

	normalizedArgs, normErr := normalizeExploreArgs(args)
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return exploreOptions{}, nil, 2, false
	}

	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return exploreOptions{}, nil, 0, false
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return exploreOptions{}, nil, 2, false
	}

	if opts.list && strings.TrimSpace(opts.show) != "" {
		fmt.Fprintln(os.Stderr, "cannot use --list with --show")
		return exploreOptions{}, nil, 2, false
	}
	if opts.list && (strings.TrimSpace(opts.title) != "" || strings.TrimSpace(opts.note) != "") {
		fmt.Fprintln(os.Stderr, "--list does not accept --title or --note")
		return exploreOptions{}, nil, 2, false
	}
	if strings.TrimSpace(opts.show) != "" && strings.TrimSpace(opts.title) != "" {
		fmt.Fprintln(os.Stderr, "--show does not accept --title")
		return exploreOptions{}, nil, 2, false
	}

	return opts, fs.Args(), 0, true
}

func normalizeExploreArgs(args []string) ([]string, error) {
	withValue := map[string]bool{"--root": true, "-root": true, "--title": true, "-title": true, "--note": true, "-note": true, "--show": true, "-show": true}
	return normalizeArgs(args, withValue)
}

func resolveExploreRoot(rawRoot string) (string, error) {
	if strings.TrimSpace(rawRoot) != "" {
		return filepath.Abs(rawRoot)
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if _, projectRoot, ok := resolvePlanRootFrom(cwd); ok {
		return projectRoot, nil
	}
	return cwd, nil
}

func runExploreCreateOrUpdate(root, slug, title, note string, lang i18n.Language) int {
	slug = strings.TrimSpace(slug)
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return 2
	}

	ideasRoot := filepath.Join(root, ".pacto", "ideas")
	ideaDir := filepath.Join(ideasRoot, slug)
	readmePath := filepath.Join(ideaDir, "README.md")
	now := time.Now().Format("2006-01-02 15:04")

	if err := os.MkdirAll(ideaDir, 0o775); err != nil {
		fmt.Fprintf(os.Stderr, "create idea dir: %v\n", err)
		return 3
	}

	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		ideaTitle := strings.TrimSpace(title)
		if ideaTitle == "" {
			ideaTitle = slugToTitle(slug)
		}
		text := buildExploreReadme(ideaTitle, now, lang)
		if strings.TrimSpace(note) != "" {
			text = appendExploreNote(text, strings.TrimSpace(note), now)
		}
		if err := os.WriteFile(readmePath, []byte(text), 0o664); err != nil {
			fmt.Fprintf(os.Stderr, "write idea readme: %v\n", err)
			return 3
		}
		fmt.Println(ui.ActionHeader(tr(lang, "Created Idea", "Idea creada"), slug))
		fmt.Println(pathLine("created", readmePath))
		return 0
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "stat idea readme: %v\n", err)
		return 3
	}

	if strings.TrimSpace(note) == "" {
		fmt.Println(ui.ActionHeader(tr(lang, "Idea Exists", "Idea existente"), slug))
		fmt.Println(pathLine("skipped", readmePath))
		return 0
	}

	b, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read idea readme: %v\n", err)
		return 3
	}
	updated := appendExploreNote(string(b), strings.TrimSpace(note), now)
	updated = setUpdatedAt(updated, now)
	if err := os.WriteFile(readmePath, []byte(updated), 0o664); err != nil {
		fmt.Fprintf(os.Stderr, "update idea readme: %v\n", err)
		return 3
	}
	fmt.Println(ui.ActionHeader(tr(lang, "Updated Idea", "Idea actualizada"), slug))
	fmt.Println(pathLine("updated", readmePath))
	return 0
}

func runExploreList(root string, lang i18n.Language) int {
	ideasRoot := filepath.Join(root, ".pacto", "ideas")
	ents, err := os.ReadDir(ideasRoot)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println(ui.Dim(tr(lang, "No ideas found.", "No se encontraron ideas.")))
			return 0
		}
		fmt.Fprintf(os.Stderr, "read ideas: %v\n", err)
		return 3
	}

	type row struct {
		slug      string
		title     string
		createdAt string
		updatedAt string
	}
	rows := make([]row, 0)
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		readmePath := filepath.Join(ideasRoot, e.Name(), "README.md")
		b, err := os.ReadFile(readmePath)
		if err != nil {
			continue
		}
		content := string(b)
		rows = append(rows, row{
			slug:      e.Name(),
			title:     extractTitle(content),
			createdAt: extractStamp(reCreatedAt, content),
			updatedAt: extractStamp(reUpdatedAt, content),
		})
	}

	if len(rows) == 0 {
		fmt.Println(ui.Dim(tr(lang, "No ideas found.", "No se encontraron ideas.")))
		return 0
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].slug < rows[j].slug })
	fmt.Println(ui.Title(tr(lang, "Ideas", "Ideas")))
	fmt.Println("")
	for _, r := range rows {
		fmt.Printf("%s\n", ui.Bullet(r.slug))
		fmt.Printf("  %s: %s\n", tr(lang, "title", "título"), r.title)
		fmt.Printf("  %s: %s\n", tr(lang, "created", "creado"), r.createdAt)
		fmt.Printf("  %s: %s\n", tr(lang, "updated", "actualizado"), r.updatedAt)
	}
	return 0
}

func runExploreShow(root, slug string, lang i18n.Language) int {
	slug = strings.TrimSpace(slug)
	if !slugRe.MatchString(slug) {
		fmt.Fprintf(os.Stderr, "invalid slug %q (use lowercase letters, numbers, dashes)\n", slug)
		return 2
	}
	readmePath := filepath.Join(root, ".pacto", "ideas", slug, "README.md")
	b, err := os.ReadFile(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "idea not found: %s\n", slug)
			return 2
		}
		fmt.Fprintf(os.Stderr, "read idea: %v\n", err)
		return 3
	}
	content := string(b)
	fmt.Printf("%s %s\n", ui.Title(tr(lang, "Idea", "Idea")), slug)
	fmt.Printf("%s: %s\n", tr(lang, "Path", "Ruta"), displayPath(readmePath))
	fmt.Printf("%s: %s\n", tr(lang, "Title", "Título"), extractTitle(content))
	fmt.Printf("%s: %s\n", tr(lang, "Created At", "Creado"), extractStamp(reCreatedAt, content))
	fmt.Printf("%s: %s\n", tr(lang, "Updated At", "Actualizado"), extractStamp(reUpdatedAt, content))
	return 0
}

func buildExploreReadme(title, now string, lang i18n.Language) string {
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString(tr(lang, "**Created At:** ", "**Creado:** ") + now + "  \n")
	b.WriteString(tr(lang, "**Updated At:** ", "**Actualizado:** ") + now + "\n\n")
	b.WriteString(tr(lang, "## Summary\n\n", "## Resumen\n\n"))
	b.WriteString(tr(lang, "Idea exploration workspace.\n\n", "Espacio de exploración de ideas.\n\n"))
	b.WriteString(tr(lang, "## Notes\n\n", "## Notas\n\n"))
	b.WriteString("- [" + now + "] " + tr(lang, "Idea created.", "Idea creada.") + "\n")
	return b.String()
}

func appendExploreNote(content, note, now string) string {
	if !strings.Contains(content, "## Notes") && !strings.Contains(content, "## Notas") {
		lang := effectiveLanguage(".")
		content = strings.TrimRight(content, "\n") + "\n\n" + tr(lang, "## Notes\n\n", "## Notas\n\n")
	}
	content = strings.TrimRight(content, "\n")
	return content + "\n- [" + now + "] " + note + "\n"
}

func setUpdatedAt(content, now string) string {
	if reUpdatedAt.MatchString(content) {
		return reUpdatedAt.ReplaceAllString(content, "**Updated At:** "+now)
	}
	updatedLabel := "**Updated At:** "
	if strings.Contains(content, "**Creado:**") || strings.Contains(content, "## Notas") {
		updatedLabel = "**Actualizado:** "
	}
	createdLabel := "**Created At:** "
	if updatedLabel == "**Actualizado:** " {
		createdLabel = "**Creado:** "
	}
	title := extractTitle(content)
	head := "# " + title + "\n\n" + createdLabel + extractStamp(reCreatedAt, content) + "  \n" + updatedLabel + now + "\n"
	body := content
	if idx := strings.Index(content, "\n"); idx >= 0 {
		body = content[idx+1:]
	}
	return strings.TrimRight(head+"\n"+body, "\n") + "\n"
}

func extractTitle(content string) string {
	for _, ln := range strings.Split(content, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(ln, "# "))
		}
	}
	return "Untitled"
}

func extractStamp(re *regexp.Regexp, content string) string {
	m := re.FindStringSubmatch(content)
	if len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return "-"
}
