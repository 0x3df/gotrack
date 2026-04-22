package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"dailytrack/db"
	"dailytrack/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// setupDoneMsg is sent when the setup wizard completes.
type setupDoneMsg struct {
	cfg *models.Config
}

type setupCanceledMsg struct{}

type setupPhase int

const (
	phaseWelcome        setupPhase = iota
	phaseDefaultAreas              // which broad areas to track
	phaseDefaultLangs              // which languages (if Languages selected)
	phaseDefaultPick               // toggle specific trackers per area
	phaseCustomInput               // custom: enter category names
	phaseCustomTrackers            // custom: build trackers for each category
	phaseTargets                   // ask for goals/targets
	phaseAppPrefs                  // app settings: theme/obsidian/background
	phaseDone
)

// setupWiz runs the first-launch setup wizard as a sub-model.
type setupWiz struct {
	phase setupPhase
	form  *huh.Form

	// form binding targets
	workspace    string
	mode         string
	areas        []string
	languages    []string
	prodPicks    []string
	healthPicks  []string
	carePicks    []string
	customCatRaw string // newline-separated category names
	targetUnits  map[string]*string
	targets      map[string]*string
	appSettings  appSettingsDraft

	// internal state
	tempConfig          *models.Config
	customCatIdx        int
	customTrackerName   string
	customTrackerType   models.TrackerType
	customTrackerUnit   string
	customTrackerTarget string
	customAddAnother    bool
	notice              string
	abortMsg            tea.Msg
}

func newSetupWiz() *setupWiz {
	return newAbortableSetupWiz(nil)
}

func newAbortableSetupWiz(abortMsg tea.Msg) *setupWiz {
	w := &setupWiz{
		workspace:         "~/.gotrack",
		targetUnits:       make(map[string]*string),
		targets:           make(map[string]*string),
		customTrackerType: models.TrackerBinary,
		appSettings: appSettingsDraft{
			Theme: models.ThemeGoTrack,
		},
		abortMsg: abortMsg,
	}
	w.buildForm()
	return w
}

func (w *setupWiz) buildForm() {
	switch w.phase {
	case phaseWelcome:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Data Directory").
				Description("Where should GoTrack store your database and config?").
				Value(&w.workspace),
			huh.NewSelect[string]().
				Title("Setup Mode").
				Description("Choose your setup mode").
				Options(
					huh.NewOption("Default — guided setup with sensible defaults", "default"),
					huh.NewOption("Custom — build your own system from scratch", "custom"),
				).
				Value(&w.mode),
		))

	case phaseDefaultAreas:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("What areas do you want to track?").
				Options(
					huh.NewOption("Productivity", "Productivity"),
					huh.NewOption("Languages", "Languages"),
					huh.NewOption("Health", "Health"),
					huh.NewOption("Personal Care", "Personal Care"),
					huh.NewOption("Reflection (day rating, journaling)", "Reflection"),
				).
				Value(&w.areas),
		))

	case phaseDefaultLangs:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Which languages are you studying?").
				Options(
					huh.NewOption("Chinese", "Chinese"),
					huh.NewOption("Japanese", "Japanese"),
					huh.NewOption("Spanish", "Spanish"),
					huh.NewOption("French", "French"),
					huh.NewOption("Korean", "Korean"),
					huh.NewOption("German", "German"),
					huh.NewOption("Portuguese", "Portuguese"),
				).
				Value(&w.languages),
		))

	case phaseDefaultPick:
		var groups []*huh.Group
		for _, area := range w.areas {
			switch area {
			case "Productivity":
				var opts []huh.Option[string]
				for _, t := range defaultProductivityTrackers {
					opts = append(opts, huh.NewOption(t.Name, t.Name))
				}
				groups = append(groups, huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Productivity — pick trackers").
						Options(opts...).
						Value(&w.prodPicks),
				))
			case "Health":
				var opts []huh.Option[string]
				for _, t := range defaultHealthTrackers {
					opts = append(opts, huh.NewOption(t.Name, t.Name))
				}
				groups = append(groups, huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Health — pick trackers").
						Options(opts...).
						Value(&w.healthPicks),
				))
			case "Personal Care":
				var opts []huh.Option[string]
				for _, t := range defaultCareTrackers {
					opts = append(opts, huh.NewOption(t.Name, t.Name))
				}
				groups = append(groups, huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Personal Care — pick trackers").
						Options(opts...).
						Value(&w.carePicks),
				))
			}
		}
		w.form = huh.NewForm(groups...)

	case phaseCustomInput:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewText().
				Title("Define your categories").
				Description("Enter one category name per line.\nYou can add trackers after setup.").
				Value(&w.customCatRaw).
				Validate(func(s string) error {
					for _, line := range strings.Split(s, "\n") {
						if strings.TrimSpace(line) != "" {
							return nil
						}
					}
					return fmt.Errorf("enter at least one category")
				}),
		))

	case phaseCustomTrackers:
		if w.tempConfig == nil || len(w.tempConfig.Categories) == 0 || w.customCatIdx >= len(w.tempConfig.Categories) {
			w.phase = phaseDone
			return
		}
		cat := w.tempConfig.Categories[w.customCatIdx]
		if !w.customTrackerType.IsValid() {
			w.customTrackerType = models.TrackerBinary
		}
		w.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Tracker name (%s)", cat.Name)).
					Description("Add one tracker at a time. Each category needs at least one tracker.").
					Value(&w.customTrackerName).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("enter a tracker name")
						}
						return nil
					}),
				huh.NewSelect[models.TrackerType]().
					Title("Tracker type").
					Options(
						huh.NewOption("Binary", models.TrackerBinary),
						huh.NewOption("Duration", models.TrackerDuration),
						huh.NewOption("Count", models.TrackerCount),
						huh.NewOption("Numeric", models.TrackerNumeric),
						huh.NewOption("Rating", models.TrackerRating),
						huh.NewOption("Text", models.TrackerText),
					).
					Value(&w.customTrackerType),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Unit").
					Description("Required for duration, count, and numeric trackers.").
					Value(&w.customTrackerUnit).
					Validate(func(s string) error {
						if trackerNeedsUnit(w.customTrackerType) && strings.TrimSpace(s) == "" {
							return fmt.Errorf("enter a unit")
						}
						return nil
					}),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.customTrackerType) }),
			huh.NewGroup(
				huh.NewInput().
					Title("Target (optional)").
					Description("Enter a numeric target or leave blank for none.").
					Value(&w.customTrackerTarget).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return nil
						}
						if _, err := strconv.ParseFloat(s, 64); err != nil {
							return fmt.Errorf("must be a number")
						}
						return nil
					}),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.customTrackerType) }),
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("Add another tracker to %s?", cat.Name)).
					Value(&w.customAddAnother),
			),
		)

	case phaseTargets:
		w.targetUnits = make(map[string]*string)
		w.targets = make(map[string]*string)
		var fields []huh.Field
		for _, cat := range w.tempConfig.Categories {
			for _, t := range cat.Trackers {
				if t.Type == models.TrackerDuration || t.Type == models.TrackerCount || t.Type == models.TrackerNumeric {
					unit := t.Unit
					target := ""
					if t.Target != nil {
						target = strconv.FormatFloat(*t.Target, 'f', -1, 64)
					}
					w.targetUnits[t.ID] = &unit
					w.targets[t.ID] = &target
					fields = append(fields, huh.NewInput().
						Title(fmt.Sprintf("Unit for %s (%s)", t.Name, cat.Name)).
						Description("Examples: minutes, count, lb, value").
						Value(&unit).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return fmt.Errorf("enter a unit")
							}
							return nil
						}))
					fields = append(fields, huh.NewInput().
						Title(fmt.Sprintf("Target for %s (%s)", t.Name, cat.Name)).
						Description("Leave blank for no target").
						Value(&target).
						Validate(func(s string) error {
							if strings.TrimSpace(s) == "" {
								return nil
							}
							_, err := strconv.ParseFloat(s, 64)
							if err != nil {
								return fmt.Errorf("must be a number")
							}
							return nil
						}))
				}
			}
		}
		if len(fields) == 0 {
			w.phase = phaseAppPrefs
			w.buildForm()
			return
		}
		w.form = huh.NewForm(huh.NewGroup(fields...))
	case phaseAppPrefs:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[models.ThemeName]().
				Title("Theme").
				Description("Pick your default UI theme.").
				Options(themeOptions()...).
				Value(&w.appSettings.Theme),
			huh.NewConfirm().
				Title("Enable Obsidian export").
				Value(&w.appSettings.ObsidianEnabled),
			huh.NewInput().
				Title("Obsidian vault path").
				Description("Required when Obsidian export is enabled.").
				Value(&w.appSettings.ObsidianVault),
			huh.NewInput().
				Title("Obsidian daily folder").
				Description("Optional subfolder inside the vault. Leave blank for vault root.").
				Value(&w.appSettings.ObsidianFolder),
			huh.NewConfirm().
				Title("Enable falling-stars background").
				Value(&w.appSettings.StarfieldEnabled),
		))
	}
}

// advance moves to the next phase after a form completes.
func (w *setupWiz) advance() {
	switch w.phase {
	case phaseWelcome:
		if w.mode == "custom" {
			w.phase = phaseCustomInput
		} else {
			w.phase = phaseDefaultAreas
		}
	case phaseDefaultAreas:
		hasLangs := false
		for _, a := range w.areas {
			if a == "Languages" {
				hasLangs = true
				break
			}
		}
		if hasLangs {
			w.phase = phaseDefaultLangs
		} else if w.needsPicker() {
			w.phase = phaseDefaultPick
		} else {
			w.tempConfig = w.buildConfig()
			w.phase = phaseTargets
		}
	case phaseDefaultLangs:
		if w.needsPicker() {
			w.phase = phaseDefaultPick
		} else {
			w.tempConfig = w.buildConfig()
			w.phase = phaseTargets
		}
	case phaseDefaultPick:
		w.tempConfig = w.buildConfig()
		w.phase = phaseTargets
	case phaseCustomInput:
		w.tempConfig = w.buildCustomConfig()
		w.customCatIdx = 0
		w.resetCustomTrackerDraft()
		w.phase = phaseCustomTrackers
	case phaseCustomTrackers:
		cat := &w.tempConfig.Categories[w.customCatIdx]
		tr := models.NewTracker(strings.TrimSpace(w.customTrackerName), w.customTrackerType)
		tr.Order = len(cat.Trackers)
		applyTrackerDetails(&tr, w.customTrackerUnit, w.customTrackerTarget)
		cat.Trackers = append(cat.Trackers, tr)
		if w.customAddAnother {
			w.resetCustomTrackerDraft()
			break
		}
		w.customCatIdx++
		if w.customCatIdx >= len(w.tempConfig.Categories) {
			w.phase = phaseAppPrefs
			break
		}
		w.resetCustomTrackerDraft()
	case phaseTargets:
		// Apply targets to tempConfig
		for _, cat := range w.tempConfig.Categories {
			for i, t := range cat.Trackers {
				if s, ok := w.targetUnits[t.ID]; ok {
					unit := strings.TrimSpace(*s)
					if unit == "" {
						unit = models.DefaultUnit(t.Name, t.Type)
					}
					cat.Trackers[i].Unit = unit
				}
				if s, ok := w.targets[t.ID]; ok && *s != "" {
					val, _ := strconv.ParseFloat(*s, 64)
					cat.Trackers[i].Target = &val
				} else {
					cat.Trackers[i].Target = nil
				}
			}
		}
		w.phase = phaseAppPrefs
	case phaseAppPrefs:
		if err := applyAppSettings(w.tempConfig, w.appSettings); err != nil {
			w.notice = fmt.Sprintf("App settings blocked: %v", err)
			break
		}
		w.phase = phaseDone
	}
	if w.phase != phaseDone {
		w.buildForm()
	}
}

// needsPicker reports whether any selected area requires a tracker picker form.
func (w *setupWiz) needsPicker() bool {
	for _, a := range w.areas {
		switch a {
		case "Productivity", "Health", "Personal Care":
			return true
		}
	}
	return false
}

// buildConfig assembles a Config from Default Q&A answers.
func (w *setupWiz) buildConfig() *models.Config {
	cfg := &models.Config{SetupComplete: true}
	catOrder := 0

	for _, area := range w.areas {
		switch area {
		case "Productivity":
			cat := models.NewCategory("Productivity", categoryColors["Productivity"])
			cat.Order = catOrder
			catOrder++
			picked := make(map[string]bool)
			for _, p := range w.prodPicks {
				picked[p] = true
			}
			tOrder := 0
			for _, t := range defaultProductivityTrackers {
				if picked[t.Name] {
					tr := models.NewTracker(t.Name, t.Type)
					tr.Order = tOrder
					tOrder++
					cat.Trackers = append(cat.Trackers, tr)
				}
			}
			if len(cat.Trackers) > 0 {
				cfg.Categories = append(cfg.Categories, cat)
			}

		case "Languages":
			for _, lang := range w.languages {
				cat := models.NewCategory(lang, categoryColors["Languages"])
				cat.Order = catOrder
				catOrder++
				for tIdx, t := range defaultLanguageTrackers(lang) {
					t.Order = tIdx
					cat.Trackers = append(cat.Trackers, t)
				}
				cfg.Categories = append(cfg.Categories, cat)
			}

		case "Health":
			cat := models.NewCategory("Health", categoryColors["Health"])
			cat.Order = catOrder
			catOrder++
			picked := make(map[string]bool)
			for _, p := range w.healthPicks {
				picked[p] = true
			}
			tOrder := 0
			for _, t := range defaultHealthTrackers {
				if picked[t.Name] {
					tr := models.NewTracker(t.Name, t.Type)
					tr.Order = tOrder
					tOrder++
					cat.Trackers = append(cat.Trackers, tr)
				}
			}
			if len(cat.Trackers) > 0 {
				cfg.Categories = append(cfg.Categories, cat)
			}

		case "Personal Care":
			cat := models.NewCategory("Personal Care", categoryColors["Personal Care"])
			cat.Order = catOrder
			catOrder++
			picked := make(map[string]bool)
			for _, p := range w.carePicks {
				picked[p] = true
			}
			tOrder := 0
			for _, t := range defaultCareTrackers {
				if picked[t.Name] {
					tr := models.NewTracker(t.Name, t.Type)
					tr.Order = tOrder
					tOrder++
					cat.Trackers = append(cat.Trackers, tr)
				}
			}
			if len(cat.Trackers) > 0 {
				cfg.Categories = append(cfg.Categories, cat)
			}

		case "Reflection":
			cat := models.NewCategory("Reflection", categoryColors["Reflection"])
			cat.Order = catOrder
			catOrder++
			for tIdx, t := range defaultReflectionTrackers {
				tr := models.NewTracker(t.Name, t.Type)
				tr.Order = tIdx
				cat.Trackers = append(cat.Trackers, tr)
			}
			cfg.Categories = append(cfg.Categories, cat)
		}
	}

	return cfg
}

// buildCustomConfig creates a config from category names only.
func (w *setupWiz) buildCustomConfig() *models.Config {
	cfg := &models.Config{SetupComplete: true}
	lines := strings.Split(strings.TrimSpace(w.customCatRaw), "\n")
	order := 0
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name == "" {
			continue
		}
		color := "#00ADD8"
		if c, ok := categoryColors[name]; ok {
			color = c
		}
		cat := models.NewCategory(name, color)
		cat.Order = order
		order++
		cfg.Categories = append(cfg.Categories, cat)
	}
	return cfg
}

func (w *setupWiz) Init() tea.Cmd {
	if w.form != nil {
		return w.form.Init()
	}
	return nil
}

func (w *setupWiz) Update(msg tea.Msg) tea.Cmd {
	if w.form == nil {
		return nil
	}
	form, cmd := w.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		w.form = f
	}
	if w.form.State == huh.StateCompleted {
		w.advance()
		if w.phase == phaseDone {
			cfg := w.tempConfig
			if err := db.SetWorkspacePath(w.workspace); err != nil {
				fmt.Fprintf(os.Stderr, "gotrack: failed to set workspace: %v\n", err)
				return nil
			}
			if err := db.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "gotrack: failed to save config: %v\n", err)
				return nil // don't proceed — user will see setup again on next launch
			}
			if err := db.InitDB(); err != nil {
				fmt.Fprintf(os.Stderr, "gotrack: failed to init db: %v\n", err)
			}
			return func() tea.Msg { return setupDoneMsg{cfg: cfg} }
		}
		if w.form != nil {
			return w.form.Init()
		}
	}
	if w.form.State == huh.StateAborted {
		if w.abortMsg != nil {
			return func() tea.Msg { return w.abortMsg }
		}
		return tea.Quit
	}
	return cmd
}

func (w *setupWiz) View() string {
	if w.form == nil {
		return "Setting up..."
	}
	header := banner + "\n\nFirst-time setup"
	if strings.TrimSpace(w.notice) == "" {
		return header + "\n\n" + w.form.View()
	}
	return header + "\n\n" + w.notice + "\n\n" + w.form.View()
}

func (w *setupWiz) resetCustomTrackerDraft() {
	w.customTrackerName = ""
	w.customTrackerType = models.TrackerBinary
	w.customTrackerUnit = ""
	w.customTrackerTarget = ""
	w.customAddAnother = false
}

func trackerNeedsUnit(t models.TrackerType) bool {
	return t == models.TrackerDuration || t == models.TrackerCount || t == models.TrackerNumeric
}

func applyTrackerDetails(tr *models.Tracker, unitInput, targetInput string) {
	applyTrackerDetailsFull(tr, unitInput, targetInput, "", "")
}

func applyTrackerDetailsFull(tr *models.Tracker, unitInput, dailyTarget, weeklyTarget, monthlyTarget string) {
	unit := strings.TrimSpace(unitInput)
	if trackerNeedsUnit(tr.Type) {
		if unit == "" {
			unit = models.DefaultUnit(tr.Name, tr.Type)
		}
		tr.Unit = unit
	} else {
		tr.Unit = ""
	}

	tr.Target = parseOptionalFloat(dailyTarget, trackerNeedsUnit(tr.Type))
	tr.WeeklyTarget = parseOptionalFloat(weeklyTarget, trackerNeedsUnit(tr.Type))
	tr.MonthlyTarget = parseOptionalFloat(monthlyTarget, trackerNeedsUnit(tr.Type))
}

func parseOptionalFloat(raw string, allowed bool) *float64 {
	raw = strings.TrimSpace(raw)
	if !allowed || raw == "" {
		return nil
	}
	if val, err := strconv.ParseFloat(raw, 64); err == nil {
		return &val
	}
	return nil
}
