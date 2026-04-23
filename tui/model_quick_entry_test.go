package tui

import (
	"testing"

	"dailytrack/db"
	"dailytrack/models"
)

func TestQuickEntryTrackers_ReturnsAllConfiguredTrackers(t *testing.T) {
	cfg := quickPomodoroTestConfig()

	got := quickEntryTrackers(cfg)

	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].ID != "deep-work" || got[1].ID != "water" || got[2].ID != "notes" {
		t.Fatalf("trackers = %#v, want config order", got)
	}
}

func TestSaveQuickEntry_LogsOneValueForToday(t *testing.T) {
	setupTUITestDB(t)
	cfg := quickPomodoroTestConfig()
	m := Model{
		config:         cfg,
		quickTrackerID: "water",
		quickValue:     "4",
	}

	if err := m.saveQuickEntry("2026-04-23"); err != nil {
		t.Fatalf("saveQuickEntry() error = %v", err)
	}

	got, err := db.GetEntryForDate("2026-04-23")
	if err != nil {
		t.Fatalf("GetEntryForDate() error = %v", err)
	}
	if got.Data["water"] != float64(4) {
		t.Fatalf("water = %v, want 4", got.Data["water"])
	}
}

func setupTUITestDB(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	workspace := t.TempDir()
	if err := db.SetWorkspacePath(workspace); err != nil {
		t.Fatalf("SetWorkspacePath() error = %v", err)
	}
	if err := db.InitDB(); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
}

func quickPomodoroTestConfig() *models.Config {
	return &models.Config{
		SetupComplete: true,
		Categories: []models.Category{{
			ID:   "work",
			Name: "Work",
			Trackers: []models.Tracker{
				{ID: "deep-work", Name: "Deep Work", Type: models.TrackerDuration},
				{ID: "water", Name: "Water", Type: models.TrackerCount},
				{ID: "notes", Name: "Notes", Type: models.TrackerText},
			},
		}},
	}
}
