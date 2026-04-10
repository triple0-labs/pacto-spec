package plugincmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/plugins"
	"pacto/internal/plugins/builtin"
)

type ListAvailableOptions struct {
	Format string
}

type InstallOptions struct {
	Root     string
	Force    bool
	NoEnable bool
}

type ListOptions struct {
	Root   string
	Format string
}

type ValidateOptions struct {
	Root     string
	PluginID string
}

type EnableOptions struct {
	Root string
}

type DisableOptions struct {
	Root string
}

func RunListAvailable(opts ListAvailableOptions) int {
	infos := builtin.ListAvailable()
	switch strings.ToLower(strings.TrimSpace(opts.Format)) {
	case "json":
		payload := map[string]any{"plugins": infos}
		enc, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(enc))
		return 0
	case "table", "":
		fmt.Println("Available Plugins")
		for _, p := range infos {
			fmt.Printf("- %s\n", p.ID)
			fmt.Printf("  summary: %s\n", p.Summary)
		}
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unsupported format %q (allowed: table|json)\n", opts.Format)
		return 2
	}
}

func RunInstall(opts InstallOptions, id string) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id = strings.ToLower(strings.TrimSpace(id))
	res, err := builtin.Install(projectRoot, id, builtin.InstallOptions{Force: opts.Force})
	if err != nil {
		fmt.Fprintf(os.Stderr, "install plugin: %v\n", err)
		if errors.Is(err, builtin.ErrUnknownPlugin) {
			return 2
		}
		return 3
	}

	fmt.Printf("Installed plugin %s\n", res.PluginID)
	for _, p := range res.Created {
		fmt.Println(pathLine("created", p))
	}
	for _, p := range res.Updated {
		fmt.Println(pathLine("updated", p))
	}
	for _, p := range res.Skipped {
		fmt.Println(pathLine("skipped", p))
	}

	if !opts.NoEnable {
		if err := plugins.Enable(projectRoot, id); err != nil {
			fmt.Fprintf(os.Stderr, "enable plugin: %v\n", err)
			return 3
		}
		fmt.Printf("Enabled plugin %s\n", id)
	}
	return 0
}

func RunList(opts ListOptions) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	activeCfg, err := plugins.ReadActiveConfig(projectRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read config: %v\n", err)
		return 2
	}
	enabled := map[string]bool{}
	for _, id := range activeCfg.Enabled {
		enabled[strings.ToLower(strings.TrimSpace(id))] = true
	}
	d := plugins.Discover(projectRoot)
	type row struct {
		ID       string `json:"id"`
		Version  string `json:"version"`
		Priority int    `json:"priority"`
		Enabled  bool   `json:"enabled"`
		Path     string `json:"path"`
	}
	rows := make([]row, 0, len(d.Plugins))
	for _, p := range d.Plugins {
		rows = append(rows, row{
			ID:       p.Manifest.Metadata.ID,
			Version:  p.Manifest.Metadata.Version,
			Priority: p.Manifest.Metadata.Priority,
			Enabled:  enabled[p.Manifest.Metadata.ID],
			Path:     p.Dir,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	if opts.Format == "json" {
		payload := map[string]any{"plugins": rows, "errors": errorStrings(d.Errors)}
		enc, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(enc))
		if len(d.Errors) > 0 {
			return 3
		}
		return 0
	}
	fmt.Println("Plugins")
	for _, r := range rows {
		state := "disabled"
		if r.Enabled {
			state = "enabled"
		}
		fmt.Printf("- %s (%s) priority=%d state=%s\n", r.ID, r.Version, r.Priority, state)
		fmt.Printf("  path: %s\n", r.Path)
	}
	for _, err := range d.Errors {
		fmt.Fprintf(os.Stderr, "plugin error: %v\n", err)
	}
	if len(d.Errors) > 0 {
		return 3
	}
	return 0
}

func RunValidate(opts ValidateOptions) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	d := plugins.Discover(projectRoot)
	if strings.TrimSpace(opts.PluginID) != "" {
		id := strings.ToLower(strings.TrimSpace(opts.PluginID))
		found := false
		for _, p := range d.Plugins {
			if p.Manifest.Metadata.ID == id {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintf(os.Stderr, "plugin not found: %s\n", id)
			return 2
		}
	}
	if len(d.Errors) > 0 {
		for _, e := range d.Errors {
			fmt.Fprintf(os.Stderr, "plugin validation error: %v\n", e)
		}
		return 3
	}
	fmt.Printf("Validated %d plugin(s)\n", len(d.Plugins))
	return 0
}

func RunEnable(opts EnableOptions, id string) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id = strings.ToLower(strings.TrimSpace(id))
	d := plugins.Discover(projectRoot)
	exists := false
	for _, p := range d.Plugins {
		if p.Manifest.Metadata.ID == id {
			exists = true
			break
		}
	}
	if !exists {
		fmt.Fprintf(os.Stderr, "plugin not found: %s\n", id)
		return 2
	}
	if err := plugins.Enable(projectRoot, id); err != nil {
		fmt.Fprintf(os.Stderr, "enable plugin: %v\n", err)
		return 3
	}
	fmt.Printf("Enabled plugin %s\n", id)
	return 0
}

func RunDisable(opts DisableOptions, id string) int {
	if strings.TrimSpace(opts.Root) == "" {
		opts.Root = "."
	}
	projectRoot, err := filepath.Abs(opts.Root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id = strings.ToLower(strings.TrimSpace(id))
	if err := plugins.Disable(projectRoot, id); err != nil {
		fmt.Fprintf(os.Stderr, "disable plugin: %v\n", err)
		return 3
	}
	fmt.Printf("Disabled plugin %s\n", id)
	return 0
}

func errorStrings(errs []error) []string {
	if len(errs) == 0 {
		return nil
	}
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.Error())
	}
	return out
}
