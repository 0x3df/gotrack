# Quick Entry And Pomodoro Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a TUI quick-entry flow and a pomodoro timer that allocates completed or ended session minutes to the selected duration tracker.

**Architecture:** Reuse the existing daily entry JSON storage. Add db helpers for single tracker merge and additive duration allocation, then wire Bubble Tea states/forms for quick entry and pomodoro. Keep timer state transient in the TUI model.

**Tech Stack:** Go, Bubble Tea, Huh forms, SQLite via existing `db` package.

---

### Task 1: Data Helpers

**Files:**
- Modify: `db/entry.go`
- Test: `db/entry_test.go`

**Step 1: Write failing tests**

Add tests that:
- `UpsertEntryLog` preserves existing fields and updates one tracker for today.
- `AddDurationToEntry` adds minutes to an existing duration tracker instead of replacing it.

**Step 2: Run tests to verify failure**

Run: `go test ./db`

Expected: FAIL because `AddDurationToEntry` does not exist.

**Step 3: Implement minimal helper**

Add `AddDurationToEntry(cfg *models.Config, date string, trackerID string, minutes float64) error`.

**Step 4: Run tests**

Run: `go test ./db`

Expected: PASS.

### Task 2: TUI Quick Entry

**Files:**
- Modify: `tui/model.go`
- Modify: `tui/help.go`
- Test: `tui/model_quick_entry_test.go`

**Step 1: Write failing model tests**

Add tests for duration tracker option listing and quick-entry save behavior using the db helper.

**Step 2: Run tests to verify failure**

Run: `go test ./tui`

Expected: FAIL because quick-entry state/helpers do not exist.

**Step 3: Implement quick-entry state**

Add dashboard key `x`, forms to select tracker and value, save via `db.UpsertEntryLog`, refresh entries, and return to dashboard.

**Step 4: Run tests**

Run: `go test ./tui`

Expected: PASS.

### Task 3: Pomodoro Timer

**Files:**
- Modify: `tui/model.go`
- Modify: `tui/help.go`
- Test: `tui/model_pomodoro_test.go`

**Step 1: Write failing model tests**

Add tests for duration tracker selection and elapsed-minute allocation on stop.

**Step 2: Run tests to verify failure**

Run: `go test ./tui`

Expected: FAIL because pomodoro state/helpers do not exist.

**Step 3: Implement timer state**

Add dashboard key `p`, form to select duration tracker and minutes, ticking command, active timer view, complete/end handling, and additive save through `db.AddDurationToEntry`.

**Step 4: Run tests**

Run: `go test ./tui`

Expected: PASS.

### Task 4: Docs And Final Verification

**Files:**
- Modify: `README.md`
- Modify: `docs/FAQ.md`

**Step 1: Update docs**

Document `x` quick entry and `p` pomodoro controls.

**Step 2: Verify full suite**

Run: `go test ./...`

Expected: PASS.
