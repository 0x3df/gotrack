package tui

import (
	"fmt"
	"os"
	"strings"

	"dailytrack/db"
	"dailytrack/models"

	"github.com/charmbracelet/huh"
	tea "github.com/charmbracelet/bubbletea"
)

// setupDoneMsg is sent when the setup wizard completes.
type setupDoneMsg struct {
	cfg *models.Config
}

type setupPhase int

const (
	phaseWelcome      setupPhase = iota
	phaseDefaultAreas            // which broad areas to track
	phaseDefaultLangs            // which languages (if Languages selected)
	phaseDefaultPick             // toggle specific trackers per area
	phaseCustomInput             // custom: enter category names
	phaseDone
)

// setupWiz runs the first-launch setup wizard as a sub-model.
type setupWiz struct {
	phase setupPhase
	form  *huh.Form

	// form binding targets
	mode         string
	areas        []string
	languages    []string
	prodPicks    []string
	healthPicks  []string
	carePicks    []string
	customCatRaw string // newline-separated category names
}

func newSetupWiz() *setupWiz {
	w := &setupWiz{}
	w.buildForm()
	return w
}

func (w *setupWiz) buildForm() {
	switch w.phase {
	case phaseWelcome:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Welcome to DailyTrack").
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
				Value(&w.customCatRaw),
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
			w.phase = phaseDone
		}
	case phaseDefaultLangs:
		if w.needsPicker() {
			w.phase = phaseDefaultPick
		} else {
			w.phase = phaseDone
		}
	case phaseDefaultPick:
		w.phase = phaseDone
	case phaseCustomInput:
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
			var cfg *models.Config
			if w.mode == "custom" {
				cfg = w.buildCustomConfig()
			} else {
				cfg = w.buildConfig()
			}
			if err := db.SaveConfig(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "dailytrack: failed to save config: %v\n", err)
				return nil // don't proceed — user will see setup again on next launch
			}
			if err := db.InitDB(); err != nil {
				fmt.Fprintf(os.Stderr, "dailytrack: failed to init db: %v\n", err)
			}
			return func() tea.Msg { return setupDoneMsg{cfg: cfg} }
		}
		if w.form != nil {
			return w.form.Init()
		}
	}
	if w.form.State == huh.StateAborted {
		return tea.Quit
	}
	return cmd
}

func (w *setupWiz) View() string {
	if w.form == nil {
		return "Setting up..."
	}
	return w.form.View()
}
