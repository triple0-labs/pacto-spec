package integrations

import "strings"

func normalize(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := normalize(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func joinSupported() string {
	return strings.Join(SupportedTools(), ",")
}
