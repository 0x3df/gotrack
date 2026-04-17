# Visual Card Framing and Omnidirectional Starfield Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make dashboard visual cards render cleanly with borders that actually frame their chart/text content, and make the starfield spawn/move from all screen edges instead of only the top-right diagonal.

**Architecture:** Keep layout responsibilities in `tui/layout.go` and star simulation responsibilities in `tui/starfield.go`. Fix card rendering by constraining/wrapping inner content to the card body width/height before border rendering, then evolve star simulation from a single hardcoded vector to direction-aware stars with edge-based spawning and matching trail rendering. Validate behavior with focused unit tests in `tui/layout_test.go` and `tui/starfield_test.go`.

**Tech Stack:** Go, Bubble Tea, Lipgloss, ansi width helpers from `github.com/charmbracelet/x/ansi`, Go `testing` package.

---

### Task 1: Lock in failing tests for card framing and content wrapping

**Files:**
- Modify: `tui/layout_test.go`
- Modify: `tui/layout.go` (later task implementation target)
- Test: `tui/layout_test.go`

- [ ] **Step 1: Write the failing test for wrapped long content staying inside card borders**

```go
func TestRenderCard_WrapsLongLinesInsideBorder(t *testing.T) {
	card := renderCard("Title", palette().Primary, "abcdefghijklmnopqrstuvwxyz", 20, 0)
	lines := strings.Split(card, "\n")
	if len(lines) < 3 {
		t.Fatalf("renderCard() produced %d lines, want at least 3", len(lines))
	}

	innerWidth := ansi.StringWidth(lines[1])
	for i, line := range lines {
		if ansi.StringWidth(line) != innerWidth {
			t.Fatalf("line %d width = %d, want %d; line=%q", i, ansi.StringWidth(line), innerWidth, line)
		}
	}
}
```

- [ ] **Step 2: Write the failing test for honoring fixed card height without content escaping**

```go
func TestRenderCard_RespectsFixedHeight(t *testing.T) {
	content := strings.Join([]string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
	}, "\n")

	card := renderCard("Title", palette().Primary, content, 30, 8)
	lines := strings.Split(card, "\n")
	if len(lines) != 8 {
		t.Fatalf("len(renderCard lines) = %d, want 8", len(lines))
	}
}
```

- [ ] **Step 3: Run tests to verify they fail before implementation**

Run: `go test ./tui -run 'TestRenderCard_(WrapsLongLinesInsideBorder|RespectsFixedHeight)' -v`

Expected: FAIL with assertions showing line widths mismatch and/or height mismatch due to current unconstrained content rendering.

- [ ] **Step 4: Commit the failing tests**

```bash
git add tui/layout_test.go
git commit -m "test: capture card wrapping and framing regressions"
```


### Task 2: Implement card content constraints so borders frame visuals cleanly

**Files:**
- Modify: `tui/layout.go`
- Test: `tui/layout_test.go`

- [ ] **Step 1: Implement minimal card-body sizing helper in `renderCard`**

```go
func renderCard(title, accentColor, content string, width, height int) string {
	p := palette()
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(p.Border)).
		Padding(1, dashboardCardInnerPadding).
		Width(width)

	contentWidth := maxInt(1, width-(dashboardCardInnerPadding*2)-2) // 2 border columns
	bodyStyle := lipgloss.NewStyle().
		Width(contentWidth).
		MaxWidth(contentWidth)

	if height > 0 {
		style = style.Height(height)
		bodyHeight := maxInt(1, height-4) // title + spacer + vertical padding + border
		bodyStyle = bodyStyle.MaxHeight(bodyHeight)
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(accentColor)).
		Bold(true)

	body := bodyStyle.Render(content)
	return style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render(title),
		"",
		body,
	))
}
```

- [ ] **Step 2: Run targeted tests**

Run: `go test ./tui -run 'TestRenderCard_(WrapsLongLinesInsideBorder|RespectsFixedHeight)' -v`

Expected: PASS for both card rendering tests.

- [ ] **Step 3: Run broader layout tests**

Run: `go test ./tui -run 'TestDashboardLayoutForWidth|TestRenderCard_' -v`

Expected: PASS for existing layout width/column behavior and new card tests.

- [ ] **Step 4: Commit the card rendering fix**

```bash
git add tui/layout.go tui/layout_test.go
git commit -m "fix: constrain card bodies so borders frame visual content"
```


### Task 3: Lock in failing tests for omnidirectional star movement and spawning

**Files:**
- Modify: `tui/starfield_test.go`
- Modify: `tui/starfield.go` (later task implementation target)
- Test: `tui/starfield_test.go`

- [ ] **Step 1: Extend tests to cover velocity-driven movement in all directions**

```go
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
}
```

- [ ] **Step 2: Add spawn tests asserting stars can originate from multiple edges**

```go
func TestMaybeSpawnStar_SpawnsFromAnyEdge(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	stars := []fallingStar{}

	for i := 0; i < 200; i++ {
		stars = maybeSpawnStar(stars, 80, 24, rng)
		stars = stepStars(stars, 80, 24)
	}

	sawLeft, sawRight, sawTop, sawBottom := false, false, false, false
	for _, s := range stars {
		if s.SpawnEdge == "left" {
			sawLeft = true
		}
		if s.SpawnEdge == "right" {
			sawRight = true
		}
		if s.SpawnEdge == "top" {
			sawTop = true
		}
		if s.SpawnEdge == "bottom" {
			sawBottom = true
		}
	}

	if !(sawLeft && sawRight && sawTop && sawBottom) {
		t.Fatalf("spawn edges seen: left=%v right=%v top=%v bottom=%v, want all true", sawLeft, sawRight, sawTop, sawBottom)
	}
}
```

- [ ] **Step 3: Run tests to verify failure**

Run: `go test ./tui -run 'TestStepStars_UsesVelocityVector|TestMaybeSpawnStar_SpawnsFromAnyEdge' -v`

Expected: FAIL because `fallingStar` has no vector/edge fields and spawn logic currently only top-right diagonal.

- [ ] **Step 4: Commit the failing starfield tests**

```bash
git add tui/starfield_test.go
git commit -m "test: define omnidirectional starfield behavior"
```


### Task 4: Implement omnidirectional star simulation and trail rendering

**Files:**
- Modify: `tui/starfield.go`
- Modify: `tui/starfield_test.go` (if compile adjustments needed)
- Test: `tui/starfield_test.go`

- [ ] **Step 1: Add direction fields and edge metadata to `fallingStar`**

```go
type fallingStar struct {
	X        int
	Y        int
	VX       int
	VY       int
	Trail    int
	Bright   bool
	SpawnEdge string
}
```

- [ ] **Step 2: Update movement and expiry to use velocity vectors**

```go
func stepStars(stars []fallingStar, width, height int) []fallingStar {
	var next []fallingStar
	for _, star := range stars {
		star.X += star.VX
		star.Y += star.VY
		if star.Y < -star.Trail || star.Y >= height+star.Trail {
			continue
		}
		if star.X < -star.Trail || star.X >= width+star.Trail {
			continue
		}
		next = append(next, star)
	}
	return next
}
```

- [ ] **Step 3: Implement multi-edge spawn logic with inward motion vectors**

```go
func maybeSpawnStar(stars []fallingStar, width, height int, rng *rand.Rand) []fallingStar {
	if rng == nil || width <= 0 || height <= 0 {
		return stars
	}
	if len(stars) >= maxInt(4, height/3) || rng.Intn(3) != 0 {
		return stars
	}

	edge := []string{"top", "right", "bottom", "left"}[rng.Intn(4)]
	star := fallingStar{Trail: 2 + rng.Intn(2), Bright: rng.Intn(2) == 0, SpawnEdge: edge}

	switch edge {
	case "top":
		star.X = rng.Intn(width)
		star.Y = 0
		star.VX = []int{-1, 0, 1}[rng.Intn(3)]
		star.VY = 1
	case "right":
		star.X = width - 1
		star.Y = rng.Intn(height)
		star.VX = -1
		star.VY = []int{-1, 0, 1}[rng.Intn(3)]
	case "bottom":
		star.X = rng.Intn(width)
		star.Y = height - 1
		star.VX = []int{-1, 0, 1}[rng.Intn(3)]
		star.VY = -1
	default: // left
		star.X = 0
		star.Y = rng.Intn(height)
		star.VX = 1
		star.VY = []int{-1, 0, 1}[rng.Intn(3)]
	}
	return append(stars, star)
}
```

- [ ] **Step 4: Update trail rendering to follow opposite of velocity direction**

```go
for i := 1; i <= star.Trail; i++ {
	tx := star.X - (star.VX * i)
	ty := star.Y - (star.VY * i)
	if tx < 0 || tx >= width || ty < 0 || ty >= height {
		continue
	}
	glyph := "."
	if i == 1 {
		glyph = trailGlyphForVector(star.VX, star.VY)
	}
	grid[ty][tx] = trail.Render(glyph)
}
```

```go
func trailGlyphForVector(vx, vy int) string {
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
```

- [ ] **Step 5: Run starfield tests**

Run: `go test ./tui -run 'TestStepStars_|TestMaybeSpawnStar_|TestApplyStarfieldOverlay_' -v`

Expected: PASS for movement/spawn/overlay test suite.

- [ ] **Step 6: Commit starfield implementation**

```bash
git add tui/starfield.go tui/starfield_test.go
git commit -m "feat: make starfield spawn and travel from all directions"
```


### Task 5: Verify end-to-end dashboard behavior and update docs text

**Files:**
- Modify: `README.md`
- Test: `tui/layout_test.go`, `tui/starfield_test.go`, full `tui` package tests

- [ ] **Step 1: Update README feature wording**

```md
- **Ambient mode:** optional omnidirectional starfield background for the dashboard.
```

- [ ] **Step 2: Run full package tests for regressions**

Run: `go test ./tui ./db ./models ./integrations -v`

Expected: PASS across all listed packages.

- [ ] **Step 3: Run full repository tests**

Run: `go test ./...`

Expected: PASS with `ok` for all packages.

- [ ] **Step 4: Manual runtime smoke test**

Run:

```bash
go run . 
```

Expected:
- Dashboard cards show borders that enclose chart/text content cleanly.
- Long chart lines no longer spill visually outside card frame.
- With starfield enabled, stars enter from top/right/bottom/left and move inward across varied trajectories.

- [ ] **Step 5: Commit documentation touch-ups and final verification state**

```bash
git add README.md
git commit -m "docs: describe omnidirectional starfield background"
```


## Self-Review Checklist

- Spec coverage:
  - Card containers looking bad / not wrapped: covered by Task 1 and Task 2 with explicit renderCard width+height constraint tests and implementation.
  - Starfall everywhere, not just top-right: covered by Task 3 and Task 4 with multi-edge spawning, vector movement, and trail updates.
  - Regression confidence: covered by Task 5 targeted and full test runs plus manual smoke test.
- Placeholder scan:
  - No `TODO`, `TBD`, or "handle appropriately" placeholders remain.
  - Each code-changing step includes concrete code blocks.
  - Each validation step includes explicit command + expected output.
- Type/signature consistency:
  - `fallingStar` additions (`VX`, `VY`, `SpawnEdge`) are introduced in Task 4 and referenced consistently in Task 3/4.
  - `renderCard` remains same signature and behavior contract, only internals change.


Plan complete and saved to `docs/superpowers/plans/2026-04-17-visual-cards-and-starfield.md`. Two execution options:

1. Subagent-Driven (recommended) - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. Inline Execution - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
