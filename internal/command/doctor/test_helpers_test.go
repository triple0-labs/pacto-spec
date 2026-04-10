package doctorcmd

import (
	"flag"
	"testing"

	installcmd "pacto/internal/command/install"
	"pacto/internal/testutil"
)

func captureOutput(t *testing.T, fn func()) (string, string) {
	t.Helper()
	return testutil.CaptureOutput(t, fn)
}

func RunInstall(args []string) int {
	fs := flag.NewFlagSet("install", flag.ContinueOnError)
	root := fs.String("root", "", "")
	tools := fs.String("tools", "", "")
	force := fs.Bool("force", false, "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return installcmd.RunInstallWithOptions(installcmd.ToolArtifactsOptions{
		Command: "install",
		Root:    *root,
		Tools:   *tools,
		Force:   *force,
	})
}

func RunDoctor(args []string) int {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	root := fs.String("root", "", "")
	tools := fs.String("tools", "", "")
	format := fs.String("format", "table", "")
	failOn := fs.String("fail-on", "none", "")
	verbose := fs.Bool("verbose", false, "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(fs.Args()) > 0 {
		return 2
	}
	return Run(Options{
		Root:    *root,
		Tools:   *tools,
		Format:  *format,
		FailOn:  *failOn,
		Verbose: *verbose,
	})
}
