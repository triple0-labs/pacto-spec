package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DiscoverResult struct {
	Plugins []Plugin
	Errors  []error
}

func Discover(projectRoot string) DiscoverResult {
	pluginsRoot := filepath.Join(projectRoot, ".pacto", "plugins")
	ents, err := os.ReadDir(pluginsRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return DiscoverResult{}
		}
		return DiscoverResult{Errors: []error{fmt.Errorf("read plugins dir: %w", err)}}
	}

	result := DiscoverResult{}
	seen := map[string]bool{}
	for _, ent := range ents {
		if !ent.IsDir() {
			continue
		}
		dir := filepath.Join(pluginsRoot, ent.Name())
		manifestPath := filepath.Join(dir, "plugin.yaml")
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}
		m, err := parseManifest(manifestPath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: parse manifest: %w", ent.Name(), err))
			continue
		}
		if err := validateManifest(m, dir); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", ent.Name(), err))
			continue
		}
		id := strings.TrimSpace(strings.ToLower(m.Metadata.ID))
		if seen[id] {
			result.Errors = append(result.Errors, fmt.Errorf("duplicate plugin id: %s", id))
			continue
		}
		seen[id] = true
		m.Metadata.ID = id
		result.Plugins = append(result.Plugins, Plugin{Dir: dir, Manifest: m})
	}

	sort.Slice(result.Plugins, func(i, j int) bool {
		pi := result.Plugins[i].Manifest.Metadata.Priority
		pj := result.Plugins[j].Manifest.Metadata.Priority
		if pi == pj {
			return result.Plugins[i].Manifest.Metadata.ID < result.Plugins[j].Manifest.Metadata.ID
		}
		return pi < pj
	})
	return result
}

func LoadActive(projectRoot string) ([]Plugin, []error) {
	cfg, err := ReadActiveConfig(projectRoot)
	if err != nil {
		return nil, []error{err}
	}
	d := Discover(projectRoot)
	if len(cfg.Enabled) == 0 {
		return nil, d.Errors
	}
	all := map[string]Plugin{}
	for _, p := range d.Plugins {
		all[p.Manifest.Metadata.ID] = p
	}
	active := make([]Plugin, 0, len(cfg.Enabled))
	errs := append([]error{}, d.Errors...)
	for _, id := range cfg.Enabled {
		p, ok := all[id]
		if !ok {
			errs = append(errs, fmt.Errorf("enabled plugin not found: %s", id))
			continue
		}
		active = append(active, p)
	}
	sort.Slice(active, func(i, j int) bool {
		pi := active[i].Manifest.Metadata.Priority
		pj := active[j].Manifest.Metadata.Priority
		if pi == pj {
			return active[i].Manifest.Metadata.ID < active[j].Manifest.Metadata.ID
		}
		return pi < pj
	})
	return active, errs
}
