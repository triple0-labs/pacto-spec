package assets

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed templates/*.md templates_es/*.md
var templateFS embed.FS

func MustTemplate(name string) string {
	b, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		panic(fmt.Sprintf("missing embedded template %q: %v", name, err))
	}
	return string(b)
}

func MustTemplateLang(lang, name string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "es" {
		if b, err := templateFS.ReadFile("templates_es/" + name); err == nil {
			return string(b)
		}
	}
	if b, err := templateFS.ReadFile("templates/" + name); err == nil {
		return string(b)
	}
	panic(fmt.Sprintf("missing embedded template %q", name))
}
