package execmd

import (
	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
)

func effectiveLanguage(projectRootHint string) i18n.Language {
	return cmdutil.EffectiveLanguage(projectRootHint)
}

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}
