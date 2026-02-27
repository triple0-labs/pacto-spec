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
)

var (
	reCreatedAt = regexp.MustCompile(`(?m)^\*\*Created At:\*\*\s*(.+)$`)
	reUpdatedAt = regexp.MustCompile(`(?m)^\*\*Updated At:\*\*\s*(.+)$`)
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

	switch {
	case opts.list:
		return runExploreList(root)
	case strings.TrimSpace(opts.show) != "":
		return runExploreShow(root, strings.TrimSpace(opts.show))
	default:
		if len(pos) != 1 {
			fmt.Fprintln(os.Stderr, "explore requires a slug, or use --list/--show")
			return 2
		}
		return runExploreCreateOrUpdate(root, pos[0], opts.title, opts.note)
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
	flags := make([]string, 0, len(args))
	pos := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "--") || strings.HasPrefix(a, "-") {
			if withValue[a] {
				if i+1 >= len(args) {
					return nil, fmt.Errorf("flag %s expects a value", a)
				}
				flags = append(flags, a, args[i+1])
				i++
				continue
			}
			flags = append(flags, a)
			continue
		}
		pos = append(pos, a)
	}
	return append(flags, pos...), nil
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

func runExploreCreateOrUpdate(root, slug, title, note string) int {
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
		text := buildExploreReadme(ideaTitle, now)
		if strings.TrimSpace(note) != "" {
			text = appendExploreNote(text, strings.TrimSpace(note), now)
		}
		if err := os.WriteFile(readmePath, []byte(text), 0o664); err != nil {
			fmt.Fprintf(os.Stderr, "write idea readme: %v\n", err)
			return 3
		}
		fmt.Printf("Created idea: %s\n", slug)
		fmt.Printf("- %s\n", readmePath)
		return 0
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "stat idea readme: %v\n", err)
		return 3
	}

	if strings.TrimSpace(note) == "" {
		fmt.Printf("Idea already exists: %s\n", slug)
		fmt.Printf("- %s\n", readmePath)
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
	fmt.Printf("Updated idea: %s\n", slug)
	fmt.Printf("- %s\n", readmePath)
	return 0
}

func runExploreList(root string) int {
	ideasRoot := filepath.Join(root, ".pacto", "ideas")
	ents, err := os.ReadDir(ideasRoot)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No ideas found.")
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
		fmt.Println("No ideas found.")
		return 0
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].slug < rows[j].slug })
	fmt.Println("Ideas")
	fmt.Println("")
	for _, r := range rows {
		fmt.Printf("- %s\n", r.slug)
		fmt.Printf("  title: %s\n", r.title)
		fmt.Printf("  created: %s\n", r.createdAt)
		fmt.Printf("  updated: %s\n", r.updatedAt)
	}
	return 0
}

func runExploreShow(root, slug string) int {
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
	fmt.Printf("Idea: %s\n", slug)
	fmt.Printf("Path: %s\n", readmePath)
	fmt.Printf("Title: %s\n", extractTitle(content))
	fmt.Printf("Created At: %s\n", extractStamp(reCreatedAt, content))
	fmt.Printf("Updated At: %s\n", extractStamp(reUpdatedAt, content))
	return 0
}

func buildExploreReadme(title, now string) string {
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")
	b.WriteString("**Created At:** " + now + "  \n")
	b.WriteString("**Updated At:** " + now + "\n\n")
	b.WriteString("## Summary\n\n")
	b.WriteString("Idea exploration workspace.\n\n")
	b.WriteString("## Notes\n\n")
	b.WriteString("- [" + now + "] Idea created.\n")
	return b.String()
}

func appendExploreNote(content, note, now string) string {
	if !strings.Contains(content, "## Notes") {
		content = strings.TrimRight(content, "\n") + "\n\n## Notes\n\n"
	}
	content = strings.TrimRight(content, "\n")
	return content + "\n- [" + now + "] " + note + "\n"
}

func setUpdatedAt(content, now string) string {
	if reUpdatedAt.MatchString(content) {
		return reUpdatedAt.ReplaceAllString(content, "**Updated At:** "+now)
	}
	title := extractTitle(content)
	head := "# " + title + "\n\n**Created At:** " + extractStamp(reCreatedAt, content) + "  \n**Updated At:** " + now + "\n"
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
