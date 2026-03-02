package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDisplayPathAbsoluteInsideCWDBecomesRelative(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(root, ".pacto", "plans", "README.md")
	want := filepath.Join(".pacto", "plans", "README.md")
	if got := displayPath(target); got != want {
		t.Fatalf("displayPath()=%q, want %q", got, want)
	}
}

func TestDisplayPathOutsideCWDStaysAbsolute(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(outside, "README.md")
	if got := displayPath(target); got != target {
		t.Fatalf("displayPath()=%q, want absolute %q", got, target)
	}
}

func TestDisplayPathRelativeInputStaysRelative(t *testing.T) {
	in := filepath.Join(".pacto", "plans", "..", "plans", "README.md")
	want := filepath.Join(".pacto", "plans", "README.md")
	if got := displayPath(in); got != want {
		t.Fatalf("displayPath()=%q, want %q", got, want)
	}
}
