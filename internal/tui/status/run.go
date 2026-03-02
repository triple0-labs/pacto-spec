package status

import (
	"pacto/internal/i18n"
	"pacto/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(report model.StatusReport, lang i18n.Language) error {
	p := tea.NewProgram(New(report, lang), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
