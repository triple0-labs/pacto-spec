package movecmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/cmdutil"
	plancontext "pacto/internal/context"
	"pacto/internal/app/move"
	"pacto/internal/ui"
	"pacto/internal/workspace"
)

type Options struct {
	Root   string
	Reason string
	Force  bool
}

func Run(opts Options, pos []string) int {
	if len(pos) != 3 {
		fmt.Fprintln(os.Stderr, "move requires <from-state> <slug> <to-state>")
		return 2
	}

	fromState := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])
	toState := strings.ToLower(strings.TrimSpace(pos[2]))

	plansRoot, err := workspace.ResolvePlansRootForAction(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}

	lang := cmdutil.EffectiveLanguage(filepath.Dir(plansRoot))

	readmePath, err := move.MovePlan(plansRoot, fromState, slug, toState, opts.Force, opts.Reason, lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "move error: %v\n", err)
		return 3
	}

	fmt.Println(ui.ActionHeader(move.Tr(lang, "Moved Plan", "Plan movido"), fmt.Sprintf("%s/%s -> %s/%s", fromState, slug, toState, slug)))
	fmt.Println(cmdutil.PathLine("updated", readmePath))

	if toState == "done" {
		planDir := filepath.Join(plansRoot, toState, slug)
		specPath := filepath.Join(planDir, "spec.md")
		domains := plancontext.ExtractDomains(specPath)
		if len(domains) > 0 {
			if err := plancontext.InitContext(plansRoot); err != nil {
				fmt.Fprintf(os.Stderr, "move error: init context: %v\n", err)
				return 3
			}
			contextDir := plancontext.ContextDirFromPlansRoot(plansRoot)
			docActions := map[string]string{}
			for _, domain := range domains {
				docPath := plancontext.DomainFolderPath(contextDir, domain)
				if _, err := os.Stat(docPath); err == nil {
					docActions[docPath] = "exists"
				} else {
					docActions[docPath] = "created"
				}
			}
			if err := plancontext.EnsureDomainFolders(contextDir, domains); err != nil {
				fmt.Fprintf(os.Stderr, "move error: update context domains: %v\n", err)
				return 3
			}
			docPaths := make([]string, 0, len(docActions))
			for path := range docActions {
				docPaths = append(docPaths, path)
			}
			sort.Strings(docPaths)
			for _, path := range docPaths {
				fmt.Println(cmdutil.PathLine(docActions[path], path))
			}

			docNames := make([]string, 0, len(docPaths))
			for _, path := range docPaths {
				docNames = append(docNames, filepath.Base(path)+"/")
			}
			fmt.Println(move.Tr(
				lang,
				fmt.Sprintf("This plan may contain decisions or constraints worth preserving. Review design.md and update the affected domain folders under `.pacto/context/domains/`: %s", strings.Join(docNames, ", ")),
				fmt.Sprintf("Este plan puede contener decisiones o restricciones que conviene preservar. Revisa design.md y actualiza las carpetas de dominio afectadas en `.pacto/context/domains/`: %s", strings.Join(docNames, ", ")),
			))
		}
	}
	return 0
}
