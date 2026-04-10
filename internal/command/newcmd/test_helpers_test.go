package newcmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	initcmd "pacto/internal/command/initcmd"
	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func RunInit(args []string) int {
	opts := initcmd.Options{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--root":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Root = args[i]
		case "--with-agents":
			opts.WithAgents = true
		case "--force":
			opts.Force = true
		case "--lang":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Lang = args[i]
		case "--no-interactive":
			opts.NoInteractive = true
		case "--tools":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Tools = args[i]
		case "--yes":
			opts.Yes = true
		case "--no-install":
			opts.NoInstall = true
		case "--dry-run":
			opts.DryRun = true
		case "--editor", "--language":
			return 2
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
				return 2
			}
			return 2
		}
	}
	return initcmd.Run(opts)
}

func RunNew(args []string) int {
	opts := Options{}
	pos := make([]string, 0, 2)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--root":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Root = args[i]
			opts.RootProvided = true
		case "--title":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Title = args[i]
		case "--owner":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Owner = args[i]
		case "--allow-minimal-root":
			opts.AllowMinimal = true
		case "--lang":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Lang = args[i]
		default:
			if strings.HasPrefix(arg, "--") {
				fmt.Fprintf(os.Stderr, "unknown flag: %s\n", arg)
				return 2
			}
			pos = append(pos, arg)
		}
	}
	if len(pos) != 2 {
		return 2
	}
	return Run(opts, pos[0], pos[1])
}

func mustCreateStateDirs(t *testing.T, root string) {
	t.Helper()
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o775); err != nil {
			t.Fatal(err)
		}
	}
}
