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
	if m.reviewMonthly {
		cards = m.buildReviewCards(time.Now(), true, layout)
	} else {
		cards = m.buildReviewCards(time.Now(), false, layout)
	}
	grid := renderCardGrid(cards, m.width)

	label := "This week"
	if m.reviewMonthly {
		label = "This month"
	}
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color(p.Primary)).
		Bold(true).
		Render(fmt.Sprintf("— %s (press 'w' to toggle) —", label))
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
	return cards
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
