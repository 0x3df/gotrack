package tui

import "testing"

func TestStepStars_AdvancesPositions(t *testing.T) {
	stars := []fallingStar{{X: 10, Y: 2, Trail: 2}}

	next := stepStars(stars, 40, 20)
	if len(next) != 1 {
		t.Fatalf("len(stepStars()) = %d, want 1", len(next))
	}
	if next[0].X != 9 || next[0].Y != 3 {
		t.Fatalf("stepStars() = %+v, want X=9 Y=3", next[0])
	}
}

func TestStepStars_RemovesExpired(t *testing.T) {
	stars := []fallingStar{{X: 0, Y: 5, Trail: 1}}

	next := stepStars(stars, 10, 5)
	if len(next) != 0 {
		t.Fatalf("len(stepStars()) = %d, want 0", len(next))
	}
}

func TestApplyStarfieldOverlay_DisabledReturnsForeground(t *testing.T) {
	fg := "  BOX  "
	bg := "..*...."

	if got := applyStarfieldOverlay(fg, bg, false); got != fg {
		t.Fatalf("applyStarfieldOverlay(disabled) = %q, want %q", got, fg)
	}
}

func TestApplyStarfieldOverlay_PreservesForegroundContent(t *testing.T) {
	fg := "  BOX  \n       "
	bg := "..*....\n***...."

	got := applyStarfieldOverlay(fg, bg, true)
	want := "..BOX..\n***...."
	if got != want {
		t.Fatalf("applyStarfieldOverlay() = %q, want %q", got, want)
	}
}
