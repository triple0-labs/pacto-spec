package integrations

import (
	"os"
	"path/filepath"
	"strings"
)

func DetectTools(projectRoot string) ([]string, error) {
	type toolCheck struct {
		id  string
		dir string
	}
	checks := []toolCheck{
		{id: "codex", dir: ".codex"},
		{id: "cursor", dir: ".cursor"},
		{id: "claude", dir: ".claude"},
		{id: "opencode", dir: ".opencode"},
	}

	out := make([]string, 0)
	for _, c := range checks {
		if dirExists(filepath.Join(projectRoot, c.dir)) {
			out = append(out, c.id)
		}
	}

	if !contains(out, "codex") {
		home := strings.TrimSpace(os.Getenv("CODEX_HOME"))
		if home == "" {
			u, err := os.UserHomeDir()
			if err == nil {
				home = filepath.Join(u, ".codex")
			}
		}
		if home != "" && dirExists(home) {
			out = append([]string{"codex"}, out...)
		}
	}

	return dedupe(out), nil
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

func contains(v []string, needle string) bool {
	for _, x := range v {
		if x == needle {
			return true
		}
	}
	return false
}

func dedupe(v []string) []string {
	out := make([]string, 0, len(v))
	seen := map[string]bool{}
	for _, x := range v {
		if !seen[x] {
			out = append(out, x)
			seen[x] = true
		}
	}
	return out
}
