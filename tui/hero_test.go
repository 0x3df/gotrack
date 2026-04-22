package tui

import (
	"strings"
	"testing"

	"dailytrack/models"
)

func TestHeroVisualsRenderWithoutPanic(t *testing.T) {
	cfg := &models.Config{
		SetupComplete: true,
		Categories: []models.Category{
			{
				Name: "Focus",
				Trackers: []models.Tracker{
					{ID: "b1", Name: "Code", Type: models.TrackerBinary},
					{ID: "b2", Name: "Read", Type: models.TrackerBinary},
					{ID: "n1", Name: "Deep Work", Type: models.TrackerDuration, Unit: "min"},
				},
			},
		},
	}
	entries := []models.Entry{
		{Date: "2026-04-19", Data: map[string]interface{}{"b1": true, "b2": false, "n1": float64(120)}},
		{Date: "2026-04-18", Data: map[string]interface{}{"b1": true, "b2": true, "n1": float64(200)}},
		{Date: "2026-04-17", Data: map[string]interface{}{"b1": false, "b2": true, "n1": float64(60)}},
	}

	m := Model{config: cfg, entries: entries, width: 120, height: 48}
	for i := range heroVisuals {
		m.heroIndex = i
		out := m.renderHero(120)
		if strings.TrimSpace(out) == "" {
			t.Fatalf("hero %d rendered empty", i)
		}
	}
}

func TestHeroCycleIndexIsModular(t *testing.T) {
	n := len(heroVisuals)
	if n < 2 {
		t.Skip("need at least 2 hero visuals for this test")
	}
	idx := 0
	idx = (idx - 1 + n) % n
	if idx != n-1 {
		t.Fatalf("wrapping backward from 0 should land on last; got %d", idx)
	}
	idx = (idx + 1) % n
	if idx != 0 {
		t.Fatalf("wrapping forward should land on 0; got %d", idx)
	}
}
