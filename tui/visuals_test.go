package tui

import (
	"strings"
	"testing"
)

func TestTrendDeltaStrip_IncludesTrendDirection(t *testing.T) {
	got := TrendDeltaStrip(12, 10)
	if !strings.Contains(got, "Recent: 12.0") {
		t.Fatalf("TrendDeltaStrip() missing recent value: %q", got)
	}
	if !strings.Contains(got, "Prev: 10.0") {
		t.Fatalf("TrendDeltaStrip() missing prev value: %q", got)
	}
	if !strings.Contains(got, "↑") {
		t.Fatalf("TrendDeltaStrip() missing upward indicator: %q", got)
	}
}

func TestTargetHitMeter_FormatsRateAndBar(t *testing.T) {
	got := TargetHitMeter(3, 4, 20)
	if !strings.Contains(got, "3/4 hits (75%)") {
		t.Fatalf("TargetHitMeter() missing summary: %q", got)
	}
	if !strings.Contains(got, "%") || !strings.Contains(got, "[") {
		t.Fatalf("TargetHitMeter() missing bar output: %q", got)
	}
}

func TestWeekdayConsistencyBars_OutputsSevenLines(t *testing.T) {
	weekday := [7]float64{10, 20, 30, 40, 50, 60, 70}
	got := WeekdayConsistencyBars(weekday)
	lines := strings.Split(got, "\n")
	if len(lines) != 7 {
		t.Fatalf("WeekdayConsistencyBars() lines = %d, want 7", len(lines))
	}
	if !strings.HasPrefix(lines[0], "S ") {
		t.Fatalf("first line prefix = %q, want S day prefix", lines[0])
	}
}

func TestDualLineChart_RendersWithTwoSeries(t *testing.T) {
	primary := []float64{1, 2, 3, 4}
	secondary := []float64{4, 3, 2, 1}
	got := DualLineChart(primary, secondary, 20, 4)
	if strings.TrimSpace(got) == "" {
		t.Fatal("DualLineChart() returned empty output")
	}
}

func TestLeaderboardBars_RendersRankedRows(t *testing.T) {
	rows := []LeaderboardRow{
		{Label: "Weight", Delta: -1.4},
		{Label: "Sleep", Delta: 0.8},
	}
	got := LeaderboardBars(rows, 10)
	if !strings.Contains(got, "Weight") || !strings.Contains(got, "Sleep") {
		t.Fatalf("LeaderboardBars() missing labels: %q", got)
	}
	if !strings.Contains(got, "↓") || !strings.Contains(got, "↑") {
		t.Fatalf("LeaderboardBars() missing trend signs: %q", got)
	}
}
