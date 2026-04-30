package move

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"pacto/internal/i18n"
)

var slugRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// MovePlan handles the business logic to move a plan from one state to another.
func MovePlan(plansRoot, fromState, slug, toState string, force bool, reason string, lang i18n.Language) (string, error) {
	fromState = strings.ToLower(strings.TrimSpace(fromState))
	slug = strings.TrimSpace(slug)
	toState = strings.ToLower(strings.TrimSpace(toState))

	if !isValidState(fromState) || !isValidState(toState) {
		return "", fmt.Errorf("invalid state transition %q -> %q (allowed: current|to-implement|done|outdated)", fromState, toState)
	}
	if fromState == toState {
		return "", errors.New("source and destination states are the same")
	}
	if !slugRe.MatchString(slug) {
		return "", fmt.Errorf("invalid slug %q (use lowercase letters, numbers, dashes)", slug)
	}

	srcDir := filepath.Join(plansRoot, fromState, slug)
	dstDir := filepath.Join(plansRoot, toState, slug)
	if _, err := os.Stat(filepath.Join(srcDir, "README.md")); err != nil {
		return "", fmt.Errorf("source plan not found: %s/%s", fromState, slug)
	}
	if _, err := os.Stat(dstDir); err == nil {
		if !force {
			return "", fmt.Errorf("destination already exists: %s (use --force to overwrite)", dstDir)
		}
		if err := os.RemoveAll(dstDir); err != nil {
			return "", fmt.Errorf("remove destination: %w", err)
		}
	}

	if err := os.Rename(srcDir, dstDir); err != nil {
		return "", fmt.Errorf("move plan directory: %w", err)
	}

	readmePath := filepath.Join(dstDir, "README.md")
	if err := rewritePlanReadmeStatus(readmePath, toState, fromState, reason); err != nil {
		return "", fmt.Errorf("update moved README: %w", err)
	}
	return readmePath, nil
}

func isValidState(s string) bool {
	return s == "current" || s == "to-implement" || s == "done" || s == "outdated"
}

func Tr(l i18n.Language, en, es string) string {
	if l == i18n.Spanish {
		return es
	}
	return en
}

func rewritePlanReadmeStatus(path, toState, fromState, reason string) error {
	lang := i18n.English
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(b)
	lines := strings.Split(text, "\n")
	newStatus := stateStatusLabel(toState, lang)
	updated := false
	for i, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if strings.HasPrefix(trimmed, "**Status:**") || strings.HasPrefix(trimmed, "**Estado:**") {
			lines[i] = Tr(lang, "**Status:** ", "**Estado:** ") + newStatus + "  "
			updated = true
			break
		}
	}
	if !updated {
		lines = append([]string{Tr(lang, "**Status:** ", "**Estado:** ") + newStatus + "  "}, lines...)
	}
	text = strings.Join(lines, "\n")
	if strings.TrimSpace(reason) != "" {
		note := fmt.Sprintf("- %s %s `%s` %s `%s`: %s", time.Now().Format("2006-01-02 15:04"), Tr(lang, "moved from", "movido de"), fromState, Tr(lang, "to", "a"), toState, strings.TrimSpace(reason))
		text = appendSectionBullet(text, Tr(lang, "## Move History", "## Historial de cambios"), note)
	}
	return os.WriteFile(path, []byte(text), 0o664)
}

func stateStatusLabel(state string, lang i18n.Language) string {
	en := map[string]string{
		"current":      "In Progress (Current)",
		"to-implement": "Pending (To Implement)",
		"done":         "Completed (Done)",
		"outdated":     "Outdated (Outdated)",
	}[state]
	es := map[string]string{
		"current":      "En ejecución (Current)",
		"to-implement": "Pendiente (To Implement)",
		"done":         "Completado (Done)",
		"outdated":     "Obsoleto (Outdated)",
	}[state]
	return Tr(lang, en, es)
}

func appendSectionBullet(text, header, bullet string) string {
	lines := strings.Split(text, "\n")
	start := -1
	for i, ln := range lines {
		if strings.TrimSpace(ln) == header {
			start = i
			break
		}
	}
	if start < 0 {
		return text + "\n" + header + "\n\n" + bullet + "\n"
	}
	out := append([]string{}, lines[:start+1]...)
	if start+1 < len(lines) && strings.TrimSpace(lines[start+1]) == "" {
		out = append(out, "")
		start++
	}
	out = append(out, bullet)
	out = append(out, lines[start+1:]...)
	return strings.Join(out, "\n")
}
