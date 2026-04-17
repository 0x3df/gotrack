package tui

import (
	"strings"
	"testing"

	"dailytrack/models"

	"github.com/charmbracelet/huh"
)

func TestSetupCustomTrackerForm_HidesUnitForNonUnitTypes(t *testing.T) {
	w := newSetupWiz()
	w.phase = phaseCustomTrackers
	w.tempConfig = &models.Config{
		Categories: []models.Category{
			{ID: "cat-1", Name: "Productivity"},
		},
	}
	w.customCatIdx = 0
	w.customTrackerType = models.TrackerBinary
	w.buildForm()

	form := batchUpdateForm(w.form, w.form.Init()).(*huh.Form)
	view := stripFormView(batchUpdateForm(form, form.NextGroup()).(*huh.Form))
	if strings.Contains(view, "Unit") || !strings.Contains(view, "Add another tracker") {
		t.Fatalf("custom tracker form should skip unit input for non-unit tracker types\n%s", view)
	}
}

func TestSetupCustomTrackerForm_ShowsUnitForUnitTypes(t *testing.T) {
	w := newSetupWiz()
	w.phase = phaseCustomTrackers
	w.tempConfig = &models.Config{
		Categories: []models.Category{
			{ID: "cat-1", Name: "Productivity"},
		},
	}
	w.customCatIdx = 0
	w.customTrackerType = models.TrackerDuration
	w.buildForm()

	form := batchUpdateForm(w.form, w.form.Init()).(*huh.Form)
	view := stripFormView(batchUpdateForm(form, form.NextGroup()).(*huh.Form))
	if !strings.Contains(view, "Unit") {
		t.Fatalf("custom tracker form should show unit input for duration/count/numeric trackers\n%s", view)
	}
}

func TestSetupTargetsAdvanceToAppPrefs(t *testing.T) {
	w := newSetupWiz()
	w.phase = phaseTargets
	w.tempConfig = &models.Config{
		SetupComplete: true,
		Categories: []models.Category{
			{
				Name: "Productivity",
				Trackers: []models.Tracker{
					{ID: "duration-1", Name: "Deep Work", Type: models.TrackerDuration, Unit: "minutes"},
				},
			},
		},
	}
	w.targetUnits = map[string]*string{"duration-1": ptrString("minutes")}
	w.targets = map[string]*string{"duration-1": ptrString("60")}

	w.advance()

	if w.phase != phaseAppPrefs {
		t.Fatalf("phase after target step = %v, want %v", w.phase, phaseAppPrefs)
	}
	if w.form == nil {
		t.Fatal("setup app prefs form was not built")
	}
}

func TestSetupView_IncludesBanner(t *testing.T) {
	w := newSetupWiz()
	view := w.View()
	if !strings.Contains(view, "______") {
		t.Fatalf("setup view missing GoTrack banner\n%s", view)
	}
	if !strings.Contains(view, "First-time setup") {
		t.Fatalf("setup view missing first-time setup heading\n%s", view)
	}
}

func ptrString(v string) *string {
	return &v
}
