package yamlutil

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

func ReadFileMap(path string) (map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return UnmarshalMap(b)
}

func UnmarshalMap(data []byte) (map[string]any, error) {
	out := map[string]any{}
	if len(strings.TrimSpace(string(data))) == 0 {
		return out, nil
	}
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = map[string]any{}
	}
	return out, nil
}

func MarshalMap(m map[string]any) ([]byte, error) {
	b, err := yaml.Marshal(m)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 || b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}
	return b, nil
}

func GetMap(m map[string]any, key string) map[string]any {
	v, ok := m[key]
	if !ok {
		return nil
	}
	child, ok := v.(map[string]any)
	if ok {
		return child
	}
	return nil
}

func EnsureMap(m map[string]any, key string) map[string]any {
	if child := GetMap(m, key); child != nil {
		return child
	}
	child := map[string]any{}
	m[key] = child
	return child
}

func ToStringSlice(v any) []string {
	if v == nil {
		return nil
	}
	switch x := v.(type) {
	case []any:
		out := make([]string, 0, len(x))
		seen := map[string]bool{}
		for _, it := range x {
			s := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", it)))
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
		}
		return out
	case []string:
		out := make([]string, 0, len(x))
		seen := map[string]bool{}
		for _, it := range x {
			s := strings.TrimSpace(strings.ToLower(it))
			if s == "" || seen[s] {
				continue
			}
			seen[s] = true
			out = append(out, s)
		}
		return out
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return nil
		}
		s = strings.TrimPrefix(s, "[")
		s = strings.TrimSuffix(s, "]")
		parts := strings.Split(s, ",")
		out := make([]string, 0, len(parts))
		seen := map[string]bool{}
		for _, p := range parts {
			t := strings.TrimSpace(strings.Trim(p, `"'`))
			t = strings.ToLower(t)
			if t == "" || seen[t] {
				continue
			}
			seen[t] = true
			out = append(out, t)
		}
		return out
	default:
		return nil
	}
}

func NormalizeStringSlice(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]bool{}
	for _, it := range items {
		s := strings.TrimSpace(strings.ToLower(it))
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func WriteMap(path string, m map[string]any) error {
	b, err := MarshalMap(m)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o775); err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o664)
}

func ReadMapOrDefault(path string) (map[string]any, error) {
	m, err := ReadFileMap(path)
	if err == nil {
		return m, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	return nil, err
}
