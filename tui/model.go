package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"dailytrack/db"
	"dailytrack/models"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type appState int

const (
	stateSetup     appState = iota // first-launch wizard
	stateDashboard                 // main tabs
	stateForm                      // daily entry
)

type Model struct {
	state   appState
	width   int
	height  int
	help    help.Model
	vp      viewport.Model

	// Setup
	setup *setupWiz

	// Dashboard
	activeTab int
	config    *models.Config
	entries   []models.Entry

	// Entry form
	form     *huh.Form
	boolPtrs map[string]*bool
	strPtrs  map[string]*string
	intPtrs  map[string]*int
	textIDs  map[string]bool
}

type keyMap struct {
	Add   key.Binding
	Quit  key.Binding
	Left  key.Binding
	Right key.Binding
	Up    key.Binding
	Down  key.Binding
}

var keys = keyMap{
	Add:   key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add entry")),
	Left:  key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev tab")),
	Right: key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next tab")),
	Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.Add, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Left, k.Right, k.Up, k.Down, k.Add, k.Quit}}
}

func InitialModel(cfg *models.Config) Model {
	m := Model{
		help: help.New(),
		vp:   viewport.New(0, 0),
	}
	if cfg == nil || !cfg.SetupComplete {
		m.state = stateSetup
		m.setup = newSetupWiz()
	} else {
		m.state = stateDashboard
		m.config = cfg
		m.entries, _ = db.GetAllEntries()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	if m.state == stateSetup {
		return m.setup.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.vp.Width = msg.Width
		m.vp.Height = msg.Height - 13 // 13 is approx header + footer height
		m.syncViewport()
	case setupDoneMsg:
		m.config = msg.cfg
		m.state = stateDashboard
		m.entries, _ = db.GetAllEntries()
		m.syncViewport()
		return m, nil
	}

	switch m.state {
	case stateSetup:
		cmd := m.setup.Update(msg)
		return m, cmd
	case stateDashboard:
		return m.updateDashboard(msg)
	case stateForm:
		return m.updateForm(msg)
	}
	return m, nil
}

func (m *Model) syncViewport() {
	if m.width == 0 || m.config == nil {
		return
	}
	var content string
	if m.activeTab == 0 {
		content = m.viewOverview()
	} else if m.activeTab == len(m.config.Categories)+1 {
		content = m.viewInsights()
	} else {
		cat := m.config.Categories[m.activeTab-1]
		content = m.viewCategory(cat)
	}
	
	m.vp.SetContent(lipgloss.NewStyle().Align(lipgloss.Center).Width(m.width).Render(content))
}

func (m Model) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Add):
			m.state = stateForm
			m.initForm()
			if m.form == nil {
				m.state = stateDashboard
				return m, nil
			}
			return m, m.form.Init()
		case key.Matches(msg, keys.Left):
			total := len(m.config.Categories) + 2
			m.activeTab = (m.activeTab - 1 + total) % total
			m.syncViewport()
			m.vp.GotoTop()
		case key.Matches(msg, keys.Right):
			total := len(m.config.Categories) + 2
			m.activeTab = (m.activeTab + 1) % total
			m.syncViewport()
			m.vp.GotoTop()
		}
	}
	
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m *Model) initForm() {
	m.boolPtrs = make(map[string]*bool)
	m.strPtrs = make(map[string]*string)
	m.intPtrs = make(map[string]*int)
	m.textIDs = make(map[string]bool)

	var groups []*huh.Group

	for _, cat := range m.config.Categories {
		var fields []huh.Field
		for _, t := range cat.Trackers {
			switch t.Type {
			case models.TrackerBinary:
				b := false
				m.boolPtrs[t.ID] = &b
				fields = append(fields, huh.NewConfirm().Title(t.Name).Value(&b))

			case models.TrackerDuration:
				s := ""
				m.strPtrs[t.ID] = &s
				label := t.Name + " (minutes)"
				if t.Target != nil {
					label = fmt.Sprintf("%s (minutes, target: %.0f)", t.Name, *t.Target)
				}
				fields = append(fields, huh.NewInput().Title(label).Value(&s).
					Validate(func(s string) error {
						if s == "" {
							return nil
						}
						v, err := strconv.ParseFloat(s, 64)
						if err != nil || v <= 0 {
							return fmt.Errorf("enter a positive number")
						}
						return nil
					}))

			case models.TrackerCount:
				s := ""
				m.strPtrs[t.ID] = &s
				fields = append(fields, huh.NewInput().Title(t.Name).Value(&s).
					Validate(func(s string) error {
						if s == "" {
							return nil
						}
						v, err := strconv.ParseFloat(s, 64)
						if err != nil || v < 0 {
							return fmt.Errorf("enter a number")
						}
						return nil
					}))

			case models.TrackerNumeric:
				s := ""
				m.strPtrs[t.ID] = &s
				fields = append(fields, huh.NewInput().Title(t.Name).Value(&s).
					Validate(func(s string) error {
						if s == "" {
							return nil
						}
						_, err := strconv.ParseFloat(s, 64)
						if err != nil {
							return fmt.Errorf("enter a number")
						}
						return nil
					}))

			case models.TrackerRating:
				v := 3
				m.intPtrs[t.ID] = &v
				fields = append(fields, huh.NewSelect[int]().
					Title(t.Name).
					Options(
						huh.NewOption("1 — rough day", 1),
						huh.NewOption("2 — below average", 2),
						huh.NewOption("3 — okay", 3),
						huh.NewOption("4 — good day", 4),
						huh.NewOption("5 — great day", 5),
					).
					Value(&v))

			case models.TrackerText:
				s := ""
				m.strPtrs[t.ID] = &s
				m.textIDs[t.ID] = true
				fields = append(fields, huh.NewText().Title(t.Name).Value(&s))
			}
		}
		if len(fields) > 0 {
			groups = append(groups, huh.NewGroup(fields...).Title(cat.Name))
		}
	}

	if len(groups) == 0 {
		m.form = nil
		return
	}
	m.form = huh.NewForm(groups...)
}

func (m Model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		m.saveEntry()
		m.state = stateDashboard
		m.entries, _ = db.GetAllEntries()
		m.syncViewport()
		return m, nil
	}
	if m.form.State == huh.StateAborted {
		m.state = stateDashboard
		return m, nil
	}
	return m, cmd
}

func (m *Model) saveEntry() {
	data := make(map[string]interface{})

	for id, ptr := range m.boolPtrs {
		data[id] = *ptr
	}
	for id, ptr := range m.strPtrs {
		if *ptr == "" {
			continue
		}
		if m.textIDs[id] {
			data[id] = *ptr
		} else if v, err := strconv.ParseFloat(*ptr, 64); err == nil {
			data[id] = v
		} else {
			data[id] = *ptr
		}
	}
	for id, ptr := range m.intPtrs {
		data[id] = float64(*ptr)
	}

	entry := &models.Entry{
		Date: time.Now().Format("2006-01-02"),
		Data: data,
	}
	if err := db.UpsertEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "dailytrack: failed to save entry: %v\n", err)
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}
	switch m.state {
	case stateSetup:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.setup.View())
	case stateForm:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.form.View())
	default:
		view := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top,
			m.dashboardView())
		// Ensure it doesn't overflow the terminal height, which causes clipping at the top
		return lipgloss.NewStyle().MaxHeight(m.height).MaxWidth(m.width).Render(view)
	}
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

const banner = `______     ______     ______   ______     ______     ______     __  __    
/\  ___\   /\  __ \   /\__  _\ /\  == \   /\  __ \   /\  ___\   /\ \/ /    
\ \ \__ \  \ \ \/\ \  \/_/\ \/ \ \  __<   \ \  __ \  \ \ \____  \ \  _"-.  
 \ \_____\  \ \_____\    \ \_\  \ \_\ \_\  \ \_\ \_\  \ \_____\  \ \_\ \_\ 
  \/_____/   \/_____/     \/_/   \/_/ /_/   \/_/\/_/   \/_____/   \/_/\/_/`

func (m Model) dashboardView() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00ADD8")).Bold(true).
		Align(lipgloss.Center).Width(m.width).MarginBottom(1).MarginTop(1)

	tabNames := []string{"Overview"}
	for _, c := range m.config.Categories {
		tabNames = append(tabNames, c.Name)
	}
	tabNames = append(tabNames, "Insights")

	var tabParts []string
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == m.activeTab {
			style = style.Background(lipgloss.Color("#00ADD8")).
				Foreground(lipgloss.Color("#000")).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color("#888"))
		}
		tabParts = append(tabParts, style.Render(name))
	}
	tabBar := lipgloss.NewStyle().MarginBottom(1).Align(lipgloss.Center).Width(m.width).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabParts...))

	helpView := lipgloss.NewStyle().MarginTop(1).Align(lipgloss.Center).Width(m.width).
		Render(m.help.View(keys))

	return lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(banner),
		tabBar,
		m.vp.View(),
		helpView,
	)
}

func box(title, content string, width, height int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444")).
		Padding(1, 2).
		Width(width)

	if height > 0 {
		style = style.Height(height)
	}

	return style.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")).Bold(true).Render(title),
		"",
		content,
	))
}

// ─── Overview Tab ─────────────────────────────────────────────────────────────

func (m Model) viewOverview() string {
	if len(m.entries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).
			Render("No entries yet. Press 'a' to log today.")
	}

	var cards []string
	const boxWidth = 40
	const boxHeight = 8

	for _, cat := range m.config.Categories {
		content := m.categorySummary(cat)
		color := cat.Color
		if color == "" {
			color = "#00ADD8"
		}
		titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		rendered := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444")).
			Padding(1, 2).
			Width(boxWidth).
			Height(boxHeight).
			Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render(cat.Name), "", content))
		cards = append(cards, rendered)
	}

	if len(cards) == 0 {
		return "No categories configured."
	}

	var rows []string
	for i := 0; i < len(cards); i += 2 {
		end := i + 2
		if end > len(cards) {
			end = len(cards)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) categorySummary(cat models.Category) string {
	var lines []string
	for _, t := range cat.Trackers {
		switch t.Type {
		case models.TrackerBinary:
			streak := db.CurrentStreak(m.entries, t.ID)
			pct := db.ConsistencyPct(m.entries, t.ID)
			lines = append(lines, fmt.Sprintf("%-22s %d streak  %.0f%%", truncate(t.Name, 22), streak, pct))

		case models.TrackerDuration:
			series := db.NumericSeries(m.entries, t.ID)
			avg := average(series)
			lines = append(lines, fmt.Sprintf("%-22s avg %.0fmin", truncate(t.Name, 22), avg))

		case models.TrackerNumeric, models.TrackerCount:
			series := db.NumericSeries(m.entries, t.ID)
			if len(series) > 0 {
				latest := series[len(series)-1]
				lines = append(lines, fmt.Sprintf("%-22s %.1f", truncate(t.Name, 22), latest))
			}

		case models.TrackerRating:
			series := db.NumericSeries(m.entries, t.ID)
			avg := average(series)
			if len(series) > 0 {
				lines = append(lines, fmt.Sprintf("%-22s avg %.1f/5", truncate(t.Name, 22), avg))
			}

		case models.TrackerText:
			for _, e := range m.entries {
				if v, ok := e.Data[t.ID].(string); ok && v != "" {
					preview := v
					if len(preview) > 28 {
						preview = preview[:25] + "..."
					}
					lines = append(lines, fmt.Sprintf("%-22s %s", truncate(t.Name, 22), preview))
					break
				}
			}
		}
	}
	if len(lines) == 0 {
		return "(no trackers)"
	}
	return strings.Join(lines, "\n")
}

// ─── Category Tab ─────────────────────────────────────────────────────────────

func (m Model) viewCategory(cat models.Category) string {
	if len(m.entries) == 0 {
		return "No entries yet."
	}

	const limit = 30
	var boxes []string

	for _, t := range cat.Trackers {
		switch t.Type {
		case models.TrackerBinary:
			heat := db.BinaryHeatmap(m.entries, t.ID)
			if len(heat) > limit {
				heat = heat[len(heat)-limit:]
			}
			pct := db.ConsistencyPct(m.entries, t.ID)
			streak := db.CurrentStreak(m.entries, t.ID)
			offset := 0
			if len(heat) > 0 {
				idx := len(m.entries) - 1
				if len(m.entries) > limit {
					idx = limit - 1
				}
				if parsed, err := time.Parse("2006-01-02", m.entries[idx].Date); err == nil {
					offset = int(parsed.Weekday())
				}
			}
			content := fmt.Sprintf("Streak: %d days  |  All-time: %.0f%%\n\n", streak, pct) +
				Heatmap(heat, offset)
			boxes = append(boxes, box(t.Name, content, 42, 14))

		case models.TrackerDuration, models.TrackerNumeric, models.TrackerCount:
			series := db.NumericSeries(m.entries, t.ID)
			if len(series) > limit {
				series = series[len(series)-limit:]
			}
			content := renderLineChart(series, t)
			boxes = append(boxes, box(t.Name, content, 42, 14))

		case models.TrackerRating:
			series := db.NumericSeries(m.entries, t.ID)
			if len(series) > 14 {
				series = series[len(series)-14:]
			}
			content := "Recent ratings:\n" + Sparkline(series, "#8B5CF6")
			if len(series) > 0 {
				content += fmt.Sprintf("\n\nAvg: %.1f / 5", average(series))
			}
			boxes = append(boxes, box(t.Name, content, 42, 14))

		case models.TrackerText:
			var logs []string
			count := 0
			for _, e := range m.entries {
				if v, ok := e.Data[t.ID].(string); ok && v != "" {
					logs = append(logs, fmt.Sprintf("[%s] %s", e.Date, v))
					count++
					if count >= 3 {
						break
					}
				}
			}
			content := strings.Join(logs, "\n\n")
			if content == "" {
				content = "(no entries yet)"
			}
			boxes = append(boxes, box(t.Name, content, 42, 14))
		}
	}

	if len(boxes) == 0 {
		return "No trackers in this category."
	}

	var rows []string
	for i := 0; i < len(boxes); i += 2 {
		end := i + 2
		if end > len(boxes) {
			end = len(boxes)
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, boxes[i:end]...))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func average(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// ─── Insights Tab ─────────────────────────────────────────────────────────────

func (m Model) viewInsights() string {
	if len(m.entries) == 0 {
		return "No data recorded yet."
	}

	var ratingTracker *models.Tracker
	var durationTracker *models.Tracker
	var binaryTracker *models.Tracker

	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			tracker := t
			if ratingTracker == nil && t.Type == models.TrackerRating {
				ratingTracker = &tracker
			}
			if durationTracker == nil && t.Type == models.TrackerDuration {
				durationTracker = &tracker
			}
			if binaryTracker == nil && t.Type == models.TrackerBinary {
				binaryTracker = &tracker
			}
		}
	}

	var scatterStr = "No numeric/duration tracker found for scatter plot."
	var compStr = "No binary tracker found for A/B impact."

	if ratingTracker != nil {
		if durationTracker != nil {
			var scatterX []float64
			var scatterY []float64
			for _, e := range m.entries {
				yVal, okY := e.Data[ratingTracker.ID].(float64)
				xVal, okX := e.Data[durationTracker.ID].(float64)
				if okX && okY {
					scatterX = append(scatterX, xVal)
					scatterY = append(scatterY, yVal)
				}
			}
			if len(scatterX) > 1 {
				scatterStr = ScatterPlot(scatterX, scatterY, durationTracker.Name, ratingTracker.Name)
			} else {
				scatterStr = "Not enough data for scatter plot."
			}
		}

		if binaryTracker != nil {
			var yesRatings []int
			var noRatings []int
			for _, e := range m.entries {
				yVal, okY := e.Data[ratingTracker.ID].(float64)
				xVal, okX := e.Data[binaryTracker.ID].(bool)
				if okX && okY {
					if xVal {
						yesRatings = append(yesRatings, int(yVal))
					} else {
						noRatings = append(noRatings, int(yVal))
					}
				}
			}
			avgCodeYes := averageInts(yesRatings)
			avgCodeNo := averageInts(noRatings)
			compStr = ComparisonBar("Done", avgCodeYes, "Skipped", avgCodeNo)
			compStr += fmt.Sprintf("\n\n(%s vs %s)", binaryTracker.Name, ratingTracker.Name)
		}
	} else {
		return "Please add a Rating Tracker to view insights."
	}

	return lipgloss.JoinHorizontal(lipgloss.Top,
		box("A/B Impact", compStr, 45, 14),
		box("Scatter Analysis", scatterStr, 45, 14),
	)
}

func averageInts(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return float64(sum) / float64(len(nums))
}
