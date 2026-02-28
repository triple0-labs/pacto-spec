package onboarding

import "strings"

var (
	KnownLanguages = []string{"go", "typescript", "javascript", "python", "rust", "java", "dotnet", "ruby", "php"}
	KnownTools     = []string{"codex", "cursor", "claude", "opencode"}
)

func IsKnownLanguage(v string) bool {
	return containsFold(KnownLanguages, v)
}

func IsKnownTool(v string) bool {
	return containsFold(KnownTools, v)
}

func NormalizeToken(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func containsFold(items []string, needle string) bool {
	n := NormalizeToken(needle)
	for _, it := range items {
		if NormalizeToken(it) == n {
			return true
		}
	}
	return false
}
