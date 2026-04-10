package statuscmd

import (
	"flag"
	"testing"

	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func RunStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	root := fs.String("root", "", "")
	plansRoot := fs.String("plans-root", "", "")
	repoRoot := fs.String("repo-root", "", "")
	mode := fs.String("mode", "compat", "")
	lang := fs.String("lang", "", "")
	format := fs.String("format", "table", "")
	configPath := fs.String("config", "", "")
	failOn := fs.String("fail-on", "none", "")
	state := fs.String("state", "all", "")
	includeArchive := fs.Bool("include-archive", false, "")
	maxNextActions := fs.Int("max-next-actions", 3, "")
	maxBlockers := fs.Int("max-blockers", 3, "")
	verbose := fs.Bool("verbose", false, "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	provided := map[string]bool{}
	fs.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	return Run(Options{
		Root:           *root,
		PlansRoot:      *plansRoot,
		RepoRoot:       *repoRoot,
		Mode:           *mode,
		Lang:           *lang,
		Format:         *format,
		ConfigPath:     *configPath,
		FailOn:         *failOn,
		State:          *state,
		IncludeArchive: *includeArchive,
		MaxNextActions: *maxNextActions,
		MaxBlockers:    *maxBlockers,
		Verbose:        *verbose,
		Provided:       provided,
	})
}
