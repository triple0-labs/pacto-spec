package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/yamlutil"
)

func ReadActiveConfig(projectRoot string) (ActiveConfig, error) {
	path := filepath.Join(projectRoot, ".pacto", "config.yaml")
	m, err := yamlutil.ReadFileMap(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ActiveConfig{}, nil
		}
		return ActiveConfig{}, err
	}

	plugins := yamlutil.GetMap(m, "plugins")
	if plugins == nil {
		return ActiveConfig{}, nil
	}
	return ActiveConfig{Enabled: yamlutil.ToStringSlice(plugins["enabled"])}, nil
}

func WriteActiveConfig(projectRoot string, enabled []string) error {
	cfgPath := filepath.Join(projectRoot, ".pacto", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o775); err != nil {
		return err
	}

	m, err := yamlutil.ReadMapOrDefault(cfgPath)
	if err != nil {
		return err
	}
	if _, ok := m["version"]; !ok {
		m["version"] = 1
	}
	plugins := yamlutil.EnsureMap(m, "plugins")
	plugins["enabled"] = yamlutil.NormalizeStringSlice(enabled)
	return yamlutil.WriteMap(cfgPath, m)
}

func Enable(projectRoot, id string) error {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return fmt.Errorf("plugin id is required")
	}
	cfg, err := ReadActiveConfig(projectRoot)
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(cfg.Enabled)+1)
	for _, x := range cfg.Enabled {
		t := strings.ToLower(strings.TrimSpace(x))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	if !seen[id] {
		out = append(out, id)
	}
	sort.Strings(out)
	return WriteActiveConfig(projectRoot, out)
}

func Disable(projectRoot, id string) error {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "" {
		return fmt.Errorf("plugin id is required")
	}
	cfg, err := ReadActiveConfig(projectRoot)
	if err != nil {
		return err
	}
	out := make([]string, 0, len(cfg.Enabled))
	for _, x := range cfg.Enabled {
		t := strings.ToLower(strings.TrimSpace(x))
		if t == "" || t == id {
			continue
		}
		out = append(out, t)
	}
	sort.Strings(out)
	return WriteActiveConfig(projectRoot, out)
}
