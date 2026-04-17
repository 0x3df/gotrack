package tui

import (
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
