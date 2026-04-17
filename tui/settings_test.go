package tui

import (
	"strings"
	"testing"

	"dailytrack/models"
)

func TestApplyAppSettings_UpdatesTheme(t *testing.T) {
	cfg := &models.Config{
		App: models.DefaultAppSettings(),
	}

	draft := appSettingsDraft{
		Theme: models.ThemeCatppuccin,
	}
	if err := applyAppSettings(cfg, draft); err != nil {
		t.Fatalf("applyAppSettings() error = %v, want nil", err)
	}

	if cfg.App.Theme != models.ThemeCatppuccin {
		t.Fatalf("cfg.App.Theme = %q, want %q", cfg.App.Theme, models.ThemeCatppuccin)
	}
}

func TestApplyAppSettings_EnablesObsidianWithVault(t *testing.T) {
	cfg := &models.Config{
		App: models.DefaultAppSettings(),
	}

	draft := appSettingsDraft{
		Theme:            models.ThemeNord,
		ObsidianEnabled:  true,
		ObsidianVault:    "/tmp/vault",
		ObsidianFolder:   "daily",
		StarfieldEnabled: true,
	}
	if err := applyAppSettings(cfg, draft); err != nil {
		t.Fatalf("applyAppSettings() error = %v, want nil", err)
	}

	if !cfg.App.Obsidian.Enabled {
		t.Fatal("cfg.App.Obsidian.Enabled = false, want true")
	}
	if cfg.App.Obsidian.VaultPath != "/tmp/vault" {
		t.Fatalf("cfg.App.Obsidian.VaultPath = %q, want %q", cfg.App.Obsidian.VaultPath, "/tmp/vault")
	}
	if cfg.App.Obsidian.DailyFolder != "daily" {
		t.Fatalf("cfg.App.Obsidian.DailyFolder = %q, want %q", cfg.App.Obsidian.DailyFolder, "daily")
	}
	if !cfg.App.Background.StarfieldEnabled {
		t.Fatal("cfg.App.Background.StarfieldEnabled = false, want true")
	}
}

func TestApplyAppSettings_BlocksBlankVaultWhenEnabled(t *testing.T) {
	cfg := &models.Config{
		App: models.DefaultAppSettings(),
	}

	err := applyAppSettings(cfg, appSettingsDraft{
		Theme:           models.ThemeGoTrack,
		ObsidianEnabled: true,
	})
	if err == nil {
		t.Fatal("applyAppSettings() error = nil, want non-nil")
	}
}

func TestSettingsEditTrackerForm_OnlyShowsSelectorAndAction(t *testing.T) {
	cfg := &models.Config{
		Categories: []models.Category{
			{
				ID:   "cat-1",
				Name: "Productivity",
				Trackers: []models.Tracker{
					{ID: "tracker-1", Name: "Deep Work", Type: models.TrackerDuration, Unit: "minutes", Order: 0},
					{ID: "tracker-2", Name: "Main Win", Type: models.TrackerText, Order: 1},
				},
			},
		},
	}

	w := newSettingsWiz(cfg, nil)
	w.phase = settingsPhaseEditTracker
	w.selectedCategory = "cat-1"
	w.selectedTracker = "tracker-1"
	w.buildForm()

	view := initFormView(w.form)
	if !strings.Contains(view, "Tracker") || !strings.Contains(view, "Action") {
		t.Fatalf("edit-tracker form missing selector/action fields\n%s", view)
	}
	if strings.Contains(view, "Tracker name") {
		t.Fatalf("edit-tracker picker should not show editable tracker fields\n%s", view)
	}
}

func TestSettingsAdvance_EditTrackerLoadsSelectedTrackerValues(t *testing.T) {
	cfg := &models.Config{
		Categories: []models.Category{
			{
				ID:   "cat-1",
				Name: "Productivity",
				Trackers: []models.Tracker{
					{ID: "tracker-1", Name: "Deep Work", Type: models.TrackerDuration, Unit: "minutes", Order: 0},
					{ID: "tracker-2", Name: "Main Win", Type: models.TrackerText, Order: 1},
				},
			},
		},
	}

	w := newSettingsWiz(cfg, nil)
	w.phase = settingsPhaseEditTracker
	w.selectedCategory = "cat-1"
	w.selectedTracker = "tracker-2"
	w.trackerAction = "edit"
	w.buildForm()

	w.advance()

	if w.phase != settingsPhaseEditTrackerDetails {
		t.Fatalf("phase after edit selection = %v, want %v", w.phase, settingsPhaseEditTrackerDetails)
	}
	if w.trackerName != "Main Win" {
		t.Fatalf("trackerName = %q, want %q", w.trackerName, "Main Win")
	}
	if w.trackerType != models.TrackerText {
		t.Fatalf("trackerType = %q, want %q", w.trackerType, models.TrackerText)
	}
}
