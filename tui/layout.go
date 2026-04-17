package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	dashboardOuterPadding     = 4
	dashboardGridGap          = 2
	dashboardMaxContentWidth  = 124
	dashboardMinContentWidth  = 40
	dashboardMinTwoColumnCard = 54
	dashboardCardHeight       = 18
	dashboardCardInnerPadding = 2
	dashboardCardVerticalGap  = 1
)

type dashboardLayout struct {
	ContentWidth int
	Columns      int
	CardWidth    int
	CardHeight   int
	Gap          int
}

func dashboardLayoutForWidth(totalWidth int) dashboardLayout {
	usableWidth := maxInt(totalWidth-dashboardOuterPadding, totalWidth-2)
	if usableWidth <= 0 {
		usableWidth = maxInt(totalWidth, 1)
	}

	contentWidth := minInt(usableWidth, dashboardMaxContentWidth)
	if usableWidth >= dashboardMinContentWidth {
		contentWidth = maxInt(contentWidth, dashboardMinContentWidth)
	}

	layout := dashboardLayout{
		ContentWidth: contentWidth,
		Columns:      1,
		CardWidth:    contentWidth,
		CardHeight:   dashboardCardHeight,
		Gap:          dashboardGridGap,
	}

	if contentWidth >= dashboardMinTwoColumnCard*2+dashboardGridGap {
		layout.Columns = 2
		layout.CardWidth = (contentWidth - dashboardGridGap) / 2
	}

	return layout
}

func renderCard(title, accentColor, content string, width, height int) string {
	p := palette()
	blockWidth := maxInt(width-2, 1)
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(1, dashboardCardInnerPadding).
		Width(blockWidth)

	if height > 0 {
		// Style.Height applies to the block before borders are drawn; total rows are inner + top + bottom border.
		innerH := height - style.GetBorderTopSize() - style.GetBorderBottomSize()
		style = style.Height(maxInt(innerH, 1))
	}

	hFrame := style.GetHorizontalFrameSize()
	contentWidth := maxInt(width-hFrame, 1)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(accentColor)).
		Bold(true).
		Width(contentWidth).
		MaxWidth(contentWidth)

	titleRendered := titleStyle.Render(title)
	titleLines := strings.Split(strings.TrimSuffix(titleRendered, "\n"), "\n")

	bodyStyle := lipgloss.NewStyle().
		Width(contentWidth).
		MaxWidth(contentWidth)

	if height > 0 {
		const titleBodyGapLines = 1 // blank JoinVertical element between title and body
		reserved := style.GetBorderTopSize() + style.GetBorderBottomSize() +
			style.GetPaddingTop() + style.GetPaddingBottom() +
			titleBodyGapLines + len(titleLines)
		bodyMax := height - reserved
		if bodyMax < 1 {
			bodyMax = 1
		}
		bodyStyle = bodyStyle.MaxHeight(bodyMax)
	}

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		titleRendered,
		"",
		bodyStyle.Render(content),
	))
}

func renderCardGrid(cards []string, totalWidth int) string {
	if len(cards) == 0 {
		return ""
	}

	layout := dashboardLayoutForWidth(totalWidth)
	var rows []string
	for i := 0; i < len(cards); i += layout.Columns {
		end := minInt(i+layout.Columns, len(cards))
		rows = append(rows, joinCardRow(cards[i:end], layout.Gap))
	}

	return strings.Join(rows, strings.Repeat("\n", dashboardCardVerticalGap+1))
}

func joinCardRow(cards []string, gap int) string {
	if len(cards) == 0 {
		return ""
	}
	if len(cards) == 1 {
		return cards[0]
	}

	parts := make([]string, 0, len(cards)*2-1)
	space := strings.Repeat(" ", gap)
	for i, card := range cards {
		if i > 0 {
			parts = append(parts, space)
		}
		parts = append(parts, card)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func dashboardChartWidth(cardWidth int) int {
	return maxInt(24, minInt(cardWidth-8, 44))
}

func dashboardProgressWidth(cardWidth int) int {
	return maxInt(20, minInt(cardWidth-10, 34))
}
