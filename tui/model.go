package tui

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"dailytrack/db"
	"dailytrack/integrations"
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
	stateEntryDate                 // pick date before entry form
	stateForm                      // daily entry
	stateSettings                  // settings/config editor
)

type Model struct {
	state  appState
	width  int
	height int
	help   help.Model
	vp     viewport.Model

	// Setup
	setup *setupWiz

	// Dashboard
	activeTab int
	config    *models.Config
	entries   []models.Entry
	settings  *settingsWiz
	stars     []fallingStar
	twinkles  []twinkleStar
	bursts    []burstParticle
	pulseTick int
	rng       *rand.Rand

	// Entry date picker
	dateForm  *huh.Form
	entryDate string

	// Entry form
	form     *huh.Form
	boolPtrs map[string]*bool
	strPtrs  map[string]*string
	intPtrs  map[string]*int
	textIDs  map[string]bool
}

type keyMap struct {
	Add      key.Binding
	Settings key.Binding
	Quit     key.Binding
	Left     key.Binding
	Right    key.Binding
	Up       key.Binding
	Down     key.Binding
}

var keys = keyMap{
	Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add/edit entry")),
	Settings: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev tab")),
	Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next tab")),
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.Add, k.Settings, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Left, k.Right, k.Up, k.Down, k.Add, k.Settings, k.Quit}}
}

func InitialModel(cfg *models.Config) Model {
	m := Model{
		help: help.New(),
		vp:   viewport.New(0, 0),
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
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
	if m.config != nil && m.config.App.Background.StarfieldEnabled {
		return starfieldTick()
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
		layout := dashboardLayoutForWidth(msg.Width)
		m.help.Width = layout.ContentWidth
		m.vp.Width = layout.ContentWidth
		m.vp.Height = msg.Height - 13 // 13 is approx header + footer height
		m.syncViewport()
	case starfieldTickMsg:
		if m.state == stateDashboard && m.config != nil && m.config.App.Background.StarfieldEnabled {
			m.stars = stepStars(m.stars, m.width, m.height)
			m.stars = maybeSpawnStar(m.stars, m.width, m.height, m.rng)
			m.twinkles = stepTwinkles(m.twinkles, m.width, m.height, m.rng)
			m.bursts = stepBurstParticles(m.bursts, m.width, m.height)
			if m.pulseTick > 0 {
				m.pulseTick--
			}
			return m, starfieldTick()
		}
	case setupDoneMsg:
		m.config = msg.cfg
		m.state = stateDashboard
		m.entries, _ = db.GetAllEntries()
		m.syncViewport()
		if m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
		return m, nil
	case setupCanceledMsg:
		m.state = stateDashboard
		m.syncViewport()
		return m, nil
	case settingsDoneMsg:
		m.state = stateDashboard
		m.settings = nil
		m.syncViewport()
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
		return m, nil
	case settingsRerunSetupMsg:
		m.settings = nil
		m.state = stateSetup
		m.setup = newAbortableSetupWiz(setupCanceledMsg{})
		return m, m.setup.Init()
	}

	switch m.state {
	case stateSetup:
		cmd := m.setup.Update(msg)
		return m, cmd
	case stateDashboard:
		return m.updateDashboard(msg)
	case stateEntryDate:
		return m.updateDateForm(msg)
	case stateForm:
		return m.updateForm(msg)
	case stateSettings:
		cmd := m.settings.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) syncViewport() {
	if m.width == 0 || m.config == nil {
		return
	}
	layout := dashboardLayoutForWidth(m.width)
	var content string
	if m.activeTab == 0 {
		content = m.viewOverview()
	} else if m.activeTab == len(m.config.Categories)+1 {
		content = m.viewInsights()
	} else {
		cat := m.config.Categories[m.activeTab-1]
		content = m.viewCategory(cat)
	}

	m.vp.Width = layout.ContentWidth
	m.vp.SetContent(lipgloss.NewStyle().Width(layout.ContentWidth).Render(content))
}

func (m Model) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Add):
			m.state = stateEntryDate
			m.initDateForm()
			if m.dateForm == nil {
				m.state = stateDashboard
				return m, nil
			}
			return m, m.dateForm.Init()
		case key.Matches(msg, keys.Settings):
			m.state = stateSettings
			m.settings = newSettingsWiz(m.config, m.entries)
			return m, m.settings.Init()
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

func (m *Model) initDateForm() {
	m.entryDate = time.Now().Format("2006-01-02")
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Entry date").
			Description("Use YYYY-MM-DD. Existing dates load for editing.").
			Value(&m.entryDate).
			Validate(validateEntryDate),
	))
}

func (m *Model) updateDateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		entry, err := db.GetEntryForDate(m.entryDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gotrack: failed to load entry: %v\n", err)
			m.state = stateDashboard
			return m, nil
		}
		m.state = stateForm
		m.initForm(entry)
		if m.form == nil {
			m.state = stateDashboard
			return m, nil
		}
		return m, m.form.Init()
	}
	if m.dateForm.State == huh.StateAborted {
		m.state = stateDashboard
		return m, nil
	}
	return m, cmd
}

func (m *Model) initForm(entry *models.Entry) {
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
				b := prefillBoolValue(entryData(entry), t.ID)
				m.boolPtrs[t.ID] = &b
				fields = append(fields, huh.NewConfirm().Title(t.Name).Value(&b))

			case models.TrackerDuration:
				s := prefillStringValue(entryData(entry), t.ID)
				m.strPtrs[t.ID] = &s
				fields = append(fields, huh.NewInput().Title(trackerInputLabel(t)).Value(&s).
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
				s := prefillStringValue(entryData(entry), t.ID)
				m.strPtrs[t.ID] = &s
				fields = append(fields, huh.NewInput().Title(trackerInputLabel(t)).Value(&s).
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
				s := prefillStringValue(entryData(entry), t.ID)
				m.strPtrs[t.ID] = &s
				fields = append(fields, huh.NewInput().Title(trackerInputLabel(t)).Value(&s).
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
				v := prefillIntValue(entryData(entry), t.ID, 3)
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
				s := prefillStringValue(entryData(entry), t.ID)
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
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}
	if m.form.State == huh.StateCompleted {
		m.saveEntry()
		m.state = stateDashboard
		m.entries, _ = db.GetAllEntries()
		m.triggerDashboardCelebration()
		m.syncViewport()
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
		return m, nil
	}
	if m.form.State == huh.StateAborted {
		m.state = stateDashboard
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
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
		Date: m.entryDate,
		Data: data,
	}
	if err := db.UpsertEntry(entry); err != nil {
		fmt.Fprintf(os.Stderr, "gotrack: failed to save entry: %v\n", err)
		return
	}
	if err := integrations.ExportObsidianEntry(m.config, entry); err != nil {
		fmt.Fprintf(os.Stderr, "gotrack: failed to export obsidian note: %v\n", err)
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Initializing..."
	}
	setActivePalette(m.config)
	switch m.state {
	case stateSetup:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.setup.View())
	case stateEntryDate:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.dateForm.View())
	case stateForm:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.form.View())
	case stateSettings:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.settings.View())
	default:
		view := lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Top,
			m.dashboardView())
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			view = applyStarfieldOverlay(view, renderStarfieldCanvas(m.width, m.height, m.stars, m.twinkles, m.bursts), true)
		}
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
	layout := dashboardLayoutForWidth(m.width)
	p := palette()
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).Bold(true).
		Align(lipgloss.Center).Width(layout.ContentWidth).MarginBottom(1).MarginTop(1)

	tabNames := []string{"Overview"}
	for _, c := range m.config.Categories {
		tabNames = append(tabNames, c.Name)
	}
	tabNames = append(tabNames, "Insights")

	var tabParts []string
	for i, name := range tabNames {
		style := lipgloss.NewStyle().Padding(0, 2)
		if i == m.activeTab {
			style = style.Background(lipgloss.Color(p.ActiveTabBg)).
				Foreground(lipgloss.Color(p.ActiveTabFg)).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color(p.InactiveTab))
		}
		tabParts = append(tabParts, style.Render(name))
	}
	tabBar := lipgloss.NewStyle().MarginBottom(1).Align(lipgloss.Center).Width(layout.ContentWidth).
		Render(lipgloss.JoinHorizontal(lipgloss.Top, tabParts...))

	helpView := lipgloss.NewStyle().MarginTop(1).Align(lipgloss.Center).Width(layout.ContentWidth).
		Render(m.help.View(keys))

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Render(banner),
		tabBar,
		m.vp.View(),
		helpView,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Align(lipgloss.Center).
		Render(content)
}

// ─── Overview Tab ─────────────────────────────────────────────────────────────

func (m Model) viewOverview() string {
	layout := dashboardLayoutForWidth(m.width)
	p := palette()
	if len(m.entries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).
			Render("No entries yet. Press 'a' to add or edit an entry.")
	}

	var cards []string
	accent := p.Primary
	if m.pulseTick > 0 && m.pulseTick%2 == 0 {
		accent = p.ChartSecondary
	}

	cards = append(cards,
		renderCard("Daily Snapshot", accent, m.overviewSnapshot(), layout.CardWidth, layout.CardHeight),
		renderCard("Momentum", accent, m.overviewMomentum(), layout.CardWidth, layout.CardHeight),
		renderCard("Personal Best", accent, m.overviewPersonalBest(), layout.CardWidth, layout.CardHeight),
	)

	for _, cat := range m.config.Categories {
		content := m.categorySummary(cat)
		color := cat.Color
		if color == "" {
			color = p.Primary
		}
		cards = append(cards, renderCard(cat.Name, color, content, layout.CardWidth, layout.CardHeight))
	}

	if len(cards) == 0 {
		return "No categories configured."
	}

	return renderCardGrid(cards, m.width)
}

func (m Model) categorySummary(cat models.Category) string {
	p := palette()
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
			color := p.Muted
			if t.Target != nil {
				if avg >= *t.Target {
					color = p.Success
				} else {
					color = p.Danger
				}
			}
			valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
			lines = append(lines, fmt.Sprintf("%-22s avg %s", truncate(t.Name, 22), valStyle.Render(formatValueWithUnit(avg, t))))

		case models.TrackerNumeric, models.TrackerCount:
			series := db.NumericSeries(m.entries, t.ID)
			if len(series) > 0 {
				latest := series[len(series)-1]
				color := p.Muted
				if t.Target != nil {
					if latest >= *t.Target {
						color = p.Success
					} else {
						color = p.Danger
					}
				}
				valStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				lines = append(lines, fmt.Sprintf("%-22s %s", truncate(t.Name, 22), valStyle.Render(formatValueWithUnit(latest, t))))
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
	layout := dashboardLayoutForWidth(m.width)
	if len(m.entries) == 0 {
		return "No entries yet."
	}

	const limit = 30
	var boxes []string

	if hitRate := m.categoryTargetHitRate(cat, limit); hitRate != "" {
		boxes = append(boxes, renderCard("Target Hit Rate (30d)", palette().Primary, hitRate, layout.CardWidth, layout.CardHeight))
	}
	if weekday := m.categoryWeekdayConsistency(cat); weekday != "" {
		boxes = append(boxes, renderCard("Consistency by Weekday", palette().Primary, weekday, layout.CardWidth, layout.CardHeight))
	}

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
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))

		case models.TrackerDuration, models.TrackerNumeric, models.TrackerCount:
			series := db.NumericSeries(m.entries, t.ID)
			content := renderLineChart(series, t, layout.CardWidth)
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))

		case models.TrackerRating:
			series := db.NumericSeries(m.entries, t.ID)
			content := renderLineChart(series, t, layout.CardWidth)
			if len(series) > 0 {
				content += fmt.Sprintf("\n\nAvg: %.1f / 5", average(series))
			}
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))

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
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))
		}
	}

	if len(boxes) == 0 {
		return "No trackers in this category."
	}

	return renderCardGrid(boxes, m.width)
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

func (m Model) overviewSnapshot() string {
	if len(m.entries) == 0 {
		return "No entries yet."
	}
	latest := m.entries[0]
	totalTrackers := 0
	recorded := 0
	activeStreaks := 0
	targetHits := 0
	targetTotal := 0

	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			totalTrackers++
			if _, ok := latest.Data[t.ID]; ok {
				recorded++
			}
			if t.Type == models.TrackerBinary {
				if db.CurrentStreak(m.entries, t.ID) > 0 {
					activeStreaks++
				}
			}
			if t.Target != nil && (t.Type == models.TrackerDuration || t.Type == models.TrackerCount || t.Type == models.TrackerNumeric) {
				hits, total, _ := db.TargetHitRate(m.entries, t.ID, *t.Target, 30)
				targetHits += hits
				targetTotal += total
			}
		}
	}

	coverage := 0.0
	if totalTrackers > 0 {
		coverage = float64(recorded) / float64(totalTrackers) * 100
	}

	return fmt.Sprintf(
		"Latest entry: %s\nCoverage: %.0f%% (%d/%d)\nActive streaks: %d\nTargets hit (30d): %d/%d",
		latest.Date, coverage, recorded, totalTrackers, activeStreaks, targetHits, targetTotal,
	)
}

func (m Model) overviewMomentum() string {
	trackerLabel := map[string]string{}
	var trackerIDs []string
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type == models.TrackerDuration || t.Type == models.TrackerCount || t.Type == models.TrackerNumeric {
				trackerIDs = append(trackerIDs, t.ID)
				trackerLabel[t.ID] = t.Name
			}
		}
	}
	rows := db.MomentumAccelerationRanking(m.entries, trackerIDs, 7)
	if len(rows) == 0 {
		return "Need at least 14 entries on numeric trackers."
	}
	var lines []string
	for i, d := range rows {
		if i >= 3 {
			break
		}
		lines = append(lines, fmt.Sprintf("%s\n%s", truncate(trackerLabel[d.TrackerID], 20), TrendDeltaStrip(d.RecentAvg, d.PrevAvg)))
	}
	return strings.Join(lines, "\n\n")
}

func (m Model) overviewPersonalBest() string {
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerDuration && t.Type != models.TrackerCount && t.Type != models.TrackerNumeric {
				continue
			}
			best, date, ok := db.PersonalBest(m.entries, t.ID)
			if !ok {
				continue
			}
			return fmt.Sprintf("%s\nBest: %s\nDate: %s", truncate(t.Name, 18), formatValueWithUnit(best, t), date)
		}
	}
	return "No numeric personal bests yet."
}

func (m Model) categoryTargetHitRate(cat models.Category, window int) string {
	var blocks []string
	for _, t := range cat.Trackers {
		if t.Target == nil {
			continue
		}
		switch t.Type {
		case models.TrackerDuration, models.TrackerCount, models.TrackerNumeric:
			hits, total, _ := db.TargetHitRate(m.entries, t.ID, *t.Target, window)
			blocks = append(blocks, fmt.Sprintf("%s\n%s", truncate(t.Name, 18), TargetHitMeter(hits, total, 22)))
		}
	}
	if len(blocks) == 0 {
		return ""
	}
	return strings.Join(blocks, "\n\n")
}

func (m Model) categoryWeekdayConsistency(cat models.Category) string {
	var binTrackers []models.Tracker
	for _, t := range cat.Trackers {
		if t.Type == models.TrackerBinary {
			binTrackers = append(binTrackers, t)
		}
	}
	if len(binTrackers) == 0 {
		return ""
	}
	// Show first binary tracker to keep card compact and readable.
	t := binTrackers[0]
	return fmt.Sprintf("%s\n\n%s", truncate(t.Name, 20), WeekdayConsistencyBars(db.BinaryWeekdayConsistency(m.entries, t.ID)))
}

func (m Model) insightBestWeekday() string {
	var totals [7]float64
	var counts [7]int
	var seen bool
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			seen = true
			w := db.BinaryWeekdayConsistency(m.entries, t.ID)
			for i := range totals {
				if w[i] > 0 {
					totals[i] += w[i]
					counts[i]++
				}
			}
		}
	}
	if !seen {
		return "No binary trackers available."
	}
	var avg [7]float64
	for i := range avg {
		if counts[i] == 0 {
			continue
		}
		avg[i] = totals[i] / float64(counts[i])
	}
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	bestIdx := 0
	for i := 1; i < 7; i++ {
		if avg[i] > avg[bestIdx] {
			bestIdx = i
		}
	}
	return fmt.Sprintf("Best day: %s\nScore: %.0f%% average\n\n%s", days[bestIdx], avg[bestIdx], WeekdayConsistencyBars(avg))
}

func (m Model) insightMomentumLeaders() string {
	trackerLabel := map[string]string{}
	var trackerIDs []string
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type == models.TrackerDuration || t.Type == models.TrackerCount || t.Type == models.TrackerNumeric {
				trackerIDs = append(trackerIDs, t.ID)
				trackerLabel[t.ID] = t.Name
			}
		}
	}
	accel := db.MomentumAccelerationRanking(m.entries, trackerIDs, 7)
	if len(accel) == 0 {
		return "No momentum data yet."
	}
	var rows []LeaderboardRow
	for _, a := range accel {
		rows = append(rows, LeaderboardRow{
			Label: trackerLabel[a.TrackerID],
			Delta: a.Delta,
		})
	}
	return LeaderboardBars(rows, 10)
}

func (m *Model) triggerDashboardCelebration() {
	if len(m.entries) == 0 || m.width <= 0 || m.height <= 0 || m.config == nil {
		return
	}
	latest := m.entries[0]
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Target == nil {
				continue
			}
			v, ok := latest.Data[t.ID].(float64)
			if !ok || v < *t.Target {
				continue
			}
			m.pulseTick = 8
			if m.rng != nil {
				m.bursts = append(m.bursts, spawnBurstParticles(m.width, m.height, m.width/2, m.height/3, m.rng)...)
			}
			return
		}
	}
}

// ─── Insights Tab ─────────────────────────────────────────────────────────────

func (m Model) viewInsights() string {
	layout := dashboardLayoutForWidth(m.width)
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
	weekdayStr := m.insightBestWeekday()
	momentumStr := m.insightMomentumLeaders()

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

	return renderCardGrid([]string{
		renderCard("A/B Impact", palette().Primary, compStr, layout.CardWidth, layout.CardHeight),
		renderCard("Scatter Analysis", palette().Primary, scatterStr, layout.CardWidth, layout.CardHeight),
		renderCard("Best Day of Week", palette().Primary, weekdayStr, layout.CardWidth, layout.CardHeight),
		renderCard("Momentum Leaders/Laggers", palette().Primary, momentumStr, layout.CardWidth, layout.CardHeight),
	}, m.width)
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
