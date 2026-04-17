package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDashboardLayoutForWidth_UsesOneColumnWhenNarrow(t *testing.T) {
	layout := dashboardLayoutForWidth(78)

	if layout.Columns != 1 {
		t.Fatalf("layout.Columns = %d, want 1", layout.Columns)
	}
	if layout.CardWidth != layout.ContentWidth {
		t.Fatalf("layout.CardWidth = %d, want content width %d for single-column layout", layout.CardWidth, layout.ContentWidth)
	}
}

func TestDashboardLayoutForWidth_UsesTwoWiderColumnsWhenWide(t *testing.T) {
	layout := dashboardLayoutForWidth(132)

	if layout.Columns != 2 {
		t.Fatalf("layout.Columns = %d, want 2", layout.Columns)
	}
	if layout.CardWidth <= 50 {
		t.Fatalf("layout.CardWidth = %d, want wider cards", layout.CardWidth)
	}
	if layout.CardHeight != dashboardCardHeight {
		t.Fatalf("layout.CardHeight = %d, want %d", layout.CardHeight, dashboardCardHeight)
	}
}

func TestDashboardLayoutForWidth_CapsContentWidth(t *testing.T) {
	layout := dashboardLayoutForWidth(200)

	if layout.ContentWidth > dashboardMaxContentWidth {
		t.Fatalf("layout.ContentWidth = %d, want <= %d", layout.ContentWidth, dashboardMaxContentWidth)
	}
}

func TestDashboardLayoutForWidth_DoesNotOverflowNarrowTerminal(t *testing.T) {
	layout := dashboardLayoutForWidth(32)

	if layout.ContentWidth > 32 {
		t.Fatalf("layout.ContentWidth = %d, want <= terminal width", layout.ContentWidth)
	}
}

func TestRenderCard_WrapsLongLinesInsideBorder(t *testing.T) {
	card := renderCard(
		"Title", palette().Primary, "abcdefghijklmnopqrstuvwxyz",
		20, // lipgloss block width (cells)
		0,  // height 0: no fixed vertical size
	)
	lines := strings.Split(card, "\n")
	if len(lines) < 3 {
		t.Fatalf("renderCard() produced %d lines, want at least 3", len(lines))
	}

	innerWidth := ansi.StringWidth(lines[0])
	for i, line := range lines {
		if ansi.StringWidth(line) != innerWidth {
			t.Fatalf("line %d width = %d, want %d; line=%q", i, ansi.StringWidth(line), innerWidth, line)
		}
	}
}

func TestRenderCard_RespectsFixedHeight(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
	}, "\n")

	card := renderCard(
		"Title", palette().Primary, content,
		30, // lipgloss block width (cells)
		8,  // fixed card height in terminal rows
	)
	lines := strings.Split(strings.TrimRight(card, "\n"), "\n")
	if len(lines) != 8 {
		t.Fatalf("len(lines) = %d, want 8", len(lines))
	}
}
