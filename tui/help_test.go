package tui

import (
	"strings"
	"testing"

	"dailytrack/models"

	tea "github.com/charmbracelet/bubbletea"
)

func TestDashboardQuestionMarkOpensKeybindPopup(t *testing.T) {
	m := Model{
		state:  stateDashboard,
		config: &models.Config{SetupComplete: true},
	}

	nextModel, _ := m.updateDashboard(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}, Alt: false})
	next := modelState(nextModel)

	if next.state != stateHelp {
		t.Fatalf("state = %v, want %v", next.state, stateHelp)
	}
	view := next.helpOverlay()
	if !strings.Contains(view, "Keyboard Reference") {
		t.Fatalf("help overlay should include Keyboard Reference, got:\n%s", view)
	}
}

func TestShortHelpAdvertisesKeybindPopup(t *testing.T) {
	found := false
	for _, binding := range keys.ShortHelp() {
		if binding.Help().Key == "?" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("short help should advertise ? keybind popup")
	}
}
