package explorecmd

import (
	"pacto/internal/cmdutil"
	"pacto/internal/i18n"
)

func tr(lang i18n.Language, en, es string) string {
	return cmdutil.Tr(lang, en, es)
}
