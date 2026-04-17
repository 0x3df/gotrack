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
