package onboarding

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"pacto/internal/yamlutil"
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

	m, err := yamlutil.ReadMapOrDefault(cfgPath)
	if err != nil {
		return "", err
	}
	mergeProfileConfig(m, profile)
	if err := yamlutil.WriteMap(cfgPath, m); err != nil {
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

func mergeProfileConfig(m map[string]any, p Profile) {
	if m == nil {
		return
	}
	if _, ok := m["version"]; !ok {
		m["version"] = 1
	}

	project := yamlutil.EnsureMap(m, "project")
	project["languages"] = normalizeForYAML(p.Languages)
	if len(p.CustomLanguages) > 0 {
		project["custom_languages"] = normalizeForYAML(p.CustomLanguages)
	} else {
		delete(project, "custom_languages")
	}
	intents := yamlutil.EnsureMap(project, "intents")
	intents["problem"] = strings.TrimSpace(p.Intents.Problem)

	tools := yamlutil.EnsureMap(m, "tools")
	tools["selected"] = normalizeForYAML(p.Tools)
	if len(p.CustomTools) > 0 {
		tools["custom"] = normalizeForYAML(p.CustomTools)
	} else {
		delete(tools, "custom")
	}
	source := yamlutil.EnsureMap(tools, "source")
	source["languages"] = strings.TrimSpace(p.Sources.Languages)
	source["tools"] = strings.TrimSpace(p.Sources.Tools)

	ui := yamlutil.EnsureMap(m, "ui")
	ui["language"] = strings.TrimSpace(p.UILanguage)
	ui["source"] = strings.TrimSpace(p.Sources.UI)

	init := yamlutil.EnsureMap(m, "init")
	init["installed_at"] = time.Now().UTC().Format(time.RFC3339)

	plugins := yamlutil.EnsureMap(m, "plugins")
	if _, ok := plugins["enabled"]; !ok {
		plugins["enabled"] = []string{}
	}
}

func normalizeForYAML(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		t := strings.TrimSpace(it)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

func renderPRDBlock(p Profile) string {
	if strings.EqualFold(strings.TrimSpace(p.UILanguage), "es") {
		return renderPRDBlockES(p)
	}
	return renderPRDBlockEN(p)
}

func renderPRDBlockEN(p Profile) string {
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

func renderPRDBlockES(p Profile) string {
	var b strings.Builder
	b.WriteString("# Documento de Requerimientos de Producto (Gestionado por Pacto)\n\n")
	b.WriteString("## Problema\n\n")
	b.WriteString(valueOrTODO(p.Intents.Problem))
	b.WriteString("\n\n## Usuarios\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Objetivos\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## No objetivos\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Métricas de éxito\n\n")
	b.WriteString("- TODO")
	b.WriteString("\n\n## Preguntas abiertas\n\n")
	b.WriteString("- Aclarar requerimientos y restricciones aún no capturados.\n")
	return b.String()
}

func valueOrTODO(s string) string {
	if strings.TrimSpace(s) == "" {
		return "TODO"
	}
	return strings.TrimSpace(s)
}
