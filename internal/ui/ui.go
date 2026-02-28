package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	dimStyle   = lipgloss.NewStyle().Faint(true)
	okStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	warnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

func supportsColor() bool {
	if strings.TrimSpace(os.Getenv("NO_COLOR")) != "" {
		return false
	}
	return true
}

func render(style lipgloss.Style, text string) string {
	if !supportsColor() {
		return text
	}
	return style.Render(text)
}

func Title(text string) string {
	return render(titleStyle, text)
}

func OK(text string) string {
	return render(okStyle, text)
}

func Warn(text string) string {
	return render(warnStyle, text)
}

func Err(text string) string {
	return render(errStyle, text)
}

func Dim(text string) string {
	return render(dimStyle, text)
}

func ActionHeader(action, target string) string {
	return fmt.Sprintf("%s %s", Title(action), target)
}

func PathLine(kind, path string) string {
	switch kind {
	case "created":
		return fmt.Sprintf("%s %s", OK("+"), path)
	case "updated":
		return fmt.Sprintf("%s %s", Warn("~"), path)
	case "skipped":
		return fmt.Sprintf("%s %s", Dim("="), path)
	default:
		return fmt.Sprintf("- %s", path)
	}
}

func Bullet(text string) string {
	return fmt.Sprintf("- %s", text)
}
