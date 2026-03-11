package explore

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"pacto/internal/i18n"
	"pacto/internal/markdown"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

type Idea struct {
	Slug      string
	Title     string
	CreatedAt string
	UpdatedAt string
	Path      string
}

// CreateOrUpdate creates a new idea, or updates an existing one if the note is provided.
func CreateOrUpdate(root, slug, title, note string, lang i18n.Language) (string, string, error) {
	slug = strings.TrimSpace(slug)
	if !slugRe.MatchString(slug) {
		return "", "", fmt.Errorf("invalid slug %q (use lowercase letters, numbers, dashes)", slug)
	}

	ideasRoot := filepath.Join(root, ".pacto", "ideas")
	ideaDir := filepath.Join(ideasRoot, slug)
	readmePath := filepath.Join(ideaDir, "README.md")
	now := time.Now().Format("2006-01-02 15:04")

	if err := os.MkdirAll(ideaDir, 0o775); err != nil {
		return "", "", fmt.Errorf("create idea dir: %w", err)
	}

	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		ideaTitle := strings.TrimSpace(title)
		if ideaTitle == "" {
			ideaTitle = slugToTitle(slug)
		}
		text := buildExploreReadme(ideaTitle, now, lang)
		if strings.TrimSpace(note) != "" {
			text = appendExploreNote(text, strings.TrimSpace(note), now, lang)
		}
		if err := os.WriteFile(readmePath, []byte(text), 0o664); err != nil {
			return "", "", fmt.Errorf("write idea readme: %w", err)
		}
		return "created", readmePath, nil
	} else if err != nil {
		return "", "", fmt.Errorf("stat idea readme: %w", err)
	}

	if strings.TrimSpace(note) == "" {
		return "skipped", readmePath, nil
	}

	b, err := os.ReadFile(readmePath)
	if err != nil {
		return "", "", fmt.Errorf("read idea readme: %w", err)
	}
	updated := appendExploreNote(string(b), strings.TrimSpace(note), now, lang)
	updated = markdown.SetUpdatedAt(updated, now)
	if err := os.WriteFile(readmePath, []byte(updated), 0o664); err != nil {
		return "", "", fmt.Errorf("update idea readme: %w", err)
	}
	return "updated", readmePath, nil
}

// ListIdeas reads the .pacto/ideas directory and parses all ideas.
func ListIdeas(root string) ([]Idea, error) {
	ideasRoot := filepath.Join(root, ".pacto", "ideas")
	ents, err := os.ReadDir(ideasRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No ideas
		}
		return nil, fmt.Errorf("read ideas: %w", err)
	}

	var rows []Idea
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		readmePath := filepath.Join(ideasRoot, e.Name(), "README.md")
		b, err := os.ReadFile(readmePath)
		if err != nil {
			continue
		}
		content := string(b)
		rows = append(rows, Idea{
			Slug:      e.Name(),
			Title:     markdown.ExtractTitle(content),
			CreatedAt: markdown.ExtractStamp(markdown.ReCreatedAt, content),
			UpdatedAt: markdown.ExtractStamp(markdown.ReUpdatedAt, content),
			Path:      readmePath,
		})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Slug < rows[j].Slug })
	return rows, nil
}

// GetIdea retrieves an idea by slug.
func GetIdea(root, slug string) (Idea, error) {
	slug = strings.TrimSpace(slug)
	if !slugRe.MatchString(slug) {
		return Idea{}, fmt.Errorf("invalid slug %q (use lowercase letters, numbers, dashes)", slug)
	}
	readmePath := filepath.Join(root, ".pacto", "ideas", slug, "README.md")
	b, err := os.ReadFile(readmePath)
	if err != nil {
		if os.IsNotExist(err) {
			return Idea{}, errors.New("idea not found")
		}
		return Idea{}, fmt.Errorf("read idea: %w", err)
	}
	content := string(b)
	return Idea{
		Slug:      slug,
		Title:     markdown.ExtractTitle(content),
		CreatedAt: markdown.ExtractStamp(markdown.ReCreatedAt, content),
		UpdatedAt: markdown.ExtractStamp(markdown.ReUpdatedAt, content),
		Path:      readmePath,
	}, nil
}

// Helper methods brought over

func slugToTitle(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func buildExploreReadme(title, now string, lang i18n.Language) string {
	var b strings.Builder
	b.WriteString("# " + title + "\n\n")

	// Quick tr util to avoid changing the original too much
	tr := func(l i18n.Language, en, es string) string {
		if l == i18n.Spanish {
			return es
		}
		return en
	}

	b.WriteString(tr(lang, "**Created At:** ", "**Creado:** ") + now + "  \n")
	b.WriteString(tr(lang, "**Updated At:** ", "**Actualizado:** ") + now + "\n\n")
	b.WriteString(tr(lang, "## Summary\n\n", "## Resumen\n\n"))
	b.WriteString(tr(lang, "Idea exploration workspace.\n\n", "Espacio de exploración de ideas.\n\n"))
	b.WriteString(tr(lang, "## Notes\n\n", "## Notas\n\n"))
	b.WriteString("- [" + now + "] " + tr(lang, "Idea created.", "Idea creada.") + "\n")
	return b.String()
}

func appendExploreNote(content, note, now string, lang i18n.Language) string {
	tr := func(l i18n.Language, en, es string) string {
		if l == i18n.Spanish {
			return es
		}
		return en
	}
	if !strings.Contains(content, "## Notes") && !strings.Contains(content, "## Notas") {
		content = strings.TrimRight(content, "\n") + "\n\n" + tr(lang, "## Notes\n\n", "## Notas\n\n")
	}
	content = strings.TrimRight(content, "\n")
	return content + "\n- [" + now + "] " + note + "\n"
}
