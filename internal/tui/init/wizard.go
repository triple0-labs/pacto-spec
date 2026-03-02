package initui

import (
	"fmt"
	"sort"
	"strings"

	"pacto/internal/i18n"
	"pacto/internal/onboarding"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	stepLanguage = iota
	stepProblem
	stepTechnologies
	stepTargets
	stepOtherTargets
)

type model struct {
	problemInput      textinput.Model
	technologiesInput textinput.Model
	otherTargetsInput textinput.Model
	langOptions       []i18n.Language
	langCursor        int
	selectedLang      i18n.Language
	targetOptions     []string
	targetCursor      int
	targetSelected    map[string]bool
	index             int
	done              bool
	cancel            bool
}

func New(p onboarding.Profile) model {
	selectedLang := i18n.NormalizeLanguage(p.UILanguage)
	langCursor := 0
	if selectedLang == i18n.Spanish {
		langCursor = 1
	}

	problem := textinput.New()
	problem.Prompt = "> "
	problem.Placeholder = "Describe the core problem"
	problem.SetValue(strings.TrimSpace(p.Intents.Problem))
	problem.CharLimit = 512
	problem.Focus()

	technologies := textinput.New()
	technologies.Prompt = "> "
	technologies.Placeholder = "Go, TypeScript, PostgreSQL"
	technologies.SetValue(strings.Join(combinedTechnologies(p), " "))
	technologies.CharLimit = 512

	otherTargets := textinput.New()
	otherTargets.Prompt = "> "
	otherTargets.Placeholder = "Only if you selected Other"
	otherTargets.SetValue(strings.Join(p.CustomTools, " "))
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
		langOptions:       []i18n.Language{i18n.English, i18n.Spanish},
		langCursor:        langCursor,
		selectedLang:      selectedLang,
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

		if m.index == stepLanguage {
			switch v.String() {
			case "up", "k", "ctrl+p":
				if m.langCursor > 0 {
					m.langCursor--
				}
				return m, nil
			case "down", "j", "ctrl+n":
				if m.langCursor < len(m.langOptions)-1 {
					m.langCursor++
				}
				return m, nil
			case "enter", "tab":
				m.selectedLang = m.langOptions[m.langCursor]
				m.index = stepProblem
				m.problemInput.Focus()
				return m, nil
			}
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
	case stepLanguage:
		return true
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
	case stepLanguage:
		m.selectedLang = m.langOptions[m.langCursor]
		m.problemInput.Focus()
		m.index = stepProblem
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
	case stepProblem:
		m.problemInput.Blur()
		m.index = stepLanguage
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
	m.problemInput.Placeholder = tr("Describe the core problem", "Describe el problema principal", m.selectedLang)
	m.technologiesInput.Placeholder = tr("Go, TypeScript, PostgreSQL", "Go, TypeScript, PostgreSQL", m.selectedLang)
	m.otherTargetsInput.Placeholder = tr("Only if you selected Other", "Solo si seleccionaste otro", m.selectedLang)

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render(tr("Pacto Init Onboarding", "Onboarding de Pacto Init", m.selectedLang))
	faint := lipgloss.NewStyle().Faint(true)
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")

	switch m.index {
	case stepLanguage:
		b.WriteString(accent.Render(tr("Step 0/3 - Language", "Paso 0/3 - Idioma", m.selectedLang)))
		b.WriteString("\n" + tr("Choose your preferred language for UI and generated docs.", "Elige el idioma para la interfaz y los documentos generados.", m.selectedLang) + "\n\n")
		for i, lang := range m.langOptions {
			cursor := " "
			if i == m.langCursor {
				cursor = ">"
			}
			label := "English"
			if lang == i18n.Spanish {
				label = "Español"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", cursor, label))
		}
	case stepProblem:
		b.WriteString(accent.Render(tr("Step 1/3 - Problem", "Paso 1/3 - Problema", m.selectedLang)))
		b.WriteString("\n" + tr("Tell us what you are building and why.", "Cuéntanos qué estás construyendo y por qué.", m.selectedLang) + "\n")
		b.WriteString(faint.Render(tr("Use your own words. We welcome custom descriptions.", "Usa tus propias palabras. Nos encantan las descripciones personalizadas.", m.selectedLang)))
		b.WriteString("\n")
		b.WriteString(m.problemInput.View())
	case stepTechnologies:
		b.WriteString(accent.Render(tr("Step 2/3 - Technologies", "Paso 2/3 - Tecnologías", m.selectedLang)))
		b.WriteString("\n" + tr("Tell us which technologies you plan to use.", "Cuéntanos qué tecnologías planeas usar.", m.selectedLang) + "\n")
		b.WriteString(faint.Render(tr("Custom technologies are welcome too.", "También puedes incluir tecnologías personalizadas.", m.selectedLang)))
		b.WriteString("\n")
		b.WriteString(faint.Render(tr("Known: ", "Conocidas: ", m.selectedLang) + strings.Join(onboarding.KnownLanguages, ", ")))
		b.WriteString("\n")
		b.WriteString(m.technologiesInput.View())
	case stepTargets, stepOtherTargets:
		b.WriteString(accent.Render(tr("Step 3/3 - Install Targets", "Paso 3/3 - Destinos de instalación", m.selectedLang)))
		b.WriteString("\n" + tr("Use Space to toggle one or many targets.", "Usa Espacio para seleccionar uno o varios destinos.", m.selectedLang) + "\n\n")
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
				label = tr("other (custom)", "otro (personalizado)", m.selectedLang)
			}
			b.WriteString(fmt.Sprintf("%s %s %s\n", cursor, checked, label))
		}
		if m.targetSelected["other"] {
			b.WriteString("\n")
			b.WriteString(faint.Render(tr("Other targets", "Otros destinos", m.selectedLang)))
			b.WriteString("\n")
			b.WriteString(m.otherTargetsInput.View())
		}
	}

	b.WriteString("\n\n")
	b.WriteString(faint.Render(tr("Enter/Tab: next - Shift+Tab: back - Esc: cancel", "Enter/Tab: siguiente - Shift+Tab: atrás - Esc: cancelar", m.selectedLang)))
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
	out.UILanguage = string(m.selectedLang)
	out.Intents.Problem = strings.TrimSpace(m.problemInput.Value())

	for _, token := range splitInputTokens(m.technologiesInput.Value()) {
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
		out.CustomTools = normalizeCSV(splitInputTokens(m.otherTargetsInput.Value()))
	}

	out.Sources.Languages = "user"
	out.Sources.Tools = "user"
	out.Sources.UI = "user"
	return out, true, nil
}

func splitInputTokens(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', '\n', '\t', '|', '/':
			return true
		case ' ':
			return true
		default:
			return false
		}
	})
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		t := strings.Trim(strings.TrimSpace(f), " .:+")
		if t == "" {
			continue
		}
		n := onboarding.NormalizeToken(t)
		switch n {
		case "and", "with", "using", "use", "the", "a", "an":
			continue
		}
		out = append(out, t)
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

func tr(en, es string, lang i18n.Language) string {
	return i18n.T(lang, en, es)
}
