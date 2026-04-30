package newplan

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

func setupWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := workspace.EnsureMinimalRoot(root, i18n.English); err != nil {
		t.Fatalf("EnsureMinimalRoot: %v", err)
	}
	return root
}

func TestCreate_HappyPath(t *testing.T) {
	root := setupWorkspace(t)

	res, err := Create(Input{
		Root:         root,
		RootProvided: true,
		State:        "current",
		Slug:         "alpha",
		Title:        "Alpha Plan",
		Owner:        "Diego",
		Lang:         i18n.English,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if !strings.HasSuffix(res.PlanDir, filepath.Join("current", "alpha")) {
		t.Fatalf("unexpected PlanDir: %s", res.PlanDir)
	}
	if len(res.CreatedFiles) != 4 {
		t.Fatalf("expected 4 files created, got %d", len(res.CreatedFiles))
	}
	for _, p := range res.CreatedFiles {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected %s to exist: %v", p, err)
		}
	}
	readme, err := os.ReadFile(filepath.Join(res.PlanDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "Alpha Plan") {
		t.Fatalf("readme missing title; got: %s", readme)
	}
}

func TestCreate_RejectsInvalidState(t *testing.T) {
	root := setupWorkspace(t)
	_, err := Create(Input{
		Root:         root,
		RootProvided: true,
		State:        "invented",
		Slug:         "x",
		Lang:         i18n.English,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestCreate_RejectsInvalidSlug(t *testing.T) {
	root := setupWorkspace(t)
	_, err := Create(Input{
		Root:         root,
		RootProvided: true,
		State:        "current",
		Slug:         "BadSlug",
		Lang:         i18n.English,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestCreate_RejectsDuplicate(t *testing.T) {
	root := setupWorkspace(t)
	in := Input{
		Root:         root,
		RootProvided: true,
		State:        "current",
		Slug:         "dup",
		Lang:         i18n.English,
	}
	if _, err := Create(in); err != nil {
		t.Fatalf("first Create: %v", err)
	}
	_, err := Create(in)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for duplicate, got %v", err)
	}
}

func TestCreate_RejectsNonPactoRoot(t *testing.T) {
	bare := t.TempDir()
	_, err := Create(Input{
		Root:         bare,
		RootProvided: true,
		State:        "current",
		Slug:         "alpha",
		Lang:         i18n.English,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for non-pacto root, got %v", err)
	}
}

func TestCreate_AllowMinimalCreatesScaffold(t *testing.T) {
	bare := t.TempDir()
	res, err := Create(Input{
		Root:         bare,
		RootProvided: true,
		State:        "to-implement",
		Slug:         "bootstrap",
		AllowMinimal: true,
		Lang:         i18n.English,
	})
	if err != nil {
		t.Fatalf("Create with AllowMinimal: %v", err)
	}
	if _, err := os.Stat(filepath.Join(res.PlanDir, "spec.md")); err != nil {
		t.Fatalf("expected spec.md after AllowMinimal: %v", err)
	}
}

func TestCreate_SpanishTemplate(t *testing.T) {
	root := setupWorkspace(t)
	res, err := Create(Input{
		Root:         root,
		RootProvided: true,
		State:        "current",
		Slug:         "es-plan",
		Lang:         i18n.Spanish,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	spec, err := os.ReadFile(filepath.Join(res.PlanDir, "spec.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(spec), "Especificación") {
		t.Fatalf("expected Spanish spec, got: %s", spec)
	}
}

func TestCreate_TitleDefaultsFromSlug(t *testing.T) {
	root := setupWorkspace(t)
	res, err := Create(Input{
		Root:         root,
		RootProvided: true,
		State:        "current",
		Slug:         "auto-title-here",
		Lang:         i18n.English,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	readme, err := os.ReadFile(filepath.Join(res.PlanDir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "Auto Title Here") {
		t.Fatalf("expected derived title, got: %s", readme)
	}
}
