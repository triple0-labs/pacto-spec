package app

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

	if code := RunExplore([]string{"leak-check", "--title", "Leak Check"}); code != 0 {
		t.Fatalf("RunExplore returned %d, want 0", code)
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
