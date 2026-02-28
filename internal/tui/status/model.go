package status

import (
	"fmt"
	"strings"

	"pacto/internal/model"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	report      model.StatusReport
	cursor      int
	width       int
	height      int
	searchInput textinput.Model
	searching   bool
	stateFilter string
}

func New(r model.StatusReport) Model {
	in := textinput.New()
	in.Placeholder = "search slug..."
	in.CharLimit = 120
	in.Prompt = "/ "
	in.Blur()
	return Model{
		report:      r,
		cursor:      0,
		searchInput: in,
		stateFilter: "all",
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.searching {
		switch k := msg.(type) {
		case tea.KeyMsg:
			switch k.String() {
			case "esc":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			case "enter":
				m.searching = false
				m.searchInput.Blur()
				m.cursor = 0
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	switch k := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = k.Width
		m.height = k.Height
		return m, nil
	case tea.KeyMsg:
		switch k.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			max := len(m.filtered()) - 1
			if m.cursor < max {
				m.cursor++
			}
			return m, nil
		case "g":
			m.cursor = 0
			return m, nil
		case "G":
			m.cursor = len(m.filtered()) - 1
			if m.cursor < 0 {
				m.cursor = 0
			}
			return m, nil
		case "/":
			m.searching = true
			m.searchInput.Focus()
			return m, nil
		case "f":
			m.stateFilter = nextFilter(m.stateFilter)
			m.cursor = 0
			return m, nil
		}
	}

	return m, nil
}

func nextFilter(state string) string {
	order := []string{"all", "current", "to-implement", "done", "outdated"}
	for i := range order {
		if order[i] == state {
			return order[(i+1)%len(order)]
		}
	}
	return "all"
}

func (m Model) filtered() []model.PlanStatus {
	query := strings.TrimSpace(strings.ToLower(m.searchInput.Value()))
	out := make([]model.PlanStatus, 0, len(m.report.Plans))
	for _, p := range m.report.Plans {
		if m.stateFilter != "all" && p.StateFolder != m.stateFilter {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Slug), query) {
			continue
		}
		out = append(out, p)
	}
	return out
}

func (m Model) selected() *model.PlanStatus {
	plans := m.filtered()
	if len(plans) == 0 {
		return nil
	}
	if m.cursor >= len(plans) {
		m.cursor = len(plans) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	p := plans[m.cursor]
	return &p
}

func (m Model) View() string {
	plans := m.filtered()
	head := fmt.Sprintf("Pacto Status  plans=%d pending=%d blocked=%d  filter=%s", len(m.report.Plans), m.report.Summary.TotalPendingTasks, m.report.Summary.TotalBlockedTasks, m.stateFilter)
	if strings.TrimSpace(m.searchInput.Value()) != "" {
		head += "  search=" + m.searchInput.Value()
	}

	leftW := 42
	if m.width > 0 && m.width < 90 {
		leftW = 30
	}
	rightW := m.width - leftW - 3
	if rightW < 20 {
		rightW = 20
	}

	left := make([]string, 0, len(plans)+2)
	left = append(left, "Plans")
	for i, p := range plans {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		line := fmt.Sprintf("%s%-12s %-24s %s/%s", cursor, p.StateFolder, truncate(p.Slug, 24), badgeForVerification(p.Verification), p.Confidence)
		left = append(left, line)
	}
	if len(plans) == 0 {
		left = append(left, "(no plans for current filter/search)")
	}

	right := []string{"Details"}
	if sel := m.selected(); sel != nil {
		right = append(right, fmt.Sprintf("slug: %s", sel.Slug))
		right = append(right, fmt.Sprintf("state: %s", sel.StateFolder))
		right = append(right, fmt.Sprintf("verification: %s", sel.Verification))
		right = append(right, fmt.Sprintf("pending: %d  blocked: %d", sel.PendingTasks, sel.BlockedTasks))
		if len(sel.NextActions) > 0 {
			right = append(right, "")
			right = append(right, "next actions:")
			for _, n := range sel.NextActions {
				right = append(right, "  - "+n)
			}
		}
		if len(sel.Blockers) > 0 {
			right = append(right, "")
			right = append(right, "blockers:")
			for _, b := range sel.Blockers {
				right = append(right, "  - "+b)
			}
		}
		if sel.ParseError != "" {
			right = append(right, "")
			right = append(right, "parse error: "+sel.ParseError)
		}
	} else {
		right = append(right, "select a plan to inspect details")
	}

	leftStyle := lipgloss.NewStyle().Width(leftW).BorderRight(true).BorderStyle(lipgloss.NormalBorder()).PaddingRight(1)
	rightStyle := lipgloss.NewStyle().Width(rightW).PaddingLeft(1)
	body := lipgloss.JoinHorizontal(lipgloss.Top, leftStyle.Render(strings.Join(left, "\n")), rightStyle.Render(strings.Join(right, "\n")))
	foot := "keys: j/k or arrows navigate • / search • f filter • q quit"
	if m.searching {
		foot = m.searchInput.View() + "  (enter apply, esc cancel)"
	}
	return strings.Join([]string{
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")).Render(head),
		"",
		body,
		"",
		lipgloss.NewStyle().Faint(true).Render(foot),
	}, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

func badgeForVerification(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "verified":
		return "ok"
	case "partial":
		return "partial"
	case "unverified":
		return "no"
	default:
		return v
	}
}
