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
