package mcpconn

import (
	"testing"
	"time"

	"dailytrack/models"
)

func TestBuildInsights_Binary(t *testing.T) {
	cfg := &models.Config{SetupComplete: true, Categories: []models.Category{
		{Name: "Life", Trackers: []models.Tracker{
			models.NewTracker("Exercise", models.TrackerBinary),
		}},
	}}
	tid := cfg.Categories[0].Trackers[0].ID
	entries := []models.Entry{
		{Date: "2026-04-02", Data: map[string]interface{}{tid: true}},
		{Date: "2026-04-01", Data: map[string]interface{}{tid: false}},
	}
	out, err := buildInsights(cfg, entries, "Exercise", 7, 30)
	if err != nil {
		t.Fatal(err)
	}
	if out.Binary == nil || out.Binary.Total != 2 || out.Binary.Done != 1 {
		t.Fatalf("binary stats = %+v", out.Binary)
	}
}

func TestBuildInsights_NumericMomentum(t *testing.T) {
	cfg := &models.Config{SetupComplete: true, Categories: []models.Category{
		{Name: "Health", Trackers: []models.Tracker{
			func() models.Tracker {
				tr := models.NewTracker("Weight", models.TrackerNumeric)
				target := 180.0
				tr.Target = &target
				return tr
			}(),
		}},
	}}
	tid := cfg.Categories[0].Trackers[0].ID
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var entries []models.Entry
	for i := 0; i < 20; i++ {
		d := base.AddDate(0, 0, i).Format("2006-01-02")
		entries = append(entries, models.Entry{
			Date: d,
			Data: map[string]interface{}{tid: float64(170 + i)},
		})
	}
	// newest first (same order as [db.GetAllEntries])
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	out, err := buildInsights(cfg, entries, "Weight", 3, 100)
	if err != nil {
		t.Fatal(err)
	}
	if out.Numeric == nil || !out.Numeric.MomentumOK {
		t.Fatalf("want momentum, got %+v", out.Numeric)
	}
	if out.Numeric.TargetHits == nil || *out.Numeric.TargetTotal == 0 {
		t.Fatal("expected target stats")
	}
}
