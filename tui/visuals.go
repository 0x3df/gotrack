package tui

import (
	"fmt"
	"math"
	"strings"

	"dailytrack/models"

	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"
)

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

	filledStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render(strings.Repeat("█", filledBlocks))
	emptyStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Border)).Render(strings.Repeat("░", emptyBlocks))

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
	max := valA
	if valB > max {
		max = valB
	}
	if max == 0 {
		max = 1
	}

	width := 20
	blocksA := int((valA / max) * float64(width))
	blocksB := int((valB / max) * float64(width))

	barA := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render(strings.Repeat("█", blocksA))
	barB := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Render(strings.Repeat("█", blocksB))

	strA := fmt.Sprintf("%-15s | %s %.1f", labelA, barA, valA)
	strB := fmt.Sprintf("%-15s | %s %.1f", labelB, barB, valB)

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

		char := "■"
		if i < offset {
			grid[row][col] = " " // Empty space for padding
		} else if done {
			grid[row][col] = activeStyle.Render(char)
		} else {
			grid[row][col] = inactiveStyle.Render(char)
		}
	}

	// Render the grid row by row
	var result []string
	dayLabels := []string{"S", "M", "T", "W", "T", "F", "S"}
	for i := 0; i < 7; i++ {
		rowStr := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(dayLabels[i] + " ")
		rowStr += strings.Join(grid[i], " ")
		result = append(result, rowStr)
	}

	return strings.Join(result, "\n")
}

// VerticalBarChart creates a column-based bar chart.
// Uses vertical block elements.
func VerticalBarChart(counts []int, labels []string) string {
	p := palette()
	if len(counts) == 0 {
		return "No data"
	}

	max := counts[0]
	for _, v := range counts {
		if v > max {
			max = v
		}
	}
	if max == 0 {
		max = 1
	}

	blocks := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	barRow := ""
	for _, v := range counts {
		idx := int(math.Round((float64(v) / float64(max)) * float64(len(blocks)-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}
		barRow += lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render(blocks[idx]) + "  "
	}

	labelRow := ""
	for _, l := range labels {
		labelRow += lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(l) + " "
	}

	return barRow + "\n" + labelRow
}

// ScatterPlot visually plots points on a 2D text grid.
// xVals and yVals must be the same length.
func ScatterPlot(xVals []float64, yVals []float64, xLabel, yLabel string) string {
	p := palette()
	if len(xVals) == 0 || len(xVals) != len(yVals) {
		return "Invalid or empty data"
	}

	gridWidth := 30
	gridHeight := 10

	var maxX, maxY float64
	for i := range xVals {
		if xVals[i] > maxX {
			maxX = xVals[i]
		}
		if yVals[i] > maxY {
			maxY = yVals[i]
		}
	}
	if maxX == 0 {
		maxX = 1
	}
	if maxY == 0 {
		maxY = 1
	}

	// Initialize empty grid
	grid := make([][]string, gridHeight)
	for i := range grid {
		grid[i] = make([]string, gridWidth)
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Plot points
	dotStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Bold(true)
	for i := range xVals {
		xIdx := int(math.Round((xVals[i] / maxX) * float64(gridWidth-1)))
		yIdx := gridHeight - 1 - int(math.Round((yVals[i]/maxY)*float64(gridHeight-1))) // invert Y so 0 is at bottom

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

		grid[yIdx][xIdx] = dotStyle.Render("•")
	}

	// Render grid with simple axes
	var result []string

	// Top Y label
	result = append(result, lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(fmt.Sprintf("%v (%.0f max)", yLabel, maxY)))

	for _, row := range grid {
		result = append(result, "│"+strings.Join(row, ""))
	}

	// X axis line
	result = append(result, "└"+strings.Repeat("─", gridWidth))

	// X axis label
	xLabelLine := fmt.Sprintf("%*s", gridWidth, fmt.Sprintf("%s (%.0f max)", xLabel, maxX))
	result = append(result, " "+lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render(xLabelLine))

	return strings.Join(result, "\n")
}

// renderLineChart renders a line chart for duration/numeric/count trackers.
func renderLineChart(series []float64, t models.Tracker, cardWidth int) string {
	chartWidth := dashboardChartWidth(cardWidth)
	progressWidth := dashboardProgressWidth(cardWidth)

	if len(series) < 2 {
		if len(series) == 1 {
			res := fmt.Sprintf("Latest: %s", formatValueWithUnit(series[0], t))
			if t.Target != nil {
				pct := (series[0] / *t.Target) * 100
				res += fmt.Sprintf("\nTarget: %s\nGoal Completion:\n%s",
					formatValueWithUnit(*t.Target, t), ProgressBar(pct, progressWidth))
			}
			return res
		}
		return "Not enough data yet."
	}

	unit := trackerUnit(t)
	label := fmt.Sprintf("Last %d entries%s", len(series), unit)
	if unit != "" {
		label = fmt.Sprintf("Last %d entries (%s)", len(series), unit)
	}
	chart := asciigraph.Plot(series, asciigraph.Height(4), asciigraph.Width(chartWidth))

	extra := ""
	if t.Target != nil {
		latest := series[len(series)-1]
		pct := (latest / *t.Target) * 100
		hitCount := 0
		for _, v := range series {
			if v >= *t.Target {
				hitCount++
			}
		}
		extra = fmt.Sprintf("\nLatest Goal: %s\nTarget %s: hit %d/%d days",
			ProgressBar(pct, progressWidth), formatValueWithUnit(*t.Target, t), hitCount, len(series))
	}

	return label + "\n" + chart + extra
}
