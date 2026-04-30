package status

import (
	"pacto/internal/domain/report"
	"pacto/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(report report.StatusReport, lang i18n.Language) error {
	p := tea.NewProgram(New(report, lang), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
