package onboarding

import (
	"sort"
	"strings"
)

type Overrides struct {
	ToolsCSV string
}

func ResolveProfile(base, answered Profile, o Overrides) (Profile, error) {
	out := base
	languageFromUser := strings.EqualFold(strings.TrimSpace(answered.Sources.Languages), "user")
	toolsFromUser := strings.EqualFold(strings.TrimSpace(answered.Sources.Tools), "user")

	if languageFromUser {
		out.Languages = normalizeList(answered.Languages)
		out.CustomLanguages = normalizeList(answered.CustomLanguages)
		out.Sources.Languages = "user"
	} else if len(answered.Languages) > 0 {
		out.Languages = normalizeList(answered.Languages)
		out.Sources.Languages = "user"
		if len(answered.CustomLanguages) > 0 {
			out.CustomLanguages = normalizeList(answered.CustomLanguages)
		}
	}
	if toolsFromUser {
		out.Tools = normalizeList(answered.Tools)
		out.CustomTools = normalizeList(answered.CustomTools)
		out.Sources.Tools = "user"
	} else if len(answered.Tools) > 0 {
		out.Tools = normalizeList(answered.Tools)
		out.Sources.Tools = "user"
		if len(answered.CustomTools) > 0 {
			out.CustomTools = normalizeList(answered.CustomTools)
		}
	}
	if strings.TrimSpace(answered.Intents.Problem) != "" {
		out.Intents.Problem = strings.TrimSpace(answered.Intents.Problem)
	}

	if strings.TrimSpace(o.ToolsCSV) != "" {
		if strings.EqualFold(strings.TrimSpace(o.ToolsCSV), "all") {
			out.Tools = append([]string{}, KnownTools...)
			out.CustomTools = nil
		} else if strings.EqualFold(strings.TrimSpace(o.ToolsCSV), "none") {
			out.Tools = nil
			out.CustomTools = nil
		} else {
			tools, custom := splitKnownUnknown(o.ToolsCSV, IsKnownTool)
			out.Tools = tools
			out.CustomTools = custom
		}
		out.Sources.Tools = "flag"
	}

	out.Languages = normalizeList(out.Languages)
	out.CustomLanguages = normalizeList(out.CustomLanguages)
	out.Tools = normalizeList(out.Tools)
	out.CustomTools = normalizeList(out.CustomTools)
	return out, nil
}

func splitCSV(raw string) []string {
	tokens := strings.Split(raw, ",")
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		t := strings.TrimSpace(strings.ToLower(tok))
		if t != "" {
			out = append(out, t)
		}
	}
	return normalizeList(out)
}

func splitKnownUnknown(raw string, knownFn func(string) bool) (known []string, unknown []string) {
	for _, tok := range splitCSV(raw) {
		if knownFn(tok) {
			known = append(known, tok)
		} else {
			unknown = append(unknown, tok)
		}
	}
	return normalizeList(known), normalizeList(unknown)
}

func normalizeList(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, x := range in {
		t := strings.TrimSpace(strings.ToLower(x))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
