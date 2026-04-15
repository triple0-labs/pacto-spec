package doctorcmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDoctorTableOK(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if code := RunInstall([]string{"--tools", "cursor"}); code != 0 {
		t.Fatalf("RunInstall returned %d", code)
	}
	stdout, stderr := captureOutput(t, func() {
		if code := RunDoctor([]string{"--tools", "cursor", "--format", "table"}); code != 0 {
			t.Fatalf("RunDoctor returned %d", code)
		}
	})
	if !strings.Contains(stdout, "SUMMARY:") {
		t.Fatalf("expected table summary output, got %q", stdout)
	}
	if strings.Contains(stderr, "warning:") {
		t.Fatalf("did not expect warnings for clean artifacts, got %q", stderr)
	}
}

func TestRunDoctorFailOnAny(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if code := RunInstall([]string{"--tools", "cursor"}); code != 0 {
		t.Fatalf("RunInstall returned %d", code)
	}
	if err := os.Remove(filepath.Join(root, ".agents", "skills", "pacto-status", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	if code := RunDoctor([]string{"--tools", "cursor", "--fail-on", "any"}); code != 1 {
		t.Fatalf("RunDoctor returned %d, want 1", code)
	}
}

func TestRunDoctorJSON(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	if code := RunInstall([]string{"--tools", "cursor"}); code != 0 {
		t.Fatalf("RunInstall returned %d", code)
	}
	stdout, _ := captureOutput(t, func() {
		if code := RunDoctor([]string{"--tools", "cursor", "--format", "json"}); code != 0 {
			t.Fatalf("RunDoctor returned %d", code)
		}
	})
	if !strings.Contains(stdout, "\"summary\"") || !strings.Contains(stdout, "\"records\"") {
		t.Fatalf("expected json output, got %q", stdout)
	}
}
