package db

import (
	"testing"

	"dailytrack/models"
)

func TestGetEntriesBetween_OrderAndFilter(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	workspace := t.TempDir()
	if err := SetWorkspacePath(workspace); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{"2026-04-01", "2026-04-15", "2026-04-30"} {
		if err := UpsertEntry(&models.Entry{Date: d, Data: map[string]interface{}{"x": 1.0}}); err != nil {
			t.Fatal(err)
		}
	}

	got, err := GetEntriesBetween("2026-04-10", "2026-04-20")
	if err != nil {
		t.Fatalf("GetEntriesBetween: %v", err)
	}
	if len(got) != 1 || got[0].Date != "2026-04-15" {
		t.Fatalf("range filter = %#v, want one row 2026-04-15", got)
	}

	all, err := GetEntriesBetween("2026-04-01", "2026-04-30")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("len = %d, want 3", len(all))
	}
	// Newest first
	if all[0].Date != "2026-04-30" || all[2].Date != "2026-04-01" {
		t.Fatalf("order = %v %v %v, want newest first", all[0].Date, all[1].Date, all[2].Date)
	}
}
