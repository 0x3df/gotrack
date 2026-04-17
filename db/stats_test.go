package db

import (
	"testing"

	"dailytrack/models"
)

func makeEntry(date string, data map[string]interface{}) models.Entry {
	return models.Entry{Date: date, Data: data}
}

func TestBinaryStats_Basic(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-03", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": false}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	done, total := BinaryStats(entries, "abc")
	if done != 2 {
		t.Errorf("want done=2, got %d", done)
	}
	if total != 3 {
		t.Errorf("want total=3, got %d", total)
	}
}

func TestBinaryStats_SkipsMissing(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-02", map[string]interface{}{}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	done, total := BinaryStats(entries, "abc")
	if total != 1 {
		t.Errorf("want total=1 (skip missing), got %d", total)
	}
	if done != 1 {
		t.Errorf("want done=1, got %d", done)
	}
}

func TestConsistencyPct_Basic(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-02", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": false}),
	}
	pct := ConsistencyPct(entries, "abc")
	if pct != 50.0 {
		t.Errorf("want 50.0, got %f", pct)
	}
}

func TestConsistencyPct_Empty(t *testing.T) {
	pct := ConsistencyPct(nil, "abc")
	if pct != 0.0 {
		t.Errorf("want 0.0 for nil entries, got %f", pct)
	}
}

func TestNumericSeries_OldestFirst(t *testing.T) {
	// entries arrive newest-first; NumericSeries must return oldest-first
	entries := []models.Entry{
		makeEntry("2026-01-03", map[string]interface{}{"abc": float64(30)}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": float64(60)}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": float64(45)}),
	}
	series := NumericSeries(entries, "abc")
	if len(series) != 3 {
		t.Fatalf("want 3 points, got %d", len(series))
	}
	if series[0] != 45 {
		t.Errorf("want oldest first: 45, got %f", series[0])
	}
	if series[2] != 30 {
		t.Errorf("want newest last: 30, got %f", series[2])
	}
}

func TestNumericSeries_SkipsMissing(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-02", map[string]interface{}{}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": float64(10)}),
	}
	series := NumericSeries(entries, "abc")
	if len(series) != 1 {
		t.Errorf("want 1 (skip missing), got %d", len(series))
	}
}

func TestBinaryHeatmap_Order(t *testing.T) {
	// entries newest-first; heatmap must be oldest-first
	entries := []models.Entry{
		makeEntry("2026-01-03", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": false}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	heat := BinaryHeatmap(entries, "abc")
	if len(heat) != 3 {
		t.Fatalf("want 3, got %d", len(heat))
	}
	if !heat[0] {
		t.Error("want heat[0]=true (oldest: 2026-01-01)")
	}
	if heat[1] {
		t.Error("want heat[1]=false (2026-01-02)")
	}
	if !heat[2] {
		t.Error("want heat[2]=true (newest: 2026-01-03)")
	}
}

func TestBinaryHeatmap_MissingIsFalse(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-02", map[string]interface{}{}), // missing tracker
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	heat := BinaryHeatmap(entries, "abc")
	if len(heat) != 2 {
		t.Fatalf("want 2, got %d", len(heat))
	}
	if !heat[0] {
		t.Error("want heat[0]=true (oldest)")
	}
	if heat[1] {
		t.Error("want heat[1]=false (missing = false)")
	}
}

func TestCurrentStreak_Basic(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-04", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-03", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": false}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	streak := CurrentStreak(entries, "abc")
	if streak != 2 {
		t.Errorf("want streak=2, got %d", streak)
	}
}

func TestCurrentStreak_Zero(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-01", map[string]interface{}{"abc": false}),
	}
	streak := CurrentStreak(entries, "abc")
	if streak != 0 {
		t.Errorf("want 0, got %d", streak)
	}
}

func TestCurrentStreak_MissingBreaksStreak(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-03", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-02", map[string]interface{}{}), // missing = breaks streak
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	streak := CurrentStreak(entries, "abc")
	if streak != 1 {
		t.Errorf("want 1 (missing breaks streak), got %d", streak)
	}
}

func TestBinaryStats_Empty(t *testing.T) {
	done, total := BinaryStats(nil, "abc")
	if done != 0 || total != 0 {
		t.Errorf("want 0,0 for nil entries, got %d,%d", done, total)
	}
}

func TestConsistencyPct_AllTrackerAbsent(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-02", map[string]interface{}{"other": true}),
		makeEntry("2026-01-01", map[string]interface{}{"other": false}),
	}
	pct := ConsistencyPct(entries, "abc")
	if pct != 0.0 {
		t.Errorf("want 0.0 when tracker absent in all entries, got %f", pct)
	}
}

func TestCurrentStreak_Empty(t *testing.T) {
	streak := CurrentStreak(nil, "abc")
	if streak != 0 {
		t.Errorf("want 0 for nil entries, got %d", streak)
	}
}

func TestCurrentStreak_AllTrue(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-03", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": true}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": true}),
	}
	streak := CurrentStreak(entries, "abc")
	if streak != 3 {
		t.Errorf("want streak=3, got %d", streak)
	}
}

func TestNumericSeries_Empty(t *testing.T) {
	series := NumericSeries(nil, "abc")
	if len(series) != 0 {
		t.Errorf("want empty slice for nil entries, got %d", len(series))
	}
}

func TestTrackerMomentum_Basic(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-06", map[string]interface{}{"abc": float64(6)}),
		makeEntry("2026-01-05", map[string]interface{}{"abc": float64(5)}),
		makeEntry("2026-01-04", map[string]interface{}{"abc": float64(4)}),
		makeEntry("2026-01-03", map[string]interface{}{"abc": float64(3)}),
		makeEntry("2026-01-02", map[string]interface{}{"abc": float64(2)}),
		makeEntry("2026-01-01", map[string]interface{}{"abc": float64(1)}),
	}
	recentAvg, prevAvg, delta, ok := TrackerMomentum(entries, "abc", 3)
	if !ok {
		t.Fatal("TrackerMomentum() ok = false, want true")
	}
	if recentAvg != 5 {
		t.Fatalf("recentAvg = %v, want 5", recentAvg)
	}
	if prevAvg != 2 {
		t.Fatalf("prevAvg = %v, want 2", prevAvg)
	}
	if delta != 3 {
		t.Fatalf("delta = %v, want 3", delta)
	}
}

func TestBinaryWeekdayConsistency(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-04", map[string]interface{}{"abc": true}),  // Sunday
		makeEntry("2026-01-11", map[string]interface{}{"abc": false}), // Sunday
		makeEntry("2026-01-05", map[string]interface{}{"abc": true}),  // Monday
		makeEntry("2026-01-12", map[string]interface{}{"abc": true}),  // Monday
	}
	weekday := BinaryWeekdayConsistency(entries, "abc")
	if weekday[0] != 50 {
		t.Fatalf("Sunday pct = %v, want 50", weekday[0])
	}
	if weekday[1] != 100 {
		t.Fatalf("Monday pct = %v, want 100", weekday[1])
	}
	if weekday[2] != 0 {
		t.Fatalf("Tuesday pct = %v, want 0", weekday[2])
	}
}

func TestTargetHitRate_WithLimit(t *testing.T) {
	entries := []models.Entry{
		makeEntry("2026-01-06", map[string]interface{}{"abc": float64(30)}),
		makeEntry("2026-01-05", map[string]interface{}{"abc": float64(20)}),
		makeEntry("2026-01-04", map[string]interface{}{"abc": float64(10)}),
		makeEntry("2026-01-03", map[string]interface{}{"abc": float64(40)}),
	}
	hits, total, pct := TargetHitRate(entries, "abc", 25, 3)
	if hits != 1 || total != 3 {
		t.Fatalf("hits/total = %d/%d, want 1/3", hits, total)
	}
	if pct < 33.3 || pct > 33.4 {
		t.Fatalf("pct = %.3f, want approx 33.333", pct)
	}
}
