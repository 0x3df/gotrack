package tui

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"dailytrack/db"
	"dailytrack/models"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewReview() string {
	layout := dashboardLayoutForWidth(m.width)
	p := palette()
	if len(m.entries) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).
			Render("No entries yet — nothing to review.")
	}

	var cards []string
	var label string
	if m.reviewCustom && !m.reviewCustomStart.IsZero() {
		cards = m.buildReviewCardsCustom(m.reviewCustomStart, m.reviewCustomEnd, layout)
		label = fmt.Sprintf("%s → %s", m.reviewCustomStart.Format("Jan 2"), m.reviewCustomEnd.Format("Jan 2"))
	} else if m.reviewMonthly {
		cards = m.buildReviewCards(time.Now(), true, layout)
		label = "This month"
	} else {
		cards = m.buildReviewCards(time.Now(), false, layout)
		label = "This week"
	}
	grid := renderCardGrid(cards, m.width)

	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Render(fmt.Sprintf("— %s (press 'w' to cycle: week/month/custom) —", label))
	return lipgloss.JoinVertical(lipgloss.Center, header, "", grid)
}

func (m Model) buildReviewCards(anchor time.Time, monthly bool, layout dashboardLayout) []string {
	p := palette()
	accent := p.Primary

	var start, end, prevStart, prevEnd time.Time
	if monthly {
		start, end = db.MonthBounds(anchor)
		prev := anchor.AddDate(0, -1, 0)
		prevStart, prevEnd = db.MonthBounds(prev)
	} else {
		start, end = db.WeekBounds(anchor)
		prev := anchor.AddDate(0, 0, -7)
		prevStart, prevEnd = db.WeekBounds(prev)
	}

	// Card 1: logging summary
	var cards []string
	logged := 0
	totalDays := int(end.Sub(start).Hours()/24) + 1
	sStr := start.Format("2006-01-02")
	eStr := end.Format("2006-01-02")
	for _, e := range m.entries {
		if e.Date >= sStr && e.Date <= eStr {
			logged++
		}
	}
	summary := fmt.Sprintf(
		"Entries logged: %d / %d\nRange: %s → %s",
		logged, totalDays,
		start.Format("Jan 2"), end.Format("Jan 2"),
	)
	if logged > 0 {
		summary += "\n" + ProgressBar(float64(logged)/float64(totalDays)*100, layout.CardWidth-8)
	}
	cards = append(cards, renderCard("Overview", accent, summary, layout.CardWidth, layout.CardHeight))

	// Card 2+: binary trackers — hit-rate with arrow.
	binNameCount := map[string]int{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type == models.TrackerBinary {
				binNameCount[t.Name]++
			}
		}
	}
	binaryLines := []string{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			h, tot := db.BinaryHitsInRange(m.entries, t.ID, start, end)
			ph, ptot := db.BinaryHitsInRange(m.entries, t.ID, prevStart, prevEnd)
			cur := 0.0
			if tot > 0 {
				cur = float64(h) / float64(tot) * 100
			}
			prev := 0.0
			if ptot > 0 {
				prev = float64(ph) / float64(ptot) * 100
			}
			label := t.Name
			if binNameCount[t.Name] > 1 {
				label = c.Name + " · " + t.Name
			}
			arrow := wowArrow(cur-prev, p, models.TrackerBinary)
			binaryLines = append(binaryLines, fmt.Sprintf("%-22s  %d/%d (%.0f%%) %s",
				truncate(label, 22), h, tot, cur, arrow))
		}
	}
	if len(binaryLines) > 0 {
		cards = append(cards, renderCard("Binary trackers", accent,
			strings.Join(binaryLines, "\n"), layout.CardWidth, layout.CardHeight))
	}

	// Card 3+: numeric trackers. Duration/Count are cumulative (sum); Numeric
	// is a measurement (avg). Labels disambiguate duplicates by category.
	type numRow struct {
		label   string
		delta   float64
		content string
	}
	nameCount := map[string]int{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			nameCount[t.Name]++
		}
	}
	var numRows []numRow
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerDuration && t.Type != models.TrackerCount && t.Type != models.TrackerNumeric {
				continue
			}
			var cur, prev float64
			var curCount int
			if t.Type == models.TrackerNumeric {
				cur, curCount = db.AvgInRange(m.entries, t.ID, start, end)
				prev, _ = db.AvgInRange(m.entries, t.ID, prevStart, prevEnd)
				if curCount == 0 && prev == 0 {
					continue
				}
			} else {
				cur = db.SumInRange(m.entries, t.ID, start, end)
				prev = db.SumInRange(m.entries, t.ID, prevStart, prevEnd)
				if cur == 0 && prev == 0 {
					continue
				}
			}
			label := t.Name
			if nameCount[t.Name] > 1 {
				label = c.Name + " · " + t.Name
			}
			arrow := wowArrow(cur-prev, p, t.Type)
			unit := trackerUnit(t)
			line := fmt.Sprintf("%-22s  %s%s  %s",
				truncate(label, 22),
				formatDisplayValue(cur, t.Type),
				unitSpace(unit),
				arrow,
			)
			if t.Type != models.TrackerNumeric {
				var target *float64
				if monthly {
					target = t.MonthlyTarget
				} else {
					target = t.WeeklyTarget
				}
				if target != nil && *target > 0 {
					pct := (cur / *target) * 100
					line += "\n  " + ProgressBar(pct, layout.CardWidth-10)
					// Pace indicator: project current pace over the full period.
					periodDays := end.Sub(start).Hours()/24 + 1
					elapsed := anchor.Sub(start).Hours()/24 + 1
					if elapsed > 0 && elapsed <= periodDays {
						projected := cur / elapsed * periodDays
						pacer := lipgloss.NewStyle()
						if projected >= *target {
							pacer = pacer.Foreground(lipgloss.Color(p.Success))
							line += "\n  " + pacer.Render(fmt.Sprintf("On pace → %.0f %s (target %.0f)", projected, trackerUnit(t), *target))
						} else {
							pacer = pacer.Foreground(lipgloss.Color(p.Danger))
							line += "\n  " + pacer.Render(fmt.Sprintf("Behind pace → %.0f %s (target %.0f)", projected, trackerUnit(t), *target))
						}
					}
				}
			}
			numRows = append(numRows, numRow{label, cur - prev, line})
		}
	}
	sort.Slice(numRows, func(i, j int) bool { return math.Abs(numRows[i].delta) > math.Abs(numRows[j].delta) })
	if len(numRows) > 0 {
		var lines []string
		for _, r := range numRows {
			lines = append(lines, r.content)
		}
		cards = append(cards, renderCard("Numeric trackers", accent,
			strings.Join(lines, "\n"), layout.CardWidth, layout.CardHeight))
	}

	// Card 4: highlight — biggest mover
	if len(numRows) > 0 {
		top := numRows[0]
		direction := "rose"
		if top.delta < 0 {
			direction = "fell"
		}
		content := fmt.Sprintf("Biggest mover:\n\n  %s %s by %s vs last %s",
			top.label, direction, formatRounded(math.Abs(top.delta), 1),
			map[bool]string{true: "month", false: "week"}[monthly])
		cards = append(cards, renderCard("Highlight", accent, content, layout.CardWidth, layout.CardHeight))
	}

	// Card 5+: journal feed — text trackers for the last 14 days.
	if journal := m.buildJournalFeed(14); journal != "" {
		cards = append(cards, renderCard("Journal", accent, journal, layout.CardWidth, layout.CardHeight))
	}

	return cards
}

func (m Model) buildReviewCardsCustom(start, end time.Time, layout dashboardLayout) []string {
	p := palette()
	accent := p.Primary

	// Prior period of equal length immediately before start.
	span := end.Sub(start)
	prevStart := start.Add(-span - 24*time.Hour)
	prevEnd := start.Add(-24 * time.Hour)

	sStr := start.Format("2006-01-02")
	eStr := end.Format("2006-01-02")
	var cards []string

	// Card 1: overview.
	logged := 0
	totalDays := int(end.Sub(start).Hours()/24) + 1
	for _, e := range m.entries {
		if e.Date >= sStr && e.Date <= eStr {
			logged++
		}
	}
	summary := fmt.Sprintf(
		"Entries logged: %d / %d\nRange: %s → %s",
		logged, totalDays,
		start.Format("Jan 2"), end.Format("Jan 2"),
	)
	if logged > 0 {
		summary += "\n" + ProgressBar(float64(logged)/float64(totalDays)*100, layout.CardWidth-8)
	}
	cards = append(cards, renderCard("Overview", accent, summary, layout.CardWidth, layout.CardHeight))

	// Binary trackers.
	binNameCount := map[string]int{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type == models.TrackerBinary {
				binNameCount[t.Name]++
			}
		}
	}
	var binaryLines []string
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerBinary {
				continue
			}
			h, tot := db.BinaryHitsInRange(m.entries, t.ID, start, end)
			ph, ptot := db.BinaryHitsInRange(m.entries, t.ID, prevStart, prevEnd)
			cur := 0.0
			if tot > 0 {
				cur = float64(h) / float64(tot) * 100
			}
			prev := 0.0
			if ptot > 0 {
				prev = float64(ph) / float64(ptot) * 100
			}
			label := t.Name
			if binNameCount[t.Name] > 1 {
				label = c.Name + " · " + t.Name
			}
			arrow := wowArrow(cur-prev, p, models.TrackerBinary)
			binaryLines = append(binaryLines, fmt.Sprintf("%-22s  %d/%d (%.0f%%) %s",
				truncate(label, 22), h, tot, cur, arrow))
		}
	}
	if len(binaryLines) > 0 {
		cards = append(cards, renderCard("Binary trackers", accent,
			strings.Join(binaryLines, "\n"), layout.CardWidth, layout.CardHeight))
	}

	// Numeric trackers.
	type numRow struct {
		label   string
		delta   float64
		content string
	}
	nameCount := map[string]int{}
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			nameCount[t.Name]++
		}
	}
	var numRows []numRow
	for _, c := range m.config.Categories {
		for _, t := range c.Trackers {
			if t.Type != models.TrackerDuration && t.Type != models.TrackerCount && t.Type != models.TrackerNumeric {
				continue
			}
			var cur, prev float64
			var curCount int
			if t.Type == models.TrackerNumeric {
				cur, curCount = db.AvgInRange(m.entries, t.ID, start, end)
				prev, _ = db.AvgInRange(m.entries, t.ID, prevStart, prevEnd)
				if curCount == 0 && prev == 0 {
					continue
				}
			} else {
				cur = db.SumInRange(m.entries, t.ID, start, end)
				prev = db.SumInRange(m.entries, t.ID, prevStart, prevEnd)
				if cur == 0 && prev == 0 {
					continue
				}
			}
			label := t.Name
			if nameCount[t.Name] > 1 {
				label = c.Name + " · " + t.Name
			}
			arrow := wowArrow(cur-prev, p, t.Type)
			unit := trackerUnit(t)
			line := fmt.Sprintf("%-22s  %s%s  %s",
				truncate(label, 22),
				formatDisplayValue(cur, t.Type),
				unitSpace(unit),
				arrow,
			)
			numRows = append(numRows, numRow{label, cur - prev, line})
		}
	}
	sort.Slice(numRows, func(i, j int) bool { return math.Abs(numRows[i].delta) > math.Abs(numRows[j].delta) })
	if len(numRows) > 0 {
		var lines []string
		for _, r := range numRows {
			lines = append(lines, r.content)
		}
		cards = append(cards, renderCard("Numeric trackers", accent,
			strings.Join(lines, "\n"), layout.CardWidth, layout.CardHeight))
	}

	if len(numRows) > 0 {
		top := numRows[0]
		direction := "rose"
		if top.delta < 0 {
			direction = "fell"
		}
		content := fmt.Sprintf("Biggest mover:\n\n  %s %s by %s vs prior period",
			top.label, direction, formatRounded(math.Abs(top.delta), 1))
		cards = append(cards, renderCard("Highlight", accent, content, layout.CardWidth, layout.CardHeight))
	}

	if journal := m.buildJournalFeedInRange(start, end); journal != "" {
		cards = append(cards, renderCard("Journal", accent, journal, layout.CardWidth, layout.CardHeight))
	}

	return cards
}

func (m Model) buildJournalFeedInRange(start, end time.Time) string {
	type textTracker struct {
		id, name string
	}
	var textTrackers []textTracker
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type == models.TrackerText {
				textTrackers = append(textTrackers, textTracker{t.ID, t.Name})
			}
		}
	}
	if len(textTrackers) == 0 {
		return ""
	}
	sStr := start.Format("2006-01-02")
	eStr := end.Format("2006-01-02")
	p := palette()
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Primary)).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	indent := strings.Repeat(" ", 12)
	var out []string
	for _, e := range m.entries {
		if e.Date < sStr || e.Date > eStr {
			continue
		}
		var fields []string
		for _, tt := range textTrackers {
			if v, ok := e.Data[tt.id].(string); ok && strings.TrimSpace(v) != "" {
				fields = append(fields, labelStyle.Render(tt.name+": ")+v)
			}
		}
		if len(fields) == 0 {
			continue
		}
		line := dateStyle.Render(e.Date) + "  " + fields[0]
		for _, f := range fields[1:] {
			line += "\n" + indent + f
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n\n")
}

func (m Model) buildJournalFeed(days int) string {
	// Collect text tracker name+ID pairs.
	type textTracker struct {
		id, name string
	}
	var textTrackers []textTracker
	for _, cat := range m.config.Categories {
		for _, t := range cat.Trackers {
			if t.Type == models.TrackerText {
				textTrackers = append(textTrackers, textTracker{t.ID, t.Name})
			}
		}
	}
	if len(textTrackers) == 0 {
		return ""
	}

	cutoff := time.Now().AddDate(0, 0, -(days - 1)).Format("2006-01-02")
	p := palette()
	dateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Primary)).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted))
	indent := strings.Repeat(" ", 12) // align to date column width

	var out []string
	for _, e := range m.entries {
		if e.Date < cutoff {
			break
		}
		var fields []string
		for _, tt := range textTrackers {
			if v, ok := e.Data[tt.id].(string); ok && strings.TrimSpace(v) != "" {
				fields = append(fields, labelStyle.Render(tt.name+": ")+v)
			}
		}
		if len(fields) == 0 {
			continue
		}
		line := dateStyle.Render(e.Date) + "  " + fields[0]
		for _, f := range fields[1:] {
			line += "\n" + indent + f
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n\n")
}

func wowArrow(delta float64, p ThemePalette, tt models.TrackerType) string {
	if math.Abs(delta) < 0.05 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Muted)).Render("→ flat")
	}
	digits := 0
	if tt == models.TrackerNumeric {
		digits = 1
	}
	if delta > 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Success)).Render(fmt.Sprintf("▲ +%s", formatRounded(delta, digits)))
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(p.Danger)).Render(fmt.Sprintf("▼ %s", formatRounded(delta, digits)))
}

func formatRounded(v float64, digits int) string {
	return strconv.FormatFloat(v, 'f', digits, 64)
}

func formatDisplayValue(v float64, tt models.TrackerType) string {
	if tt == models.TrackerNumeric {
		return formatRounded(roundAverageUp2(v), 2)
	}
	return formatRounded(v, 0)
}

func unitSpace(unit string) string {
	if strings.TrimSpace(unit) == "" {
		return ""
	}
	return " " + unit
}
