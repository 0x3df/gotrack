package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"dailytrack/db"
	"dailytrack/models"

	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

// goalProgressLines returns a short stacked summary of weekly/monthly
// goal progress for a tracker, or "" when neither goal is configured.
func goalProgressLines(entries []models.Entry, t models.Tracker, width int) string {
	if t.WeeklyTarget == nil && t.MonthlyTarget == nil {
		return ""
	}
	var out []string
	unit := trackerUnit(t)
	if t.WeeklyTarget != nil {
		cur := db.WeeklyProgress(entries, t.ID, time.Time{})
		out = append(out, renderGoalProgress("Week", cur, *t.WeeklyTarget, unit, width))
	}
	if t.MonthlyTarget != nil {
		cur := db.MonthlyProgress(entries, t.ID, time.Time{})
		out = append(out, renderGoalProgress("Month", cur, *t.MonthlyTarget, unit, width))
	}
	return strings.Join(out, "\n")
}

// renderGoalProgress returns a single-line "Label  current / target unit [bar] pct%"
// row suitable for stacking beside other visuals.
func renderGoalProgress(label string, current, target float64, unit string, width int) string {
	p := palette()
	pct := 0.0
	if target > 0 {
		pct = (current / target) * 100
	}
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	numerator := formatFloatValue(current)
	denom := formatFloatValue(target)
	summary := fmt.Sprintf("%s / %s", numerator, denom)
	if strings.TrimSpace(unit) != "" {
		summary += " " + unit
	}
	barWidth := width - len(label) - len(summary) - 4
	if barWidth < 12 {
		barWidth = 12
	}
	return fmt.Sprintf("%s  %s  %s",
		lipgloss.NewStyle().Bold(true).Render(label),
		muted.Render(summary),
		ProgressBar(pct, barWidth),
	)
}

// glyphFilled, glyphEmpty, glyphDone, glyphPartial, glyphMissed return
// appropriate cell characters for the active theme. Themes with ASCIIOnly
// set use 7-bit-safe characters for screen readers and basic terminals.
func glyphFilled() string {
	if palette().ASCIIOnly {
		return "#"
	}
	return "█"
}

func glyphEmpty() string {
	if palette().ASCIIOnly {
		return "-"
	}
	return "░"
}

func glyphDone() string {
	if palette().ASCIIOnly {
		return "X"
	}
	return "■"
}

func glyphPartial() string {
	if palette().ASCIIOnly {
		return "o"
	}
	return "▪"
}

func glyphMissed() string {
	if palette().ASCIIOnly {
		return "."
	}
	return "·"
}

// ProgressBar returns a string like "[████████░░] 80%"
func ProgressBar(percent float64, width int) string {
	p := palette()
	if math.IsNaN(percent) || math.IsInf(percent, 0) {
		percent = 0
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	totalBlocks := width - 8 // account for brackets and percentage text
	if totalBlocks < 1 {
		totalBlocks = 10
	}

	filledBlocks := int(math.Round((percent / 100.0) * float64(totalBlocks)))
	emptyBlocks := totalBlocks - filledBlocks

	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render(strings.Repeat(glyphFilled(), filledBlocks))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Border)).Render(strings.Repeat(glyphEmpty(), emptyBlocks))

	return fmt.Sprintf("[%s%s] %3.0f%%", filledStr, emptyStr, percent)
}

// Sparkline creates a mini bar chart from a slice of floats
func Sparkline(data []float64, color string) string {
	if len(data) == 0 {
		return "No data"
	}

	max := data[0]
	for _, v := range data {
		if v > max {
			max = v
		}
	}

	if max == 0 {
		max = 1 // avoid div by zero
	}

	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var result []string

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	for _, v := range data {
		idx := int(math.Round((v / max) * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		result = append(result, style.Render(blocks[idx]))
	}

	return strings.Join(result, " ")
}

// ComparisonBar returns a side-by-side comparison visualization
func ComparisonBar(labelA string, valA float64, labelB string, valB float64) string {
	p := palette()
	ra := roundAverageUp2(valA)
	rb := roundAverageUp2(valB)
	max := ra
	if rb > max {
		max = rb
	}
	if max == 0 {
		max = 1
	}

	width := 20
	blocksA := int((ra / max) * float64(width))
	blocksB := int((rb / max) * float64(width))

	barA := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render(strings.Repeat("█", blocksA))
	barB := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Render(strings.Repeat("█", blocksB))

	strA := fmt.Sprintf("%-15s | %s %s", labelA, barA, formatAverageFixed2(valA))
	strB := fmt.Sprintf("%-15s | %s %s", labelB, barB, formatAverageFixed2(valB))

	return strA + "\n" + strB
}

// Heatmap returns a Github-style contribution grid (7 rows representing days of week, columns are weeks)
func Heatmap(data []bool, offset int) string {
	p := palette()
	if len(data) == 0 {
		return "No data"
	}

	// Pad data with false based on offset to align with Sunday
	padded := make([]bool, offset)
	padded = append(padded, data...)

	// Calculate how many full/partial weeks we have
	numCols := int(math.Ceil(float64(len(padded)) / 7.0))
	grid := make([][]string, 7)
	for i := range grid {
		grid[i] = make([]string, numCols)
		for j := range grid[i] {
			grid[i][j] = " " // Fill with empty spaces initially
		}
	}

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapActive))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapInactive))

	// Map chronological data into columns.
	for i, done := range padded {
		col := i / 7
		row := i % 7 // Sunday = 0

		char := glyphDone()
		if i < offset {
			grid[row][col] = " " // Empty space for padding
		} else if done {
			grid[row][col] = activeStyle.Render(char)
		} else {
			grid[row][col] = inactiveStyle.Render(glyphMissed())
		}
	}

	// Render the grid row by row
	var result []string
	dayLabels := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	for i := 0; i < 7; i++ {
		rowStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(dayLabels[i] + " ")
		rowStr += strings.Join(grid[i], " ")
		result = append(result, rowStr)
	}

	return strings.Join(result, "\n")
}

// ScatterPlot visually plots points on a 2D text grid.
// xVals and yVals must be the same length.
func ScatterPlot(xVals []float64, yVals []float64, xLabel, yLabel string) string {
	p := palette()
	if len(xVals) == 0 || len(xVals) != len(yVals) {
		return "Invalid or empty data"
	}

	gridWidth := 30
	gridHeight := 8

	minX, maxX := xVals[0], xVals[0]
	minY, maxY := yVals[0], yVals[0]
	for i := range xVals {
		if xVals[i] < minX {
			minX = xVals[i]
		}
		if xVals[i] > maxX {
			maxX = xVals[i]
		}
		if yVals[i] < minY {
			minY = yVals[i]
		}
		if yVals[i] > maxY {
			maxY = yVals[i]
		}
	}
	rangeX := maxX - minX
	rangeY := maxY - minY
	if rangeX == 0 {
		rangeX = 1
	}
	if rangeY == 0 {
		rangeY = 1
	}

	// Initialize empty grid
	grid := make([][]string, gridHeight)
	for i := range grid {
		grid[i] = make([]string, gridWidth)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Plot points; overlapping points use "●" so density is visible.
	dotStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Bold(true)
	denseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Bold(true)
	rawCounts := make([][]int, gridHeight)
	for i := range rawCounts {
		rawCounts[i] = make([]int, gridWidth)
	}
	for i := range xVals {
		xIdx := int(math.Round(((xVals[i] - minX) / rangeX) * float64(gridWidth-1)))
		yIdx := gridHeight - 1 - int(math.Round(((yVals[i]-minY)/rangeY)*float64(gridHeight-1)))
		if xIdx < 0 {
			xIdx = 0
		}
		if xIdx >= gridWidth {
			xIdx = gridWidth - 1
		}
		if yIdx < 0 {
			yIdx = 0
		}
		if yIdx >= gridHeight {
			yIdx = gridHeight - 1
		}
		rawCounts[yIdx][xIdx]++
	}
	for y := 0; y < gridHeight; y++ {
		for x := 0; x < gridWidth; x++ {
			switch {
			case rawCounts[y][x] >= 2:
				grid[y][x] = denseStyle.Render("●")
			case rawCounts[y][x] == 1:
				grid[y][x] = dotStyle.Render("•")
			}
		}
	}

	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	var result []string
	// Y-axis top tick (max) → bottom (min)
	result = append(result, muted.Render(fmt.Sprintf("%s", truncateLabel(yLabel, gridWidth))))
	for i, row := range grid {
		prefix := "│ "
		switch i {
		case 0:
			prefix = muted.Render(fmt.Sprintf("%4.0f ┤", maxY))
		case gridHeight - 1:
			prefix = muted.Render(fmt.Sprintf("%4.0f ┤", minY))
		default:
			prefix = muted.Render("     │")
		}
		result = append(result, prefix+strings.Join(row, ""))
	}
	result = append(result, muted.Render("     └"+strings.Repeat("─", gridWidth)))
	xTicks := muted.Render(fmt.Sprintf("      %-*.0f%*.0f", gridWidth/2, minX, gridWidth-gridWidth/2, maxX))
	result = append(result, xTicks)
	result = append(result, muted.Render(fmt.Sprintf("     %s →", truncateLabel(xLabel, gridWidth))))
	return strings.Join(result, "\n")
}

// renderLineChart renders a legible numeric/duration/count tracker card.
// Few-entry cases show an explicit recent-values list rather than a noisy
// 2-point line chart; richer histories add a sparkline and summary stats.
func renderLineChart(statsSeries, trendSeries []float64, t models.Tracker, cardWidth int) string {
	progressWidth := dashboardProgressWidth(cardWidth)
	p := palette()

	if len(statsSeries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("No entries yet.")
	}

	latest := statsSeries[len(statsSeries)-1]
	min, max, sum := statsSeries[0], statsSeries[0], 0.0
	for _, v := range statsSeries {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}
	avg := sum / float64(len(statsSeries))

	var lines []string
	lines = append(lines, fmt.Sprintf("Latest: %s", formatValueWithUnit(latest, t)))
	if len(statsSeries) >= 2 {
		prev := statsSeries[len(statsSeries)-2]
		arrow, color := deltaArrow(latest - prev)
		arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		lines = append(lines, fmt.Sprintf("vs prev: %s %s",
			arrowStyle.Render(arrow), formatValueWithUnit(latest-prev, t)))
	}
	lines = append(lines, fmt.Sprintf("Avg: %s   Range: %s – %s",
		formatAverageWithUnit(avg, t),
		formatValueWithUnit(min, t), formatValueWithUnit(max, t)))

	// Short history: show recent values as an explicit list, which is far
	// easier to read than a 2-point asciigraph.
	if len(statsSeries) < 4 {
		lines = append(lines, "")
		lines = append(lines, "Recent values:")
		for i := len(statsSeries) - 1; i >= 0; i-- {
			marker := "·"
			if t.Target != nil && statsSeries[i] >= *t.Target {
				marker = "✓"
			}
			lines = append(lines, fmt.Sprintf("  %s %s", marker, formatValueWithUnit(statsSeries[i], t)))
		}
	} else {
		sparkColor := p.ChartPrimary
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Trend (last %d days):", len(trendSeries)))
		lines = append(lines, Sparkline(trendSeries, sparkColor))
		if len(statsSeries) >= 7 {
			rolling := db.RollingAverageSeries(statsSeries, 7)
			if len(rolling) >= 2 {
				first := rolling[0]
				last := rolling[len(rolling)-1]
				arrow, color := deltaArrow(last - first)
				arrowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
				lines = append(lines, fmt.Sprintf("7d avg: %s %s %s",
					formatAverageWithUnit(first, t),
					arrowStyle.Render(arrow),
					formatAverageWithUnit(last, t)))
			}
		}
	}

	if t.Target != nil {
		pct := (latest / *t.Target) * 100
		hitCount := 0
		for _, v := range statsSeries {
			if v >= *t.Target {
				hitCount++
			}
		}
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Target: %s — hit %d/%d",
			formatValueWithUnit(*t.Target, t), hitCount, len(statsSeries)))
		lines = append(lines, ProgressBar(pct, progressWidth))
	}
	return strings.Join(lines, "\n")
}

// renderRatingCard renders a rating-tracker card emphasizing the current
// rating and the full 1–5 distribution so the shape is readable at a glance.
func renderRatingCard(statsSeries, trendSeries []float64, t models.Tracker, cardWidth int) string {
	p := palette()
	if len(statsSeries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("No ratings yet.")
	}
	latest := int(math.Round(statsSeries[len(statsSeries)-1]))
	if latest < 1 {
		latest = 1
	}
	if latest > 5 {
		latest = 5
	}
	sum := 0.0
	var counts [6]int // index 1..5
	for _, v := range statsSeries {
		iv := int(math.Round(v))
		if iv < 1 {
			iv = 1
		}
		if iv > 5 {
			iv = 5
		}
		counts[iv]++
		sum += v
	}
	avg := sum / float64(len(statsSeries))

	stars := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Render(strings.Repeat("★", latest)) +
		lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(strings.Repeat("☆", 5-latest))

	maxCount := 0
	for i := 1; i <= 5; i++ {
		if counts[i] > maxCount {
			maxCount = counts[i]
		}
	}
	if maxCount == 0 {
		maxCount = 1
	}
	barWidth := dashboardProgressWidth(cardWidth) - 6
	if barWidth < 6 {
		barWidth = 6
	}
	var distLines []string
	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary))
	for i := 5; i >= 1; i-- { // show 5 → 1 so "great" is on top
		n := int(math.Round((float64(counts[i]) / float64(maxCount)) * float64(barWidth)))
		bar := barStyle.Render(strings.Repeat("█", n))
		distLines = append(distLines, fmt.Sprintf("%d★ %s %d", i, bar, counts[i]))
	}

	header := []string{
		fmt.Sprintf("Latest: %s  (%d/5)", stars, latest),
		fmt.Sprintf("Avg: %s / 5   n=%d", formatAverageFixed2(avg), len(statsSeries)),
		"",
		fmt.Sprintf("Trend (last %d days):", len(trendSeries)),
		Sparkline(trendSeries, p.ChartPrimary),
		"",
		"Distribution:",
	}
	return strings.Join(header, "\n") + "\n" + strings.Join(distLines, "\n")
}

func deltaArrow(delta float64) (string, string) {
	p := palette()
	if math.Abs(delta) < 1e-9 {
		return "→", p.Muted
	}
	if delta > 0 {
		return "↑", p.Success
	}
	return "↓", p.Danger
}

func unitSuffix(unit string) string {
	if unit == "" {
		return ""
	}
	return " " + unit
}

// DualLineChart renders two series in the same chart area.
func DualLineChart(primary, secondary []float64, width, height int) string {
	if len(primary) == 0 {
		return "Not enough data yet."
	}
	if width <= 0 {
		width = 24
	}
	if height <= 0 {
		height = 4
	}
	if len(secondary) == 0 {
		return asciigraph.Plot(primary, asciigraph.Width(width), asciigraph.Height(height))
	}
	return asciigraph.PlotMany(
		[][]float64{primary, secondary},
		asciigraph.Width(width),
		asciigraph.Height(height),
	)
}

// LeaderboardBars renders ranked momentum rows.
func LeaderboardBars(rows []LeaderboardRow, barWidth int) string {
	if len(rows) == 0 {
		return "No momentum data yet."
	}
	if barWidth <= 0 {
		barWidth = 12
	}
	maxAbs := 0.0
	for _, r := range rows {
		if math.Abs(r.Delta) > maxAbs {
			maxAbs = math.Abs(r.Delta)
		}
	}
	if maxAbs == 0 {
		maxAbs = 1
	}
	p := palette()
	var lines []string
	for i, r := range rows {
		if i >= 5 {
			break
		}
		n := int(math.Round((math.Abs(r.Delta) / maxAbs) * float64(barWidth)))
		if n < 1 {
			n = 1
		}
		if n > barWidth {
			n = barWidth
		}
		bar := strings.Repeat("█", n)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success))
		sign := "↑"
		if r.Delta < 0 {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger))
			sign = "↓"
		}
		lines = append(lines, fmt.Sprintf("%-16s %s %s %.2f", truncateLabel(r.Label, 16), sign, style.Render(bar), r.Delta))
	}
	return strings.Join(lines, "\n")
}

type LeaderboardRow struct {
	Label string
	Delta float64
}

func truncateLabel(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:1]
	}
	return s[:max-1] + "…"
}

// TrendDeltaStrip renders a compact delta summary for recent vs previous windows.
func TrendDeltaStrip(recentAvg, prevAvg float64) string {
	p := palette()
	rRecent := roundAverageUp2(recentAvg)
	rPrev := roundAverageUp2(prevAvg)
	delta := rRecent - rPrev
	sign := "↑"
	color := p.Success
	if delta < 0 {
		sign = "↓"
		color = p.Danger
	}
	if math.Abs(delta) < 0.0001 {
		sign = "→"
		color = p.Muted
	}
	deltaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	return fmt.Sprintf("Recent: %s  Prev: %s\nTrend: %s %s",
		formatAverageFixed2(recentAvg), formatAverageFixed2(prevAvg),
		deltaStyle.Render(sign), formatAverageFixed2(math.Abs(delta)))
}

// TargetHitMeter renders a compact meter for hit-rate cards.
func TargetHitMeter(hits, total, width int) string {
	if total <= 0 {
		return "No target data."
	}
	pct := float64(hits) / float64(total) * 100
	if width <= 0 {
		width = 24
	}
	return fmt.Sprintf("%d/%d hits (%.0f%%)\n%s", hits, total, pct, ProgressBar(pct, width))
}

// CorrelationReadout renders a one-line summary of a Pearson r with sample size.
func CorrelationReadout(r float64, n int, xLabel, yLabel string) string {
	p := palette()
	strength := "weak"
	color := p.Muted
	abs := math.Abs(r)
	switch {
	case abs >= 0.7:
		strength = "strong"
		color = p.Success
	case abs >= 0.4:
		strength = "moderate"
		color = p.ChartPrimary
	}
	if r < 0 {
		color = p.Danger
	}
	rStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
	return fmt.Sprintf("%s vs %s\nr = %s  (%s, n=%d)",
		truncateLabel(xLabel, 20), truncateLabel(yLabel, 20),
		rStyle.Render(fmt.Sprintf("%+.2f", r)), strength, n)
}

// ABCompareRow renders a two-row group comparison with sample sizes.
func ABCompareRow(labelYes string, yesVals []float64, labelNo string, noVals []float64, metricLabel string) string {
	avgYes := 0.0
	if len(yesVals) > 0 {
		for _, v := range yesVals {
			avgYes += v
		}
		avgYes /= float64(len(yesVals))
	}
	avgNo := 0.0
	if len(noVals) > 0 {
		for _, v := range noVals {
			avgNo += v
		}
		avgNo /= float64(len(noVals))
	}
	bar := ComparisonBar(fmt.Sprintf("%s (n=%d)", labelYes, len(yesVals)), avgYes,
		fmt.Sprintf("%s (n=%d)", labelNo, len(noVals)), avgNo)
	delta := roundAverageUp2(avgYes) - roundAverageUp2(avgNo)
	return fmt.Sprintf("%s\n\nΔ %s: %+.2f", bar, metricLabel, delta)
}

// LastWeekStrip renders a compact 7-day status row for a binary tracker:
// one block per day ending today (oldest → newest = left → right).
func LastWeekStrip(boolOldestFirst []bool) string {
	p := palette()
	if len(boolOldestFirst) == 0 {
		return "—"
	}
	start := len(boolOldestFirst) - 7
	if start < 0 {
		start = 0
	}
	slice := boolOldestFirst[start:]
	active := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapActive))
	inactive := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapInactive))
	var cells []string
	for _, b := range slice {
		if b {
			cells = append(cells, active.Render(glyphDone()))
		} else {
			cells = append(cells, inactive.Render(glyphMissed()))
		}
	}
	return strings.Join(cells, " ")
}

// NumericLastWeekStrip renders a 7-day calendar-aligned strip for numeric
// trackers. Missing days stay visible as gap markers rather than collapsing
// the most recent logged value into today's slot.
func NumericLastWeekStrip(values []float64, present []bool, color string) string {
	p := palette()
	if len(values) == 0 {
		return "—"
	}
	if len(present) != len(values) {
		return "—"
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	missStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapInactive))

	var cells []string
	for i := 0; i < len(values); i++ {
		if present[i] {
			cells = append(cells, style.Render(glyphPartial()))
		} else {
			cells = append(cells, missStyle.Render(glyphMissed()))
		}
	}
	return strings.Join(cells, " ")
}

// WeekdayConsistencyBars renders weekday completion percentages as compact bars.
func WeekdayConsistencyBars(weekdayPct [7]float64) string {
	p := palette()
	days := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	var lines []string
	for i, pct := range weekdayPct {
		level := int(math.Round((pct / 100) * 8))
		if level < 0 {
			level = 0
		}
		if level > 8 {
			level = 8
		}
		bar := strings.Repeat("█", level) + strings.Repeat("░", 8-level)
		barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary))
		lines = append(lines, fmt.Sprintf("%s %s %.0f%%", days[i], barStyle.Render(bar), pct))
	}
	return strings.Join(lines, "\n")
}
