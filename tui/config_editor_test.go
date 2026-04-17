package tui

import (
	"testing"

	"dailytrack/models"
)

func TestDeleteTracker_BlocksHistoricalData(t *testing.T) {
	cfg := &models.Config{
		Categories: []models.Category{
			{
				ID: "cat-1",
				Trackers: []models.Tracker{
					{ID: "tracker-1", Name: "Deep Work", Type: models.TrackerDuration, Order: 0},
				},
			},
		},
	}
	entries := []models.Entry{
		{Date: "2026-04-16", Data: map[string]interface{}{"tracker-1": float64(45)}},
	}

	if err := deleteTracker(cfg, entries, "cat-1", "tracker-1"); err == nil {
		t.Fatal("deleteTracker() error = nil, want non-nil for historical data")
	}
}

func TestMoveCategory_ReordersAndRefreshesOrder(t *testing.T) {
	cfg := &models.Config{
		Categories: []models.Category{
			{ID: "a", Name: "A", Order: 0},
			{ID: "b", Name: "B", Order: 1},
			{ID: "c", Name: "C", Order: 2},
		},
	}

	if ok := moveCategory(cfg, "c", -1); !ok {
		t.Fatal("moveCategory() = false, want true")
	}

	if cfg.Categories[1].ID != "c" {
		t.Fatalf("category at index 1 = %q, want %q", cfg.Categories[1].ID, "c")
	}
	if cfg.Categories[1].Order != 1 {
		t.Fatalf("moved category order = %d, want 1", cfg.Categories[1].Order)
	}
}

func TestDeleteCategory_RemovesCategoryWithoutData(t *testing.T) {
	cfg := &models.Config{
		Categories: []models.Category{
			{ID: "cat-1", Name: "One", Order: 0},
			{ID: "cat-2", Name: "Two", Order: 1},
		},
	}

	if err := deleteCategory(cfg, nil, "cat-1"); err != nil {
		t.Fatalf("deleteCategory() error = %v, want nil", err)
	}
	if len(cfg.Categories) != 1 || cfg.Categories[0].ID != "cat-2" {
		t.Fatalf("remaining categories = %#v, want only cat-2", cfg.Categories)
	}
}
