package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	explorecmd "pacto/internal/command/explore"
	initcmd "pacto/internal/command/initcmd"
	statuscmd "pacto/internal/command/status"
)

func TestChangedFlagsTracksStatusFormatFlag(t *testing.T) {
	cmd := newRootCommand([]string{"status", "--format", "table"})
	cmd.SetArgs([]string{"status", "--format", "table"})
	statusCmd, _, err := cmd.Find([]string{"status", "--format", "table"})
	if err != nil {
		t.Fatal(err)
	}
	statusCmd.SetArgs([]string{"--format", "table"})
	if err := statusCmd.ParseFlags([]string{"--format", "table"}); err != nil {
		t.Fatal(err)
	}

	provided := changedFlags(statusCmd, []string{"format"})
	if !provided["format"] {
		t.Fatal("expected --format to be marked as provided")
	}
}

func TestExecuteArgsPassesExplicitStatusFormatAsProvided(t *testing.T) {
	workspace := t.TempDir()
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(workspace, ".pacto", "plans", st), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}

	origRunStatusCommand := runStatusCommand
	t.Cleanup(func() { runStatusCommand = origRunStatusCommand })

	var got statuscmd.Options
	runStatusCommand = func(opts statuscmd.Options) int {
		got = opts
		return 0
	}

	code := ExecuteArgs([]string{"status", "--root", workspace, "--repo-root", workspace, "--format", "table"})
	if code != 0 {
		t.Fatalf("ExecuteArgs returned %d, want 0", code)
	}
	if !got.Provided["format"] {
		t.Fatalf("expected status options to mark format as provided, got %+v", got.Provided)
	}
}

func TestRunExecRequiresArgs(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		code := ExecuteArgs([]string{"exec"})
		if code != 2 {
			t.Fatalf("ExecuteArgs returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "exec requires <state> <slug>") && !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected missing args message, got stdout:%q stderr:%q", stdout, stderr)
	}
}

func TestRunNewMissingArgsShowsUsage(t *testing.T) {
	workspace := t.TempDir()
	_, _ = captureOutput(t, func() {
		if code := initcmd.Run(initcmd.Options{Root: workspace, NoInteractive: true, NoInstall: true}); code != 0 {
			t.Fatalf("RunInit returned %d", code)
		}
	})
	plansRoot := filepath.Join(workspace, ".pacto", "plans")
	stdout, stderr := captureOutput(t, func() {
		code := ExecuteArgs([]string{"new", "--root", plansRoot})
		if code == 0 {
			t.Fatalf("ExecuteArgs returned 0, want non-zero")
		}
	})
	if !strings.Contains(stderr, "Usage:") && !strings.Contains(stdout, "Usage:") {
		t.Fatalf("expected usage for missing new args, got stdout:%q stderr:%q", stdout, stderr)
	}
}

func TestRunUnknownCommandShowsHelp(t *testing.T) {
	_, stderr := captureOutput(t, func() {
		code := ExecuteArgs([]string{"missing-cmd"})
		if code != 1 {
			t.Fatalf("ExecuteArgs returned %d, want 1", code)
		}
	})
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr)
	}
}

func TestRunAcceptsLangOverride(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		code := ExecuteArgs([]string{"--lang=es", "version"})
		if code != 0 {
			t.Fatalf("ExecuteArgs returned %d, want 0", code)
		}
	})
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected no warning/error, got %q", stderr)
	}
	if !strings.Contains(stdout, "pacto version") {
		t.Fatalf("expected version output, got %q", stdout)
	}
}

func TestRunLangOverrideDoesNotLeakAcrossExecutions(t *testing.T) {
	if code := ExecuteArgs([]string{"--lang=es", "version"}); code != 0 {
		t.Fatalf("first ExecuteArgs returned %d, want 0", code)
	}

	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	if code := explorecmd.Run(explorecmd.Options{Title: "Leak Check"}, []string{"leak-check"}); code != 0 {
		t.Fatalf("Run returned %d, want 0", code)
	}
	b, err := os.ReadFile(filepath.Join(root, ".pacto", "ideas", "leak-check", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if strings.Contains(content, "**Creado:**") || strings.Contains(content, "**Actualizado:**") {
		t.Fatalf("unexpected leaked Spanish metadata in README: %q", content)
	}
}

func captureOutput(t *testing.T, fn func()) (stdout string, stderr string) {
	t.Helper()
	oldOut := os.Stdout
	oldErr := os.Stderr

	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = outW
	os.Stderr = errW
	defer func() {
		os.Stdout = oldOut
		os.Stderr = oldErr
	}()

	fn()

	_ = outW.Close()
	_ = errW.Close()

	outB, err := io.ReadAll(outR)
	if err != nil {
		t.Fatal(err)
	}
	errB, err := io.ReadAll(errR)
	if err != nil {
		t.Fatal(err)
	}
	return string(outB), string(errB)
}
