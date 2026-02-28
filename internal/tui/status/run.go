package status

import (
	"pacto/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func Run(report model.StatusReport) error {
	p := tea.NewProgram(New(report), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
