package onboarding

import "strings"

type ValidationResult struct {
	Errors   []string
	Warnings []string
}

func ValidateProfile(p Profile) ValidationResult {
	res := ValidationResult{
		Errors:   []string{},
		Warnings: []string{},
	}

	if len(p.Languages) == 0 && len(p.CustomLanguages) == 0 {
		res.Errors = append(res.Errors, "at least one technology is required")
	}
	if strings.TrimSpace(p.Intents.Problem) == "" {
		res.Errors = append(res.Errors, "problem is required")
	}
	if len(p.Tools) == 0 && len(p.CustomTools) == 0 {
		res.Warnings = append(res.Warnings, "no install targets selected")
	}
	return res
}
