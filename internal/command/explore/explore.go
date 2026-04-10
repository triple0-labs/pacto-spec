package explorecmd

import (
	"fmt"
	"os"
	"strings"

	"pacto/internal/cmdutil"
	"pacto/internal/explore"
	"pacto/internal/i18n"
	"pacto/internal/ui"
	"pacto/internal/workspace"
)

type Options struct {
	Root  string
	Title string
	Note  string
	List  bool
	Show  string
}

func Run(opts Options, pos []string) int {
	root, err := workspace.ResolveExploreRoot(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	lang := cmdutil.EffectiveLanguage(root)
	if opts.List && strings.TrimSpace(opts.Show) != "" {
		fmt.Fprintln(os.Stderr, "cannot use --list with --show")
		return 2
	}
	if opts.List && (strings.TrimSpace(opts.Title) != "" || strings.TrimSpace(opts.Note) != "") {
		fmt.Fprintln(os.Stderr, "--list does not accept --title or --note")
		return 2
	}
	if strings.TrimSpace(opts.Show) != "" && strings.TrimSpace(opts.Title) != "" {
		fmt.Fprintln(os.Stderr, "--show does not accept --title")
		return 2
	}

	switch {
	case opts.List:
		return runExploreList(root, lang)
	case strings.TrimSpace(opts.Show) != "":
		return runExploreShow(root, strings.TrimSpace(opts.Show), lang)
	default:
		if len(pos) != 1 {
			fmt.Fprintln(os.Stderr, tr(lang, "explore requires a slug, or use --list/--show", "explore requiere un slug, o usar --list/--show"))
			return 2
		}
		return runExploreCreateOrUpdate(root, pos[0], opts.Title, opts.Note, lang)
	}
}

func runExploreCreateOrUpdate(root, slug, title, note string, lang i18n.Language) int {
	action, readmePath, err := explore.CreateOrUpdate(root, slug, title, note, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "explore: %v\n", err)
		return 3
	}

	switch action {
	case "created":
		fmt.Println(ui.ActionHeader(cmdutil.Tr(lang, "Created Idea", "Idea creada"), slug))
		fmt.Println(cmdutil.PathLine("created", readmePath))
	case "updated":
		fmt.Println(ui.ActionHeader(cmdutil.Tr(lang, "Updated Idea", "Idea actualizada"), slug))
		fmt.Println(cmdutil.PathLine("updated", readmePath))
	case "skipped":
		fmt.Println(ui.ActionHeader(cmdutil.Tr(lang, "Idea Exists", "Idea existente"), slug))
		fmt.Println(cmdutil.PathLine("skipped", readmePath))
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
		fmt.Println(ui.Dim(cmdutil.Tr(lang, "No ideas found.", "No se encontraron ideas.")))
		return 0
	}

	fmt.Println(ui.Title(cmdutil.Tr(lang, "Ideas", "Ideas")))
	fmt.Println("")
	for _, r := range ideas {
		fmt.Printf("%s\n", ui.Bullet(r.Slug))
		fmt.Printf("  %s: %s\n", cmdutil.Tr(lang, "title", "título"), r.Title)
		fmt.Printf("  %s: %s\n", cmdutil.Tr(lang, "created", "creado"), r.CreatedAt)
		fmt.Printf("  %s: %s\n", cmdutil.Tr(lang, "updated", "actualizado"), r.UpdatedAt)
	}
	return 0
}

func runExploreShow(root, slug string, lang i18n.Language) int {
	idea, err := explore.GetIdea(root, slug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "show idea: %v\n", err)
		return 2
	}

	fmt.Printf("%s %s\n", ui.Title(cmdutil.Tr(lang, "Idea", "Idea")), idea.Slug)
	fmt.Printf("%s: %s\n", cmdutil.Tr(lang, "Path", "Ruta"), cmdutil.DisplayPath(idea.Path))
	fmt.Printf("%s: %s\n", cmdutil.Tr(lang, "Title", "Título"), idea.Title)
	fmt.Printf("%s: %s\n", cmdutil.Tr(lang, "Created At", "Creado"), idea.CreatedAt)
	fmt.Printf("%s: %s\n", cmdutil.Tr(lang, "Updated At", "Actualizado"), idea.UpdatedAt)
	return 0
}
