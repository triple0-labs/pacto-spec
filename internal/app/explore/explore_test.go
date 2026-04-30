package explore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pacto/internal/i18n"
)

func TestCreateOrUpdate_CreatesNewIdea(t *testing.T) {
	root := t.TempDir()

	action, path, err := CreateOrUpdate(root, "dark-mode", "Dark Mode", "", i18n.English)
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	if action != "created" {
		t.Fatalf("action=%q want created", action)
	}
	want := filepath.Join(root, ".pacto", "ideas", "dark-mode", "README.md")
	if path != want {
		t.Fatalf("path=%q want %q", path, want)
	}
	b, _ := os.ReadFile(path)
	body := string(b)
	if !strings.Contains(body, "# Dark Mode") {
		t.Fatalf("expected title in body:\n%s", body)
	}
	if !strings.Contains(body, "## Summary") {
		t.Fatalf("expected Summary section:\n%s", body)
	}
}

func TestCreateOrUpdate_DerivesTitleFromSlug(t *testing.T) {
	root := t.TempDir()
	_, path, err := CreateOrUpdate(root, "two-words", "", "", i18n.English)
	if err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}
	b, _ := os.ReadFile(path)
	if !strings.Contains(string(b), "# Two Words") {
		t.Fatalf("expected derived title 'Two Words':\n%s", b)
	}
}

func TestCreateOrUpdate_AppendsNotesOnUpdate(t *testing.T) {
	root := t.TempDir()
	if _, _, err := CreateOrUpdate(root, "topic", "Topic", "first note", i18n.English); err != nil {
		t.Fatal(err)
	}

	action, path, err := CreateOrUpdate(root, "topic", "", "second note", i18n.English)
	if err != nil {
		t.Fatalf("CreateOrUpdate update: %v", err)
	}
	if action != "updated" {
		t.Fatalf("action=%q want updated", action)
	}
	b, _ := os.ReadFile(path)
	body := string(b)
	if !strings.Contains(body, "first note") || !strings.Contains(body, "second note") {
		t.Fatalf("expected both notes:\n%s", body)
	}
}

func TestCreateOrUpdate_SkipsWhenNoNoteOnExisting(t *testing.T) {
	root := t.TempDir()
	if _, _, err := CreateOrUpdate(root, "topic", "Topic", "", i18n.English); err != nil {
		t.Fatal(err)
	}
	action, _, err := CreateOrUpdate(root, "topic", "", "", i18n.English)
	if err != nil {
		t.Fatal(err)
	}
	if action != "skipped" {
		t.Fatalf("action=%q want skipped", action)
	}
}

func TestCreateOrUpdate_RejectsInvalidSlug(t *testing.T) {
	_, _, err := CreateOrUpdate(t.TempDir(), "Bad Slug", "", "", i18n.English)
	if err == nil || !strings.Contains(err.Error(), "invalid slug") {
		t.Fatalf("expected invalid slug error, got %v", err)
	}
}

func TestListIdeas_ReturnsNilWhenNoIdeasDir(t *testing.T) {
	rows, err := ListIdeas(t.TempDir())
	if err != nil {
		t.Fatalf("ListIdeas: %v", err)
	}
	if rows != nil {
		t.Fatalf("expected nil, got %v", rows)
	}
}

func TestListIdeas_SortedBySlug(t *testing.T) {
	root := t.TempDir()
	for _, s := range []string{"zeta", "alpha", "mu"} {
		if _, _, err := CreateOrUpdate(root, s, "", "", i18n.English); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := ListIdeas(root)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, r := range rows {
		got = append(got, r.Slug)
	}
	want := []string{"alpha", "mu", "zeta"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestGetIdea_NotFound(t *testing.T) {
	_, err := GetIdea(t.TempDir(), "ghost")
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetIdea_RejectsInvalidSlug(t *testing.T) {
	_, err := GetIdea(t.TempDir(), "Bad")
	if err == nil || !strings.Contains(err.Error(), "invalid slug") {
		t.Fatalf("expected invalid slug, got %v", err)
	}
}

func TestGetIdea_RoundTrip(t *testing.T) {
	root := t.TempDir()
	if _, _, err := CreateOrUpdate(root, "alpha", "Alpha", "", i18n.English); err != nil {
		t.Fatal(err)
	}
	idea, err := GetIdea(root, "alpha")
	if err != nil {
		t.Fatalf("GetIdea: %v", err)
	}
	if idea.Slug != "alpha" || idea.Title != "Alpha" {
		t.Fatalf("unexpected idea: %+v", idea)
	}
}

func TestCreateOrUpdate_SpanishLanguageHeaders(t *testing.T) {
	root := t.TempDir()
	_, path, err := CreateOrUpdate(root, "tema", "Tema", "", i18n.Spanish)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(path)
	body := string(b)
	if !strings.Contains(body, "## Resumen") || !strings.Contains(body, "## Notas") {
		t.Fatalf("expected Spanish headers:\n%s", body)
	}
}
