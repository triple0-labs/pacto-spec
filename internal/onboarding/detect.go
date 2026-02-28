package onboarding

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DetectProfile(projectRoot string) Profile {
	langs := detectLanguages(projectRoot)
	tools := detectLocalTools(projectRoot)
	sort.Strings(tools)
	return Profile{
		Languages: langs,
		Tools:     tools,
		Sources: Sources{
			Languages: "auto",
			Tools:     "auto",
		},
	}
}

func detectLocalTools(root string) []string {
	type toolDir struct {
		id  string
		dir string
	}
	checks := []toolDir{
		{id: "codex", dir: ".codex"},
		{id: "cursor", dir: ".cursor"},
		{id: "claude", dir: ".claude"},
		{id: "opencode", dir: ".opencode"},
	}
	out := make([]string, 0, len(checks))
	for _, c := range checks {
		if dirExists(filepath.Join(root, c.dir)) {
			out = append(out, c.id)
		}
	}
	return out
}

func detectLanguages(root string) []string {
	found := map[string]bool{}
	checks := map[string][]string{
		"go":         {"go.mod"},
		"typescript": {"tsconfig.json"},
		"javascript": {"package.json"},
		"python":     {"pyproject.toml", "requirements.txt"},
		"rust":       {"Cargo.toml"},
		"java":       {"pom.xml", "build.gradle", "build.gradle.kts"},
		"dotnet":     {"*.csproj", "*.sln"},
		"ruby":       {"Gemfile"},
		"php":        {"composer.json"},
	}

	for lang, files := range checks {
		for _, f := range files {
			if strings.Contains(f, "*") {
				m, _ := filepath.Glob(filepath.Join(root, f))
				if len(m) > 0 {
					found[lang] = true
					break
				}
				continue
			}
			if fileExists(filepath.Join(root, f)) {
				found[lang] = true
				break
			}
		}
	}

	out := make([]string, 0, len(found))
	for lang := range found {
		out = append(out, lang)
	}
	sort.Strings(out)
	return out
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}
