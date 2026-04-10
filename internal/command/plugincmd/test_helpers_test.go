package plugincmd_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"pacto/internal/cli"
	initcmd "pacto/internal/command/initcmd"
	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func Execute() int {
	return cli.Execute()
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
