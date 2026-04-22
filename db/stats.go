package db

import (
	"math"
	"sort"
	"strings"
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

// RollingAverageSeries returns a trailing-window moving average for each point
// in the oldest-first series. It returns nil if window <= 0.
func RollingAverageSeries(series []float64, window int) []float64 {
	if window <= 0 {
		return nil
	}
	if len(series) == 0 {
		return []float64{}
	}
	out := make([]float64, len(series))
	sum := 0.0
	for i := 0; i < len(series); i++ {
		sum += series[i]
		if i >= window {
			sum -= series[i-window]
		}
		denom := i + 1
		if denom > window {
			denom = window
		}
		out[i] = sum / float64(denom)
	}
	return out
}

// PersonalBest finds the highest numeric value and its date.
func PersonalBest(entries []models.Entry, trackerID string) (best float64, date string, ok bool) {
	best = 0
	date = ""
	ok = false
	for i := len(entries) - 1; i >= 0; i-- { // oldest -> newest for deterministic date tie-break
		v, found := entries[i].Data[trackerID]
		if !found {
			continue
		}
		var val float64
		switch t := v.(type) {
		case float64:
			val = t
		case int:
			val = float64(t)
		default:
			continue
		}
		if !ok || val > best {
			best = val
			date = entries[i].Date
			ok = true
		}
	}
	return best, date, ok
}

// PersonalBestEntry pairs a tracker with its best value.
type PersonalBestEntry struct {
	TrackerID string
	Value     float64
	Date      string
}

// TopPersonalBests returns up to N personal bests across the given trackers,
// ranked by value/target ratio (or raw value when target is nil).
func TopPersonalBests(entries []models.Entry, trackers []models.Tracker, limit int) []PersonalBestEntry {
	type scored struct {
		entry PersonalBestEntry
		score float64
	}
	var rows []scored
	for _, t := range trackers {
		best, date, ok := PersonalBest(entries, t.ID)
		if !ok {
			continue
		}
		score := best
		if t.Target != nil && *t.Target > 0 {
			score = best / *t.Target
		}
		rows = append(rows, scored{PersonalBestEntry{TrackerID: t.ID, Value: best, Date: date}, score})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].score > rows[j].score })
	if limit > 0 && len(rows) > limit {
		rows = rows[:limit]
	}
	out := make([]PersonalBestEntry, len(rows))
	for i, r := range rows {
		out[i] = r.entry
	}
	return out
}

// PearsonCorrelation returns Pearson's r for two equal-length series, and
// ok=false when the input is too small or has zero variance.
func PearsonCorrelation(xs, ys []float64) (float64, bool) {
	n := len(xs)
	if n < 2 || n != len(ys) {
		return 0, false
	}
	var sumX, sumY float64
	for i := 0; i < n; i++ {
		sumX += xs[i]
		sumY += ys[i]
	}
	meanX := sumX / float64(n)
	meanY := sumY / float64(n)
	var num, dxSq, dySq float64
	for i := 0; i < n; i++ {
		dx := xs[i] - meanX
		dy := ys[i] - meanY
		num += dx * dy
		dxSq += dx * dx
		dySq += dy * dy
	}
	if dxSq == 0 || dySq == 0 {
		return 0, false
	}
	return num / math.Sqrt(dxSq*dySq), true
}

// LongestStreak scans entries (newest-first) and returns the longest run of
// consecutive days where trackerID == true.
func LongestStreak(entries []models.Entry, trackerID string) int {
	longest, cur := 0, 0
	for i := len(entries) - 1; i >= 0; i-- {
		v, ok := entries[i].Data[trackerID]
		if !ok {
			cur = 0
			continue
		}
		if b, ok := v.(bool); ok && b {
			cur++
			if cur > longest {
				longest = cur
			}
		} else {
			cur = 0
		}
	}
	return longest
}

// DaysSinceLastEntry returns how many full days ago the most recent entry
// was logged. Returns -1 when there are no entries.
func DaysSinceLastEntry(entries []models.Entry) int {
	newest := time.Time{}
	for _, e := range entries {
		t, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue
		}
		if t.After(newest) {
			newest = t
		}
	}
	if newest.IsZero() {
		return -1
	}
	today := time.Now()
	a := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(newest.Year(), newest.Month(), newest.Day(), 0, 0, 0, 0, time.UTC)
	return int(a.Sub(b).Hours() / 24)
}

// HasEntryForToday returns true when an entry exists for today's date.
func HasEntryForToday(entries []models.Entry) bool {
	today := time.Now().Format("2006-01-02")
	for _, e := range entries {
		if e.Date == today {
			return true
		}
	}
	return false
}

// SumInRange totals a numeric tracker's values for all entries whose ISO
// date falls in [start, end] inclusive.
func SumInRange(entries []models.Entry, trackerID string, start, end time.Time) float64 {
	sum := 0.0
	s := start.Format("2006-01-02")
	e := end.Format("2006-01-02")
	for _, entry := range entries {
		if entry.Date < s || entry.Date > e {
			continue
		}
		v, ok := entry.Data[trackerID]
		if !ok {
			continue
		}
		switch val := v.(type) {
		case float64:
			sum += val
		case int:
			sum += float64(val)
		case bool:
			if val {
				sum += 1
			}
		}
	}
	return sum
}

// AvgInRange averages non-missing numeric values for the tracker across the
// inclusive date range. Returns (0, 0) when no entries carry the tracker.
func AvgInRange(entries []models.Entry, trackerID string, start, end time.Time) (float64, int) {
	sum := 0.0
	count := 0
	s := start.Format("2006-01-02")
	e := end.Format("2006-01-02")
	for _, entry := range entries {
		if entry.Date < s || entry.Date > e {
			continue
		}
		v, ok := entry.Data[trackerID]
		if !ok {
			continue
		}
		switch val := v.(type) {
		case float64:
			sum += val
			count++
		case int:
			sum += float64(val)
			count++
		}
	}
	if count == 0 {
		return 0, 0
	}
	return sum / float64(count), count
}

// WeekBounds returns the Monday 00:00 and Sunday 23:59:59 of the week
// containing t (ISO week — Monday as week start).
func WeekBounds(t time.Time) (start, end time.Time) {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	start = time.Date(t.Year(), t.Month(), t.Day()-(wd-1), 0, 0, 0, 0, t.Location())
	end = start.AddDate(0, 0, 6)
	return
}

// MonthBounds returns the first and last day of the month containing t.
func MonthBounds(t time.Time) (start, end time.Time) {
	start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	end = start.AddDate(0, 1, -1)
	return
}

// WeeklyProgress returns the sum of tracker values for the week containing
// anchor (defaults to today when zero).
func WeeklyProgress(entries []models.Entry, trackerID string, anchor time.Time) float64 {
	if anchor.IsZero() {
		anchor = time.Now()
	}
	s, e := WeekBounds(anchor)
	return SumInRange(entries, trackerID, s, e)
}

// MonthlyProgress returns the sum of tracker values for the month containing
// anchor (defaults to today when zero).
func MonthlyProgress(entries []models.Entry, trackerID string, anchor time.Time) float64 {
	if anchor.IsZero() {
		anchor = time.Now()
	}
	s, e := MonthBounds(anchor)
	return SumInRange(entries, trackerID, s, e)
}

// minuteTrackerIDs returns the set of tracker IDs across the given categories
// whose type is Duration and unit is "minutes". Lookup is case-insensitive on
// category name. Used to scope category-level time rollups.
func minuteTrackerIDs(cfg *models.Config, categoryNames []string) map[string]bool {
	ids := make(map[string]bool)
	if cfg == nil {
		return ids
	}
	wanted := make(map[string]bool, len(categoryNames))
	for _, n := range categoryNames {
		wanted[strings.ToLower(strings.TrimSpace(n))] = true
	}
	for _, c := range cfg.Categories {
		if !wanted[strings.ToLower(c.Name)] {
			continue
		}
		for _, t := range c.Trackers {
			if t.Type == models.TrackerDuration && strings.EqualFold(t.Unit, "minutes") {
				ids[t.ID] = true
			}
		}
	}
	return ids
}

// SumCategoryMinutesOn totals minute-typed Duration trackers across the named
// categories for a single ISO date.
func SumCategoryMinutesOn(entries []models.Entry, cfg *models.Config, categoryNames []string, isoDate string) float64 {
	ids := minuteTrackerIDs(cfg, categoryNames)
	if len(ids) == 0 {
		return 0
	}
	for _, e := range entries {
		if e.Date != isoDate {
			continue
		}
		total := 0.0
		for id := range ids {
			if v, ok := e.Data[id].(float64); ok {
				total += v
			}
		}
		return total
	}
	return 0
}

// SumCategoryMinutesInRange totals minute-typed Duration trackers across the
// named categories for every entry whose date falls in [start,end] inclusive.
func SumCategoryMinutesInRange(entries []models.Entry, cfg *models.Config, categoryNames []string, start, end time.Time) float64 {
	ids := minuteTrackerIDs(cfg, categoryNames)
	if len(ids) == 0 {
		return 0
	}
	s := start.Format("2006-01-02")
	e := end.Format("2006-01-02")
	total := 0.0
	for _, entry := range entries {
		if entry.Date < s || entry.Date > e {
			continue
		}
		for id := range ids {
			if v, ok := entry.Data[id].(float64); ok {
				total += v
			}
		}
	}
	return total
}

// WeeklyCategoryMinutes sums minute-typed Duration trackers across the named
// categories for the calendar week containing anchor (defaults to today).
func WeeklyCategoryMinutes(entries []models.Entry, cfg *models.Config, categoryNames []string, anchor time.Time) float64 {
	if anchor.IsZero() {
		anchor = time.Now()
	}
	s, e := WeekBounds(anchor)
	return SumCategoryMinutesInRange(entries, cfg, categoryNames, s, e)
}

// BinaryByDate returns a dense bool slice of length `days`, ending on `end`
// (inclusive). Each index corresponds to a specific calendar date; missing
// dates are false. Unlike BinaryHeatmap, gaps in logging preserve calendar
// position — the rightmost cell is always `end`.
func BinaryByDate(entries []models.Entry, trackerID string, end time.Time, days int) []bool {
	if days <= 0 {
		return nil
	}
	result := make([]bool, days)
	byDate := make(map[string]bool, len(entries))
	for _, e := range entries {
		if v, ok := e.Data[trackerID].(bool); ok && v {
			byDate[e.Date] = true
		}
	}
	for i := 0; i < days; i++ {
		d := end.AddDate(0, 0, -(days - 1 - i)).Format("2006-01-02")
		result[i] = byDate[d]
	}
	return result
}

// NumericByDate returns dense value/present slices of length `days`, ending on
// `end` (inclusive). Missing days get zero in values and false in present,
// preserving calendar alignment for sparklines and line charts.
func NumericByDate(entries []models.Entry, trackerID string, end time.Time, days int) (values []float64, present []bool) {
	if days <= 0 {
		return nil, nil
	}
	values = make([]float64, days)
	present = make([]bool, days)
	byDate := make(map[string]float64, len(entries))
	for _, e := range entries {
		raw, ok := e.Data[trackerID]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			byDate[e.Date] = v
		case int:
			byDate[e.Date] = float64(v)
		case bool:
			if v {
				byDate[e.Date] = 1
			} else {
				byDate[e.Date] = 0
			}
		}
	}
	for i := 0; i < days; i++ {
		d := end.AddDate(0, 0, -(days - 1 - i)).Format("2006-01-02")
		if v, ok := byDate[d]; ok {
			values[i] = v
			present[i] = true
		}
	}
	return
}

// BinaryHitsInRange counts how many entries in [start,end] have trackerID==true.
func BinaryHitsInRange(entries []models.Entry, trackerID string, start, end time.Time) (hits, total int) {
	s := start.Format("2006-01-02")
	e := end.Format("2006-01-02")
	for _, entry := range entries {
		if entry.Date < s || entry.Date > e {
			continue
		}
		total++
		if b, ok := entry.Data[trackerID].(bool); ok && b {
			hits++
		}
	}
	return
}

type MomentumAcceleration struct {
	TrackerID string
	RecentAvg float64
	PrevAvg   float64
	Delta     float64
}

// MomentumAccelerationRanking computes momentum deltas and sorts descending.
func MomentumAccelerationRanking(entries []models.Entry, trackerIDs []string, window int) []MomentumAcceleration {
	var rows []MomentumAcceleration
	for _, trackerID := range trackerIDs {
		recent, prev, delta, ok := TrackerMomentum(entries, trackerID, window)
		if !ok {
			continue
		}
		rows = append(rows, MomentumAcceleration{
			TrackerID: trackerID,
			RecentAvg: recent,
			PrevAvg:   prev,
			Delta:     delta,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].Delta > rows[j].Delta
	})
	return rows
}
