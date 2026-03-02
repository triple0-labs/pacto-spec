package app

import (
	"os"
	"path/filepath"
	"strings"

	"pacto/internal/i18n"
	"pacto/internal/yamlutil"
)

var globalLangOverride string

func setGlobalLangOverride(raw string) {
	globalLangOverride = strings.TrimSpace(raw)
}

func effectiveLanguage(projectRootHint string) i18n.Language {
	if strings.TrimSpace(globalLangOverride) != "" {
		if lang, ok := i18n.ParseLanguage(globalLangOverride); ok {
			return lang
		}
	}
	for _, cfgPath := range candidateConfigPaths(projectRootHint) {
		cfg, err := yamlutil.ReadFileMap(cfgPath)
		if err != nil {
			continue
		}
		ui := yamlutil.GetMap(cfg, "ui")
		if ui == nil {
			continue
		}
		if lang, ok := i18n.ParseLanguage(anyString(ui["language"])); ok {
			return lang
		}
	}
	return i18n.English
}

func candidateConfigPaths(projectRootHint string) []string {
	seen := map[string]struct{}{}
	paths := make([]string, 0, 4)
	add := func(p string) {
		p = cleanAbs(p)
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		paths = append(paths, filepath.Join(p, ".pacto", "config.yaml"))
	}
	addDerived := func(p string) {
		p = cleanAbs(p)
		if p == "" {
			return
		}
		if st, err := os.Stat(filepath.Join(p, "config.yaml")); err == nil && !st.IsDir() && filepath.Base(p) == ".pacto" {
			add(filepath.Dir(p))
		}
		if filepath.Base(p) == "plans" && filepath.Base(filepath.Dir(p)) == ".pacto" {
			add(filepath.Dir(filepath.Dir(p)))
		}
	}

	if strings.TrimSpace(projectRootHint) != "" {
		add(projectRootHint)
		addDerived(projectRootHint)
		return paths
	}
	cwd := cleanAbs(".")
	add(cwd)
	addDerived(cwd)
	if planRoot, foundProjectRoot, ok := resolvePlanRootFrom(cwd); ok {
		add(foundProjectRoot)
		add(planRoot)
		addDerived(planRoot)
	}
	return paths
}

func tr(lang i18n.Language, en, es string) string {
	return i18n.T(lang, en, es)
}

func anyString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return ""
	}
}
