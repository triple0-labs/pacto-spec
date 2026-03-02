package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"pacto/internal/yamlutil"
)

type Config struct {
	Root            string
	PlansRoot       string
	RepoRoot        string
	Mode            string
	Format          string
	FailOn          string
	State           string
	IncludeArchive  bool
	MaxNextActions  int
	MaxBlockers     int
	ClaimsPaths     bool
	ClaimsSymbols   bool
	ClaimsEndpoints bool
	ClaimsTestRefs  bool
}

func Defaults(_ string) Config {
	return Config{
		Root:            "",
		PlansRoot:       "",
		RepoRoot:        "",
		Mode:            "compat",
		Format:          "table",
		FailOn:          "none",
		State:           "all",
		IncludeArchive:  false,
		MaxNextActions:  3,
		MaxBlockers:     3,
		ClaimsPaths:     true,
		ClaimsSymbols:   true,
		ClaimsEndpoints: true,
		ClaimsTestRefs:  true,
	}
}

func Load(configPath, root string) (Config, []string, error) {
	cfg := Defaults(root)
	warnings := make([]string, 0)

	path := configPath
	if strings.TrimSpace(path) == "" {
		path = filepath.Join(root, ".pacto-engine.yaml")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	path = filepath.Clean(path)
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, warnings, nil
		}
		return cfg, warnings, err
	}
	if st.IsDir() {
		return cfg, warnings, fmt.Errorf("config path is a directory: %s", path)
	}

	raw, err := yamlutil.ReadFileMap(path)
	if err != nil {
		return cfg, warnings, err
	}
	vals := flatten(raw)
	known := map[string]bool{
		"root":                           true,
		"pacto_root":                     true,
		"plans_root":                     true,
		"repo_root":                      true,
		"mode":                           true,
		"format":                         true,
		"fail_on":                        true,
		"state":                          true,
		"include_archive":                true,
		"limits.max_next_actions":        true,
		"limits.max_blockers":            true,
		"verification.claims.paths":      true,
		"verification.claims.symbols":    true,
		"verification.claims.endpoints":  true,
		"verification.claims.test_refs":  true,
	}

	for k, v := range vals {
		switch k {
		case "root":
			cfg.Root = resolveRootValue(asString(v), filepath.Dir(path))
		case "pacto_root":
			cfg.Root = resolveRootValue(asString(v), filepath.Dir(path))
		case "plans_root":
			cfg.PlansRoot = resolveRootValue(asString(v), filepath.Dir(path))
			warnings = append(warnings, "config key 'plans_root' is deprecated for status; use 'root' and 'repo_root'")
		case "repo_root":
			cfg.RepoRoot = resolveRootValue(asString(v), filepath.Dir(path))
		case "mode":
			cfg.Mode = asString(v)
		case "format":
			cfg.Format = asString(v)
		case "fail_on":
			cfg.FailOn = asString(v)
		case "state":
			cfg.State = asString(v)
		case "include_archive":
			b, e := parseBoolAny(v)
			if e != nil {
				warnings = append(warnings, fmt.Sprintf("invalid include_archive: %q", asString(v)))
				continue
			}
			cfg.IncludeArchive = b
		case "limits.max_next_actions":
			n, e := parseIntAny(v)
			if e == nil {
				cfg.MaxNextActions = n
			}
		case "limits.max_blockers":
			n, e := parseIntAny(v)
			if e == nil {
				cfg.MaxBlockers = n
			}
		case "verification.claims.paths":
			if b, e := parseBoolAny(v); e == nil {
				cfg.ClaimsPaths = b
			}
		case "verification.claims.symbols":
			if b, e := parseBoolAny(v); e == nil {
				cfg.ClaimsSymbols = b
			}
		case "verification.claims.endpoints":
			if b, e := parseBoolAny(v); e == nil {
				cfg.ClaimsEndpoints = b
			}
		case "verification.claims.test_refs":
			if b, e := parseBoolAny(v); e == nil {
				cfg.ClaimsTestRefs = b
			}
		default:
			if !known[k] {
				warnings = append(warnings, fmt.Sprintf("unknown config key: %s", k))
			}
		}
	}

	return cfg, warnings, nil
}

func resolveRootValue(v, baseDir string) string {
	root := strings.TrimSpace(v)
	if root == "" {
		return root
	}
	if filepath.IsAbs(root) {
		return filepath.Clean(root)
	}
	return filepath.Clean(filepath.Join(baseDir, root))
}

func flatten(in map[string]any) map[string]any {
	out := map[string]any{}
	var walk func(prefix string, v any)
	walk = func(prefix string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, child := range x {
				key := k
				if prefix != "" {
					key = prefix + "." + k
				}
				walk(key, child)
			}
		default:
			if strings.TrimSpace(prefix) != "" {
				out[prefix] = v
			}
		}
	}
	for k, v := range in {
		walk(k, v)
	}
	return out
}

func parseBoolAny(v any) (bool, error) {
	switch x := v.(type) {
	case bool:
		return x, nil
	}
	s := strings.ToLower(strings.TrimSpace(asString(v)))
	switch s {
	case "true", "yes", "1", "on":
		return true, nil
	case "false", "no", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", s)
	}
}

func parseIntAny(v any) (int, error) {
	switch x := v.(type) {
	case int:
		return x, nil
	case int64:
		return int(x), nil
	case float64:
		return int(x), nil
	}
	return strconv.Atoi(strings.TrimSpace(asString(v)))
}

func asString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", v)
	}
}
