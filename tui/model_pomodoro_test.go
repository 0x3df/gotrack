package tui

import (
	"testing"
	"time"

	"dailytrack/db"
	"dailytrack/models"
)

func TestDurationTrackers_ReturnsOnlyDurationTrackers(t *testing.T) {
	cfg := quickPomodoroTestConfig()

	got := durationTrackers(cfg)

	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].ID != "deep-work" {
		t.Fatalf("tracker = %#v, want deep-work", got[0])
	}
}

func TestCompletePomodoro_AllocatesElapsedMinutes(t *testing.T) {
	setupTUITestDB(t)
	cfg := quickPomodoroTestConfig()
	if err := db.UpsertEntry(&models.Entry{
		Date: "2026-04-23",
		Data: map[string]interface{}{"deep-work": float64(30)},
	}); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}
	m := Model{
		config:            cfg,
		pomodoroTrackerID: "deep-work",
		pomodoroStarted:   time.Date(2026, 4, 23, 12, 0, 0, 0, time.UTC),
	}

	if err := m.completePomodoro("2026-04-23", time.Date(2026, 4, 23, 12, 25, 0, 0, time.UTC)); err != nil {
		t.Fatalf("completePomodoro() error = %v", err)
	}

	got, err := db.GetEntryForDate("2026-04-23")
	if err != nil {
		t.Fatalf("GetEntryForDate() error = %v", err)
	}
	if got.Data["deep-work"] != float64(55) {
		t.Fatalf("deep-work = %v, want 55", got.Data["deep-work"])
	}
}
