package execmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"pacto/internal/app/execplan"
	"pacto/internal/ui"
)

type Options struct {
	Root     string
	Step     string
	Note     string
	Blocker  string
	Evidence string
	DryRun   bool
}

func Run(opts Options, pos []string) int {
	lang := effectiveLanguage(opts.Root)
	if len(pos) != 2 {
		fmt.Fprintln(os.Stderr, tr(lang, "exec requires <state> <slug>", "exec requiere <state> <slug>"))
		return 2
	}

	state := strings.ToLower(strings.TrimSpace(pos[0]))
	slug := strings.TrimSpace(pos[1])

	res, err := execplan.Apply(execplan.Input{
		Root:     opts.Root,
		State:    state,
		Slug:     slug,
		Step:     opts.Step,
		Note:     opts.Note,
		Blocker:  opts.Blocker,
		Evidence: opts.Evidence,
		DryRun:   opts.DryRun,
		Lang:     lang,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if errors.Is(err, execplan.ErrInvalid) {
			return 2
		}
		return 3
	}

	if res.NoChange {
		fmt.Println(ui.Dim(tr(lang, "No execution changes to apply.", "No hay cambios de ejecución para aplicar.")))
		return 0
	}

	if res.DryRun {
		fmt.Println(ui.ActionHeader(tr(lang, "Dry Run", "Simulación"), tr(lang, "execution update", "actualización de ejecución")))
		fmt.Println(pathLine("updated", res.PlanPath))
		for _, a := range res.Actions {
			fmt.Println(ui.Bullet(a))
		}
		return 0
	}

	fmt.Println(ui.ActionHeader(tr(lang, "Executed Plan", "Plan ejecutado"), state+"/"+slug))
	fmt.Println(pathLine("updated", res.PlanPath))
	for _, a := range res.Actions {
		fmt.Println(ui.Bullet(a))
	}
	return 0
}
