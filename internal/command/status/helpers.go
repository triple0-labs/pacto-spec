package statuscmd

import (
	"io"

	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

var isTerminalFn = cmdutil.IsTerminal

func setGlobalLangOverride(raw string) {
	cmdutil.SetGlobalLangOverride(raw)
}

func effectiveLanguage(projectRootHint string) i18n.Language {
	return cmdutil.EffectiveLanguage(projectRootHint)
}

func isTerminal(w io.Writer) bool {
	return isTerminalFn(w)
}

func resolvePlanRoot(path string) (string, bool) {
	return workspace.ResolvePlanRoot(path)
}

func resolvePlanRootFrom(path string) (string, string, bool) {
	return workspace.ResolvePlanRootFrom(path)
}

func hasStateDirs(path string) bool {
	return workspace.HasStateDirs(path)
}
