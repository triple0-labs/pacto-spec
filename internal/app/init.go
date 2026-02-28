package app

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/assets"
)

const (
	agentsManagedStart = "<!-- pacto:init:start -->"
	agentsManagedEnd   = "<!-- pacto:init:end -->"
)

func RunInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage:")
		fmt.Fprintln(os.Stderr, "  pacto init [--root .] [--with-agents] [--force]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Options:")
		fs.PrintDefaults()
	}

	root := fs.String("root", ".", "Project root path")
	withAgents := fs.Bool("with-agents", false, "Create or update optional AGENTS.md hand-off block (canonical guidance remains in PACTO.md)")
	force := fs.Bool("force", false, "Overwrite init-managed files when they already exist")
	lang := fs.String("lang", "", "Deprecated: ignored, CLI output is English-only")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if strings.TrimSpace(*lang) != "" || hasLangArg(args) {
		fmt.Fprintln(os.Stderr, "warning: --lang is deprecated and ignored; CLI output is English-only")
	}
	if len(fs.Args()) > 0 {
		fs.Usage()
		return 2
	}

	projectRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	plansRoot := filepath.Join(projectRoot, ".pacto", "plans")

	var created, updated, skipped []string
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		p := filepath.Join(plansRoot, st)
		if info, err := os.Stat(p); err == nil {
			if !info.IsDir() {
				fmt.Fprintf(os.Stderr, "state path exists but is not a directory: %s\n", p)
				return 3
			}
			skipped = append(skipped, p)
			continue
		}
		if err := os.MkdirAll(p, 0o775); err != nil {
			fmt.Fprintf(os.Stderr, "create state dir %q: %v\n", p, err)
			return 3
		}
		created = append(created, p)
	}

	workspaceFiles := map[string]string{
		filepath.Join(plansRoot, "README.md"):               assets.MustTemplate("README.md"),
		filepath.Join(plansRoot, "PACTO.md"):                assets.MustTemplate("PACTO.md"),
		filepath.Join(plansRoot, "PLANTILLA_PACTO_PLAN.md"): assets.MustTemplate("PLANTILLA_PACTO_PLAN.md"),
		filepath.Join(plansRoot, "SLASH_COMMANDS.md"):       assets.MustTemplate("SLASH_COMMANDS.md"),
	}

	for path, content := range workspaceFiles {
		wc, wu, ws, werr := writeManagedFile(path, content, *force)
		if werr != nil {
			fmt.Fprintf(os.Stderr, "write file %q: %v\n", path, werr)
			return 3
		}
		if wc {
			created = append(created, path)
		}
		if wu {
			updated = append(updated, path)
		}
		if ws {
			skipped = append(skipped, path)
		}
	}

	if *withAgents {
		agentsPath := filepath.Join(projectRoot, "AGENTS.md")
		act, aerr := writeAgentsManagedBlock(agentsPath, assets.MustTemplate("AGENTS.md"))
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "update AGENTS.md: %v\n", aerr)
			return 3
		}
		switch act {
		case "created":
			created = append(created, agentsPath)
		case "updated":
			updated = append(updated, agentsPath)
		case "skipped":
			skipped = append(skipped, agentsPath)
		}
	}

	sort.Strings(created)
	sort.Strings(updated)
	sort.Strings(skipped)

	fmt.Printf("Initialized Pacto workspace: %s\n", plansRoot)
	fmt.Printf("Created: %d  Updated: %d  Skipped: %d\n", len(created), len(updated), len(skipped))
	for _, p := range created {
		fmt.Printf("+ %s\n", p)
	}
	for _, p := range updated {
		fmt.Printf("~ %s\n", p)
	}
	for _, p := range skipped {
		fmt.Printf("= %s\n", p)
	}
	return 0
}

func writeManagedFile(path, content string, force bool) (created, updated, skipped bool, err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return false, false, false, err
	}
	_, statErr := os.Stat(path)
	exists := statErr == nil
	if exists && !force {
		return false, false, true, nil
	}
	if err := os.WriteFile(path, []byte(content), 0o664); err != nil {
		return false, false, false, err
	}
	if exists {
		return false, true, false, nil
	}
	return true, false, false, nil
}

func writeAgentsManagedBlock(path, template string) (string, error) {
	block := agentsManagedStart + "\n" + strings.TrimSpace(template) + "\n" + agentsManagedEnd + "\n"

	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(block), 0o664); err != nil {
				return "", err
			}
			return "created", nil
		}
		return "", err
	}

	s := string(b)
	start := strings.Index(s, agentsManagedStart)
	end := strings.Index(s, agentsManagedEnd)
	if start >= 0 && end >= 0 && end > start {
		end += len(agentsManagedEnd)
		next := s[:start] + block + s[end:]
		if next == s {
			return "skipped", nil
		}
		if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
			return "", err
		}
		return "updated", nil
	}

	trimmed := strings.TrimRight(s, "\n")
	next := trimmed + "\n\n" + block
	if next == s {
		return "skipped", nil
	}
	if err := os.WriteFile(path, []byte(next), 0o664); err != nil {
		return "", err
	}
	return "updated", nil
}
