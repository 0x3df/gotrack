package tui

import (
	"testing"
	"time"
)

func TestNormalizeEntryDate_Shortcuts(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	if got, err := normalizeEntryDate("t"); err != nil || got != today {
		t.Fatalf("normalizeEntryDate(t) = %q, %v; want %q, nil", got, err, today)
	}
	if got, err := normalizeEntryDate("today"); err != nil || got != today {
		t.Fatalf("normalizeEntryDate(today) = %q, %v", got, err)
	}
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	if got, err := normalizeEntryDate("y"); err != nil || got != yesterday {
		t.Fatalf("normalizeEntryDate(y) = %q, %v; want %q", got, err, yesterday)
	}
	threeAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
	if got, err := normalizeEntryDate("-3"); err != nil || got != threeAgo {
		t.Fatalf("normalizeEntryDate(-3) = %q, %v; want %q", got, err, threeAgo)
	}
	if _, err := normalizeEntryDate("tomorrow"); err == nil {
		t.Fatal("expected error for unknown shortcut")
	}
}

func TestValidateEntryDate(t *testing.T) {
	if err := validateEntryDate("2026-04-17"); err != nil {
		t.Fatalf("validateEntryDate(valid) error = %v, want nil", err)
	}
	if err := validateEntryDate("04/17/2026"); err == nil {
		t.Fatal("validateEntryDate(bad format) error = nil, want non-nil")
	}
	if err := validateEntryDate("2026-02-31"); err == nil {
		t.Fatal("validateEntryDate(impossible date) error = nil, want non-nil")
	}
}

func TestPrefillEntryHelpers(t *testing.T) {
	data := map[string]interface{}{
		"bool":   true,
		"num":    float64(42),
		"rating": float64(4),
		"text":   "note",
	}

	if !prefillBoolValue(data, "bool") {
		t.Fatal("prefillBoolValue() = false, want true")
	}
	if got := prefillStringValue(data, "num"); got != "42" {
		t.Fatalf("prefillStringValue(num) = %q, want %q", got, "42")
	}
	if got := prefillStringValue(data, "text"); got != "note" {
		t.Fatalf("prefillStringValue(text) = %q, want %q", got, "note")
	}
	if got := prefillIntValue(data, "rating", 3); got != 4 {
		t.Fatalf("prefillIntValue() = %d, want 4", got)
	}
	if got := prefillIntValue(data, "missing", 3); got != 3 {
		t.Fatalf("prefillIntValue(missing) = %d, want 3", got)
	}
}
