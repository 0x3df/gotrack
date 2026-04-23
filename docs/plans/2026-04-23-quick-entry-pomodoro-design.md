# Quick Entry And Pomodoro Design

## Goal

Add fast item logging and a pomodoro timer that allocates elapsed work time to a selected duration tracker.

## Recommended Design

Quick entry is a TUI dashboard shortcut for logging one tracker value to today without opening the full daily entry form. The flow is: press `x`, choose a tracker, enter a value, and save. It reuses the same tracker lookup and coercion semantics as `gotrack log`, then refreshes the dashboard.

Pomodoro is a TUI dashboard shortcut for starting a timed work session against one duration tracker. The flow is: press `p`, choose a duration tracker, choose or enter a session length, run the timer, and either let it complete or press a key to end/pause. When the session stops, GoTrack adds elapsed minutes to today's existing value for that tracker. It never replaces prior time.

## Data Flow

GoTrack already stores daily entries as a JSON map keyed by tracker ID, and duration trackers are stored as `float64` minutes. Both features should use that existing storage. A small db helper should merge values into today's entry and, for pomodoro, add elapsed minutes to the existing duration value.

## UI

Dashboard keys:

- `x`: quick entry
- `p`: pomodoro

Both flows use existing Huh forms where possible. The active pomodoro view shows selected tracker, remaining time, elapsed time, and key help for end/cancel.

## Testing

Add data-layer tests for additive duration allocation and single-value quick logging. Add model-level tests only where behavior is deterministic without running a real terminal timer.
