package execmd

import (
	"regexp"

	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

var slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

func isValidState(state string) bool {
	switch state {
	case "current", "to-implement", "done", "outdated":
		return true
	default:
		return false
	}
}

func effectiveLanguage(projectRootHint string) i18n.Language {
	return cmdutil.EffectiveLanguage(projectRootHint)
}

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}

func resolvePlanRoot(path string) (string, bool) {
	return workspace.ResolvePlanRoot(path)
}

func resolvePlanRootFrom(path string) (string, string, bool) {
	return workspace.ResolvePlanRootFrom(path)
}
