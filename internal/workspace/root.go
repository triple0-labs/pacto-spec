package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/i18n"
)

func ResolvePlanRoot(path string) (string, bool) {
	if HasStateDirs(path) {
		return path, true
	}
	cand := filepath.Join(path, ".pacto", "plans")
	if HasStateDirs(cand) {
		return cand, true
	}
	return path, false
}

func ResolvePlanRootFrom(path string) (string, string, bool) {
	cur := CleanAbs(path)
	for {
		if resolved, ok := ResolvePlanRoot(cur); ok {
			return resolved, cur, true
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			break
		}
		cur = parent
	}
	return "", "", false
}

func HasStateDirs(path string) bool {
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		info, err := os.Stat(filepath.Join(path, st))
		if err != nil || !info.IsDir() {
			return false
		}
	}
	return true
}

func CleanAbs(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func ResolvePlansRootForAction(rawRoot string) (string, error) {
	if strings.TrimSpace(rawRoot) != "" {
		abs, err := filepath.Abs(rawRoot)
		if err != nil {
			return "", err
		}
		if resolved, ok := ResolvePlanRoot(abs); ok {
			return resolved, nil
		}
		return "", fmt.Errorf("could not resolve plans root from %s (expected .pacto/plans)", abs)
	}

	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if resolved, _, ok := ResolvePlanRootFrom(cwd); ok {
		return resolved, nil
	}
	return "", fmt.Errorf("could not resolve plans root from %s or parents (expected .pacto/plans)", cwd)
}

func ResolveExploreRoot(rawRoot string) (string, error) {
	if strings.TrimSpace(rawRoot) != "" {
		return filepath.Abs(rawRoot)
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		return "", err
	}
	if _, projectRoot, ok := ResolvePlanRootFrom(cwd); ok {
		return projectRoot, nil
	}
	return cwd, nil
}

func ValidateRoot(root string) error {
	for _, p := range []string{"README.md", "PACTO.md"} {
		if _, err := os.Stat(filepath.Join(root, p)); err != nil {
			return fmt.Errorf("missing %s", p)
		}
	}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if _, err := os.Stat(filepath.Join(root, st)); err != nil {
			return fmt.Errorf("missing state folder %s", st)
		}
	}
	return nil
}

func EnsureMinimalRoot(root string, lang i18n.Language) error {
	if err := os.MkdirAll(root, 0o775); err != nil {
		return err
	}
	for _, st := range []string{"current", "to-implement", "done", "outdated"} {
		if err := os.MkdirAll(filepath.Join(root, st), 0o775); err != nil {
			return err
		}
	}
	readmePath := filepath.Join(root, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		if err := os.WriteFile(readmePath, []byte(defaultRootReadme(lang)), 0o664); err != nil {
			return err
		}
	}
	pactoPath := filepath.Join(root, "PACTO.md")
	if _, err := os.Stat(pactoPath); err != nil {
		content := "# Pacto\n\nMinimal root created by pacto CLI.\n"
		if lang == i18n.Spanish {
			content = "# Pacto\n\nRaiz minima creada por la CLI de pacto.\n"
		}
		if err := os.WriteFile(pactoPath, []byte(content), 0o664); err != nil {
			return err
		}
	}
	return nil
}

func defaultRootReadme(lang i18n.Language) string {
	if lang == i18n.Spanish {
		return "# Planes de Pacto\n\n" +
			"## Resumen\n\n" +
			"El estado de planes se deriva de carpetas y documentos mediante `pacto status`.\n" +
			"Este README es un resumen, no una fuente de verdad de estado.\n\n" +
			"---\n\n" +
			"## 📜 Pacto\n\n" +
			"- [PACTO.md](./PACTO.md)\n"
	}
	return "# Pacto Plans\n\n" +
		"## Summary\n\n" +
		"Plan status is derived from plan folders and documents via `pacto status`.\n" +
		"This README is an overview, not a status source of truth.\n\n" +
		"---\n\n" +
		"## 📜 Pacto\n\n" +
		"- [PACTO.md](./PACTO.md)\n"
}
