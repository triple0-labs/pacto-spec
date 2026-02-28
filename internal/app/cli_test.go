package app

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunExecRequiresArgs(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		code := Run([]string{"exec"})
		if code != 2 {
			t.Fatalf("Run returned %d, want 2", code)
		}
	})
	if stdout != "" {
		t.Fatalf("unexpected stdout: %q", stdout)
	}
	if !strings.Contains(stderr, "exec requires <state> <slug>") {
		t.Fatalf("expected missing args message, got %q", stderr)
	}
}

func TestRunUnknownCommandShowsHelp(t *testing.T) {
	stdout, stderr := captureOutput(t, func() {
		code := Run([]string{"missing-cmd"})
		if code != 2 {
			t.Fatalf("Run returned %d, want 2", code)
		}
	})
	if !strings.Contains(stderr, "unknown command") {
		t.Fatalf("expected unknown command error, got %q", stderr)
	}
	if !strings.Contains(stdout, "Commands:") {
		t.Fatalf("expected root help on stdout, got %q", stdout)
	}
}

func TestRunWarnsOnDeprecatedLang(t *testing.T) {
	_, stderr := captureOutput(t, func() {
		code := Run([]string{"--lang", "es", "version"})
		if code != 0 {
			t.Fatalf("Run returned %d, want 0", code)
		}
	})
	if !strings.Contains(stderr, "--lang is deprecated") {
		t.Fatalf("expected deprecation warning, got %q", stderr)
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
