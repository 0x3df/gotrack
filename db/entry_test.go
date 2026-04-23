package db

import (
	"testing"

	"dailytrack/models"
)

func setupEntryTestDB(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	workspace := t.TempDir()
	if err := SetWorkspacePath(workspace); err != nil {
		t.Fatalf("SetWorkspacePath() error = %v", err)
	}
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func TestUpsertEntryLog_MergesOneTrackerValue(t *testing.T) {
	setupEntryTestDB(t)
	deepWork := models.Tracker{ID: "deep-work", Name: "Deep Work", Type: models.TrackerDuration}
	water := models.Tracker{ID: "water", Name: "Water", Type: models.TrackerCount}
	cfg := &models.Config{
		SetupComplete: true,
		Categories: []models.Category{{
			ID:       "work",
			Name:     "Work",
			Trackers: []models.Tracker{deepWork, water},
		}},
	}
	if err := UpsertEntry(&models.Entry{
		Date: "2026-04-23",
		Data: map[string]interface{}{
			deepWork.ID: float64(30),
		},
	}); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	if err := UpsertEntryLog(cfg, "2026-04-23", map[string]interface{}{"Water": "3"}); err != nil {
		t.Fatalf("UpsertEntryLog() error = %v", err)
	}

	got, err := GetEntryForDate("2026-04-23")
	if err != nil {
		t.Fatalf("GetEntryForDate() error = %v", err)
	}
	if got.Data[deepWork.ID] != float64(30) {
		t.Fatalf("deep work = %v, want preserved 30", got.Data[deepWork.ID])
	}
	if got.Data[water.ID] != float64(3) {
		t.Fatalf("water = %v, want 3", got.Data[water.ID])
	}
}

func TestAddDurationToEntry_AddsMinutesToExistingValue(t *testing.T) {
	setupEntryTestDB(t)
	deepWork := models.Tracker{ID: "deep-work", Name: "Deep Work", Type: models.TrackerDuration}
	cfg := &models.Config{
		SetupComplete: true,
		Categories: []models.Category{{
			ID:       "work",
			Name:     "Work",
			Trackers: []models.Tracker{deepWork},
		}},
	}
	if err := UpsertEntry(&models.Entry{
		Date: "2026-04-23",
		Data: map[string]interface{}{
			deepWork.ID: float64(30),
		},
	}); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	if err := AddDurationToEntry(cfg, "2026-04-23", deepWork.ID, 25); err != nil {
		t.Fatalf("AddDurationToEntry() error = %v", err)
	}

	got, err := GetEntryForDate("2026-04-23")
	if err != nil {
		t.Fatalf("GetEntryForDate() error = %v", err)
	}
	if got.Data[deepWork.ID] != float64(55) {
		t.Fatalf("deep work = %v, want 55", got.Data[deepWork.ID])
	}
}
