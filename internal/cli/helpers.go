package cli

import (
	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
	"pacto/internal/workspace"
)

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}

func cleanAbs(path string) string {
	return workspace.CleanAbs(path)
}
