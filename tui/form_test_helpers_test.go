package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/x/ansi"
)

func initFormView(form *huh.Form) string {
	return stripFormView(batchUpdateForm(form, form.Init()).(*huh.Form))
}

func stripFormView(form *huh.Form) string {
	return ansi.Strip(form.View())
}

func batchUpdateForm(m tea.Model, cmd tea.Cmd) tea.Model {
	if cmd == nil {
		return m
	}
	msg := cmd()
	m, cmd = m.Update(msg)
	if cmd == nil {
		return m
	}
	msg = cmd()
	m, _ = m.Update(msg)
	return m
}
