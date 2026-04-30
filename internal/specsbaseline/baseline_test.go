package specsbaseline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitBaseline_createsDirAndReadme(t *testing.T) {
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	if err := os.MkdirAll(plansRoot, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := InitBaseline(plansRoot); err != nil {
		t.Fatalf("init: %v", err)
	}
	specsDir := SpecsDirFromPlansRoot(plansRoot)
	if _, err := os.Stat(specsDir); err != nil {
		t.Fatalf("specs dir not created: %v", err)
	}
	if _, err := os.Stat(filepath.Join(specsDir, "README.md")); err != nil {
		t.Fatalf("README not created: %v", err)
	}
}

func TestInitBaseline_idempotent(t *testing.T) {
	root := t.TempDir()
	plansRoot := filepath.Join(root, ".pacto", "plans")
	if err := os.MkdirAll(plansRoot, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := InitBaseline(plansRoot); err != nil {
		t.Fatal(err)
	}
	readme := filepath.Join(SpecsDirFromPlansRoot(plansRoot), "README.md")
	custom := []byte("# Custom override\n")
	if err := os.WriteFile(readme, custom, 0o664); err != nil {
		t.Fatal(err)
	}
	// also create a capability directory and ensure init preserves it
	capDir := filepath.Join(SpecsDirFromPlansRoot(plansRoot), "auth")
	if err := os.MkdirAll(capDir, 0o775); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(capDir, "spec.md"), []byte("# auth\n"), 0o664); err != nil {
		t.Fatal(err)
	}
	if err := InitBaseline(plansRoot); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(readme)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(custom) {
		t.Fatalf("README was overwritten; got %s", got)
	}
	if _, err := os.Stat(filepath.Join(capDir, "spec.md")); err != nil {
		t.Fatalf("capability spec was removed: %v", err)
	}
}
