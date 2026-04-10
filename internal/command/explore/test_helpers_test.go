package explorecmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func RunExplore(args []string) int {
	opts := Options{}
	pos := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--root":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "missing value for --root")
				return 2
			}
			opts.Root = args[i]
		case "--title":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "missing value for --title")
				return 2
			}
			opts.Title = args[i]
		case "--note":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "missing value for --note")
				return 2
			}
			opts.Note = args[i]
		case "--show":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "missing value for --show")
				return 2
			}
			opts.Show = args[i]
		case "--list":
			opts.List = true
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
				return 2
			}
			pos = append(pos, arg)
		}
	}
	return Run(opts, pos)
}
