package tui

import (
	"fmt"
	"strings"

	"dailytrack/models"
)

type appSettingsDraft struct {
	Theme            models.ThemeName
	ObsidianEnabled  bool
	ObsidianVault    string
	ObsidianFolder   string
	StarfieldEnabled bool
}

func applyAppSettings(cfg *models.Config, draft appSettingsDraft) error {
	if cfg == nil {
		return fmt.Errorf("missing config")
	}
	if !draft.Theme.IsValid() {
		draft.Theme = models.ThemeGoTrack
	}
	draft.ObsidianVault = strings.TrimSpace(draft.ObsidianVault)
	draft.ObsidianFolder = strings.TrimSpace(draft.ObsidianFolder)
	if draft.ObsidianEnabled && draft.ObsidianVault == "" {
		return fmt.Errorf("obsidian vault path is required when export is enabled")
	}

	cfg.App.Theme = draft.Theme
	cfg.App.Obsidian.Enabled = draft.ObsidianEnabled
	cfg.App.Obsidian.VaultPath = draft.ObsidianVault
	cfg.App.Obsidian.DailyFolder = draft.ObsidianFolder
	cfg.App.Background.StarfieldEnabled = draft.StarfieldEnabled
	models.NormalizeAppSettings(&cfg.App)
	return nil
}
