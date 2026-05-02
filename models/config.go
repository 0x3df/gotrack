package models

import (
	"strings"

	"github.com/google/uuid"
)

type TrackerType string
type ThemeName string

const (
	TrackerBinary   TrackerType = "binary"
	TrackerDuration TrackerType = "duration" // stored as float64 minutes
	TrackerCount    TrackerType = "count"    // stored as float64
	TrackerNumeric  TrackerType = "numeric"  // stored as float64
	TrackerRating   TrackerType = "rating"   // stored as float64, 1-5
	TrackerText     TrackerType = "text"     // stored as string

	ThemeGoTrack    ThemeName = "gotrack"
	ThemeCatppuccin ThemeName = "catppuccin"
	ThemeNord       ThemeName = "nord"
	ThemeAccessible ThemeName = "accessible"
)

type Tracker struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Type          TrackerType `json:"type"`
	Unit          string      `json:"unit,omitempty"`
	Target        *float64    `json:"target,omitempty"`         // per-day target; nil = none
	WeeklyTarget  *float64    `json:"weekly_target,omitempty"`  // sum over calendar week; nil = none
	MonthlyTarget *float64    `json:"monthly_target,omitempty"` // sum over calendar month; nil = none
	Order         int         `json:"order"`
}

func (t TrackerType) IsValid() bool {
	switch t {
	case TrackerBinary, TrackerDuration, TrackerCount, TrackerNumeric, TrackerRating, TrackerText:
		return true
	}
	return false
}

func (t ThemeName) IsValid() bool {
	switch t {
	case ThemeGoTrack, ThemeCatppuccin, ThemeNord, ThemeAccessible:
		return true
	}
	return false
}

func NewTracker(name string, t TrackerType) Tracker {
	if !t.IsValid() {
		panic("invalid TrackerType: " + string(t))
	}
	return Tracker{
		ID:   uuid.New().String(),
		Name: name,
		Type: t,
		Unit: DefaultUnit(name, t),
	}
}

type Category struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Color    string    `json:"color"` // lipgloss color string
	Order    int       `json:"order"`
	Trackers []Tracker `json:"trackers"`
	// Category-level minute targets. Sum of all Duration (unit=minutes)
	// sub-activities in this category rolls up against these.
	DailyMinMinutes *float64 `json:"daily_min_minutes,omitempty"`
	DailyMaxMinutes *float64 `json:"daily_max_minutes,omitempty"`
	WeeklyMinutes   *float64 `json:"weekly_minutes,omitempty"`
	MonthlyMinutes  *float64 `json:"monthly_minutes,omitempty"`
}

// NonNegotiableGroup defines one row on the Overview Non-Negotiables card.
// Minutes are summed across all Duration trackers (unit=minutes) inside any
// of the listed categories. If RequiredTrackerNames is set, each named binary
// tracker must be true today for the row to show ✓.
type NonNegotiableGroup struct {
	Label                string   `json:"label"`
	Categories           []string `json:"categories"`
	DailyMinMinutes      *float64 `json:"daily_min_minutes,omitempty"`
	DailyMaxMinutes      *float64 `json:"daily_max_minutes,omitempty"`
	WeeklyMinutes        *float64 `json:"weekly_minutes,omitempty"`
	RequiredTrackerNames []string `json:"required_trackers,omitempty"`
}

func NewCategory(name, color string) Category {
	return Category{
		ID:    uuid.New().String(),
		Name:  name,
		Color: color,
	}
}

type Config struct {
	SetupComplete  bool                 `json:"setup_complete"`
	App            AppSettings          `json:"app,omitempty"`
	Categories     []Category           `json:"categories"`
	NonNegotiables []NonNegotiableGroup `json:"non_negotiables,omitempty"`
}

type AppSettings struct {
	Theme      ThemeName          `json:"theme,omitempty"`
	Obsidian   ObsidianSettings   `json:"obsidian,omitempty"`
	Background BackgroundSettings `json:"background,omitempty"`
	BackupCmd  string             `json:"backup_cmd,omitempty"`
}

type ObsidianSettings struct {
	Enabled     bool   `json:"enabled,omitempty"`
	VaultPath   string `json:"vault_path,omitempty"`
	DailyFolder string `json:"daily_folder,omitempty"`
}

type BackgroundSettings struct {
	StarfieldEnabled bool `json:"starfield_enabled,omitempty"`
}

func DefaultAppSettings() AppSettings {
	return AppSettings{
		Theme: ThemeGoTrack,
	}
}

func DefaultUnit(name string, t TrackerType) string {
	switch t {
	case TrackerDuration:
		return "minutes"
	case TrackerCount:
		return "count"
	case TrackerNumeric:
		if strings.EqualFold(strings.TrimSpace(name), "Weight") {
			return "lb"
		}
		return "value"
	default:
		return ""
	}
}

func NormalizeConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	NormalizeAppSettings(&cfg.App)
	for catIdx := range cfg.Categories {
		for trackerIdx := range cfg.Categories[catIdx].Trackers {
			tracker := &cfg.Categories[catIdx].Trackers[trackerIdx]
			if strings.TrimSpace(tracker.Unit) != "" {
				continue
			}
			tracker.Unit = DefaultUnit(tracker.Name, tracker.Type)
		}
	}
}

func NormalizeAppSettings(app *AppSettings) {
	if app == nil {
		return
	}
	if !app.Theme.IsValid() {
		app.Theme = ThemeGoTrack
	}
	app.Obsidian.VaultPath = strings.TrimSpace(app.Obsidian.VaultPath)
	app.Obsidian.DailyFolder = strings.TrimSpace(app.Obsidian.DailyFolder)
}

type Entry struct {
	Date string `json:"date"`
	// Data maps tracker UUID to its value.
	// Values must be exactly: bool (binary), float64 (duration/count/numeric/rating), string (text).
	// Never store int — JSON round-trips will convert it to float64, breaking type assertions.
	Data map[string]interface{} `json:"data"`
}
