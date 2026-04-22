package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func TestUpdateForm_EscReturnsToDashboard(t *testing.T) {
	m := Model{
		state: stateForm,
		form: huh.NewForm(
			huh.NewGroup(
				huh.NewInput().Title("Field"),
			),
		),
	}

	nextModel, _ := m.updateForm(tea.KeyMsg{Type: tea.KeyEsc})
	next := modelState(nextModel)
	if next.state != stateDashboard {
		t.Fatalf("state after esc = %v, want %v", next.state, stateDashboard)
	}
}

func TestUpdateDateForm_EscReturnsToDashboard(t *testing.T) {
	m := Model{
		state: stateEntryDate,
	}
	m.initDateForm("Entry date", "desc")

	nextModel, _ := m.updateDateForm(tea.KeyMsg{Type: tea.KeyEsc})
	next := modelState(nextModel)
	if next.state != stateDashboard {
		t.Fatalf("state after esc = %v, want %v", next.state, stateDashboard)
	}
}

func modelState(m tea.Model) Model {
	switch v := m.(type) {
	case Model:
		return v
	case *Model:
		return *v
	default:
		panic("unexpected model type")
	}
}
