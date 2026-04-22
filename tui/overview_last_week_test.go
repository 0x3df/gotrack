package tui

import (
	"strings"
	"testing"
	"time"

	"dailytrack/models"

	"github.com/charmbracelet/x/ansi"
)

func TestOverviewLastWeek_BinaryTrackersPreserveCalendarGaps(t *testing.T) {
	today := time.Now()
	tracker := models.Tracker{ID: "code", Name: "Code", Type: models.TrackerBinary}
	cfg := &models.Config{
		SetupComplete: true,
		Categories: []models.Category{
			{Name: "Focus", Trackers: []models.Tracker{tracker}},
		},
	}
	setActivePalette(cfg)

	m := Model{
		config: cfg,
		entries: []models.Entry{
			{
				Date: today.AddDate(0, 0, -2).Format("2006-01-02"),
				Data: map[string]interface{}{"code": true},
			},
		},
	}

	out := ansi.Strip(m.overviewLastWeek())
	want := strings.Join([]string{
		glyphMissed(),
		glyphMissed(),
		glyphMissed(),
		glyphMissed(),
		glyphDone(),
		glyphMissed(),
		glyphMissed(),
	}, " ")
	if !strings.Contains(out, want) {
		t.Fatalf("overviewLastWeek() should keep the completed day at its actual calendar position\nwant strip: %q\n got: %q", want, out)
	}
}

func TestOverviewLastWeek_NumericTrackersPreserveCalendarGaps(t *testing.T) {
	today := time.Now()
	tracker := models.Tracker{ID: "work", Name: "Work", Type: models.TrackerDuration, Unit: "min"}
	cfg := &models.Config{
		SetupComplete: true,
		Categories: []models.Category{
			{Name: "Focus", Trackers: []models.Tracker{tracker}},
		},
	}
	setActivePalette(cfg)

	m := Model{
		config: cfg,
		entries: []models.Entry{
			{
				Date: today.AddDate(0, 0, -2).Format("2006-01-02"),
				Data: map[string]interface{}{"work": float64(90)},
			},
		},
	}

	out := ansi.Strip(m.overviewLastWeek())
	fields := strings.Fields(out)
	if len(fields) < 8 {
		t.Fatalf("overviewLastWeek() should render a 7-day numeric strip; got %q", out)
	}
	cells := fields[len(fields)-7:]
	if cells[4] == glyphMissed() {
		t.Fatalf("expected the numeric value to appear at the 2-days-ago slot, got %q", out)
	}
	if cells[5] != glyphMissed() || cells[6] != glyphMissed() {
		t.Fatalf("expected yesterday and today to remain empty for missing numeric entries, got %q", out)
	}
}
