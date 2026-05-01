package tui

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
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
	stateSetup          appState = iota // first-launch wizard
	stateDashboard                      // main tabs
	stateEntryDate                      // pick date before entry form
	stateForm                           // daily entry
	stateSettings                       // settings/config editor
	stateDeleteDate                     // pick date to delete
	stateDeleteConfirm                  // confirm deletion
	stateHelp                           // full-screen help overlay
	stateEditPick                       // pick an existing entry to edit
	stateQuickPick                      // pick tracker for quick entry
	stateQuickValue                     // enter one tracker value
	statePomodoroSetup                  // choose tracker and duration
	statePomodoroActive                 // running pomodoro timer
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
	activeTab         int
	heroIndex         int
	reviewMonthly     bool
	dismissedReminder bool
	undoSnapshot      *models.Entry
	undoHad           bool
	config            *models.Config
	entries           []models.Entry
	settings          *settingsWiz
	stars             []fallingStar
	twinkles          []twinkleStar
	bursts            []burstParticle
	pulseTick         int
	rng               *rand.Rand

	// Entry date picker
	dateForm        *huh.Form
	entryDate       string
	deleteConfirmed *bool

	// Entry form
	form        *huh.Form
	boolPtrs    map[string]*bool
	strPtrs     map[string]*string
	intPtrs     map[string]*int
	textIDs     map[string]bool
	catDonePtrs map[string]*bool // per-category "did you do this today?" gate
	catTrackers map[string][]string // catID -> tracker IDs in that category

	// Quick entry
	quickTrackerID string
	quickValue     string

	// Pomodoro
	pomodoroTrackerID string
	pomodoroMinutes   string
	pomodoroStarted   time.Time
	pomodoroDuration  time.Duration
	pomodoroNow       time.Time
}

type keyMap struct {
	Add      key.Binding
	Delete   key.Binding
	Settings key.Binding
	Quit     key.Binding
	Left     key.Binding
	Right    key.Binding
	Up       key.Binding
	Down     key.Binding
	HeroPrev key.Binding
	HeroNext key.Binding
	Review   key.Binding
	Help     key.Binding
	Undo     key.Binding
	Dismiss  key.Binding
	Edit     key.Binding
	Quick    key.Binding
	Pomodoro key.Binding
}

var keys = keyMap{
	Add:      key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add/edit entry")),
	Delete:   key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete entry")),
	Settings: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
	Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev tab")),
	Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next tab")),
	Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
	Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
	HeroPrev: key.NewBinding(key.WithKeys("["), key.WithHelp("[", "prev visual")),
	HeroNext: key.NewBinding(key.WithKeys("]"), key.WithHelp("]", "next visual")),
	Review:   key.NewBinding(key.WithKeys("w"), key.WithHelp("w", "week/month toggle")),
	Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Undo:     key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "undo last save")),
	Dismiss:  key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "dismiss reminder")),
	Edit:     key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit recent entry")),
	Quick:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "quick entry")),
	Pomodoro: key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "pomodoro")),
	Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Left, k.Right, k.Up, k.Down, k.HeroPrev, k.HeroNext, k.Add, k.Quick, k.Pomodoro, k.Settings, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Left, k.Right, k.Up, k.Down, k.HeroPrev, k.HeroNext, k.Add, k.Quick, k.Pomodoro, k.Delete, k.Settings, k.Quit}}
}

type pomodoroTickMsg time.Time

func pomodoroTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return pomodoroTickMsg(t)
	})
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
	case pomodoroTickMsg:
		if m.state == statePomodoroActive {
			m.pomodoroNow = time.Time(msg)
			if !m.pomodoroStarted.IsZero() && m.pomodoroNow.Sub(m.pomodoroStarted) >= m.pomodoroDuration {
				if err := m.completePomodoro(time.Now().Format("2006-01-02"), m.pomodoroStarted.Add(m.pomodoroDuration)); err != nil {
					fmt.Fprintf(os.Stderr, "gotrack: failed to save pomodoro: %v\n", err)
				}
				m.state = stateDashboard
				m.entries, _ = db.GetAllEntries()
				m.triggerDashboardCelebration()
				m.syncViewport()
				if m.config != nil && m.config.App.Background.StarfieldEnabled {
					return m, starfieldTick()
				}
				return m, nil
			}
			return m, pomodoroTick()
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
	case stateEntryDate, stateDeleteDate:
		return m.updateDateForm(msg)
	case stateEditPick:
		return m.updateEditPick(msg)
	case stateQuickPick, stateQuickValue:
		return m.updateQuickEntry(msg)
	case statePomodoroSetup:
		return m.updatePomodoroSetup(msg)
	case statePomodoroActive:
		return m.updatePomodoroActive(msg)
	case stateDeleteConfirm:
		return m.updateDeleteConfirm(msg)
	case stateHelp:
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.Type == tea.KeyEsc || keyMsg.String() == "?" || keyMsg.String() == "q" {
				m.state = stateDashboard
			}
		}
		return m, nil
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
	} else if m.activeTab == len(m.config.Categories)+2 {
		content = m.viewReview()
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
			m.initDateForm("Entry date", "YYYY-MM-DD, MM-DD, t, y, -N. Existing dates load for editing.")
			if m.dateForm == nil {
				m.state = stateDashboard
				return m, nil
			}
			return m, m.dateForm.Init()
		case key.Matches(msg, keys.Delete):
			if len(m.entries) == 0 {
				return m, nil
			}
			m.state = stateDeleteDate
			m.initDateForm("Delete entry", "YYYY-MM-DD, MM-DD, t, y, or -N.")
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
			total := len(m.config.Categories) + 3
			m.activeTab = (m.activeTab - 1 + total) % total
			m.syncViewport()
			m.vp.GotoTop()
		case key.Matches(msg, keys.Right):
			total := len(m.config.Categories) + 3
			m.activeTab = (m.activeTab + 1) % total
			m.syncViewport()
			m.vp.GotoTop()
		case key.Matches(msg, keys.HeroPrev):
			if m.activeTab == 0 {
				n := len(heroVisuals)
				m.heroIndex = (m.heroIndex - 1 + n) % n
				m.syncViewport()
				m.vp.GotoTop()
			}
		case key.Matches(msg, keys.HeroNext):
			if m.activeTab == 0 {
				n := len(heroVisuals)
				m.heroIndex = (m.heroIndex + 1) % n
				m.syncViewport()
				m.vp.GotoTop()
			}
		case key.Matches(msg, keys.Help):
			m.state = stateHelp
			return m, nil
		case key.Matches(msg, keys.Review):
			if m.activeTab == len(m.config.Categories)+2 {
				m.reviewMonthly = !m.reviewMonthly
				m.syncViewport()
				m.vp.GotoTop()
			}
		case key.Matches(msg, keys.Dismiss):
			m.dismissedReminder = true
			m.syncViewport()
		case key.Matches(msg, keys.Undo):
			if m.undoHad {
				if m.undoSnapshot != nil {
					_ = db.UpsertEntry(m.undoSnapshot)
				} else {
					_ = db.DeleteEntry(m.entryDate)
				}
				m.entries, _ = db.GetAllEntries()
				m.undoHad = false
				m.undoSnapshot = nil
				m.syncViewport()
			}
		case key.Matches(msg, keys.Edit):
			if len(m.entries) == 0 {
				return m, nil
			}
			m.initEditPick()
			if m.dateForm == nil {
				return m, nil
			}
			m.state = stateEditPick
			return m, m.dateForm.Init()
		case key.Matches(msg, keys.Quick):
			m.initQuickPick()
			if m.dateForm == nil {
				return m, nil
			}
			m.state = stateQuickPick
			return m, m.dateForm.Init()
		case key.Matches(msg, keys.Pomodoro):
			m.initPomodoroSetup()
			if m.dateForm == nil {
				return m, nil
			}
			m.state = statePomodoroSetup
			return m, m.dateForm.Init()
		}
	}

	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func quickEntryTrackers(cfg *models.Config) []models.Tracker {
	if cfg == nil {
		return nil
	}
	var trackers []models.Tracker
	for _, cat := range cfg.Categories {
		trackers = append(trackers, cat.Trackers...)
	}
	return trackers
}

func durationTrackers(cfg *models.Config) []models.Tracker {
	if cfg == nil {
		return nil
	}
	var trackers []models.Tracker
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			if t.Type == models.TrackerDuration {
				trackers = append(trackers, t)
			}
		}
	}
	return trackers
}

func trackerByID(cfg *models.Config, id string) (models.Tracker, bool) {
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			if t.ID == id {
				return t, true
			}
		}
	}
	return models.Tracker{}, false
}

func trackerSelectOptions(trackers []models.Tracker) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(trackers))
	for _, t := range trackers {
		opts = append(opts, huh.NewOption(t.Name, t.ID))
	}
	return opts
}

func (m *Model) initQuickPick() {
	trackers := quickEntryTrackers(m.config)
	if len(trackers) == 0 {
		m.dateForm = nil
		return
	}
	m.quickTrackerID = trackers[0].ID
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Key("quickTrackerID").
			Title("Quick entry").
			Description("Pick one tracker to log for today.").
			Options(trackerSelectOptions(trackers)...).
			Value(&m.quickTrackerID),
	))
}

func (m *Model) initQuickValue() {
	t, _ := trackerByID(m.config, m.quickTrackerID)
	m.quickValue = ""
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Key("quickValue").
			Title(trackerInputLabel(t)).
			Description("Enter one value. Existing fields for today are preserved.").
			Value(&m.quickValue).
			Validate(func(s string) error {
				_, err := db.CoerceValue(t, s)
				return err
			}),
	))
}

func (m Model) updateQuickEntry(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		if m.state == stateQuickPick {
			if selected := m.dateForm.GetString("quickTrackerID"); selected != "" {
				m.quickTrackerID = selected
			}
			m.state = stateQuickValue
			m.initQuickValue()
			return m, m.dateForm.Init()
		}
		if value := m.dateForm.GetString("quickValue"); value != "" {
			m.quickValue = value
		}
		if err := m.saveQuickEntry(time.Now().Format("2006-01-02")); err != nil {
			fmt.Fprintf(os.Stderr, "gotrack: failed to save quick entry: %v\n", err)
		}
		m.entries, _ = db.GetAllEntries()
		m.triggerDashboardCelebration()
		m.state = stateDashboard
		m.syncViewport()
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
		return m, nil
	}
	if m.dateForm.State == huh.StateAborted {
		m.state = stateDashboard
		return m, nil
	}
	return m, cmd
}

func (m Model) saveQuickEntry(date string) error {
	if err := db.UpsertEntryLog(m.config, date, map[string]interface{}{m.quickTrackerID: m.quickValue}); err != nil {
		return err
	}
	entry, err := db.GetEntryForDate(date)
	if err != nil {
		return err
	}
	return integrations.ExportObsidianEntry(m.config, entry)
}

func (m *Model) initPomodoroSetup() {
	trackers := durationTrackers(m.config)
	if len(trackers) == 0 {
		m.dateForm = nil
		return
	}
	m.pomodoroTrackerID = trackers[0].ID
	m.pomodoroMinutes = "25"
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Key("pomodoroTrackerID").
			Title("Pomodoro tracker").
			Description("Pick the duration tracker that should receive this time.").
			Options(trackerSelectOptions(trackers)...).
			Value(&m.pomodoroTrackerID),
		huh.NewInput().
			Key("pomodoroMinutes").
			Title("Session length (minutes)").
			Value(&m.pomodoroMinutes).
			Validate(func(s string) error {
				v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
				if err != nil || v <= 0 {
					return fmt.Errorf("enter a positive number of minutes")
				}
				return nil
			}),
	))
}

func (m Model) updatePomodoroSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		if selected := m.dateForm.GetString("pomodoroTrackerID"); selected != "" {
			m.pomodoroTrackerID = selected
		}
		if minutes := m.dateForm.GetString("pomodoroMinutes"); minutes != "" {
			m.pomodoroMinutes = minutes
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(m.pomodoroMinutes), 64)
		if err != nil || v <= 0 {
			m.state = stateDashboard
			return m, nil
		}
		m.pomodoroDuration = time.Duration(v * float64(time.Minute))
		m.pomodoroStarted = time.Now()
		m.pomodoroNow = m.pomodoroStarted
		m.state = statePomodoroActive
		return m, pomodoroTick()
	}
	if m.dateForm.State == huh.StateAborted {
		m.state = stateDashboard
		return m, nil
	}
	return m, cmd
}

func (m Model) updatePomodoroActive(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "e", "enter", " ":
			if err := m.completePomodoro(time.Now().Format("2006-01-02"), time.Now()); err != nil {
				fmt.Fprintf(os.Stderr, "gotrack: failed to save pomodoro: %v\n", err)
			}
			m.entries, _ = db.GetAllEntries()
			m.triggerDashboardCelebration()
			m.state = stateDashboard
			m.syncViewport()
			if m.config != nil && m.config.App.Background.StarfieldEnabled {
				return m, starfieldTick()
			}
			return m, nil
		case "esc":
			m.state = stateDashboard
			if m.config != nil && m.config.App.Background.StarfieldEnabled {
				return m, starfieldTick()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m Model) completePomodoro(date string, ended time.Time) error {
	elapsed := ended.Sub(m.pomodoroStarted).Minutes()
	if elapsed <= 0 {
		return nil
	}
	if err := db.AddDurationToEntry(m.config, date, m.pomodoroTrackerID, elapsed); err != nil {
		return err
	}
	entry, err := db.GetEntryForDate(date)
	if err != nil {
		return err
	}
	return integrations.ExportObsidianEntry(m.config, entry)
}

func (m *Model) initDateForm(title, description string) {
	m.entryDate = time.Now().Format("2006-01-02")
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Key("entryDate").
			Title(title).
			Description(description).
			Placeholder("e.g. t, y, 4-18, 2026-04-18").
			Value(&m.entryDate).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("date cannot be empty")
				}
				return validateEntryDate(s)
			}),
	))
}

func (m *Model) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		if m.deleteConfirmed != nil && *m.deleteConfirmed {
			// Snapshot for undo.
			prior, _ := db.GetEntryForDate(m.entryDate)
			m.undoSnapshot = prior
			m.undoHad = true
			if err := db.DeleteEntry(m.entryDate); err != nil {
				fmt.Fprintf(os.Stderr, "gotrack: failed to delete entry: %v\n", err)
			}
			m.entries, _ = db.GetAllEntries()
		}
		m.state = stateDashboard
		m.syncViewport()
		if m.config != nil && m.config.App.Background.StarfieldEnabled {
			return m, starfieldTick()
		}
		return m, nil
	}
	return m, cmd
}

func (m *Model) initEditPick() {
	if len(m.entries) == 0 {
		m.dateForm = nil
		return
	}
	sorted := make([]models.Entry, len(m.entries))
	copy(sorted, m.entries)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Date > sorted[j].Date })
	limit := 20
	if len(sorted) < limit {
		limit = len(sorted)
	}
	opts := make([]huh.Option[string], 0, limit)
	for _, e := range sorted[:limit] {
		label := e.Date
		if summary := entrySummary(e, m.config); summary != "" {
			label = fmt.Sprintf("%s  %s", e.Date, summary)
		}
		opts = append(opts, huh.NewOption(label, e.Date))
	}
	m.entryDate = sorted[0].Date
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Key("entryDate").
			Title("Edit which entry?").
			Description("Pick a recent entry to open in the edit form.").
			Options(opts...).
			Value(&m.entryDate),
	))
}

func entrySummary(e models.Entry, cfg *models.Config) string {
	if cfg == nil || e.Data == nil {
		return ""
	}
	filled := 0
	total := 0
	for _, cat := range cfg.Categories {
		for _, t := range cat.Trackers {
			total++
			if v, ok := e.Data[t.ID]; ok && v != nil && v != "" && v != false {
				filled++
			}
		}
	}
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("(%d/%d filled)", filled, total)
}

func (m *Model) updateEditPick(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		input := m.dateForm.GetString("entryDate")
		if input == "" {
			input = m.entryDate
		}
		m.entryDate = input
		entry, err := db.GetEntryForDate(m.entryDate)
		if err != nil {
			m.state = stateDashboard
			return m, nil
		}
		m.initForm(entry)
		if m.form == nil {
			m.state = stateDashboard
			return m, nil
		}
		m.state = stateForm
		return m, m.form.Init()
	}
	return m, cmd
}

func (m *Model) initDeleteConfirm() {
	confirm := false
	m.dateForm = huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title(fmt.Sprintf("Delete entry for %s?", m.entryDate)).
			Description("This cannot be undone easily. Press y to confirm, n to cancel.").
			Affirmative("Delete").
			Negative("Cancel").
			Value(&confirm),
	))
	// stash confirm pointer on the model via a per-state field
	m.deleteConfirmed = &confirm
}

func (m *Model) updateDateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.Type == tea.KeyEsc {
		m.state = stateDashboard
		return m, nil
	}
	deleting := m.state == stateDeleteDate
	form, cmd := m.dateForm.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.dateForm = f
	}
	if m.dateForm.State == huh.StateCompleted {
		// Fetch value securely from the form, bypassing the value-receiver copy of m.entryDate.
		input := m.dateForm.GetString("entryDate")
		if input == "" {
			input = m.entryDate // Fallback if Key isn't populated
		}

		fmt.Fprintf(os.Stderr, "gotrack: debug: raw input date: '%s'\n", input)
		normalized, err := normalizeEntryDate(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gotrack: debug: normalize error: %v\n", err)
			m.state = stateDashboard
			return m, nil
		}
		fmt.Fprintf(os.Stderr, "gotrack: debug: normalized date: '%s'\n", normalized)
		m.entryDate = normalized
		if deleting {
			m.state = stateDeleteConfirm
			m.initDeleteConfirm()
			if m.dateForm == nil {
				m.state = stateDashboard
				return m, nil
			}
			return m, m.dateForm.Init()
		}
		entry, err := db.GetEntryForDate(m.entryDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gotrack: failed to load entry: %v\n", err)
			m.state = stateDashboard
			return m, nil
		}
		m.state = stateForm
		m.initForm(entry)
		if m.form == nil {
			fmt.Fprintf(os.Stderr, "gotrack: debug: initForm returned nil\n")
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
	m.catDonePtrs = make(map[string]*bool)
	m.catTrackers = make(map[string][]string)

	// Split categories: reflection last, rest first.
	var mainCats, reflectionCats []models.Category
	for _, cat := range m.config.Categories {
		if strings.EqualFold(cat.Name, "reflection") {
			reflectionCats = append(reflectionCats, cat)
		} else {
			mainCats = append(mainCats, cat)
		}
	}
	orderedCats := append(mainCats, reflectionCats...)

	var groups []*huh.Group

	for _, cat := range orderedCats {
		cat := cat // capture loop var for closures
		fields := m.buildCategoryFields(entry, cat)
		if len(fields) == 0 {
			continue
		}

		isReflection := strings.EqualFold(cat.Name, "reflection")
		if isReflection {
			groups = append(groups, huh.NewGroup(fields...).Title(cat.Name))
		} else {
			done := entry != nil && m.catHasData(entry, cat)
			m.catDonePtrs[cat.ID] = &done
			donePtr := m.catDonePtrs[cat.ID]

			gateGroup := huh.NewGroup(
				huh.NewConfirm().
					Title(cat.Name).
					Description("Did you do this today?").
					Value(donePtr),
			)
			trackerGroup := huh.NewGroup(fields...).
				Title(cat.Name).
				WithHideFunc(func() bool { return !*donePtr })

			groups = append(groups, gateGroup, trackerGroup)
		}
	}

	if len(groups) == 0 {
		fmt.Fprintf(os.Stderr, "gotrack: debug: no groups created (categories: %d)\n", len(m.config.Categories))
		m.form = nil
		return
	}
	m.form = huh.NewForm(groups...)
}

func (m *Model) buildCategoryFields(entry *models.Entry, cat models.Category) []huh.Field {
	var fields []huh.Field
	var trackerIDs []string
	for _, t := range cat.Trackers {
		t := t
		trackerIDs = append(trackerIDs, t.ID)
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
	m.catTrackers[cat.ID] = trackerIDs
	return fields
}

// catHasData reports whether any tracker in cat has a value in entry.
func (m *Model) catHasData(entry *models.Entry, cat models.Category) bool {
	if entry == nil {
		return false
	}
	for _, t := range cat.Trackers {
		if _, ok := entry.Data[t.ID]; ok {
			return true
		}
	}
	return false
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
	fmt.Fprintf(os.Stderr, "gotrack: debug: saving entry for date: '%s'\n", m.entryDate)
	prior, _ := db.GetEntryForDate(m.entryDate)
	m.undoSnapshot = prior
	m.undoHad = true
	data := make(map[string]interface{})

	// Build set of tracker IDs that belong to skipped (gated-off) categories.
	skippedTrackers := make(map[string]bool)
	for catID, donePtr := range m.catDonePtrs {
		if !*donePtr {
			for _, tid := range m.catTrackers[catID] {
				skippedTrackers[tid] = true
			}
		}
	}

	for id, ptr := range m.boolPtrs {
		if skippedTrackers[id] {
			continue
		}
		data[id] = *ptr
	}
	for id, ptr := range m.strPtrs {
		if skippedTrackers[id] || *ptr == "" {
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
		if skippedTrackers[id] {
			continue
		}
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
	case stateEntryDate, stateDeleteDate, stateDeleteConfirm, stateEditPick, stateQuickPick, stateQuickValue, statePomodoroSetup:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.dateForm.View())
	case statePomodoroActive:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.pomodoroView())
	case stateHelp:
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
			m.helpOverlay())
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

func (m Model) pomodoroView() string {
	p := palette()
	now := m.pomodoroNow
	if now.IsZero() {
		now = time.Now()
	}
	elapsed := now.Sub(m.pomodoroStarted)
	if elapsed < 0 {
		elapsed = 0
	}
	remaining := m.pomodoroDuration - elapsed
	if remaining < 0 {
		remaining = 0
	}
	t, _ := trackerByID(m.config, m.pomodoroTrackerID)
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Render("Pomodoro")
	body := strings.Join([]string{
		fmt.Sprintf("Tracking: %s", t.Name),
		fmt.Sprintf("Remaining: %s", formatPomodoroDuration(remaining)),
		fmt.Sprintf("Elapsed:   %s", formatPomodoroDuration(elapsed)),
		"",
		"Press e, enter, or space to end and log elapsed time.",
		"Press esc to cancel without logging.",
	}, "\n")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.Primary)).
		Padding(1, 3).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body))
}

func formatPomodoroDuration(d time.Duration) string {
	total := int(d.Round(time.Second).Seconds())
	minutes := total / 60
	seconds := total % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// ─── Dashboard ────────────────────────────────────────────────────────────────

const banner = `______     ______     ______   ______     ______     ______     __  __
/\  ___\   /\  __ \   /\__  _\ /\  == \   /\  __ \   /\  ___\   /\ \/ /
\ \ \__ \  \ \ \/\ \  \/_/\ \/ \ \  __<   \ \  __ \  \ \ \____  \ \  _"-.
 \ \_____\  \ \_____\    \ \_\  \ \_\ \_\  \ \_\ \_\  \ \_____\  \ \_\ \_\
  \/_____/   \/_____/     \/_/   \/_/ /_/   \/_/\/_/   \/_____/   \/_/\/_/`

const bannerCompact = "── GoTrack ──"

// bannerForWidth picks the full ASCII banner when the terminal is wide enough
// to render it without wrapping, and a compact label otherwise.
func bannerForWidth(w int) string {
	if w >= 76 {
		return banner
	}
	return bannerCompact
}

// renderTabsSingleLine lays out tabs on a single row. Spacing between tabs
// is always consistent (padding 2). When labels + padding exceed available
// width, tabs split into pages and only the page containing the active tab
// renders, with a "1/3" indicator and ‹/› markers for neighboring pages.
func (m Model) renderTabsSingleLine(names []string, maxWidth int, p ThemePalette) string {
	const pad = 2

	renderTab := func(idx int) string {
		style := lipgloss.NewStyle().Padding(0, pad)
		if idx == m.activeTab {
			style = style.Background(lipgloss.Color(p.ActiveTabBg)).
				Foreground(lipgloss.Color(p.ActiveTabFg)).Bold(true)
		} else {
			style = style.Foreground(lipgloss.Color(p.InactiveTab))
		}
		return style.Render(names[idx])
	}

	// Fast path: if everything fits, render all tabs.
	full := make([]string, len(names))
	for i := range names {
		full[i] = renderTab(i)
	}
	joined := lipgloss.JoinHorizontal(lipgloss.Top, full...)
	if lipgloss.Width(joined) <= maxWidth {
		return joined
	}

	// Paging: greedy-split tabs into pages that each fit maxWidth minus
	// headroom for ‹/› markers and the page counter.
	const reserve = 10 // room for "‹ " + " ›" + " 2/3"
	budget := maxWidth - reserve
	if budget < 10 {
		budget = 10
	}

	var pages [][]int // each page = list of tab indices
	cur := []int{}
	curWidth := 0
	for i := range names {
		w := lipgloss.Width(full[i])
		if len(cur) > 0 && curWidth+w > budget {
			pages = append(pages, cur)
			cur = nil
			curWidth = 0
		}
		cur = append(cur, i)
		curWidth += w
	}
	if len(cur) > 0 {
		pages = append(pages, cur)
	}

	// Find the page containing the active tab.
	activePage := 0
	for pi, page := range pages {
		for _, idx := range page {
			if idx == m.activeTab {
				activePage = pi
				break
			}
		}
	}

	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.InactiveTab))
	parts := make([]string, 0, len(pages[activePage]))
	for _, idx := range pages[activePage] {
		parts = append(parts, renderTab(idx))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, parts...)

	prefix := "  "
	suffix := "  "
	if activePage > 0 {
		prefix = muted.Render("‹ ")
	}
	if activePage < len(pages)-1 {
		suffix = muted.Render(" ›")
	}
	counter := muted.Render(fmt.Sprintf(" %d/%d", activePage+1, len(pages)))
	return prefix + row + suffix + counter
}

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
	tabNames = append(tabNames, "Insights", "Review")

	tabBar := lipgloss.NewStyle().MarginBottom(1).Align(lipgloss.Center).Width(layout.ContentWidth).
		Render(m.renderTabsSingleLine(tabNames, layout.ContentWidth, p))

	helpView := lipgloss.NewStyle().MarginTop(1).Align(lipgloss.Center).Width(layout.ContentWidth).
		Render(m.help.View(keys))

	var banner string
	if !m.dismissedReminder && !db.HasEntryForToday(m.entries) && len(m.entries) > 0 {
		banner = lipgloss.NewStyle().
			Foreground(lipgloss.Color(p.Danger)).
			Bold(true).
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(p.Danger)).
			Padding(0, 1).
			Width(layout.ContentWidth - 2).
			Align(lipgloss.Center).
			Render(fmt.Sprintf("! No entry for %s — press 'a' to log, 'D' to dismiss",
				time.Now().Format("2006-01-02")))
	}

	parts := []string{titleStyle.Render(bannerForWidth(layout.ContentWidth))}
	if banner != "" {
		parts = append(parts, banner)
	}
	parts = append(parts, tabBar, m.vp.View(), helpView)
	content := lipgloss.JoinVertical(lipgloss.Center, parts...)

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

	accent := p.Primary
	if m.pulseTick > 0 && m.pulseTick%2 == 0 {
		accent = p.ChartSecondary
	}

	hero := m.renderHero(layout.ContentWidth)

	var cards []string
	if nn := m.overviewNonNegotiables(layout.CardWidth - 6); nn != "" {
		cards = append(cards, renderCard("Non-Negotiables", accent, nn, layout.CardWidth, layout.CardHeight))
	}
	cards = append(cards,
		renderCard("Daily Snapshot", accent, m.overviewSnapshot(), layout.CardWidth, layout.CardHeight),
		renderCard("Last 7 Days", accent, m.overviewLastWeek(), layout.CardWidth, layout.CardHeight),
		renderCard("Momentum", accent, m.overviewMomentum(), layout.CardWidth, layout.CardHeight),
		renderCard("Streaks", accent, m.overviewStreaks(), layout.CardWidth, layout.CardHeight),
		renderCard("Personal Bests", accent, m.overviewPersonalBest(), layout.CardWidth, layout.CardHeight),
	)
	for _, cat := range m.config.Categories {
		content := m.categorySummary(cat)
		color := cat.Color
		if color == "" {
			color = p.Primary
		}
		cards = append(cards, renderCard(cat.Name, color, content, layout.CardWidth, layout.CardHeight))
	}

	grid := renderCardGrid(cards, m.width)
	return lipgloss.JoinVertical(lipgloss.Center, hero, "", grid)
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
			lines = append(lines, fmt.Sprintf("%-22s avg %s", truncate(t.Name, 22), valStyle.Render(formatAverageWithUnit(avg, t))))

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
				lines = append(lines, fmt.Sprintf("%-22s avg %s/5", truncate(t.Name, 22), formatAverageFixed2(avg)))
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
		boxes = append(boxes, renderCard(fmt.Sprintf("Target Hit Rate (last %d)", limit), palette().Primary, hitRate, layout.CardWidth, layout.CardHeight))
	}
	if weekday := m.categoryWeekdayConsistency(cat); weekday != "" {
		boxes = append(boxes, renderCard("Consistency by Weekday", palette().Primary, weekday, layout.CardWidth, layout.CardHeight))
	}

	for _, t := range cat.Trackers {
		anchor := time.Now()
		switch t.Type {
		case models.TrackerBinary:
			heat := db.BinaryByDate(m.entries, t.ID, anchor, limit)
			pct := db.ConsistencyPct(m.entries, t.ID)
			streak := db.CurrentStreak(m.entries, t.ID)
			// Align the first visible heatmap cell with its real weekday.
			firstShown := anchor.AddDate(0, 0, -(limit - 1))
			offset := int(firstShown.Weekday())

			content := fmt.Sprintf("Streak: %d days  |  All-time: %.0f%%\n\n", streak, pct) +
				Heatmap(heat, offset)
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))

		case models.TrackerDuration, models.TrackerNumeric, models.TrackerCount:
			// For stats (average, target hits), we use the compressed series of logged days.
			statsSeries := db.NumericSeries(m.entries, t.ID)
			// For the visual trend, we use a date-aligned series to show gaps.
			trendSeries, _ := db.NumericByDate(m.entries, t.ID, anchor, limit)

			content := renderLineChart(statsSeries, trendSeries, t, layout.CardWidth)
			if goals := goalProgressLines(m.entries, t, layout.CardWidth-4); goals != "" {
				content = content + "\n\n" + goals
			}
			boxes = append(boxes, renderCard(t.Name, palette().Primary, content, layout.CardWidth, layout.CardHeight))

		case models.TrackerRating:
			statsSeries := db.NumericSeries(m.entries, t.ID)
			trendSeries, _ := db.NumericByDate(m.entries, t.ID, anchor, limit)
			content := renderRatingCard(statsSeries, trendSeries, t, layout.CardWidth)
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
	today := time.Now().Format("2006-01-02")
	hasToday := db.HasEntryForToday(m.entries)

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

	status := "Recorded"
	if !hasToday {
		status = "Missing"
	}

	return fmt.Sprintf(
		"Today (%s): %s\nLatest entry: %s\nCoverage (latest): %.0f%% (%d/%d)\nActive streaks: %d\nTargets hit: %d/%d (last 30 logged)",
		today, status, latest.Date, coverage, recorded, totalTrackers, activeStreaks, targetHits, targetTotal,
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
		maxLen := 0
		for _, id := range trackerIDs {
			n := len(db.NumericSeries(m.entries, id))
			if n > maxLen {
				maxLen = n
			}
		}
		need := 14 - maxLen
		if need < 0 {
			need = 0
		}
		return fmt.Sprintf("Need %d more numeric entries (have %d / 14).", need, maxLen)
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
	var numericTrackers []models.Tracker
	lookup := map[string]models.Tracker{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerDuration && t.Type != models.TrackerCount && t.Type != models.TrackerNumeric {
				continue
			}
			numericTrackers = append(numericTrackers, t)
			lookup[t.ID] = t
		}
	}
	rows := db.TopPersonalBests(m.entries, numericTrackers, 3)
	if len(rows) == 0 {
		return "No numeric personal bests yet."
	}
	var lines []string
	for i, r := range rows {
		t := lookup[r.TrackerID]
		lines = append(lines, fmt.Sprintf("%d. %s\n   %s  (%s)",
			i+1, truncate(t.Name, 22), formatValueWithUnit(r.Value, t), r.Date))
	}
	return strings.Join(lines, "\n")
}

// overviewNonNegotiables renders one line per configured non-negotiable group,
// summing minutes across all Duration trackers in the group's categories and
// checking any required binary gates. Returns "" if no groups are configured.
func (m Model) overviewNonNegotiables(width int) string {
	if m.config == nil || len(m.config.NonNegotiables) == 0 {
		return ""
	}
	p := palette()
	today := time.Now().Format("2006-01-02")
	var todayEntry *models.Entry
	for i := range m.entries {
		if m.entries[i].Date == today {
			todayEntry = &m.entries[i]
			break
		}
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	success := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Bold(true)
	danger := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger))

	var out []string
	for _, g := range m.config.NonNegotiables {
		todayMin := db.SumCategoryMinutesOn(m.entries, m.config, g.Categories, today)
		binariesOK := true
		var binaryParts []string
		for _, name := range g.RequiredTrackerNames {
			ok := findBinaryTrue(todayEntry, m.config, g.Categories, name)
			if !ok {
				binariesOK = false
			}
			mark := danger.Render("✗")
			if ok {
				mark = success.Render("✓")
			}
			short := strings.ToLower(strings.Fields(name)[0])
			binaryParts = append(binaryParts, fmt.Sprintf("%s %s", short, mark))
		}
		dailyMin := 0.0
		if g.DailyMinMinutes != nil {
			dailyMin = *g.DailyMinMinutes
		}
		dailyMax := dailyMin
		if g.DailyMaxMinutes != nil {
			dailyMax = *g.DailyMaxMinutes
		}
		status := danger.Render("◦")
		if todayMin >= dailyMin && binariesOK {
			status = success.Render("✓")
		}
		summary := fmt.Sprintf("%s / %s min today", formatFloatValue(todayMin), formatFloatValue(dailyMax))
		if g.WeeklyMinutes != nil {
			wk := db.WeeklyCategoryMinutes(m.entries, m.config, g.Categories, time.Time{})
			summary += muted.Render(fmt.Sprintf("  (wk %s/%s)",
				formatFloatValue(wk), formatFloatValue(*g.WeeklyMinutes)))
		}
		if len(binaryParts) > 0 {
			summary += "  " + strings.Join(binaryParts, " ")
		}
		line := fmt.Sprintf("%s %-12s %s", status, truncate(g.Label, 12), summary)
		if g.WeeklyMinutes != nil && *g.WeeklyMinutes > 0 {
			wk := db.WeeklyCategoryMinutes(m.entries, m.config, g.Categories, time.Time{})
			pct := (wk / *g.WeeklyMinutes) * 100
			barW := width - lipgloss.Width(line) - 2
			if barW < 8 {
				barW = 8
			}
			line += " " + ProgressBar(pct, barW)
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func findBinaryTrue(entry *models.Entry, cfg *models.Config, categoryNames []string, trackerName string) bool {
	if entry == nil || cfg == nil {
		return false
	}
	allowed := map[string]bool{}
	for _, n := range categoryNames {
		allowed[strings.ToLower(n)] = true
	}
	for _, c := range cfg.Categories {
		if !allowed[strings.ToLower(c.Name)] {
			continue
		}
		for _, t := range c.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			if !strings.EqualFold(t.Name, trackerName) {
				continue
			}
			if v, ok := entry.Data[t.ID].(bool); ok && v {
				return true
			}
		}
	}
	return false
}

func (m Model) overviewStreaks() string {
	p := palette()
	type row struct {
		name          string
		current, best int
	}
	var rows []row
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			cur := db.CurrentStreak(m.entries, t.ID)
			best := db.LongestStreak(m.entries, t.ID)
			if cur == 0 && best == 0 {
				continue
			}
			rows = append(rows, row{t.Name, cur, best})
		}
	}
	if len(rows) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("Log a binary tracker to start a streak.")
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].current > rows[j].current })
	if len(rows) > 4 {
		rows = rows[:4]
	}

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Bold(true)
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	var out []string
	for _, r := range rows {
		milestone := ""
		switch {
		case r.current >= 100:
			milestone = "  ★ 100!"
		case r.current >= 60:
			milestone = "  ★ 60"
		case r.current >= 30:
			milestone = "  ★ 30"
		case r.current >= 14:
			milestone = "  · 14"
		case r.current >= 7:
			milestone = "  · 7"
		}
		line := fmt.Sprintf("%-18s %s%s",
			truncate(r.name, 18),
			activeStyle.Render(fmt.Sprintf("%3d", r.current)),
			muted.Render(fmt.Sprintf("d  best %d", r.best)),
		)
		if milestone != "" {
			line += activeStyle.Render(milestone)
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func (m Model) overviewLastWeek() string {
	if len(m.entries) == 0 {
		return "No data yet."
	}
	p := palette()
	anchor := time.Now()
	var lines []string
	shown := 0
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if shown >= 6 {
				break
			}
			switch t.Type {
			case models.TrackerBinary:
				strip := LastWeekStrip(db.BinaryByDate(m.entries, t.ID, anchor, 7))
				lines = append(lines, fmt.Sprintf("%-20s %s", truncate(t.Name, 20), strip))
				shown++
			case models.TrackerDuration, models.TrackerNumeric, models.TrackerCount, models.TrackerRating:
				values, present := db.NumericByDate(m.entries, t.ID, anchor, 7)
				strip := NumericLastWeekStrip(values, present, p.ChartPrimary)
				lines = append(lines, fmt.Sprintf("%-20s %s", truncate(t.Name, 20), strip))
				shown++
			}
		}
	}
	if len(lines) == 0 {
		return "No trackers configured."
	}
	return strings.Join(lines, "\n")
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
	// Show the binary tracker with the richest data so the card isn't empty
	// when the first-declared tracker has no history.
	best := binTrackers[0]
	bestCount := 0
	for _, t := range binTrackers {
		n, _ := db.BinaryStats(m.entries, t.ID)
		if n > bestCount {
			bestCount = n
			best = t
		}
	}
	return fmt.Sprintf("%s\n\n%s", truncate(best.Name, 20), WeekdayConsistencyBars(db.BinaryWeekdayConsistency(m.entries, best.ID)))
}

// insightBestWeekday picks the binary tracker with the highest overall
// consistency and reports its weekday breakdown — mixing unrelated habits
// into one average would be noise, not signal.
func (m Model) insightBestWeekday() string {
	var bestTracker *models.Tracker
	bestPct := -1.0
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			tracker := t
			pct := db.ConsistencyPct(m.entries, t.ID)
			if pct > bestPct {
				bestPct = pct
				bestTracker = &tracker
			}
		}
	}
	if bestTracker == nil {
		return "No binary trackers available."
	}
	w := db.BinaryWeekdayConsistency(m.entries, bestTracker.ID)
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	bestIdx := 0
	for i := 1; i < 7; i++ {
		if w[i] > w[bestIdx] {
			bestIdx = i
		}
	}
	return fmt.Sprintf("%s\nBest: %s (%.0f%%)\n\n%s",
		truncate(bestTracker.Name, 24), days[bestIdx], w[bestIdx], WeekdayConsistencyBars(w))
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

	var numericTrackers []models.Tracker
	var binaryTrackers []models.Tracker
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			switch t.Type {
			case models.TrackerNumeric, models.TrackerDuration, models.TrackerCount, models.TrackerRating:
				numericTrackers = append(numericTrackers, t)
			case models.TrackerBinary:
				binaryTrackers = append(binaryTrackers, t)
			}
		}
	}

	scatterStr := m.insightStrongestCorrelation(numericTrackers)
	compStr := m.insightStrongestABImpact(binaryTrackers, numericTrackers)
	weekdayStr := m.insightBestWeekday()
	momentumStr := m.insightMomentumLeaders()

	return renderCardGrid([]string{
		renderCard("A/B Impact", palette().Primary, compStr, layout.CardWidth, layout.CardHeight),
		renderCard("Strongest Correlation", palette().Primary, scatterStr, layout.CardWidth, layout.CardHeight),
		renderCard("Best Day of Week", palette().Primary, weekdayStr, layout.CardWidth, layout.CardHeight),
		renderCard("Momentum Leaders/Laggers", palette().Primary, momentumStr, layout.CardWidth, layout.CardHeight),
	}, m.width)
}

// insightStrongestCorrelation scans every numeric pair and returns the
// strongest Pearson r with a scatter plot and r readout.
func (m Model) insightStrongestCorrelation(numericTrackers []models.Tracker) string {
	if len(numericTrackers) < 2 {
		return "Need two or more numeric trackers."
	}
	type pair struct {
		x, y     models.Tracker
		xs, ys   []float64
		r        float64
		okCorrel bool
	}
	var best *pair
	for i := 0; i < len(numericTrackers); i++ {
		for j := i + 1; j < len(numericTrackers); j++ {
			var xs, ys []float64
			for _, e := range m.entries {
				xv, okX := e.Data[numericTrackers[i].ID].(float64)
				yv, okY := e.Data[numericTrackers[j].ID].(float64)
				if okX && okY {
					xs = append(xs, xv)
					ys = append(ys, yv)
				}
			}
			if len(xs) < 3 {
				continue
			}
			r, ok := db.PearsonCorrelation(xs, ys)
			if !ok {
				continue
			}
			cand := pair{numericTrackers[i], numericTrackers[j], xs, ys, r, true}
			if best == nil || mathAbs(r) > mathAbs(best.r) {
				c := cand
				best = &c
			}
		}
	}
	if best == nil {
		return "Not enough overlapping numeric data yet."
	}
	plot := ScatterPlot(best.xs, best.ys, best.x.Name, best.y.Name)
	readout := CorrelationReadout(best.r, len(best.xs), best.x.Name, best.y.Name)
	return readout + "\n\n" + plot
}

// insightStrongestABImpact picks the binary × numeric pair with the largest
// mean-difference between "done" and "skipped" groups.
func (m Model) insightStrongestABImpact(binaryTrackers, numericTrackers []models.Tracker) string {
	if len(binaryTrackers) == 0 || len(numericTrackers) == 0 {
		return "Need one binary and one numeric tracker."
	}
	type candidate struct {
		bin, num models.Tracker
		yes, no  []float64
		delta    float64
	}
	var best *candidate
	for _, b := range binaryTrackers {
		for _, n := range numericTrackers {
			var yes, no []float64
			for _, e := range m.entries {
				numv, okN := e.Data[n.ID].(float64)
				binv, okB := e.Data[b.ID].(bool)
				if !okN || !okB {
					continue
				}
				if binv {
					yes = append(yes, numv)
				} else {
					no = append(no, numv)
				}
			}
			if len(yes) < 1 || len(no) < 1 {
				continue
			}
			ay := meanFloat(yes)
			an := meanFloat(no)
			d := mathAbs(ay - an)
			if best == nil || d > best.delta {
				c := candidate{b, n, yes, no, d}
				best = &c
			}
		}
	}
	if best == nil {
		return "Not enough A/B data yet."
	}
	return fmt.Sprintf("%s → %s\n\n%s",
		truncate(best.bin.Name, 24), truncate(best.num.Name, 24),
		ABCompareRow("Done", best.yes, "Skipped", best.no, best.num.Name))
}

func mathAbs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func meanFloat(xs []float64) float64 {
	if len(xs) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range xs {
		s += x
	}
	return s / float64(len(xs))
}
