package installcmd

import (
	"flag"
	"testing"

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
	if len(fs.Args()) > 0 {
		return 2
	}
	return RunInstallWithOptions(ToolArtifactsOptions{
		Command: "install",
		Root:    *root,
		Tools:   *tools,
		Force:   *force,
	})
}

func RunUpdate(args []string) int {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	artifacts := fs.Bool("artifacts", false, "")
	root := fs.String("root", "", "")
	tools := fs.String("tools", "", "")
	force := fs.Bool("force", false, "")
	checkOnly := fs.Bool("check", false, "")
	yes := fs.Bool("yes", false, "")
	version := fs.String("version", "", "")
	repo := fs.String("repo", DefaultUpdateRepo, "")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(fs.Args()) > 0 {
		return 2
	}
	return RunUpdateWithOptions(UpdateOptions{
		Artifacts: *artifacts,
		Root:      *root,
		Tools:     *tools,
		Force:     *force,
		CheckOnly: *checkOnly,
		Yes:       *yes,
		Version:   *version,
		Repo:      *repo,
	})
}
