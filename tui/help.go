package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) helpOverlay() string {
	p := palette()
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.ActiveTabBg)).
		Bold(true).
		Render("GoTrack — Keyboard Reference")

	sections := []struct {
		header string
		rows   [][2]string
	}{
		{"Navigation", [][2]string{
			{"← / h", "previous tab"},
			{"→ / l", "next tab"},
			{"↑ / k", "scroll up"},
			{"↓ / j", "scroll down"},
			{"[  ]", "cycle overview hero visuals"},
			{"w", "toggle week/month in Review"},
		}},
		{"Entries", [][2]string{
			{"a", "add or edit an entry (by date)"},
			{"x", "quick entry for one tracker today"},
			{"p", "pomodoro timer for duration tracking"},
			{"e", "pick a recent entry to edit"},
			{"d", "delete an entry (with confirm)"},
			{"u", "undo last save or delete"},
			{"D", "dismiss today's reminder banner"},
		}},
		{"Config & app", [][2]string{
			{"s", "open settings"},
			{"?", "toggle this help"},
			{"q / ctrl+c", "quit"},
		}},
		{"Command line", [][2]string{
			{"gotrack log k=v …", "quick-log entry"},
			{"gotrack export --format json", "backup to stdout"},
			{"gotrack export --format csv", "CSV export"},
			{"gotrack import file.json", "restore from backup"},
		}},
	}

	var blocks []string
	for _, sec := range sections {
		header := lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Primary)).
			Bold(true).
			Render(sec.header)
		var lines []string
		lines = append(lines, header)
		for _, row := range sec.rows {
			key := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Render(fmt.Sprintf("%-26s", row[0]))
			desc := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(row[1])
			lines = append(lines, "  "+key+"  "+desc)
		}
		blocks = append(blocks, strings.Join(lines, "\n"))
	}

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Muted)).
		Italic(true).
		Render("Press ? or esc to close")

	body := strings.Join(blocks, "\n\n")
	card := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(p.ActiveTabBg)).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body, "", footer))
	return card
}
