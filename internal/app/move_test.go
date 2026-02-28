package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunMoveTransitionsStateAndUpdatesIndex(t *testing.T) {
	root := t.TempDir()
	if code := RunInit([]string{"--root", root}); code != 0 {
		t.Fatalf("RunInit returned %d", code)
	}

	if code := RunNew([]string{"to-implement", "move-sample", "--root", root}); code != 0 {
		t.Fatalf("RunNew returned %d", code)
	}

	_, stderr := captureOutput(t, func() {
		code := RunMove([]string{"to-implement", "move-sample", "done", "--root", root, "--reason", "completed work"})
		if code != 0 {
			t.Fatalf("RunMove returned %d", code)
		}
	})
	if stderr != "" {
		t.Fatalf("unexpected stderr: %q", stderr)
	}

	plansRoot := filepath.Join(root, ".pacto", "plans")
	dstReadme := filepath.Join(plansRoot, "done", "move-sample", "README.md")
	if _, err := os.Stat(dstReadme); err != nil {
		t.Fatalf("expected moved README: %v", err)
	}
	if _, err := os.Stat(filepath.Join(plansRoot, "to-implement", "move-sample")); !os.IsNotExist(err) {
		t.Fatalf("expected source dir removed")
	}

	b, err := os.ReadFile(dstReadme)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "**Status:** Completed (Done)") {
		t.Fatalf("expected status updated to done, got %q", text)
	}
	if !strings.Contains(text, "## Move History") || !strings.Contains(text, "completed work") {
		t.Fatalf("expected move history note, got %q", text)
	}

	idx, err := os.ReadFile(filepath.Join(plansRoot, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	index := string(idx)
	if !strings.Contains(index, "| 🟡 **To Implement** | 0 |") || !strings.Contains(index, "| ✅ **Done** | 1 |") {
		t.Fatalf("expected counts updated, got %q", index)
	}
	if strings.Contains(index, "./to-implement/move-sample/") {
		t.Fatalf("expected old state link removed, got %q", index)
	}
	if !strings.Contains(index, "./done/move-sample/") {
		t.Fatalf("expected done link added, got %q", index)
	}
}
