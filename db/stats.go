package db

import (
	"time"

	"dailytrack/models"
)

// BinaryStats returns (done, total) for a binary tracker across entries.
// Entries missing the tracker ID are skipped (not counted in total).
func BinaryStats(entries []models.Entry, trackerID string) (int, int) {
	done, total := 0, 0
	for _, e := range entries {
		v, ok := e.Data[trackerID]
		if !ok {
			continue
		}
		total++
		if b, ok := v.(bool); ok && b {
			done++
		}
	}
	return done, total
}

// ConsistencyPct returns done/total * 100 for a binary tracker.
// Returns 0 for empty or tracker-absent entries.
func ConsistencyPct(entries []models.Entry, trackerID string) float64 {
	done, total := BinaryStats(entries, trackerID)
	if total == 0 {
		return 0
	}
	return float64(done) / float64(total) * 100
}

// NumericSeries returns a float64 slice oldest-first for numeric/duration/count/rating trackers.
// Entries missing the tracker ID are skipped.
// Input entries must be newest-first (standard DB order).
func NumericSeries(entries []models.Entry, trackerID string) []float64 {
	var result []float64
	// Iterate from oldest (high index) to newest (low index)
	for i := len(entries) - 1; i >= 0; i-- {
		v, ok := entries[i].Data[trackerID]
		if !ok {
			continue
		}
		switch val := v.(type) {
		case float64:
			result = append(result, val)
		case int:
			result = append(result, float64(val))
		}
	}
	return result
}

// BinaryHeatmap returns a bool slice oldest-first for all entries.
// Missing tracker in an entry is treated as false.
// Input entries must be newest-first (standard DB order).
func BinaryHeatmap(entries []models.Entry, trackerID string) []bool {
	n := len(entries)
	result := make([]bool, n)
	for i, e := range entries {
		// Map newest-first index i to oldest-first index (n-1-i)
		revIdx := n - 1 - i
		v, ok := e.Data[trackerID]
		if !ok {
			result[revIdx] = false // explicit: missing tracker treated as false (same as zero value, here for clarity)
			continue
		}
		if b, ok := v.(bool); ok {
			result[revIdx] = b
		}
	}
	return result
}

// CurrentStreak counts consecutive entries (newest first) where tracker = true.
// A missing tracker ID breaks the streak.
func CurrentStreak(entries []models.Entry, trackerID string) int {
	streak := 0
	for _, e := range entries {
		v, ok := e.Data[trackerID]
		if !ok {
			break
		}
		if b, ok := v.(bool); ok && b {
			streak++
		} else {
			break
		}
	}
	return streak
}

// TrackerMomentum returns the average of the most recent window vs the previous
// window and their delta (recent - previous). ok=false means not enough points.
func TrackerMomentum(entries []models.Entry, trackerID string, window int) (recentAvg, prevAvg, delta float64, ok bool) {
	if window <= 0 {
		return 0, 0, 0, false
	}
	series := NumericSeries(entries, trackerID) // oldest-first
	if len(series) < window*2 {
		return 0, 0, 0, false
	}

	recent := series[len(series)-window:]
	prev := series[len(series)-window*2 : len(series)-window]
	recentAvg = averageFloatSlice(recent)
	prevAvg = averageFloatSlice(prev)
	delta = recentAvg - prevAvg
	return recentAvg, prevAvg, delta, true
}

// BinaryWeekdayConsistency returns weekday completion percentages (Sunday=0).
// Missing tracker values are counted as false for that entry date.
func BinaryWeekdayConsistency(entries []models.Entry, trackerID string) [7]float64 {
	var totals [7]int
	var hits [7]int
	var out [7]float64

	for _, e := range entries {
		parsed, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue
		}
		idx := int(parsed.Weekday())
		totals[idx]++
		if v, ok := e.Data[trackerID].(bool); ok && v {
			hits[idx]++
		}
	}

	for i := 0; i < len(out); i++ {
		if totals[i] == 0 {
			continue
		}
		out[i] = float64(hits[i]) / float64(totals[i]) * 100
	}
	return out
}

// TargetHitRate returns hit count/total/pct over up to the most recent limit
// points. limit<=0 uses all available points.
func TargetHitRate(entries []models.Entry, trackerID string, target float64, limit int) (hits, total int, pct float64) {
	series := NumericSeries(entries, trackerID) // oldest-first
	if len(series) == 0 {
		return 0, 0, 0
	}
	if limit > 0 && len(series) > limit {
		series = series[len(series)-limit:]
	}

	for _, v := range series {
		total++
		if v >= target {
			hits++
		}
	}
	if total == 0 {
		return 0, 0, 0
	}
	return hits, total, float64(hits) / float64(total) * 100
}

func averageFloatSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}
