package tui

import (
	"testing"

	"dailytrack/models"
)

func TestPaletteForTheme_KnownThemesNonEmpty(t *testing.T) {
	for _, theme := range []models.ThemeName{models.ThemeGoTrack, models.ThemeCatppuccin, models.ThemeNord} {
		p := paletteForTheme(theme)
		if p.Name != theme {
			t.Fatalf("paletteForTheme(%q).Name = %q, want %q", theme, p.Name, theme)
		}
		if p.Primary == "" || p.Border == "" || p.Muted == "" || p.Success == "" || p.Danger == "" {
			t.Fatalf("paletteForTheme(%q) has empty required colors: %#v", theme, p)
		}
	}
}

func TestCurrentPalette_FallsBackToGoTrack(t *testing.T) {
	cfg := &models.Config{
		App: models.AppSettings{
			Theme: models.ThemeName("mystery"),
		},
	}

	p := currentPalette(cfg)
	if p.Name != models.ThemeGoTrack {
		t.Fatalf("currentPalette(unknown).Name = %q, want %q", p.Name, models.ThemeGoTrack)
	}
}

func TestPaletteForTheme_DiffersAcrossBuiltins(t *testing.T) {
	got := paletteForTheme(models.ThemeGoTrack)
	cat := paletteForTheme(models.ThemeCatppuccin)
	nord := paletteForTheme(models.ThemeNord)

	if got.Primary == cat.Primary || got.Primary == nord.Primary || cat.Primary == nord.Primary {
		t.Fatalf("theme primaries should differ: gotrack=%q catppuccin=%q nord=%q", got.Primary, cat.Primary, nord.Primary)
	}
}
