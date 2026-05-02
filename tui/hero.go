package tui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"dailytrack/db"
	"dailytrack/models"

	"github.com/charmbracelet/lipgloss"
)

// heroVisual describes one striking full-width overview visualization the
// user can cycle through with [ and ].
type heroVisual struct {
	Key    string
	Title  string
	Render func(m Model, innerWidth, innerHeight int) string
}

// heroVisuals lists the cycleable big-container visualizations. Order here
// is the cycle order — the first one is what new users see first.
var heroVisuals = []heroVisual{
	{"wall", "Tracker Wall (last 4 weeks)", renderHeroTrackerWall},
	{"pulse", "Yearly Pulse", renderHeroYearPulse},
	{"rhythm", "Weekday Rhythm", renderHeroWeekdayRhythm},
	{"podium", "Momentum Podium", renderHeroMomentumPodium},
	{"month", "Month at a Glance", renderHeroMonthCalendar},
}

// renderHero renders the currently selected hero visual, wrapped in a
// prominently styled card that fills most of the overview viewport.
func (m Model) renderHero(outerWidth int) string {
	p := palette()
	idx := m.heroIndex % len(heroVisuals)
	if idx < 0 {
		idx = 0
	}
	viz := heroVisuals[idx]

	// Inner dimensions: border + padding + title + footer.
	const heroHeight = 22
	const framePadding = 2
	innerWidth := outerWidth - 2*framePadding - 2 // border
	if innerWidth < 20 {
		innerWidth = 20
	}
	innerHeight := heroHeight - 2 - 2 - 3 // border+padding+title+footer
	if innerHeight < 6 {
		innerHeight = 6
	}

	body := viz.Render(m, innerWidth, innerHeight)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.ActiveTabBg)).
		Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))

	dots := make([]string, len(heroVisuals))
	for i := range heroVisuals {
		if i == idx {
			dots[i] = titleStyle.Render("●")
		} else {
			dots[i] = mutedStyle.Render("○")
		}
	}
	footer := mutedStyle.Render(fmt.Sprintf("◀ [  ] ▶   %s   %d/%d",
		strings.Join(dots, " "), idx+1, len(heroVisuals)))
	footer = lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(footer)

	content := lipgloss.JoinVertical(lipgloss.Center,
		titleStyle.Width(innerWidth).Align(lipgloss.Center).Render(viz.Title),
		"",
		lipgloss.NewStyle().Width(innerWidth).Align(lipgloss.Center).Render(body),
	)

	card := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(p.ActiveTabBg)).
		Padding(1, framePadding).
		Width(outerWidth - 2).
		Height(heroHeight - 2).
		Render(lipgloss.JoinVertical(lipgloss.Center, content, "", footer))

	return card
}

// ─── Hero visual #1: Yearly Pulse ────────────────────────────────────────
//
// Shows every logged day as a colored cell, intensity = fraction of binary
// trackers completed that day. Gives a calendar-style "year at a glance".

func renderHeroYearPulse(m Model, innerWidth, innerHeight int) string {
	p := palette()
	if len(m.entries) == 0 {
		return mutedLine("No entries yet. Press 'a' to log your first day.")
	}

	// Build completion-percent-by-date for binary trackers only.
	byDate := map[string]float64{}
	binaryIDs := collectBinaryIDs(m.config)
	if len(binaryIDs) == 0 {
		return mutedLine("Add a binary tracker to see the yearly pulse.")
	}

	var newest, oldest time.Time
	first := true
	for _, e := range m.entries {
		parsed, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue
		}
		done := 0
		total := 0
		for _, id := range binaryIDs {
			if v, ok := e.Data[id]; ok {
				total++
				if b, ok := v.(bool); ok && b {
					done++
				}
			}
		}
		frac := 0.0
		if total > 0 {
			frac = float64(done) / float64(total)
		}
		byDate[e.Date] = frac
		if first || parsed.After(newest) {
			newest = parsed
		}
		if first || parsed.Before(oldest) {
			oldest = parsed
		}
		first = false
	}

	// How many columns (weeks) fit? Each cell is 2 chars wide ("■ ").
	// Leave room for the weekday labels on the left (3 chars).
	cols := (innerWidth - 3) / 2
	if cols < 10 {
		cols = 10
	}
	totalDays := cols * 7
	end := newest
	start := end.AddDate(0, 0, -totalDays+1)
	if start.Before(oldest.AddDate(0, 0, -7)) {
		// show slightly more than data range so empty cells are obvious
		start = oldest.AddDate(0, 0, -int(oldest.Weekday()))
	}

	// Build the grid: rows=7 (Sun..Sat), cols=cols
	grid := make([][]string, 7)
	for i := range grid {
		grid[i] = make([]string, cols)
		for j := range grid[i] {
			grid[i][j] = "  "
		}
	}
	monthMarks := make([]string, cols)
	for i := range monthMarks {
		monthMarks[i] = " "
	}

	cur := start
	col := 0
	for col < cols {
		for wd := 0; wd < 7; wd++ {
			d := cur.AddDate(0, 0, col*7+wd-int(start.Weekday()))
			if d.Before(start) || d.After(end) {
				continue
			}
			if d.Day() <= 7 {
				monthMarks[col] = shortMonth(d.Month())
			}
			iso := d.Format("2006-01-02")
			frac, hasEntry := byDate[iso]
			grid[int(d.Weekday())][col] = pulseCellChar(frac, hasEntry, p)
		}
		col++
	}

	days := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))

	var rows []string
	// Month header row: thin, just markers where the month changes
	monthRow := "   "
	for _, m := range monthMarks {
		if m == " " {
			monthRow += "  "
		} else {
			monthRow += mutedStyle.Render(m)
		}
	}
	rows = append(rows, monthRow)
	for i := 0; i < 7; i++ {
		line := mutedStyle.Render(days[i] + " ")
		line += strings.Join(grid[i], "")
		rows = append(rows, line)
	}

	legend := "   " + mutedStyle.Render("less ") +
		pulseCellChar(0, true, p) +
		pulseCellChar(0.25, true, p) +
		pulseCellChar(0.5, true, p) +
		pulseCellChar(0.75, true, p) +
		pulseCellChar(1, true, p) +
		mutedStyle.Render(" more   ") +
		mutedStyle.Render(fmt.Sprintf("(%d binary trackers, %d logged days)",
			len(binaryIDs), len(byDate)))
	rows = append(rows, "")
	rows = append(rows, legend)
	return strings.Join(rows, "\n")
}

func pulseCellChar(frac float64, hasEntry bool, p ThemePalette) string {
	if !hasEntry {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapInactive)).Render("· ")
	}
	// Map fraction to a 5-level scale of block density/color.
	levels := []string{"░ ", "▒ ", "▓ ", "█ ", "█ "}
	colors := []string{p.HeatmapInactive, p.ChartPrimary, p.ChartPrimary, p.HeatmapActive, p.Success}
	idx := int(math.Round(frac * 4))
	if idx < 0 {
		idx = 0
	}
	if idx > 4 {
		idx = 4
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(colors[idx])).Render(levels[idx])
}

func shortMonth(m time.Month) string {
	return m.String()[:1]
}

// ─── Hero visual #2: Tracker Wall ────────────────────────────────────────
//
// One row per tracker, showing the last N days as a strip of cells. Binary
// trackers use solid/empty dots; numeric trackers use graded bar blocks
// relative to that tracker's own max.

func renderHeroTrackerWall(m Model, innerWidth, innerHeight int) string {
	p := palette()
	if len(m.entries) == 0 {
		return mutedLine("No entries yet.")
	}

	const weeks = 4
	const days = weeks * 7
	const labelWidth = 20
	const summaryWidth = 10
	const weekGap = 2

	// Build the 28-date window ending on today, aligned so the far-right
	// column is today. Group into 4 consecutive week-blocks of 7 cells each.
	today := time.Now()
	dates := make([]string, days)
	for i := 0; i < days; i++ {
		dates[days-1-i] = today.AddDate(0, 0, -i).Format("2006-01-02")
	}
	byDate := map[string]models.Entry{}
	for _, e := range m.entries {
		byDate[e.Date] = e
	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	missStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapInactive))
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.HeatmapActive))
	partialStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success))

	// Week-start date row: one label per 7-cell block at its first column.
	weekHeader := strings.Repeat(" ", labelWidth+2)
	weekdayHeader := strings.Repeat(" ", labelWidth+2)
	for w := 0; w < weeks; w++ {
		first, _ := time.Parse("2006-01-02", dates[w*7])
		weekHeader += mutedStyle.Render(fmt.Sprintf("%-7s", first.Format("1/02")))
		if w < weeks-1 {
			weekHeader += strings.Repeat(" ", weekGap)
		}
		for d := 0; d < 7; d++ {
			parsed, _ := time.Parse("2006-01-02", dates[w*7+d])
			weekdayHeader += mutedStyle.Render(weekdayInitial(parsed.Weekday()))
		}
		if w < weeks-1 {
			weekdayHeader += strings.Repeat(" ", weekGap)
		}
	}
	weekdayHeader += "  " + mutedStyle.Render(fmt.Sprintf("%-*s", summaryWidth, "summary"))

	var rows []string
	rows = append(rows, weekHeader)
	rows = append(rows, weekdayHeader)

	shown := 0
	maxRows := innerHeight - 4 // week header + weekday header + blank + legend
	if maxRows < 3 {
		maxRows = 3
	}
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if shown >= maxRows {
				break
			}
			if t.Type == models.TrackerText {
				continue
			}
			icon := trackerTypeIcon(t.Type)
			label := fmt.Sprintf("%s %-*s", icon, labelWidth-2, truncate(t.Name, labelWidth-3))
			line := label + "  "

			logged, hit, total := 0, 0, 0
			var sum float64
			for i, d := range dates {
				if i > 0 && i%7 == 0 {
					line += strings.Repeat(" ", weekGap)
				}
				cell := " "
				e, has := byDate[d]
				switch t.Type {
				case models.TrackerBinary:
					if has {
						total++
						if v, ok := e.Data[t.ID].(bool); ok && v {
							cell = activeStyle.Render("■")
							logged++
							hit++
						} else {
							cell = missStyle.Render("·")
						}
					} else {
						cell = mutedStyle.Render("·")
					}
				case models.TrackerDuration, models.TrackerNumeric, models.TrackerCount, models.TrackerRating:
					if has {
						if v, ok := e.Data[t.ID].(float64); ok && v > 0 {
							total++
							logged++
							sum += v
							if t.Target != nil && v >= *t.Target {
								cell = successStyle.Render("■")
								hit++
							} else {
								cell = partialStyle.Render("▪")
							}
						} else if _, ok := e.Data[t.ID]; ok {
							total++
							cell = missStyle.Render("·")
						} else {
							cell = mutedStyle.Render("·")
						}
					} else {
						cell = mutedStyle.Render("·")
					}
				}
				line += cell
			}

			// Summary column.
			var summary string
			switch t.Type {
			case models.TrackerBinary:
				summary = fmt.Sprintf("%d/%d", hit, days)
			default:
				if logged == 0 {
					summary = "—"
				} else {
					avg := sum / float64(logged)
					summary = fmt.Sprintf("avg %s", formatAverageFixed2(avg))
				}
			}
			line += "  " + mutedStyle.Render(fmt.Sprintf("%-*s", summaryWidth, summary))
			rows = append(rows, line)
			shown++
		}
		if shown >= maxRows {
			break
		}
	}

	// Legend.
	rows = append(rows, "")
	legend := strings.Repeat(" ", labelWidth+2) +
		activeStyle.Render("■") + mutedStyle.Render(" done  ") +
		successStyle.Render("■") + mutedStyle.Render(" target met  ") +
		partialStyle.Render("▪") + mutedStyle.Render(" logged  ") +
		missStyle.Render("·") + mutedStyle.Render(" missed / no entry  ") +
		mutedStyle.Render("→ today")
	rows = append(rows, legend)
	return strings.Join(rows, "\n")
}

func trackerTypeIcon(t models.TrackerType) string {
	switch t {
	case models.TrackerBinary:
		return "◉"
	case models.TrackerDuration:
		return "◷"
	case models.TrackerNumeric:
		return "#"
	case models.TrackerCount:
		return "Σ"
	case models.TrackerRating:
		return "★"
	}
	return " "
}

func formatNumberShort(v float64) string {
	if v >= 100 {
		return fmt.Sprintf("%.0f", v)
	}
	if v >= 10 {
		return fmt.Sprintf("%.1f", v)
	}
	return fmt.Sprintf("%.2f", v)
}

func weekdayInitial(w time.Weekday) string {
	return []string{"S", "M", "T", "W", "T", "F", "S"}[int(w)]
}

// ─── Hero visual #3: Weekday Rhythm ──────────────────────────────────────
//
// Tracker × weekday grid. Each cell is a heat block colored by completion
// fraction. Reveals "I'm bad at Mondays" / "great at Sundays" patterns.

func renderHeroWeekdayRhythm(m Model, innerWidth, innerHeight int) string {
	p := palette()
	if len(m.entries) == 0 {
		return mutedLine("No entries yet.")
	}

	labelWidth := 22
	cellWidth := 5
	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))

	days := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	header := strings.Repeat(" ", labelWidth+2)
	for _, d := range days {
		header += mutedStyle.Render(centerPad(d, cellWidth))
	}

	var rows []string
	rows = append(rows, header)
	shown := 0
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			if shown >= innerHeight-3 {
				break
			}
			w := db.BinaryWeekdayConsistency(m.entries, t.ID)
			label := fmt.Sprintf("%-*s", labelWidth, truncate(t.Name, labelWidth-1))
			line := label + "  "
			for i := 0; i < 7; i++ {
				line += rhythmCell(w[i], cellWidth, p)
			}
			rows = append(rows, line)
			shown++
		}
	}
	if shown == 0 {
		return mutedLine("Add a binary tracker to see the weekday rhythm.")
	}

	// Legend row
	rows = append(rows, "")
	legend := strings.Repeat(" ", labelWidth+2) +
		mutedStyle.Render("0%  ") +
		rhythmCell(0, cellWidth, p) +
		rhythmCell(25, cellWidth, p) +
		rhythmCell(50, cellWidth, p) +
		rhythmCell(75, cellWidth, p) +
		rhythmCell(100, cellWidth, p) +
		mutedStyle.Render(" 100%")
	rows = append(rows, legend)
	return strings.Join(rows, "\n")
}

func rhythmCell(pct float64, width int, p ThemePalette) string {
	// Graded intensity: empty → deep green via palette.
	levels := []string{" ", "░", "▒", "▓", "█"}
	colors := []string{p.HeatmapInactive, p.HeatmapInactive, p.ChartPrimary, p.HeatmapActive, p.Success}
	idx := int(math.Round((pct / 100) * 4))
	if idx < 0 {
		idx = 0
	}
	if idx > 4 {
		idx = 4
	}
	inner := strings.Repeat(levels[idx], width-1)
	styled := lipgloss.NewStyle().Foreground(lipgloss.Color(colors[idx])).Render(inner)
	return styled + " "
}

func centerPad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	pad := width - len(s)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

// ─── Hero visual #4: Momentum Podium ─────────────────────────────────────
//
// Big leaderboard of numeric trackers by recent-vs-prior delta, with fat
// bars so the up/down magnitudes read at a glance.

func renderHeroMomentumPodium(m Model, innerWidth, innerHeight int) string {
	p := palette()
	trackerLabel := map[string]string{}
	trackerLookup := map[string]models.Tracker{}
	var trackerIDs []string
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type == models.TrackerDuration || t.Type == models.TrackerCount || t.Type == models.TrackerNumeric {
				trackerIDs = append(trackerIDs, t.ID)
				trackerLabel[t.ID] = t.Name
				trackerLookup[t.ID] = t
			}
		}
	}
	rows := db.MomentumAccelerationRanking(m.entries, trackerIDs, 7)
	if len(rows) == 0 {
		return mutedLine("Need at least 14 entries on numeric trackers for momentum.")
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Delta > rows[j].Delta })
	maxAbs := 0.0
	for _, r := range rows {
		if math.Abs(r.Delta) > maxAbs {
			maxAbs = math.Abs(r.Delta)
		}
	}
	if maxAbs == 0 {
		maxAbs = 1
	}

	labelW := 24
	barWidth := innerWidth - labelW - 20
	if barWidth < 8 {
		barWidth = 8
	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	var out []string
	out = append(out, mutedStyle.Render(fmt.Sprintf("%-*s  %s  %-*s  %s",
		labelW, "Tracker", " ", barWidth, "Δ (7d vs prior 7d)", "value")))
	maxRows := innerHeight - 2
	for i, r := range rows {
		if i >= maxRows {
			break
		}
		t := trackerLookup[r.TrackerID]
		n := int(math.Round((math.Abs(r.Delta) / maxAbs) * float64(barWidth)))
		if n < 1 {
			n = 1
		}
		if n > barWidth {
			n = barWidth
		}
		color := p.Success
		sign := "▲"
		if r.Delta < 0 {
			color = p.Danger
			sign = "▼"
		}
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true)
		bar := style.Render(strings.Repeat("█", n)) + strings.Repeat(" ", barWidth-n)
		value := fmt.Sprintf("%+.1f %s", r.Delta, trackerUnit(t))
		out = append(out, fmt.Sprintf("%-*s  %s  %s  %s",
			labelW, truncate(trackerLabel[r.TrackerID], labelW),
			style.Render(sign), bar, mutedStyle.Render(value)))
	}
	return strings.Join(out, "\n")
}

// ─── Hero visual #5: Month at a Glance ───────────────────────────────────
//
// Calendar grid for the current month. Each day is colored by how many
// binary trackers were completed (none → muted, all → success).

func renderHeroMonthCalendar(m Model, innerWidth, innerHeight int) string {
	p := palette()
	if m.config == nil {
		return mutedLine("No config loaded.")
	}

	binaryIDs := collectBinaryIDs(m.config)
	now := time.Now()
	year, month, _ := now.Date()

	// Build a set of date → completion fraction.
	firstOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()

	// Precompute per-day binary fraction.
	fracByDate := make(map[string]float64, daysInMonth)
	entryByDate := make(map[string]bool, daysInMonth)
	for _, e := range m.entries {
		if len(e.Date) >= 7 && e.Date[:7] == fmt.Sprintf("%04d-%02d", year, int(month)) {
			entryByDate[e.Date] = true
			if len(binaryIDs) == 0 {
				fracByDate[e.Date] = 1.0
				continue
			}
			done := 0
			for _, id := range binaryIDs {
				if v, ok := e.Data[id].(bool); ok && v {
					done++
				}
			}
			fracByDate[e.Date] = float64(done) / float64(len(binaryIDs))
		}
	}

	// Intensity colors: none logged → muted; partial → primary; all → success.
	cellColor := func(date string) string {
		if !entryByDate[date] {
			return p.Muted
		}
		frac := fracByDate[date]
		switch {
		case frac >= 1.0:
			return p.Success
		case frac >= 0.66:
			return p.Primary
		case frac >= 0.33:
			return p.ChartPrimary
		default:
			return p.Danger
		}
	}

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Primary)).Bold(true)

	title := headerStyle.Render(fmt.Sprintf("%s %d", month.String(), year))

	weekdays := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	header := mutedStyle.Render(strings.Join(weekdays, "  "))

	// startDow: weekday of first day of month (0=Sun).
	startDow := int(firstOfMonth.Weekday())
	var rows []string
	var cells []string

	// Blank leading cells.
	for i := 0; i < startDow; i++ {
		cells = append(cells, "  ")
	}

	today := now.Format("2006-01-02")
	for day := 1; day <= daysInMonth; day++ {
		date := fmt.Sprintf("%04d-%02d-%02d", year, int(month), day)
		label := fmt.Sprintf("%2d", day)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(cellColor(date)))
		if date == today {
			style = style.Bold(true).Underline(true)
		}
		cells = append(cells, style.Render(label))

		if (startDow+day)%7 == 0 || day == daysInMonth {
			// Pad last row.
			for len(cells) < 7 {
				cells = append(cells, "  ")
			}
			rows = append(rows, strings.Join(cells, "  "))
			cells = nil
		}
	}

	var out []string
	out = append(out, title)
	out = append(out, header)
	for _, r := range rows {
		out = append(out, r)
	}
	// Legend.
	out = append(out, "")
	legend := mutedStyle.Render("● ") +
		lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger)).Render("low") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartPrimary)).Render("partial") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color(p.Primary)).Render("good") + "  " +
		lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Render("all done")
	out = append(out, legend)
	return strings.Join(out, "\n")
}

// ─── helpers ─────────────────────────────────────────────────────────────

func mutedLine(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(palette().Muted)).Render(s)
}

func collectBinaryIDs(cfg *models.Config) []string {
	if cfg == nil {
		return nil
	}
	var out []string
	for _, c := range cfg.Categories {
		for _, t := range c.Trackers {
			if t.Type == models.TrackerBinary {
				out = append(out, t.ID)
			}
		}
	}
	return out
}

func newestEntryDate(entries []models.Entry) time.Time {
	var newest time.Time
	for i, e := range entries {
		parsed, err := time.Parse("2006-01-02", e.Date)
		if err != nil {
			continue
		}
		if i == 0 || parsed.After(newest) {
			newest = parsed
		}
	}
	if newest.IsZero() {
		newest = time.Now()
	}
	return newest
}
