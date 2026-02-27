package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	vals, parseWarnings, err := parseSimpleYAML(path)
	warnings = append(warnings, parseWarnings...)
	if err != nil {
		return cfg, warnings, err
	}
	for k, v := range vals {
		switch k {
		case "root":
			cfg.Root = resolveRootValue(v, filepath.Dir(path))
			warnings = append(warnings, "config key 'root' is deprecated for status; use 'plans_root' and 'repo_root'")
		case "plans_root":
			cfg.PlansRoot = resolveRootValue(v, filepath.Dir(path))
		case "repo_root":
			cfg.RepoRoot = resolveRootValue(v, filepath.Dir(path))
		case "mode":
			cfg.Mode = v
		case "format":
			cfg.Format = v
		case "fail_on":
			cfg.FailOn = v
		case "state":
			cfg.State = v
		case "include_archive":
			b, e := parseBool(v)
			if e != nil {
				warnings = append(warnings, fmt.Sprintf("invalid include_archive: %q", v))
				continue
			}
			cfg.IncludeArchive = b
		case "limits.max_next_actions":
			n, e := strconv.Atoi(v)
			if e == nil {
				cfg.MaxNextActions = n
			}
		case "limits.max_blockers":
			n, e := strconv.Atoi(v)
			if e == nil {
				cfg.MaxBlockers = n
			}
		case "verification.claims.paths":
			if b, e := parseBool(v); e == nil {
				cfg.ClaimsPaths = b
			}
		case "verification.claims.symbols":
			if b, e := parseBool(v); e == nil {
				cfg.ClaimsSymbols = b
			}
		case "verification.claims.endpoints":
			if b, e := parseBool(v); e == nil {
				cfg.ClaimsEndpoints = b
			}
		case "verification.claims.test_refs":
			if b, e := parseBool(v); e == nil {
				cfg.ClaimsTestRefs = b
			}
		default:
			warnings = append(warnings, fmt.Sprintf("unknown config key: %s", k))
		}
	}

	return cfg, warnings, nil
}

func parseSimpleYAML(path string) (map[string]string, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	out := make(map[string]string)
	warnings := make([]string, 0)
	stack := []string{}

	s := bufio.NewScanner(f)
	for s.Scan() {
		raw := s.Text()
		line := stripComment(raw)
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := countIndent(line)
		trimmed := strings.TrimSpace(line)
		key, val, ok := splitKeyVal(trimmed)
		if !ok {
			warnings = append(warnings, fmt.Sprintf("could not parse line: %s", raw))
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		level := indent / 2
		if level < len(stack) {
			stack = stack[:level]
		}

		if val == "" {
			stack = append(stack, key)
			continue
		}
		full := append(append([]string{}, stack...), key)
		out[strings.Join(full, ".")] = strings.Trim(val, "\"'")
	}
	if err := s.Err(); err != nil {
		return nil, warnings, err
	}
	return out, warnings, nil
}

func stripComment(line string) string {
	inSingle := false
	inDouble := false
	escaped := false
	for i, ch := range line {
		switch ch {
		case '\\':
			if inDouble {
				escaped = !escaped
			}
			continue
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !escaped {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return line[:i]
			}
		}
		escaped = false
	}
	return line
}

func splitKeyVal(line string) (key string, val string, ok bool) {
	inSingle := false
	inDouble := false
	escaped := false
	for i, ch := range line {
		switch ch {
		case '\\':
			if inDouble {
				escaped = !escaped
			}
			continue
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle && !escaped {
				inDouble = !inDouble
			}
		case ':':
			if !inSingle && !inDouble {
				return line[:i], line[i+1:], true
			}
		}
		escaped = false
	}
	return "", "", false
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

func countIndent(s string) int {
	n := 0
	for _, ch := range s {
		if ch == ' ' {
			n++
			continue
		}
		break
	}
	return n
}

func parseBool(s string) (bool, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "true", "yes", "1", "on":
		return true, nil
	case "false", "no", "0", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", s)
	}
}
