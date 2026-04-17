package tui

import "testing"

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
