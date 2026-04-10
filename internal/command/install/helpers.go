package installcmd

import (
	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

var Version = cmdutil.Version

func effectiveLanguage(projectRootHint string) i18n.Language {
	return cmdutil.EffectiveLanguage(projectRootHint)
}

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}

func cleanAbs(path string) string {
	return workspace.CleanAbs(path)
}

func displayPath(path string) string {
	return cmdutil.DisplayPath(path)
}

func pathLine(kind, path string) string {
	return cmdutil.PathLine(kind, path)
}

func promptYesNo(defaultYes bool) bool {
	return cmdutil.PromptYesNo(defaultYes)
}
