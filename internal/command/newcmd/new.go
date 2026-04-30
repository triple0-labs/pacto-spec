package newcmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"pacto/internal/app/newplan"
	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/ui"
)

// Options mirrors the CLI flag surface for `pacto new`. RootProvided records
// whether --root was explicitly supplied by the user (so the use case knows
// whether to anchor at that directory or walk upward).
type Options struct {
	Root         string
	Title        string
	Owner        string
	AllowMinimal bool
	Lang         string
	RootProvided bool
}

// Run is the thin Cobra-side wrapper: it validates the language flag, applies
// any global language override for downstream packages, delegates to the
// newplan use case, and prints user-facing output.
func Run(opts Options, state, slug string) int {
	if strings.TrimSpace(opts.Lang) != "" {
		if _, ok := i18n.ParseLanguage(opts.Lang); !ok {
			fmt.Fprintf(os.Stderr, "invalid --lang value %q (allowed: en|es)\n", opts.Lang)
			return 2
		}
		cmdutil.SetGlobalLangOverride(opts.Lang)
		defer cmdutil.SetGlobalLangOverride("")
	}
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}

	lang := cmdutil.EffectiveLanguage(opts.Root)
	res, err := newplan.Create(newplan.Input{
		Root:         opts.Root,
		RootProvided: opts.RootProvided,
		State:        state,
		Slug:         slug,
		Title:        opts.Title,
		Owner:        opts.Owner,
		AllowMinimal: opts.AllowMinimal,
		Lang:         lang,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if errors.Is(err, newplan.ErrInvalid) {
			return 2
		}
		return 3
	}

	// Re-evaluate language now that the plan root may have been resolved
	// (matters when the workspace was just initialised via --allow-minimal).
	lang = cmdutil.EffectiveLanguage(res.PlanDir)
	fmt.Println(ui.ActionHeader(cmdutil.Tr(lang, "Created Plan", "Plan creado"), state+"/"+slug))
	for _, p := range res.CreatedFiles {
		fmt.Println(cmdutil.PathLine("created", p))
	}
	return 0
}
