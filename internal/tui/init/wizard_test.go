package initui

import (
	"strings"
	"testing"

	"pacto/internal/i18n"
	"pacto/internal/onboarding"

	tea "github.com/charmbracelet/bubbletea"
)

// key returns a synthetic tea.KeyMsg for the given key string. We use this
// instead of constructing tea.KeyMsg literals directly because the field
// shape varies across bubbletea releases; routing by string is stable.
func key(k string) tea.KeyMsg {
	switch k {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "shift+tab":
		return tea.KeyMsg{Type: tea.KeyShiftTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

func sendKeys(t *testing.T, m tea.Model, keys ...tea.KeyMsg) tea.Model {
	t.Helper()
	for _, k := range keys {
		var cmd tea.Cmd
		m, cmd = m.Update(k)
		_ = cmd // discard quit/blink commands; we inspect state directly
	}
	return m
}

func TestNewSeedsLanguageFromProfile(t *testing.T) {
	m := New(onboarding.Profile{UILanguage: "es"})
	if m.selectedLang != i18n.Spanish {
		t.Errorf("expected Spanish, got %v", m.selectedLang)
	}
	if m.langCursor != 1 {
		t.Errorf("expected cursor at 1 for Spanish, got %d", m.langCursor)
	}
}

func TestNewSeedsToolsFromProfile(t *testing.T) {
	m := New(onboarding.Profile{
		Tools:       []string{"codex", "cursor"},
		CustomTools: []string{"my-ide"},
	})
	if !m.targetSelected["codex"] || !m.targetSelected["cursor"] {
		t.Errorf("known tools not selected: %+v", m.targetSelected)
	}
	if !m.targetSelected["other"] {
		t.Error("custom tools should mark 'other' as selected")
	}
}

func TestEscCancelsImmediately(t *testing.T) {
	m := sendKeys(t, New(onboarding.Profile{}), key("esc"))
	wm := m.(model)
	if !wm.cancel {
		t.Error("expected cancel flag set")
	}
	if wm.done {
		t.Error("expected done flag false on cancel")
	}
}

func TestProblemStepBlocksAdvanceWhenEmpty(t *testing.T) {
	// Walk: language step -> enter -> problem step -> enter (with empty input)
	m := sendKeys(t, New(onboarding.Profile{}), key("enter"), key("enter"))
	wm := m.(model)
	if wm.index != stepProblem {
		t.Errorf("expected to remain on problem step, got %d", wm.index)
	}
}

func TestCanAdvanceRespectsRequiredFields(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepProblem
	if m.canAdvance() {
		t.Error("expected false when problem empty")
	}
	m.problemInput.SetValue("ship it")
	if !m.canAdvance() {
		t.Error("expected true when problem set")
	}
	m.index = stepTechnologies
	if m.canAdvance() {
		t.Error("expected false when technologies empty")
	}
	m.technologiesInput.SetValue("go")
	if !m.canAdvance() {
		t.Error("expected true when technologies set")
	}
	m.index = stepLanguage
	if !m.canAdvance() {
		t.Error("language step always advances")
	}
}

func TestBackNavigatesThroughSteps(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepTargets
	res, _ := m.back()
	wm := res.(model)
	if wm.index != stepTechnologies {
		t.Errorf("expected stepTechnologies, got %d", wm.index)
	}
	res, _ = wm.back()
	wm = res.(model)
	if wm.index != stepProblem {
		t.Errorf("expected stepProblem, got %d", wm.index)
	}
	res, _ = wm.back()
	wm = res.(model)
	if wm.index != stepLanguage {
		t.Errorf("expected stepLanguage, got %d", wm.index)
	}
}

func TestUpdateTargetsTogglesSelection(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepTargets
	m.targetCursor = 0 // codex
	res, _ := m.updateTargets(key(" "))
	wm := res.(model)
	if !wm.targetSelected["codex"] {
		t.Error("expected codex selected after space")
	}
	res, _ = wm.updateTargets(key(" "))
	if res.(model).targetSelected["codex"] {
		t.Error("expected codex unselected after second space")
	}
}

func TestUpdateTargetsCursorBounds(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepTargets
	res, _ := m.updateTargets(key("up"))
	if res.(model).targetCursor != 0 {
		t.Error("cursor should clamp at 0")
	}
	wm := res.(model)
	for range wm.targetOptions {
		res, _ = wm.updateTargets(key("down"))
		wm = res.(model)
	}
	if wm.targetCursor != len(wm.targetOptions)-1 {
		t.Errorf("cursor should clamp at last option, got %d", wm.targetCursor)
	}
}

func TestUpdateTargetsEnterFinishesWhenNoOther(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepTargets
	m.targetSelected["codex"] = true
	res, _ := m.updateTargets(key("enter"))
	wm := res.(model)
	if !wm.done {
		t.Error("expected done after enter on targets without 'other'")
	}
}

func TestUpdateTargetsEnterAdvancesToOtherTargets(t *testing.T) {
	m := New(onboarding.Profile{})
	m.index = stepTargets
	m.targetSelected["other"] = true
	res, _ := m.updateTargets(key("enter"))
	wm := res.(model)
	if wm.index != stepOtherTargets {
		t.Errorf("expected stepOtherTargets, got %d", wm.index)
	}
	if wm.done {
		t.Error("should not be done yet")
	}
}

func TestSplitInputTokensFiltersStopwords(t *testing.T) {
	got := splitInputTokens("go and python with the typescript")
	want := []string{"go", "python", "typescript"}
	if !equalStrSlices(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
}

func TestSplitInputTokensSplitsOnPunctuation(t *testing.T) {
	got := splitInputTokens("go, python; rust|c++")
	if len(got) != 4 {
		t.Errorf("expected 4 tokens, got %v", got)
	}
}

func TestNormalizeCSVDedupesAndSorts(t *testing.T) {
	got := normalizeCSV([]string{"go", "Go", "python", "go"})
	if len(got) != 2 || got[0] != "go" || got[1] != "python" {
		t.Errorf("got %v", got)
	}
}

func TestViewRendersForEachStep(t *testing.T) {
	// Smoke test: View must not panic at any step. Required strings are
	// step labels (stable across i18n layouts).
	m := New(onboarding.Profile{})
	for _, step := range []int{stepLanguage, stepProblem, stepTechnologies, stepTargets, stepOtherTargets} {
		m.index = step
		out := m.View()
		if strings.TrimSpace(out) == "" {
			t.Errorf("step %d rendered empty view", step)
		}
	}
}

func equalStrSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
