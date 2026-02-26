package claims

import (
	"regexp"
	"strings"

	"pacto/internal/model"
	"pacto/internal/parser"
)

var (
	reBacktick = regexp.MustCompile("`([^`]+)`")
	reMDLink   = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	reVerbEP   = regexp.MustCompile(`(?i)\b(GET|POST|PUT|PATCH|DELETE)\s+(/[-A-Za-z0-9_/{}/.:]+)`)
	reAPIPath  = regexp.MustCompile(`\b(/api/[-A-Za-z0-9_/{}/.:]+)`)
)

type Options struct {
	Paths     bool
	Symbols   bool
	Endpoints bool
	TestRefs  bool
}

func Extract(p parser.ParsedPlan, opts Options) []model.ClaimResult {
	claims := make([]model.ClaimResult, 0)
	text := p.RawText

	if opts.Paths {
		for _, m := range reMDLink.FindAllStringSubmatch(text, -1) {
			v := strings.TrimSpace(m[1])
			if looksLikePath(v) {
				claims = append(claims, model.ClaimResult{ClaimType: model.ClaimPath, SourceText: v, Evidence: "markdown_link"})
			}
		}
	}

	for _, m := range reBacktick.FindAllStringSubmatch(text, -1) {
		v := strings.TrimSpace(m[1])
		if v == "" {
			continue
		}
		if opts.Paths && looksLikePath(v) {
			claims = append(claims, model.ClaimResult{ClaimType: model.ClaimPath, SourceText: v, Evidence: "inline_code"})
			continue
		}
		if opts.TestRefs && looksLikeTestRef(v) {
			claims = append(claims, model.ClaimResult{ClaimType: model.ClaimTestRef, SourceText: v, Evidence: "inline_code"})
			continue
		}
		if opts.Symbols && looksLikeSymbol(v) {
			claims = append(claims, model.ClaimResult{ClaimType: model.ClaimSymbol, SourceText: v, Evidence: "inline_code"})
		}
	}

	if opts.Endpoints {
		for _, m := range reVerbEP.FindAllStringSubmatch(text, -1) {
			claims = append(claims, model.ClaimResult{ClaimType: model.ClaimEndpoint, SourceText: strings.ToUpper(m[1]) + " " + m[2], Evidence: "verb_endpoint"})
		}
		for _, m := range reAPIPath.FindAllStringSubmatch(text, -1) {
			claims = append(claims, model.ClaimResult{ClaimType: model.ClaimEndpoint, SourceText: m[1], Evidence: "api_path"})
		}
	}

	return dedupe(claims)
}

func dedupe(in []model.ClaimResult) []model.ClaimResult {
	seen := map[string]struct{}{}
	out := make([]model.ClaimResult, 0, len(in))
	for _, c := range in {
		k := string(c.ClaimType) + "::" + c.SourceText
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, c)
		if len(out) >= 120 {
			break
		}
	}
	return out
}

func looksLikePath(v string) bool {
	if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") {
		return false
	}
	if strings.Contains(v, " ") {
		return false
	}
	return strings.Contains(v, "/") || strings.HasPrefix(v, "./") || strings.HasPrefix(v, "../") || strings.HasSuffix(v, ".md") || strings.HasSuffix(v, ".go") || strings.HasSuffix(v, ".py") || strings.HasSuffix(v, ".ts") || strings.HasSuffix(v, ".tsx")
}

func looksLikeSymbol(v string) bool {
	if strings.Contains(v, "/") || strings.Contains(v, " ") {
		return false
	}
	if len(v) < 4 {
		return false
	}
	if strings.Contains(strings.ToLower(v), "fase") || strings.Contains(strings.ToLower(v), "plan") {
		return false
	}
	return strings.Contains(v, "_") || strings.Contains(v, ".") || strings.Contains(v, "(") || strings.ContainsAny(v, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
}

func looksLikeTestRef(v string) bool {
	l := strings.ToLower(v)
	return strings.Contains(l, "test") || strings.Contains(l, "pytest") || strings.Contains(l, "npm run") || strings.Contains(l, "go test")
}
