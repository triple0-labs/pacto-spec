package builtin

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

var ErrUnknownPlugin = errors.New("unknown built-in plugin")

//go:embed assets
var assetsFS embed.FS

type PluginInfo struct {
	ID      string `json:"id"`
	Summary string `json:"summary"`
}

type InstallOptions struct {
	Force bool
}

type InstallResult struct {
	PluginID string
	Created  []string
	Updated  []string
	Skipped  []string
}

var catalog = map[string]PluginInfo{
	"git-sync": {
		ID:      "git-sync",
		Summary: "Sync git fetch/pull context before pacto status.",
	},
}

func ListAvailable() []PluginInfo {
	out := make([]PluginInfo, 0, len(catalog))
	for _, p := range catalog {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func Install(projectRoot, rawID string, opts InstallOptions) (InstallResult, error) {
	id := strings.ToLower(strings.TrimSpace(rawID))
	if id == "" {
		return InstallResult{}, fmt.Errorf("plugin id is required")
	}
	if _, ok := catalog[id]; !ok {
		return InstallResult{}, fmt.Errorf("%w: %s", ErrUnknownPlugin, id)
	}

	pluginRoot := filepath.Join(projectRoot, ".pacto", "plugins", id)
	if err := os.MkdirAll(pluginRoot, 0o775); err != nil {
		return InstallResult{}, err
	}

	prefix := path.Join("assets", id)
	result := InstallResult{PluginID: id}
	if err := fs.WalkDir(assetsFS, prefix, func(assetPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel := strings.TrimPrefix(assetPath, prefix)
		rel = strings.TrimPrefix(rel, "/")
		if rel == "" {
			return nil
		}
		dst := filepath.Join(pluginRoot, filepath.FromSlash(rel))
		if d.IsDir() {
			return os.MkdirAll(dst, 0o775)
		}

		b, err := assetsFS.ReadFile(assetPath)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o775); err != nil {
			return err
		}
		if _, err := os.Stat(dst); err == nil && !opts.Force {
			result.Skipped = append(result.Skipped, dst)
			return nil
		}
		mode := os.FileMode(0o664)
		if strings.HasSuffix(strings.ToLower(dst), ".sh") {
			mode = 0o775
		}
		_, existedErr := os.Stat(dst)
		existed := existedErr == nil
		if err := os.WriteFile(dst, b, mode); err != nil {
			return err
		}
		if existed {
			result.Updated = append(result.Updated, dst)
		} else {
			result.Created = append(result.Created, dst)
		}
		return nil
	}); err != nil {
		return InstallResult{}, err
	}

	examplePath := filepath.Join(pluginRoot, "config.env.example")
	configPath := filepath.Join(pluginRoot, "config.env")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		b, readErr := os.ReadFile(examplePath)
		if readErr != nil {
			return InstallResult{}, readErr
		}
		if err := os.WriteFile(configPath, b, 0o664); err != nil {
			return InstallResult{}, err
		}
		result.Created = append(result.Created, configPath)
	}

	sort.Strings(result.Created)
	sort.Strings(result.Updated)
	sort.Strings(result.Skipped)
	return result, nil
}
