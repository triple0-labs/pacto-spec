package markdown

import (
	"regexp"
	"strings"
)

var (
	ReCreatedAt = regexp.MustCompile(`(?m)^\*\*(?:Created At|Creado):\*\*\s*(.+)$`)
	ReUpdatedAt = regexp.MustCompile(`(?m)^\*\*(?:Updated At|Actualizado):\*\*\s*(.+)$`)
)

// ExtractTitle finds the first heading level 1 (# ...)
func ExtractTitle(content string) string {
	for _, ln := range strings.Split(content, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(ln, "# "))
		}
	}
	return "Untitled"
}

// ExtractStamp gets the date stamp matching the provided regex.
func ExtractStamp(re *regexp.Regexp, content string) string {
	m := re.FindStringSubmatch(content)
	if len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return "-"
}

// SetUpdatedAt replaces or inserts an updated at timestamp
func SetUpdatedAt(content, now string) string {
	if ReUpdatedAt.MatchString(content) {
		return ReUpdatedAt.ReplaceAllString(content, "**Updated At:** "+now)
	}
	updatedLabel := "**Updated At:** "
	if strings.Contains(content, "**Creado:**") || strings.Contains(content, "## Notas") {
		updatedLabel = "**Actualizado:** "
	}
	createdLabel := "**Created At:** "
	if updatedLabel == "**Actualizado:** " {
		createdLabel = "**Creado:** "
	}
	title := ExtractTitle(content)
	head := "# " + title + "\n\n" + createdLabel + ExtractStamp(ReCreatedAt, content) + "  \n" + updatedLabel + now + "\n"
	body := content
	if idx := strings.Index(content, "\n"); idx >= 0 {
		body = content[idx+1:]
	}
	return strings.TrimRight(head+"\n"+body, "\n") + "\n"
}
