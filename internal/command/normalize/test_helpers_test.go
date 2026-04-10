package normalizecmd

import (
	"flag"
	"testing"

	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func RunNormalize(args []string) int {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	root := fs.String("root", "", "")
	state := fs.String("state", "all", "")
	includeArchive := fs.Bool("include-archive", false, "")
	write := fs.Bool("write", false, "")
	format := fs.String("format", "table", "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(fs.Args()) > 0 {
		return 2
	}
	return Run(Options{
		Root:           *root,
		State:          *state,
		IncludeArchive: *includeArchive,
		Write:          *write,
		Format:         *format,
	})
}
