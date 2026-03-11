package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/explore"
	"pacto/internal/i18n"
	"pacto/internal/ui"
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
	return runExplore(opts, pos)
}

func runExplore(opts exploreOptions, pos []string) int {
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
	action, readmePath, err := explore.CreateOrUpdate(root, slug, title, note, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explore: %v\n", err)
		return 3
	}

	switch action {
	case "created":
		fmt.Println(ui.ActionHeader(tr(lang, "Created Idea", "Idea creada"), slug))
		fmt.Println(pathLine("created", readmePath))
	case "updated":
		fmt.Println(ui.ActionHeader(tr(lang, "Updated Idea", "Idea actualizada"), slug))
		fmt.Println(pathLine("updated", readmePath))
	case "skipped":
		fmt.Println(ui.ActionHeader(tr(lang, "Idea Exists", "Idea existente"), slug))
		fmt.Println(pathLine("skipped", readmePath))
	}

	return 0
}

func runExploreList(root string, lang i18n.Language) int {
	ideas, err := explore.ListIdeas(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read ideas: %v\n", err)
		return 3
	}

	if len(ideas) == 0 {
		fmt.Println(ui.Dim(tr(lang, "No ideas found.", "No se encontraron ideas.")))
		return 0
	}

	fmt.Println(ui.Title(tr(lang, "Ideas", "Ideas")))
	fmt.Println("")
	for _, r := range ideas {
		fmt.Printf("%s\n", ui.Bullet(r.Slug))
		fmt.Printf("  %s: %s\n", tr(lang, "title", "título"), r.Title)
		fmt.Printf("  %s: %s\n", tr(lang, "created", "creado"), r.CreatedAt)
		fmt.Printf("  %s: %s\n", tr(lang, "updated", "actualizado"), r.UpdatedAt)
	}
	return 0
}

func runExploreShow(root, slug string, lang i18n.Language) int {
	idea, err := explore.GetIdea(root, slug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "show idea: %v\n", err)
		return 2
	}

	fmt.Printf("%s %s\n", ui.Title(tr(lang, "Idea", "Idea")), idea.Slug)
	fmt.Printf("%s: %s\n", tr(lang, "Path", "Ruta"), displayPath(idea.Path))
	fmt.Printf("%s: %s\n", tr(lang, "Title", "Título"), idea.Title)
	fmt.Printf("%s: %s\n", tr(lang, "Created At", "Creado"), idea.CreatedAt)
	fmt.Printf("%s: %s\n", tr(lang, "Updated At", "Actualizado"), idea.UpdatedAt)
	return 0
}
