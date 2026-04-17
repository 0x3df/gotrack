package tui

import (
	"math/rand"
	"testing"
)

func TestStepStars_AdvancesPositions(t *testing.T) {
	stars := []fallingStar{{X: 10, Y: 2, VX: -1, VY: 1, Trail: 2}}

	next := stepStars(stars, 40, 20)
	if len(next) != 1 {
		t.Fatalf("len(stepStars()) = %d, want 1", len(next))
	}
	if next[0].X != 9 || next[0].Y != 3 {
		t.Fatalf("stepStars() = %+v, want X=9 Y=3", next[0])
	}
}

func TestStepStars_UsesVelocityVector(t *testing.T) {
	stars := []fallingStar{
		{X: 5, Y: 5, VX: 1, VY: 0, Trail: 2},
		{X: 5, Y: 5, VX: 0, VY: -1, Trail: 2},
	}

	next := stepStars(stars, 20, 20)
	if len(next) != 2 {
		t.Fatalf("len(stepStars()) = %d, want 2", len(next))
	}
	if next[0].X != 6 || next[0].Y != 5 {
		t.Fatalf("first star moved to (%d,%d), want (6,5)", next[0].X, next[0].Y)
	}
	if next[1].X != 5 || next[1].Y != 4 {
		t.Fatalf("second star moved to (%d,%d), want (5,4)", next[1].X, next[1].Y)
	}
	if next[0].VX != 1 || next[0].VY != 0 {
		t.Fatalf("first star velocity = (%d,%d), want (1,0)", next[0].VX, next[0].VY)
	}
	if next[1].VX != 0 || next[1].VY != -1 {
		t.Fatalf("second star velocity = (%d,%d), want (0,-1)", next[1].VX, next[1].VY)
	}
}

func TestMaybeSpawnStar_SpawnsFromAnyEdge(t *testing.T) {
	const width, height = 80, 24
	rng := rand.New(rand.NewSource(7))
	stars := []fallingStar{}

	var sawTop, sawRight, sawBottom, sawLeft bool

	recordNewStar := func(s fallingStar) {
		t.Helper()
		switch s.SpawnEdge {
		case SpawnEdgeTop:
			sawTop = true
			if s.Y != 0 {
				t.Fatalf("SpawnEdgeTop but Y=%d, want 0", s.Y)
			}
		case SpawnEdgeRight:
			sawRight = true
			if s.X != width-1 {
				t.Fatalf("SpawnEdgeRight but X=%d, want %d", s.X, width-1)
			}
		case SpawnEdgeBottom:
			sawBottom = true
			if s.Y != height-1 {
				t.Fatalf("SpawnEdgeBottom but Y=%d, want %d", s.Y, height-1)
			}
		case SpawnEdgeLeft:
			sawLeft = true
			if s.X != 0 {
				t.Fatalf("SpawnEdgeLeft but X=%d, want 0", s.X)
			}
		default:
			t.Fatalf("unexpected SpawnEdge %v for new star %+v", s.SpawnEdge, s)
		}
	}

	for range 200 {
		before := len(stars)
		stars = maybeSpawnStar(stars, width, height, rng)
		if len(stars) > before {
			recordNewStar(stars[len(stars)-1])
		}
		stars = stepStars(stars, width, height)
	}

	if !sawTop || !sawRight || !sawBottom || !sawLeft {
		t.Fatalf("SpawnEdge coverage: top=%v right=%v bottom=%v left=%v, want all true",
			sawTop, sawRight, sawBottom, sawLeft)
	}
}

func TestStepStars_RemovesExpired(t *testing.T) {
	stars := []fallingStar{{X: 0, Y: 5, VX: -1, VY: 1, Trail: 1}}

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
