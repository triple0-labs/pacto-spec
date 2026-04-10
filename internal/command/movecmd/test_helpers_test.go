package movecmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	initcmd "pacto/internal/command/initcmd"
	newcmd "pacto/internal/command/newcmd"
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
			fmt.Fprintf(os.Stderr, "flag %s is no longer supported; use interactive onboarding for technologies and install targets\n", arg)
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
	opts := newcmd.Options{}
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
	return newcmd.Run(opts, pos[0], pos[1])
}

func RunMove(args []string) int {
	opts := Options{}
	pos := make([]string, 0, 3)
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--root":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Root = args[i]
		case "--reason":
			i++
			if i >= len(args) {
				return 2
			}
			opts.Reason = args[i]
		case "--force":
			opts.Force = true
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
