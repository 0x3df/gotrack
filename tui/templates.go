package tui

import "dailytrack/models"

// defaultLanguageTrackers returns the standard three trackers for a language.
func defaultLanguageTrackers(lang string) []models.Tracker {
	return []models.Tracker{
		models.NewTracker(lang+" Anki", models.TrackerBinary),
		models.NewTracker(lang+" Immersion", models.TrackerDuration),
		models.NewTracker(lang+" Active Study", models.TrackerDuration),
	}
}

// defaultProductivityTrackers lists all available productivity tracker definitions.
var defaultProductivityTrackers = []struct {
	Name string
	Type models.TrackerType
}{
	{"Coursework", models.TrackerBinary},
	{"Coding", models.TrackerBinary},
	{"Real Project Progress", models.TrackerBinary},
	{"Content Creation", models.TrackerBinary},
	{"Deep Work", models.TrackerDuration},
}

// defaultHealthTrackers lists all available health tracker definitions.
var defaultHealthTrackers = []struct {
	Name string
	Type models.TrackerType
}{
	{"Weightlifting", models.TrackerBinary},
	{"Cardio", models.TrackerBinary},
	{"Diet On Track", models.TrackerBinary},
	{"Weight", models.TrackerNumeric},
}

// defaultCareTrackers lists personal care tracker definitions.
var defaultCareTrackers = []struct {
	Name string
	Type models.TrackerType
}{
	{"AM Skincare", models.TrackerBinary},
	{"PM Skincare", models.TrackerBinary},
}

// defaultReflectionTrackers lists reflection tracker definitions (always all included).
var defaultReflectionTrackers = []struct {
	Name string
	Type models.TrackerType
}{
	{"Day Rating", models.TrackerRating},
	{"Main Win", models.TrackerText},
	{"Main Blocker", models.TrackerText},
	{"Tomorrow Top 3", models.TrackerText},
}

// categoryColors maps standard category names to lipgloss color strings.
var categoryColors = map[string]string{
	"Productivity":  "#00ADD8",
	"Languages":     "#F59E0B",
	"Health":        "#10B981",
	"Personal Care": "#EC4899",
	"Reflection":    "#8B5CF6",
}
