package tui

import "dailytrack/models"

type ThemePalette struct {
	Name            models.ThemeName
	Primary         string
	Border          string
	Muted           string
	Success         string
	Danger          string
	ActiveTabBg     string
	ActiveTabFg     string
	InactiveTab     string
	ChartPrimary    string
	ChartSecondary  string
	HeatmapActive   string
	HeatmapInactive string
	StarDim         string
	StarBright      string
	StarTrail       string
}

var activeThemePalette = paletteForTheme(models.ThemeGoTrack)

func paletteForTheme(name models.ThemeName) ThemePalette {
	switch name {
	case models.ThemeCatppuccin:
		return ThemePalette{
			Name:            models.ThemeCatppuccin,
			Primary:         "#74c7ec",
			Border:          "#45475a",
			Muted:           "#a6adc8",
			Success:         "#a6e3a1",
			Danger:          "#f38ba8",
			ActiveTabBg:     "#74c7ec",
			ActiveTabFg:     "#1e1e2e",
			InactiveTab:     "#a6adc8",
			ChartPrimary:    "#74c7ec",
			ChartSecondary:  "#f38ba8",
			HeatmapActive:   "#a6e3a1",
			HeatmapInactive: "#45475a",
			StarDim:         "#6c7086",
			StarBright:      "#f5e0dc",
			StarTrail:       "#89dceb",
		}
	case models.ThemeNord:
		return ThemePalette{
			Name:            models.ThemeNord,
			Primary:         "#88c0d0",
			Border:          "#4c566a",
			Muted:           "#81a1c1",
			Success:         "#a3be8c",
			Danger:          "#bf616a",
			ActiveTabBg:     "#88c0d0",
			ActiveTabFg:     "#2e3440",
			InactiveTab:     "#81a1c1",
			ChartPrimary:    "#88c0d0",
			ChartSecondary:  "#bf616a",
			HeatmapActive:   "#a3be8c",
			HeatmapInactive: "#4c566a",
			StarDim:         "#5e81ac",
			StarBright:      "#eceff4",
			StarTrail:       "#81a1c1",
		}
	default:
		return ThemePalette{
			Name:            models.ThemeGoTrack,
			Primary:         "#00ADD8",
			Border:          "#444444",
			Muted:           "#888888",
			Success:         "#00D855",
			Danger:          "#FF5F87",
			ActiveTabBg:     "#00ADD8",
			ActiveTabFg:     "#000000",
			InactiveTab:     "#888888",
			ChartPrimary:    "#00ADD8",
			ChartSecondary:  "#FF5F87",
			HeatmapActive:   "#00D855",
			HeatmapInactive: "#444444",
			StarDim:         "#6b7280",
			StarBright:      "#ffffff",
			StarTrail:       "#00ADD8",
		}
	}
}

func currentPalette(cfg *models.Config) ThemePalette {
	if cfg == nil {
		return paletteForTheme(models.ThemeGoTrack)
	}
	return paletteForTheme(cfg.App.Theme)
}

func setActivePalette(cfg *models.Config) ThemePalette {
	activeThemePalette = currentPalette(cfg)
	return activeThemePalette
}

func palette() ThemePalette {
	return activeThemePalette
}
