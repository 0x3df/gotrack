package db

import "dailytrack/models"

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
