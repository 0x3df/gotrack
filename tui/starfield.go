package tui

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type starfieldTickMsg time.Time

// SpawnEdge records which screen edge a star originated from.
type SpawnEdge int

const (
	SpawnEdgeNone SpawnEdge = iota
	SpawnEdgeTop
	SpawnEdgeRight
	SpawnEdgeBottom
	SpawnEdgeLeft
)

type fallingStar struct {
	X         int
	Y         int
	VX        int
	VY        int
	Trail     int
	Bright    bool
	SpawnEdge SpawnEdge
}

type twinkleStar struct {
	X     int
	Y     int
	Phase int
}

type burstParticle struct {
	X      int
	Y      int
	VX     int
	VY     int
	Life   int
	Bright bool
}

func starfieldTick() tea.Cmd {
	return tea.Tick(125*time.Millisecond, func(t time.Time) tea.Msg {
		return starfieldTickMsg(t)
	})
}

func stepStars(stars []fallingStar, width, height int) []fallingStar {
	var next []fallingStar
	for _, star := range stars {
		star.X += star.VX
		star.Y += star.VY
		t := star.Trail
		if star.Y < -t || star.Y >= height+t {
			continue
		}
		if star.X < -t || star.X >= width+t {
			continue
		}
		next = append(next, star)
	}
	return next
}

func maybeSpawnStar(stars []fallingStar, width, height int, rng *rand.Rand) []fallingStar {
	if rng == nil || width <= 0 || height <= 0 {
		return stars
	}
	if len(stars) >= maxInt(4, height/3) {
		return stars
	}
	if rng.Intn(3) != 0 {
		return stars
	}

	edge := SpawnEdge(int(SpawnEdgeTop) + rng.Intn(4))
	star := fallingStar{
		Trail:     2 + rng.Intn(2),
		Bright:    rng.Intn(2) == 0,
		SpawnEdge: edge,
	}

	switch edge {
	case SpawnEdgeTop:
		star.X = rng.Intn(width)
		star.Y = 0
		star.VX = []int{-1, 0, 1}[rng.Intn(3)]
		star.VY = 1
	case SpawnEdgeRight:
		star.X = width - 1
		star.Y = rng.Intn(height)
		star.VX = -1
		star.VY = []int{-1, 0, 1}[rng.Intn(3)]
	case SpawnEdgeBottom:
		star.X = rng.Intn(width)
		star.Y = height - 1
		star.VX = []int{-1, 0, 1}[rng.Intn(3)]
		star.VY = -1
	default: // SpawnEdgeLeft
		star.X = 0
		star.Y = rng.Intn(height)
		star.VX = 1
		star.VY = []int{-1, 0, 1}[rng.Intn(3)]
	}

	return append(stars, star)
}

func trailGlyphForVector(vx, vy int) string {
	if vx == 0 && vy == 0 {
		return "."
	}
	if vx == 0 {
		return "|"
	}
	if vy == 0 {
		return "-"
	}
	if (vx > 0 && vy > 0) || (vx < 0 && vy < 0) {
		return "\\"
	}
	return "/"
}

func stepTwinkles(twinkles []twinkleStar, width, height int, rng *rand.Rand) []twinkleStar {
	if width <= 0 || height <= 0 {
		return nil
	}
	if rng == nil {
		return twinkles
	}
	next := twinkles[:0]
	for _, t := range twinkles {
		if t.X < 0 || t.X >= width || t.Y < 0 || t.Y >= height {
			continue
		}
		t.Phase = (t.Phase + 1) % 12
		next = append(next, t)
	}
	// Keep ambient twinkles sparse.
	target := maxInt(6, (width*height)/550)
	for len(next) < target {
		next = append(next, twinkleStar{
			X:     rng.Intn(width),
			Y:     rng.Intn(height),
			Phase: rng.Intn(12),
		})
	}
	return next
}

func spawnBurstParticles(width, height, cx, cy int, rng *rand.Rand) []burstParticle {
	if rng == nil || width <= 0 || height <= 0 {
		return nil
	}
	if cx < 0 || cx >= width {
		cx = width / 2
	}
	if cy < 0 || cy >= height {
		cy = height / 3
	}
	particles := make([]burstParticle, 0, 10)
	dirs := [][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}, {1, 1}, {-1, -1}, {1, -1}, {-1, 1}}
	for i := 0; i < len(dirs); i++ {
		v := dirs[i]
		particles = append(particles, burstParticle{
			X:      cx,
			Y:      cy,
			VX:     v[0],
			VY:     v[1],
			Life:   3 + rng.Intn(2),
			Bright: i%2 == 0,
		})
	}
	return particles
}

func stepBurstParticles(parts []burstParticle, width, height int) []burstParticle {
	var next []burstParticle
	for _, p := range parts {
		p.X += p.VX
		p.Y += p.VY
		p.Life--
		if p.Life <= 0 || p.X < 0 || p.X >= width || p.Y < 0 || p.Y >= height {
			continue
		}
		next = append(next, p)
	}
	return next
}

func renderStarfieldCanvas(width, height int, stars []fallingStar, twinkles []twinkleStar, bursts []burstParticle) string {
	if width <= 0 || height <= 0 {
		return ""
	}

	p := palette()
	grid := make([][]string, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]string, width)
		for x := 0; x < width; x++ {
			grid[y][x] = " "
		}
	}

	dim := lipgloss.NewStyle().Foreground(lipgloss.Color(p.StarDim))
	bright := lipgloss.NewStyle().Foreground(lipgloss.Color(p.StarBright)).Bold(true)
	trailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.StarTrail))
	burstStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(p.ChartSecondary)).Bold(true)

	for _, t := range twinkles {
		if t.X < 0 || t.X >= width || t.Y < 0 || t.Y >= height {
			continue
		}
		g := "."
		switch {
		case t.Phase%6 == 0:
			g = "*"
		case t.Phase%3 == 0:
			g = "+"
		}
		grid[t.Y][t.X] = dim.Render(g)
	}

	for _, star := range stars {
		if star.X >= 0 && star.X < width && star.Y >= 0 && star.Y < height {
			glyph := "."
			if star.Bright {
				glyph = "*"
				grid[star.Y][star.X] = bright.Render(glyph)
			} else {
				grid[star.Y][star.X] = dim.Render(glyph)
			}
		}
		for i := 1; i <= star.Trail; i++ {
			tx := star.X - star.VX*i
			ty := star.Y - star.VY*i
			if tx < 0 || tx >= width || ty < 0 || ty >= height {
				continue
			}
			glyph := "."
			if i == 1 {
				glyph = trailGlyphForVector(star.VX, star.VY)
			}
			grid[ty][tx] = trailStyle.Render(glyph)
		}
	}

	for _, b := range bursts {
		if b.X < 0 || b.X >= width || b.Y < 0 || b.Y >= height {
			continue
		}
		g := "."
		if b.Bright {
			g = "*"
		}
		grid[b.Y][b.X] = burstStyle.Render(g)
	}

	lines := make([]string, height)
	for y := range grid {
		lines[y] = strings.Join(grid[y], "")
	}
	return strings.Join(lines, "\n")
}

func applyStarfieldOverlay(foreground, background string, enabled bool) string {
	if !enabled || background == "" {
		return foreground
	}

	fgLines := strings.Split(foreground, "\n")
	bgLines := strings.Split(background, "\n")
	height := maxInt(len(fgLines), len(bgLines))
	out := make([]string, height)

	for i := 0; i < height; i++ {
		fg := lineAt(fgLines, i)
		bg := lineAt(bgLines, i)
		stripped := ansi.Strip(fg)
		if strings.TrimSpace(stripped) == "" {
			out[i] = bg
			continue
		}

		lead := countLeadingSpaces(stripped)
		trail := countTrailingSpaces(stripped)
		fgWidth := ansi.StringWidth(fg)
		midWidth := fgWidth - lead - trail
		if midWidth < 0 {
			midWidth = 0
		}

		prefix := ansi.Cut(bg, 0, lead)
		middle := ansi.Cut(fg, lead, lead+midWidth)
		suffix := ansi.Cut(bg, fgWidth-trail, fgWidth)
		out[i] = prefix + middle + suffix
	}

	return strings.Join(out, "\n")
}

func lineAt(lines []string, idx int) string {
	if idx < 0 || idx >= len(lines) {
		return ""
	}
	return lines[idx]
}

func countLeadingSpaces(s string) int {
	count := 0
	for _, r := range s {
		if r != ' ' {
			break
		}
		count++
	}
	return count
}

func countTrailingSpaces(s string) int {
	count := 0
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] != ' ' {
			break
		}
		count++
	}
	return count
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
