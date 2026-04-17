package tui

import (
	"math/rand"
	"testing"

	"dailytrack/models"
)

func TestStarfieldTick_DoesNotAnimateOutsideDashboard(t *testing.T) {
	m := Model{
		state: stateSetup,
		setup: newSetupWiz(),
		config: &models.Config{
			App: models.AppSettings{
				Background: models.BackgroundSettings{StarfieldEnabled: true},
			},
		},
		stars:    []fallingStar{{X: 3, Y: 3, VX: 1, VY: 1, Trail: 2}},
		twinkles: []twinkleStar{{X: 1, Y: 1, Phase: 1}},
		bursts:   []burstParticle{{X: 2, Y: 2, VX: 1, VY: 0, Life: 2}},
	}

	got, _ := m.Update(starfieldTickMsg{})
	next := got.(Model)
	if next.stars[0].X != 3 || next.stars[0].Y != 3 {
		t.Fatalf("stars moved outside dashboard state: %+v", next.stars[0])
	}
	if next.twinkles[0].Phase != 1 {
		t.Fatalf("twinkle phase changed outside dashboard state: %d", next.twinkles[0].Phase)
	}
	if next.bursts[0].Life != 2 {
		t.Fatalf("burst changed outside dashboard state: %+v", next.bursts[0])
	}
}

func TestTriggerDashboardCelebration_OnTargetHit(t *testing.T) {
	target := 60.0
	m := Model{
		width:  80,
		height: 24,
		rng:    rand.New(rand.NewSource(3)),
		config: &models.Config{
			Categories: []models.Category{
				{
					Trackers: []models.Tracker{
						{ID: "deep-work", Name: "Deep Work", Type: models.TrackerDuration, Target: &target},
					},
				},
			},
		},
		entries: []models.Entry{
			{Date: "2026-05-01", Data: map[string]interface{}{"deep-work": float64(75)}},
		},
	}

	m.triggerDashboardCelebration()
	if m.pulseTick == 0 {
		t.Fatal("pulseTick should be set on target hit")
	}
	if len(m.bursts) == 0 {
		t.Fatal("burst particles should spawn on target hit")
	}
}
