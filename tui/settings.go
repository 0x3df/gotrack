package tui

import (
	"fmt"
	"strconv"
	"strings"

	"dailytrack/db"
	"dailytrack/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type settingsDoneMsg struct{}
type settingsRerunSetupMsg struct{}

type settingsPhase int

const (
	settingsPhaseMenu settingsPhase = iota
	settingsPhaseTrackingMenu
	settingsPhaseAddCategory
	settingsPhaseEditCategory
	settingsPhaseAddTracker
	settingsPhaseEditTrackerCategory
	settingsPhaseEditTracker
	settingsPhaseEditTrackerDetails
	settingsPhaseApplyPack
	settingsPhaseAppMenu
)

type settingsWiz struct {
	phase settingsPhase
	form  *huh.Form

	config  *models.Config
	entries []models.Entry

	notice      string
	workspace   string
	pointerPath string

	mainChoice           string
	trackingChoice       string
	appChoice            string
	categoryAction       string
	trackerAction        string
	selectedCategory     string
	selectedTracker      string
	selectedPack         string
	categoryName         string
	trackerName          string
	trackerType          models.TrackerType
	trackerUnit          string
	trackerTarget        string
	trackerWeeklyTarget  string
	trackerMonthlyTarget string
	appSettings          appSettingsDraft
	appAction            string
}

func newSettingsWiz(cfg *models.Config, entries []models.Entry) *settingsWiz {
	workspace, _ := db.GetWorkspacePath()
	ptr, _ := db.GetPointerFilePath()
	w := &settingsWiz{
		phase:          settingsPhaseMenu,
		config:         cfg,
		entries:        entries,
		workspace:      workspace,
		pointerPath:    ptr,
		trackerType:    models.TrackerBinary,
		mainChoice:     "tracking",
		appChoice:      "save",
		trackingChoice: "add-category",
		appSettings: appSettingsDraft{
			Theme:            cfg.App.Theme,
			ObsidianEnabled:  cfg.App.Obsidian.Enabled,
			ObsidianVault:    cfg.App.Obsidian.VaultPath,
			ObsidianFolder:   cfg.App.Obsidian.DailyFolder,
			StarfieldEnabled: cfg.App.Background.StarfieldEnabled,
			BackupCmd:        cfg.App.BackupCmd,
			SyncCmd:          cfg.App.SyncCmd,
		},
	}
	w.buildForm()
	return w
}

func (w *settingsWiz) Init() tea.Cmd {
	if w.form != nil {
		return w.form.Init()
	}
	return nil
}

func (w *settingsWiz) Update(msg tea.Msg) tea.Cmd {
	if w.form == nil {
		return nil
	}
	form, cmd := w.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		w.form = f
	}
	if w.form.State == huh.StateCompleted {
		return w.advance()
	}
	if w.form.State == huh.StateAborted {
		return func() tea.Msg { return settingsDoneMsg{} }
	}
	return cmd
}

func (w *settingsWiz) View() string {
	if w.form == nil {
		return "Loading settings..."
	}
	var parts []string
	if strings.TrimSpace(w.notice) != "" {
		parts = append(parts, w.notice)
	}
	if w.phase == settingsPhaseAppMenu {
		parts = append(parts,
			fmt.Sprintf("Workspace: %s\nPointer file: %s", w.workspace, w.pointerPath))
	}
	parts = append(parts, w.form.View())
	return strings.Join(parts, "\n\n")
}

func (w *settingsWiz) buildForm() {
	switch w.phase {
	case settingsPhaseMenu:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Settings").
				Options(
					huh.NewOption("Tracking Setup", "tracking"),
					huh.NewOption("App", "app"),
					huh.NewOption("Back to Dashboard", "back"),
				).
				Value(&w.mainChoice),
		))

	case settingsPhaseTrackingMenu:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Tracking Setup").
				Options(
					huh.NewOption("Add Category", "add-category"),
					huh.NewOption("Edit Category", "edit-category"),
					huh.NewOption("Add Tracker", "add-tracker"),
					huh.NewOption("Edit Tracker", "edit-tracker"),
					huh.NewOption("Apply Preset Pack", "apply-pack"),
					huh.NewOption("Back", "back"),
				).
				Value(&w.trackingChoice),
		))

	case settingsPhaseAddCategory:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Category name").
				Value(&w.categoryName).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("enter a category name")
					}
					return nil
				}),
		))

	case settingsPhaseEditCategory:
		if len(w.config.Categories) == 0 {
			w.notice = "No categories configured yet."
			w.phase = settingsPhaseTrackingMenu
			w.buildForm()
			return
		}
		if findCategory(w.config, w.selectedCategory) == nil {
			w.selectedCategory = w.config.Categories[0].ID
		}
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Category").
				Options(categoryOptions(w.config)...).
				Value(&w.selectedCategory),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Rename", "rename"),
					huh.NewOption("Move Up", "move-up"),
					huh.NewOption("Move Down", "move-down"),
					huh.NewOption("Delete", "delete"),
				).
				Value(&w.categoryAction),
			huh.NewInput().
				Title("New name (rename only)").
				Value(&w.categoryName),
		))

	case settingsPhaseAddTracker:
		if len(w.config.Categories) == 0 {
			w.notice = "Add a category before adding trackers."
			w.phase = settingsPhaseTrackingMenu
			w.buildForm()
			return
		}
		if findCategory(w.config, w.selectedCategory) == nil {
			w.selectedCategory = w.config.Categories[0].ID
		}
		w.form = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Category").
					Options(categoryOptions(w.config)...).
					Value(&w.selectedCategory),
				huh.NewInput().
					Title("Tracker name").
					Value(&w.trackerName).
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
					Value(&w.trackerType),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Unit").
					Description("Required for duration, count, and numeric trackers.").
					Value(&w.trackerUnit).
					Validate(func(s string) error {
						if trackerNeedsUnit(w.trackerType) && strings.TrimSpace(s) == "" {
							return fmt.Errorf("enter a unit")
						}
						return nil
					}),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.trackerType) }),
			huh.NewGroup(
				huh.NewInput().
					Title("Daily target (optional)").
					Value(&w.trackerTarget).
					Validate(optionalFloatValidator),
				huh.NewInput().
					Title("Weekly target (optional, total across 7 days)").
					Value(&w.trackerWeeklyTarget).
					Validate(optionalFloatValidator),
				huh.NewInput().
					Title("Monthly target (optional, total for the month)").
					Value(&w.trackerMonthlyTarget).
					Validate(optionalFloatValidator),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.trackerType) }),
		)

	case settingsPhaseEditTrackerCategory:
		if len(w.config.Categories) == 0 {
			w.notice = "No categories configured yet."
			w.phase = settingsPhaseTrackingMenu
			w.buildForm()
			return
		}
		if findCategory(w.config, w.selectedCategory) == nil {
			w.selectedCategory = w.config.Categories[0].ID
		}
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose category").
				Options(categoryOptions(w.config)...).
				Value(&w.selectedCategory),
		))

	case settingsPhaseEditTracker:
		cat := findCategory(w.config, w.selectedCategory)
		if cat == nil || len(cat.Trackers) == 0 {
			w.notice = "Selected category has no trackers yet."
			w.phase = settingsPhaseTrackingMenu
			w.buildForm()
			return
		}
		if findTracker(cat, w.selectedTracker) == nil {
			w.selectedTracker = cat.Trackers[0].ID
		}
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Tracker").
				Options(trackerOptions(cat)...).
				Value(&w.selectedTracker),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Edit", "edit"),
					huh.NewOption("Move Up", "move-up"),
					huh.NewOption("Move Down", "move-down"),
					huh.NewOption("Delete", "delete"),
				).
				Value(&w.trackerAction),
		))

	case settingsPhaseEditTrackerDetails:
		cat := findCategory(w.config, w.selectedCategory)
		if cat == nil || findTracker(cat, w.selectedTracker) == nil {
			w.notice = "Tracker not found."
			w.phase = settingsPhaseTrackingMenu
			w.buildForm()
			return
		}
		w.form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Tracker name").
					Value(&w.trackerName).
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
					Value(&w.trackerType),
			),
			huh.NewGroup(
				huh.NewInput().
					Title("Unit").
					Description("Required for duration, count, and numeric trackers.").
					Value(&w.trackerUnit).
					Validate(func(s string) error {
						if trackerNeedsUnit(w.trackerType) && strings.TrimSpace(s) == "" {
							return fmt.Errorf("enter a unit")
						}
						return nil
					}),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.trackerType) }),
			huh.NewGroup(
				huh.NewInput().
					Title("Daily target (optional)").
					Value(&w.trackerTarget).
					Validate(optionalFloatValidator),
				huh.NewInput().
					Title("Weekly target (optional, total across 7 days)").
					Value(&w.trackerWeeklyTarget).
					Validate(optionalFloatValidator),
				huh.NewInput().
					Title("Monthly target (optional, total for the month)").
					Value(&w.trackerMonthlyTarget).
					Validate(optionalFloatValidator),
			).WithHideFunc(func() bool { return !trackerNeedsUnit(w.trackerType) }),
		)

	case settingsPhaseApplyPack:
		var opts []huh.Option[string]
		for _, p := range Packs {
			opts = append(opts, huh.NewOption(fmt.Sprintf("%s — %s", p.Name, p.Description), p.Name))
		}
		if w.selectedPack == "" && len(Packs) > 0 {
			w.selectedPack = Packs[0].Name
		}
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a preset pack to append").
				Description("Categories already in your config will be skipped.").
				Options(opts...).
				Value(&w.selectedPack),
		))

	case settingsPhaseAppMenu:
		w.form = huh.NewForm(huh.NewGroup(
			huh.NewSelect[models.ThemeName]().
				Title("Theme").
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
			huh.NewInput().
				Title("Backup command (optional)").
				Description("Shell command run after every save and on app close. e.g. git -C ~/.gotrack add -A && git -C ~/.gotrack commit -m 'backup' && git -C ~/.gotrack push").
				Value(&w.appSettings.BackupCmd),
			huh.NewInput().
				Title("Sync command (optional)").
				Description("Shell command run on app open. e.g. git -C ~/.gotrack pull").
				Value(&w.appSettings.SyncCmd),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Save app settings", "save"),
					huh.NewOption("Rerun setup", "rerun"),
					huh.NewOption("Back without saving", "back"),
				).
				Value(&w.appChoice),
		))
	}
}

func (w *settingsWiz) advance() tea.Cmd {
	w.notice = ""

	switch w.phase {
	case settingsPhaseMenu:
		switch w.mainChoice {
		case "tracking":
			w.phase = settingsPhaseTrackingMenu
		case "app":
			w.phase = settingsPhaseAppMenu
		default:
			return func() tea.Msg { return settingsDoneMsg{} }
		}
	case settingsPhaseTrackingMenu:
		switch w.trackingChoice {
		case "add-category":
			w.phase = settingsPhaseAddCategory
			w.categoryName = ""
		case "edit-category":
			w.phase = settingsPhaseEditCategory
			w.categoryAction = "rename"
			w.categoryName = ""
		case "add-tracker":
			w.phase = settingsPhaseAddTracker
			w.resetTrackerDraft()
		case "edit-tracker":
			w.phase = settingsPhaseEditTrackerCategory
		case "apply-pack":
			w.phase = settingsPhaseApplyPack
		default:
			w.phase = settingsPhaseMenu
		}
	case settingsPhaseAddCategory:
		cat := models.NewCategory(strings.TrimSpace(w.categoryName), categoryColorForName(w.categoryName))
		cat.Order = len(w.config.Categories)
		w.config.Categories = append(w.config.Categories, cat)
		refreshCategoryOrders(w.config)
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
		} else {
			w.notice = fmt.Sprintf("Added category %q.", cat.Name)
		}
		w.phase = settingsPhaseTrackingMenu
	case settingsPhaseEditCategory:
		w.applyCategoryAction()
		w.phase = settingsPhaseTrackingMenu
	case settingsPhaseAddTracker:
		w.addTracker()
		w.phase = settingsPhaseTrackingMenu
	case settingsPhaseEditTrackerCategory:
		cat := findCategory(w.config, w.selectedCategory)
		if cat != nil && len(cat.Trackers) > 0 && findTracker(cat, w.selectedTracker) == nil {
			w.selectedTracker = cat.Trackers[0].ID
		}
		w.trackerAction = "edit"
		w.phase = settingsPhaseEditTracker
	case settingsPhaseEditTracker:
		if w.trackerAction == "edit" {
			w.loadSelectedTracker()
			w.phase = settingsPhaseEditTrackerDetails
		} else {
			w.applyTrackerAction()
			w.phase = settingsPhaseTrackingMenu
		}
	case settingsPhaseEditTrackerDetails:
		w.applyTrackerAction()
		w.phase = settingsPhaseTrackingMenu
	case settingsPhaseApplyPack:
		pack := PackByName(w.selectedPack)
		if pack == nil {
			w.notice = "Pack not found."
		} else {
			added := ApplyPack(w.config, *pack)
			refreshCategoryOrders(w.config)
			if added == 0 {
				w.notice = fmt.Sprintf("Pack %q — nothing added (categories already exist).", pack.Name)
			} else if err := db.SaveConfig(w.config); err != nil {
				w.notice = fmt.Sprintf("Failed to save config: %v", err)
			} else {
				w.notice = fmt.Sprintf("Applied %q — added %d categor%s.", pack.Name, added, plural(added, "y", "ies"))
			}
		}
		w.phase = settingsPhaseTrackingMenu
	case settingsPhaseAppMenu:
		switch w.appChoice {
		case "rerun":
			return func() tea.Msg { return settingsRerunSetupMsg{} }
		case "save":
			if err := applyAppSettings(w.config, w.appSettings); err != nil {
				w.notice = fmt.Sprintf("App settings blocked: %v", err)
				w.buildForm()
				if w.form != nil {
					return w.form.Init()
				}
				return nil
			}
			if err := db.SaveConfig(w.config); err != nil {
				w.notice = fmt.Sprintf("Failed to save config: %v", err)
				w.buildForm()
				if w.form != nil {
					return w.form.Init()
				}
				return nil
			}
			w.workspace, _ = db.GetWorkspacePath()
			w.pointerPath, _ = db.GetPointerFilePath()
			w.notice = "App settings saved."
			db.LogEvent("config_saved", "app settings")
			runBackupCmd(w.config)
		}
		w.phase = settingsPhaseMenu
	}

	w.buildForm()
	if w.form != nil {
		return w.form.Init()
	}
	return nil
}

func (w *settingsWiz) applyCategoryAction() {
	cat := findCategory(w.config, w.selectedCategory)
	if cat == nil {
		w.notice = "Category not found."
		return
	}
	switch w.categoryAction {
	case "rename":
		name := strings.TrimSpace(w.categoryName)
		if name == "" {
			w.notice = "Enter a new category name to rename it."
			return
		}
		oldName := cat.Name
		renameLanguageTemplateTrackers(cat, oldName, name)
		cat.Name = name
		cat.Color = categoryColorForName(name)
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = fmt.Sprintf("Renamed category to %q.", name)
	case "move-up":
		if !moveCategory(w.config, w.selectedCategory, -1) {
			w.notice = "Category cannot move up."
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Category moved up."
	case "move-down":
		if !moveCategory(w.config, w.selectedCategory, 1) {
			w.notice = "Category cannot move down."
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Category moved down."
	case "delete":
		if err := deleteCategory(w.config, w.entries, w.selectedCategory); err != nil {
			w.notice = fmt.Sprintf("Delete blocked: %v", err)
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Category deleted."
	}
}

func (w *settingsWiz) addTracker() {
	cat := findCategory(w.config, w.selectedCategory)
	if cat == nil {
		w.notice = "Category not found."
		return
	}
	tr := models.NewTracker(strings.TrimSpace(w.trackerName), w.trackerType)
	tr.Order = len(cat.Trackers)
	applyTrackerDetailsFull(&tr, w.trackerUnit, w.trackerTarget, w.trackerWeeklyTarget, w.trackerMonthlyTarget)
	cat.Trackers = append(cat.Trackers, tr)
	if err := db.SaveConfig(w.config); err != nil {
		w.notice = fmt.Sprintf("Failed to save config: %v", err)
		return
	}
	w.notice = fmt.Sprintf("Added tracker %q.", tr.Name)
}

func (w *settingsWiz) applyTrackerAction() {
	cat := findCategory(w.config, w.selectedCategory)
	if cat == nil {
		w.notice = "Category not found."
		return
	}
	tr := findTracker(cat, w.selectedTracker)
	if tr == nil {
		w.notice = "Tracker not found."
		return
	}

	switch w.trackerAction {
	case "edit":
		hasHistory := trackerHasData(w.entries, tr.ID)
		name := strings.TrimSpace(w.trackerName)
		if name == "" {
			w.notice = "Tracker name cannot be blank."
			return
		}
		if hasHistory && w.trackerType != tr.Type {
			w.notice = "Type changes are blocked for trackers with historical data."
			return
		}
		tr.Name = name
		if !hasHistory {
			tr.Type = w.trackerType
		}
		applyTrackerDetailsFull(tr, w.trackerUnit, w.trackerTarget, w.trackerWeeklyTarget, w.trackerMonthlyTarget)
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		if hasHistory {
			w.notice = "Tracker updated. Type remained locked because data exists."
		} else {
			w.notice = "Tracker updated."
		}
	case "move-up":
		if !moveTracker(w.config, w.selectedCategory, w.selectedTracker, -1) {
			w.notice = "Tracker cannot move up."
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Tracker moved up."
	case "move-down":
		if !moveTracker(w.config, w.selectedCategory, w.selectedTracker, 1) {
			w.notice = "Tracker cannot move down."
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Tracker moved down."
	case "delete":
		if err := deleteTracker(w.config, w.entries, w.selectedCategory, w.selectedTracker); err != nil {
			w.notice = fmt.Sprintf("Delete blocked: %v", err)
			return
		}
		if err := db.SaveConfig(w.config); err != nil {
			w.notice = fmt.Sprintf("Failed to save config: %v", err)
			return
		}
		w.notice = "Tracker deleted."
	}
}

func (w *settingsWiz) loadSelectedTracker() {
	cat := findCategory(w.config, w.selectedCategory)
	tr := findTracker(cat, w.selectedTracker)
	if tr == nil {
		return
	}
	w.trackerName = tr.Name
	w.trackerType = tr.Type
	w.trackerUnit = tr.Unit
	w.trackerTarget = ""
	if tr.Target != nil {
		w.trackerTarget = formatFloatValue(*tr.Target)
	}
	w.trackerWeeklyTarget = ""
	if tr.WeeklyTarget != nil {
		w.trackerWeeklyTarget = formatFloatValue(*tr.WeeklyTarget)
	}
	w.trackerMonthlyTarget = ""
	if tr.MonthlyTarget != nil {
		w.trackerMonthlyTarget = formatFloatValue(*tr.MonthlyTarget)
	}
}

func (w *settingsWiz) resetTrackerDraft() {
	w.trackerName = ""
	w.trackerType = models.TrackerBinary
	w.trackerUnit = ""
	w.trackerTarget = ""
	w.trackerWeeklyTarget = ""
	w.trackerMonthlyTarget = ""
	w.trackerAction = "edit"
}

func plural(n int, singular, many string) string {
	if n == 1 {
		return singular
	}
	return many
}

func optionalFloatValidator(s string) error {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return fmt.Errorf("must be a number")
	}
	return nil
}

func categoryOptions(cfg *models.Config) []huh.Option[string] {
	var opts []huh.Option[string]
	for _, cat := range cfg.Categories {
		opts = append(opts, huh.NewOption(cat.Name, cat.ID))
	}
	return opts
}

func themeOptions() []huh.Option[models.ThemeName] {
	return []huh.Option[models.ThemeName]{
		huh.NewOption("GoTrack", models.ThemeGoTrack),
		huh.NewOption("Catppuccin", models.ThemeCatppuccin),
		huh.NewOption("Nord", models.ThemeNord),
		huh.NewOption("Accessible (ASCII, high contrast)", models.ThemeAccessible),
	}
}

func trackerOptions(cat *models.Category) []huh.Option[string] {
	var opts []huh.Option[string]
	for _, tracker := range cat.Trackers {
		opts = append(opts, huh.NewOption(tracker.Name, tracker.ID))
	}
	return opts
}

func categoryColorForName(name string) string {
	if color, ok := categoryColors[strings.TrimSpace(name)]; ok {
		return color
	}
	return palette().Primary
}

func renameLanguageTemplateTrackers(cat *models.Category, oldName, newName string) {
	if cat == nil {
		return
	}
	oldName = strings.TrimSpace(oldName)
	newName = strings.TrimSpace(newName)
	if oldName == "" || newName == "" || oldName == newName {
		return
	}

	replacements := map[string]string{
		oldName + " Anki":         newName + " Anki",
		oldName + " Immersion":    newName + " Immersion",
		oldName + " Active Study": newName + " Active Study",
	}

	for i := range cat.Trackers {
		if renamed, ok := replacements[cat.Trackers[i].Name]; ok {
			cat.Trackers[i].Name = renamed
		}
	}
}
