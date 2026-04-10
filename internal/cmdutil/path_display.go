package cmdutil

import (
	"path/filepath"
	"strings"

	"pacto/internal/ui"
)

func PathLine(kind, path string) string {
	return ui.PathLine(kind, DisplayPath(path))
}

func DisplayPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return path
	}
	cleaned := filepath.Clean(trimmed)
	if !filepath.IsAbs(cleaned) {
		return cleaned
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return cleaned
	}
	rel, err := filepath.Rel(cwd, cleaned)
	if err != nil {
		return cleaned
	}
	if rel == "." {
		return "."
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return cleaned
	}
	return filepath.Clean(rel)
}
