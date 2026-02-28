package onboarding

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PRDManagedStart = "<!-- pacto:prd:start -->"
	PRDManagedEnd   = "<!-- pacto:prd:end -->"
)

func WriteConfig(projectRoot string, profile Profile) (string, error) {
	cfgPath := filepath.Join(projectRoot, ".pacto", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o775); err != nil {
		return "", err
	}
	body := renderConfig(profile)
	if err := os.WriteFile(cfgPath, []byte(body), 0o664); err != nil {
		return "", err
	}
	return cfgPath, nil
}

func WritePRD(projectRoot string, profile Profile) (string, bool, error) {
	prdPath := filepath.Join(projectRoot, "prd.md")
	managed := renderPRDBlock(profile)
	block := PRDManagedStart + "\n" + strings.TrimSpace(managed) + "\n" + PRDManagedEnd + "\n"

	b, err := os.ReadFile(prdPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.WriteFile(prdPath, []byte(block), 0o664); err != nil {
				return "", false, err
			}
			return prdPath, true, nil
		}
		return "", false, err
	}

	s := string(b)
	start := strings.Index(s, PRDManagedStart)
	end := strings.Index(s, PRDManagedEnd)
	if start >= 0 && end > start {
		end += len(PRDManagedEnd)
		next := s[:start] + block + s[end:]
		if next == s {
			return prdPath, false, nil
		}
		if err := os.WriteFile(prdPath, []byte(next), 0o664); err != nil {
			return "", false, err
		}
		return prdPath, true, nil
	}

	next := strings.TrimRight(s, "\n")
	if next != "" {
		next += "\n\n"
	}
	next += block
	if err := os.WriteFile(prdPath, []byte(next), 0o664); err != nil {
		return "", false, err
	}
	return prdPath, true, nil
}

func renderConfig(p Profile) string {
	var b strings.Builder
	b.WriteString("version: 1\n")
	b.WriteString("project:\n")
	b.WriteString(fmt.Sprintf("  languages: [%s]\n", yamlList(p.Languages)))
	if len(p.CustomLanguages) > 0 {
		b.WriteString(fmt.Sprintf("  custom_languages: [%s]\n", yamlList(p.CustomLanguages)))
	}
	b.WriteString("  intents:\n")
	b.WriteString(fmt.Sprintf("    problem: %q\n", p.Intents.Problem))
	b.WriteString("tools:\n")
	b.WriteString(fmt.Sprintf("  selected: [%s]\n", yamlList(p.Tools)))
	if len(p.CustomTools) > 0 {
		b.WriteString(fmt.Sprintf("  custom: [%s]\n", yamlList(p.CustomTools)))
	}
	b.WriteString("  source:\n")
	b.WriteString(fmt.Sprintf("    languages: %q\n", p.Sources.Languages))
	b.WriteString(fmt.Sprintf("    tools: %q\n", p.Sources.Tools))
	b.WriteString("init:\n")
	b.WriteString(fmt.Sprintf("  installed_at: %q\n", time.Now().UTC().Format(time.RFC3339)))
	return b.String()
}

func renderPRDBlock(p Profile) string {
	var b strings.Builder
	b.WriteString("# Product Requirements Document (Pacto Managed)\n\n")
	b.WriteString("## Problem\n\n")
	b.WriteString(valueOrTODO(p.Intents.Problem))
	b.WriteString("\n\n## Users\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Goals\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Non-goals\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Success Metrics\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Open Questions\n\n")
	b.WriteString("- Clarify requirements and constraints not captured yet.\n")
	return b.String()
}

func valueOrTODO(s string) string {
	if strings.TrimSpace(s) == "" {
		return "TODO"
	}
	return strings.TrimSpace(s)
}

func yamlList(items []string) string {
	if len(items) == 0 {
		return ""
	}
	escaped := make([]string, 0, len(items))
	for _, it := range items {
		t := strings.TrimSpace(it)
		if t == "" {
			continue
		}
		escaped = append(escaped, fmt.Sprintf("%q", t))
	}
	return strings.Join(escaped, ", ")
}
