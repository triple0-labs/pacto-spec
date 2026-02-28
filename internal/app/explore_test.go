package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunExploreCreateAndNote(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	if code := RunExplore([]string{"my-idea", "--title", "My Idea"}); code != 0 {
		t.Fatalf("RunExplore create returned %d", code)
	}
	path := filepath.Join(root, ".pacto", "ideas", "my-idea", "README.md")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.Contains(content, "**Created At:**") || !strings.Contains(content, "**Updated At:**") {
		t.Fatalf("expected timestamps in README, got %q", content)
	}

	if code := RunExplore([]string{"my-idea", "--note", "compare options"}); code != 0 {
		t.Fatalf("RunExplore note returned %d", code)
	}
	b, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "compare options") {
		t.Fatalf("expected note appended, got %q", string(b))
	}
}

func TestRunExploreListAndShow(t *testing.T) {
	root := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}

	if code := RunExplore([]string{"idea-a"}); code != 0 {
		t.Fatalf("RunExplore create returned %d", code)
	}

	stdout, _ := captureOutput(t, func() {
		if code := RunExplore([]string{"--list"}); code != 0 {
			t.Fatalf("RunExplore list returned %d", code)
		}
	})
	if !strings.Contains(stdout, "idea-a") {
		t.Fatalf("expected idea slug in list output, got %q", stdout)
	}

	stdout, _ = captureOutput(t, func() {
		if code := RunExplore([]string{"--show", "idea-a"}); code != 0 {
			t.Fatalf("RunExplore show returned %d", code)
		}
	})
	if !strings.Contains(stdout, "Idea idea-a") {
		t.Fatalf("expected show output, got %q", stdout)
	}
}
