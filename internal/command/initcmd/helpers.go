package initcmd

import (
	"io"

	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
)

func setGlobalLangOverride(raw string) {
	cmdutil.SetGlobalLangOverride(raw)
}

func effectiveLanguage(projectRootHint string) i18n.Language {
	return cmdutil.EffectiveLanguage(projectRootHint)
}

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}

func displayPath(path string) string {
	return cmdutil.DisplayPath(path)
}

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}

func isTerminal(w io.Writer) bool {
	return cmdutil.IsTerminal(w)
}
