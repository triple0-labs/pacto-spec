package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"pacto/internal/plugins"
	"pacto/internal/plugins/builtin"
)

func RunPlugin(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: pacto plugin <list|list-available|install|validate|enable|disable> [options]")
		return 2
	}
	sub := strings.ToLower(strings.TrimSpace(args[0]))
	rest := args[1:]
	switch sub {
	case "list":
		return runPluginList(rest)
	case "list-available":
		return runPluginListAvailable(rest)
	case "install":
		return runPluginInstall(rest)
	case "validate":
		return runPluginValidate(rest)
	case "enable":
		return runPluginEnable(rest)
	case "disable":
		return runPluginDisable(rest)
	default:
		fmt.Fprintf(os.Stderr, "unknown plugin subcommand: %s\n", sub)
		return 2
	}
}

func runPluginListAvailable(args []string) int {
	fs := flag.NewFlagSet("plugin list-available", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	format := fs.String("format", "table", "Output format: table|json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "plugin list-available does not accept positional args")
		return 2
	}
	infos := builtin.ListAvailable()
	switch strings.ToLower(strings.TrimSpace(*format)) {
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
		fmt.Fprintf(os.Stderr, "unsupported format %q (allowed: table|json)\n", *format)
		return 2
	}
}

func runPluginInstall(args []string) int {
	fs := flag.NewFlagSet("plugin install", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "Project root path")
	force := fs.Bool("force", false, "Overwrite existing plugin files")
	noEnable := fs.Bool("no-enable", false, "Install files but do not auto-enable plugin")
	normalizedArgs, normErr := normalizeArgs(args, map[string]bool{"--root": true, "-root": true})
	if normErr != nil {
		fmt.Fprintf(os.Stderr, "parse args: %v\n", normErr)
		return 2
	}
	if err := fs.Parse(normalizedArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pacto plugin install <id> [--root <path>] [--force] [--no-enable]")
		return 2
	}
	projectRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id := strings.ToLower(strings.TrimSpace(fs.Args()[0]))
	res, err := builtin.Install(projectRoot, id, builtin.InstallOptions{Force: *force})
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

	if !*noEnable {
		if err := plugins.Enable(projectRoot, id); err != nil {
			fmt.Fprintf(os.Stderr, "enable plugin: %v\n", err)
			return 3
		}
		fmt.Printf("Enabled plugin %s\n", id)
	}
	return 0
}

func runPluginList(args []string) int {
	fs := flag.NewFlagSet("plugin list", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "Project root path")
	format := fs.String("format", "table", "Output format: table|json")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "plugin list does not accept positional args")
		return 2
	}
	projectRoot, err := filepath.Abs(*root)
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
	if *format == "json" {
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

func runPluginValidate(args []string) int {
	fs := flag.NewFlagSet("plugin validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "Project root path")
	pluginID := fs.String("plugin", "", "Validate a specific plugin ID")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	projectRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	d := plugins.Discover(projectRoot)
	if strings.TrimSpace(*pluginID) != "" {
		id := strings.ToLower(strings.TrimSpace(*pluginID))
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

func runPluginEnable(args []string) int {
	fs := flag.NewFlagSet("plugin enable", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "Project root path")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pacto plugin enable <id> [--root <path>]")
		return 2
	}
	projectRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id := strings.ToLower(strings.TrimSpace(fs.Args()[0]))
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

func runPluginDisable(args []string) int {
	fs := flag.NewFlagSet("plugin disable", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "Project root path")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		return 2
	}
	if len(fs.Args()) != 1 {
		fmt.Fprintln(os.Stderr, "usage: pacto plugin disable <id> [--root <path>]")
		return 2
	}
	projectRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve root: %v\n", err)
		return 2
	}
	id := strings.ToLower(strings.TrimSpace(fs.Args()[0]))
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
