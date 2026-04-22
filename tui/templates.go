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
	"Productivity":    "#00ADD8",
	"Languages":       "#F59E0B",
	"Health":          "#10B981",
	"Personal Care":   "#EC4899",
	"Reflection":      "#8B5CF6",
	"Focus":           "#00ADD8",
	"Fitness":         "#10B981",
	"Mood":            "#8B5CF6",
	"Sleep":           "#6366F1",
	"Study":           "#F59E0B",
	"Non-Negotiables": "#00ADD8",
	"Coursework":      "#00ADD8",
	"Chinese":         "#EF4444",
	"Japanese":        "#F97316",
	"Programming":     "#06B6D4",
	"Art":             "#EC4899",
	"Game Dev":        "#8B5CF6",
	"Content":         "#EAB308",
	"Optional":        "#64748B",
}

func ptrF(v float64) *float64 { return &v }

func trackerWithUnit(name string, t models.TrackerType, unit string) models.Tracker {
	tr := models.NewTracker(name, t)
	tr.Unit = unit
	return tr
}

func hoursTracker(name string, weekly, monthly float64) models.Tracker {
	tr := trackerWithUnit(name, models.TrackerDuration, "hours")
	if weekly > 0 {
		tr.WeeklyTarget = ptrF(weekly)
	}
	if monthly > 0 {
		tr.MonthlyTarget = ptrF(monthly)
	}
	return tr
}

// TrackerPack is a named preset the user can adopt whole-cloth. Build
// returns fresh Categories (with fresh UUIDs) each invocation.
type TrackerPack struct {
	Name        string
	Description string
	Build       func() []models.Category
}

// Packs is the ordered set of presets offered in the setup wizard and in
// settings → tracking → "Apply a preset pack".
var Packs = []TrackerPack{
	{
		Name:        "Power",
		Description: "Deep AJATT-style language + programming + art + gamedev framework with per-area hour targets.",
		Build:       buildPowerPack,
	},
	{
		Name:        "Focus",
		Description: "Deep work time + the daily boolean habits that support it.",
		Build: func() []models.Category {
			focus := models.NewCategory("Focus", categoryColors["Focus"])
			focus.Trackers = []models.Tracker{
				models.NewTracker("Deep Work", models.TrackerDuration),
				models.NewTracker("Coding", models.TrackerBinary),
				models.NewTracker("Writing", models.TrackerBinary),
				models.NewTracker("Reading", models.TrackerBinary),
				models.NewTracker("Distractions", models.TrackerCount),
			}
			reflection := buildReflectionCategory()
			return []models.Category{focus, reflection}
		},
	},
	{
		Name:        "Fitness",
		Description: "Workouts, cardio minutes, weight, and diet adherence.",
		Build: func() []models.Category {
			fit := models.NewCategory("Fitness", categoryColors["Fitness"])
			fit.Trackers = []models.Tracker{
				models.NewTracker("Weightlifting", models.TrackerBinary),
				models.NewTracker("Cardio", models.TrackerDuration),
				models.NewTracker("Steps", models.TrackerNumeric),
				models.NewTracker("Diet On Track", models.TrackerBinary),
				models.NewTracker("Weight", models.TrackerNumeric),
			}
			return []models.Category{fit, buildReflectionCategory()}
		},
	},
	{
		Name:        "Mood",
		Description: "Day rating, energy, sleep hours, and short reflection prompts.",
		Build: func() []models.Category {
			mood := models.NewCategory("Mood", categoryColors["Mood"])
			mood.Trackers = []models.Tracker{
				models.NewTracker("Day Rating", models.TrackerRating),
				models.NewTracker("Energy", models.TrackerRating),
				models.NewTracker("Anxiety", models.TrackerRating),
				models.NewTracker("Gratitude", models.TrackerText),
				models.NewTracker("Highlight", models.TrackerText),
			}
			return []models.Category{mood}
		},
	},
	{
		Name:        "Sleep",
		Description: "Hours slept, in-bed time, sleep quality, and wake mood.",
		Build: func() []models.Category {
			sleep := models.NewCategory("Sleep", categoryColors["Sleep"])
			sleep.Trackers = []models.Tracker{
				models.NewTracker("Hours Slept", models.TrackerDuration),
				models.NewTracker("Lights Out", models.TrackerDuration),
				models.NewTracker("Sleep Quality", models.TrackerRating),
				models.NewTracker("Morning Mood", models.TrackerRating),
				models.NewTracker("Caffeine After 2pm", models.TrackerBinary),
			}
			return []models.Category{sleep}
		},
	},
	{
		Name:        "Student",
		Description: "Study minutes, Anki, and coursework — plus reflection.",
		Build: func() []models.Category {
			study := models.NewCategory("Study", categoryColors["Study"])
			study.Trackers = []models.Tracker{
				models.NewTracker("Study Time", models.TrackerDuration),
				models.NewTracker("Anki", models.TrackerBinary),
				models.NewTracker("Coursework", models.TrackerBinary),
				models.NewTracker("Classes Attended", models.TrackerCount),
			}
			return []models.Category{study, buildReflectionCategory()}
		},
	},
}

// PackByName finds a pack by case-insensitive display name.
func PackByName(name string) *TrackerPack {
	for i := range Packs {
		if Packs[i].Name == name {
			return &Packs[i]
		}
	}
	return nil
}

// ApplyPack appends a pack's categories to cfg, skipping any whose name
// already exists. Returns the number of categories added.
func ApplyPack(cfg *models.Config, pack TrackerPack) int {
	if cfg == nil {
		return 0
	}
	existing := map[string]bool{}
	for _, c := range cfg.Categories {
		existing[c.Name] = true
	}
	added := 0
	for _, cat := range pack.Build() {
		if existing[cat.Name] {
			continue
		}
		cat.Order = len(cfg.Categories)
		cfg.Categories = append(cfg.Categories, cat)
		added++
	}
	return added
}

func minutesTracker(name string) models.Tracker {
	return trackerWithUnit(name, models.TrackerDuration, "minutes")
}

// powerPack is the materialized Power preset: categories and non-negotiables
// groups built together so names stay in sync. Exposed via buildPowerPack +
// powerNonNegotiables so cmd/init.go can wire both into the config.
func powerPack() ([]models.Category, []models.NonNegotiableGroup) {
	reflection := models.NewCategory("Reflection", categoryColors["Reflection"])
	reflection.Trackers = []models.Tracker{
		models.NewTracker("Day Rating", models.TrackerRating),
		models.NewTracker("Main Win", models.TrackerText),
		models.NewTracker("Main Blocker", models.TrackerText),
		models.NewTracker("Tomorrow Top 3", models.TrackerText),
		models.NewTracker("Notes", models.TrackerText),
	}

	coursework := models.NewCategory("Coursework", categoryColors["Coursework"])
	coursework.DailyMinMinutes = ptrF(120)
	coursework.DailyMaxMinutes = ptrF(180)
	coursework.WeeklyMinutes = ptrF(1020)
	coursework.MonthlyMinutes = ptrF(4200)
	coursework.Trackers = []models.Tracker{
		minutesTracker("Class / Lectures"),
		minutesTracker("Homework / Assignments"),
		minutesTracker("Reading"),
		minutesTracker("Notes / Review"),
	}

	chinese := models.NewCategory("Chinese", categoryColors["Chinese"])
	chinese.DailyMinMinutes = ptrF(165)
	chinese.DailyMaxMinutes = ptrF(260)
	chinese.WeeklyMinutes = ptrF(1400)
	chinese.Trackers = []models.Tracker{
		minutesTracker("CI / Input"),
		minutesTracker("Rewatch"),
		minutesTracker("Active Listening"),
		trackerWithUnit("Anki New Cards", models.TrackerCount, "cards"),
		minutesTracker("Anki Reviews"),
		minutesTracker("Shadowing / Chorusing"),
	}

	japanese := models.NewCategory("Japanese", categoryColors["Japanese"])
	japanese.DailyMinMinutes = ptrF(130)
	japanese.DailyMaxMinutes = ptrF(220)
	japanese.WeeklyMinutes = ptrF(1100)
	japanese.Trackers = []models.Tracker{
		minutesTracker("CI / Input"),
		minutesTracker("Rewatch"),
		minutesTracker("Active Listening"),
		trackerWithUnit("Anki New Cards", models.TrackerCount, "cards"),
		minutesTracker("Anki Reviews"),
		minutesTracker("Shadowing / Chorusing"),
		minutesTracker("Class Homework"),
		minutesTracker("Class Review"),
	}

	programming := models.NewCategory("Programming", categoryColors["Programming"])
	programming.DailyMinMinutes = ptrF(120)
	programming.DailyMaxMinutes = ptrF(180)
	programming.WeeklyMinutes = ptrF(1020)
	programming.Trackers = []models.Tracker{
		minutesTracker("C++ Study"),
		minutesTracker("DS&A / LeetCode"),
		minutesTracker("Real Coding Reps"),
		minutesTracker("Engine / Gameplay Work"),
		minutesTracker("Graphics / Shaders"),
		minutesTracker("Python Tooling"),
		minutesTracker("Refactor / Cleanup"),
	}

	art := models.NewCategory("Art", categoryColors["Art"])
	art.DailyMinMinutes = ptrF(60)
	art.DailyMaxMinutes = ptrF(120)
	art.WeeklyMinutes = ptrF(600)
	art.Trackers = []models.Tracker{
		minutesTracker("Drawing Fundamentals"),
		minutesTracker("Visual Study"),
		minutesTracker("Pixel / Sprite Work"),
		minutesTracker("UI / Icon / Asset Work"),
		minutesTracker("Concept / Ideation"),
		minutesTracker("Animation Study"),
		minutesTracker("Blender Work"),
	}

	gamedev := models.NewCategory("Game Dev", categoryColors["Game Dev"])
	gamedev.DailyMinMinutes = ptrF(90)
	gamedev.DailyMaxMinutes = ptrF(150)
	gamedev.WeeklyMinutes = ptrF(840)
	gamedev.Trackers = []models.Tracker{
		minutesTracker("Feature / System Moved"),
		minutesTracker("Code + Visuals Integrated"),
		minutesTracker("Polish / Feel Pass"),
		minutesTracker("Animation / VFX / Feedback"),
		minutesTracker("Testing / Bug Fixing"),
		models.NewTracker("Kept Playable", models.TrackerBinary),
		models.NewTracker("Scoped Task Discipline", models.TrackerBinary),
	}

	content := models.NewCategory("Content", categoryColors["Content"])
	content.Trackers = []models.Tracker{
		minutesTracker("Editing Time"),
		models.NewTracker("Progress Clip Captured", models.TrackerBinary),
		models.NewTracker("Dev / Learning Note", models.TrackerBinary),
	}

	health := models.NewCategory("Health", categoryColors["Health"])
	health.DailyMinMinutes = ptrF(45)
	health.DailyMaxMinutes = ptrF(90)
	health.Trackers = []models.Tracker{
		minutesTracker("Training Minutes"),
		models.NewTracker("Diet On Track", models.TrackerBinary),
		models.NewTracker("Skin / Hygiene", models.TrackerBinary),
		trackerWithUnit("Weight", models.TrackerNumeric, "lb"),
	}

	optional := models.NewCategory("Optional", categoryColors["Optional"])
	optional.Trackers = []models.Tracker{
		minutesTracker("Audio / Music Experiment"),
		minutesTracker("VFX / Shader Experiment"),
		minutesTracker("Repo Cleanup / Organization"),
	}

	cats := []models.Category{
		reflection, coursework, chinese, japanese,
		programming, art, gamedev, content, health, optional,
	}

	groups := []models.NonNegotiableGroup{
		{Label: "Coursework", Categories: []string{"Coursework"},
			DailyMinMinutes: ptrF(120), DailyMaxMinutes: ptrF(180), WeeklyMinutes: ptrF(1020)},
		{Label: "Programming", Categories: []string{"Programming"},
			DailyMinMinutes: ptrF(120), DailyMaxMinutes: ptrF(180), WeeklyMinutes: ptrF(1020)},
		{Label: "Art", Categories: []string{"Art"},
			DailyMinMinutes: ptrF(60), DailyMaxMinutes: ptrF(120), WeeklyMinutes: ptrF(600)},
		{Label: "Game Dev", Categories: []string{"Game Dev"},
			DailyMinMinutes: ptrF(90), DailyMaxMinutes: ptrF(150), WeeklyMinutes: ptrF(840)},
		{Label: "Languages", Categories: []string{"Chinese", "Japanese"},
			DailyMinMinutes: ptrF(195), DailyMaxMinutes: ptrF(300), WeeklyMinutes: ptrF(1700)},
		{Label: "Health", Categories: []string{"Health"},
			DailyMinMinutes: ptrF(45), DailyMaxMinutes: ptrF(90),
			RequiredTrackerNames: []string{"Diet On Track", "Skin / Hygiene"}},
	}
	return cats, groups
}

func buildPowerPack() []models.Category {
	cats, _ := powerPack()
	return cats
}

// PowerNonNegotiables returns the non-negotiable groups shipped with the
// Power pack. cmd/init.go wires these onto Config.NonNegotiables when the
// pack is applied.
func PowerNonNegotiables() []models.NonNegotiableGroup {
	_, groups := powerPack()
	return groups
}

func buildReflectionCategory() models.Category {
	cat := models.NewCategory("Reflection", categoryColors["Reflection"])
	cat.Trackers = []models.Tracker{
		models.NewTracker("Day Rating", models.TrackerRating),
		models.NewTracker("Main Win", models.TrackerText),
		models.NewTracker("Main Blocker", models.TrackerText),
	}
	return cat
}
