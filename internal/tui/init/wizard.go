package initui

import (
	"fmt"
	"sort"
	"strings"

	"pacto/internal/onboarding"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	stepProblem = iota
	stepTechnologies
	stepTargets
	stepOtherTargets
)

type model struct {
	problemInput      textinput.Model
	technologiesInput textinput.Model
	otherTargetsInput textinput.Model
	targetOptions     []string
	targetCursor      int
	targetSelected    map[string]bool
	index             int
	done              bool
	cancel            bool
}

func New(p onboarding.Profile) model {
	problem := textinput.New()
	problem.Prompt = "> "
	problem.Placeholder = "Describe the core problem"
	problem.SetValue(strings.TrimSpace(p.Intents.Problem))
	problem.CharLimit = 512
	problem.Focus()

	technologies := textinput.New()
	technologies.Prompt = "> "
	technologies.Placeholder = "go,typescript,postgresql"
	technologies.SetValue(strings.Join(combinedTechnologies(p), ","))
	technologies.CharLimit = 512

	otherTargets := textinput.New()
	otherTargets.Prompt = "> "
	otherTargets.Placeholder = "Only if you selected Other (csv)"
	otherTargets.SetValue(strings.Join(p.CustomTools, ","))
	otherTargets.CharLimit = 512

	selected := map[string]bool{}
	for _, tool := range p.Tools {
		if onboarding.IsKnownTool(tool) {
			selected[tool] = true
		}
	}
	if len(p.CustomTools) > 0 {
		selected["other"] = true
	}

	return model{
		problemInput:      problem,
		technologiesInput: technologies,
		otherTargetsInput: otherTargets,
		targetOptions:     []string{"codex", "cursor", "claude", "opencode", "other"},
		targetSelected:    selected,
	}
}

func combinedTechnologies(p onboarding.Profile) []string {
	all := append([]string{}, p.Languages...)
	all = append(all, p.CustomLanguages...)
	all = normalizeCSV(all)
	return all
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "esc":
			m.cancel = true
			return m, tea.Quit
		}

		if m.index == stepTargets {
			return m.updateTargets(v)
		}

		switch v.String() {
		case "enter", "tab", "down", "ctrl+n":
			if !m.canAdvance() {
				return m, nil
			}
			return m.advance()
		case "shift+tab", "up", "ctrl+p":
			return m.back()
		}
	}

	var cmd tea.Cmd
	switch m.index {
	case stepProblem:
		m.problemInput, cmd = m.problemInput.Update(msg)
	case stepTechnologies:
		m.technologiesInput, cmd = m.technologiesInput.Update(msg)
	case stepOtherTargets:
		m.otherTargetsInput, cmd = m.otherTargetsInput.Update(msg)
	}
	return m, cmd
}

func (m model) updateTargets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k", "ctrl+p":
		if m.targetCursor > 0 {
			m.targetCursor--
		}
		return m, nil
	case "down", "j", "ctrl+n":
		if m.targetCursor < len(m.targetOptions)-1 {
			m.targetCursor++
		}
		return m, nil
	case " ":
		key := m.targetOptions[m.targetCursor]
		m.targetSelected[key] = !m.targetSelected[key]
		if key == "other" && !m.targetSelected[key] {
			m.otherTargetsInput.SetValue("")
		}
		return m, nil
	case "shift+tab":
		return m.back()
	case "enter", "tab":
		if m.targetSelected["other"] {
			m.index = stepOtherTargets
			m.otherTargetsInput.Focus()
			return m, nil
		}
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) canAdvance() bool {
	switch m.index {
	case stepProblem:
		return strings.TrimSpace(m.problemInput.Value()) != ""
	case stepTechnologies:
		return strings.TrimSpace(m.technologiesInput.Value()) != ""
	default:
		return true
	}
}

func (m model) advance() (tea.Model, tea.Cmd) {
	switch m.index {
	case stepProblem:
		m.problemInput.Blur()
		m.technologiesInput.Focus()
		m.index = stepTechnologies
	case stepTechnologies:
		m.technologiesInput.Blur()
		m.index = stepTargets
	case stepOtherTargets:
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m model) back() (tea.Model, tea.Cmd) {
	switch m.index {
	case stepTechnologies:
		m.technologiesInput.Blur()
		m.problemInput.Focus()
		m.index = stepProblem
	case stepTargets:
		m.technologiesInput.Focus()
		m.index = stepTechnologies
	case stepOtherTargets:
		m.otherTargetsInput.Blur()
		m.index = stepTargets
	}
	return m, nil
}

func (m model) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render("Pacto Init Onboarding")
	faint := lipgloss.NewStyle().Faint(true)
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.index {
	case stepProblem:
		b.WriteString(accent.Render("Step 1/3 - Problem"))
		b.WriteString("\nDescribe what you are building and why.\n")
		b.WriteString(m.problemInput.View())
	case stepTechnologies:
		b.WriteString(accent.Render("Step 2/3 - Technologies"))
		b.WriteString("\nComma-separated technologies. Known values are suggested below.\n")
		b.WriteString(faint.Render("Known: " + strings.Join(onboarding.KnownLanguages, ", ")))
		b.WriteString("\n")
		b.WriteString(m.technologiesInput.View())
	case stepTargets, stepOtherTargets:
		b.WriteString(accent.Render("Step 3/3 - Install Targets"))
		b.WriteString("\nUse Space to toggle one or many targets.\n\n")
		for i, option := range m.targetOptions {
			cursor := " "
			if i == m.targetCursor && m.index == stepTargets {
				cursor = ">"
			}
			checked := "[ ]"
			if m.targetSelected[option] {
				checked = "[x]"
			}
			label := option
			if option == "other" {
				label = "other (custom)"
			}
			b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checked, label))
		}
		if m.targetSelected["other"] {
			b.WriteString("\n")
			b.WriteString(faint.Render("Other targets (csv)"))
			b.WriteString("\n")
			b.WriteString(m.otherTargetsInput.View())
		}
	}

	b.WriteString("\n\n")
	b.WriteString(faint.Render("Enter/Tab: next - Shift+Tab: back - Esc: cancel"))
	return b.String()
}

func Run(initial onboarding.Profile) (onboarding.Profile, bool, error) {
	p := tea.NewProgram(New(initial), tea.WithAltScreen())
	res, err := p.Run()
	if err != nil {
		return onboarding.Profile{}, false, err
	}
	m := res.(model)
	if m.cancel || !m.done {
		return onboarding.Profile{}, false, nil
	}

	out := onboarding.Profile{}
	out.Intents.Problem = strings.TrimSpace(m.problemInput.Value())

	for _, token := range splitCSV(m.technologiesInput.Value()) {
		if onboarding.IsKnownLanguage(token) {
			out.Languages = append(out.Languages, token)
		} else {
			out.CustomLanguages = append(out.CustomLanguages, token)
		}
	}
	out.Languages = normalizeCSV(out.Languages)
	out.CustomLanguages = normalizeCSV(out.CustomLanguages)

	for _, option := range m.targetOptions {
		if option == "other" || !m.targetSelected[option] {
			continue
		}
		out.Tools = append(out.Tools, option)
	}
	out.Tools = normalizeCSV(out.Tools)
	if m.targetSelected["other"] {
		out.CustomTools = normalizeCSV(splitCSV(m.otherTargetsInput.Value()))
	}

	out.Sources.Languages = "user"
	out.Sources.Tools = "user"
	return out, true, nil
}

func splitCSV(raw string) []string {
	tokens := strings.Split(raw, ",")
	out := make([]string, 0, len(tokens))
	for _, tok := range tokens {
		t := strings.TrimSpace(tok)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func normalizeCSV(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, item := range in {
		t := onboarding.NormalizeToken(item)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}
