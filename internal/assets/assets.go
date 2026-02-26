package assets

import (
	"embed"
	"fmt"
)

//go:embed templates/*.md
var templateFS embed.FS

func MustTemplate(name string) string {
	b, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		panic(fmt.Sprintf("missing embedded template %q: %v", name, err))
	}
	return string(b)
}
