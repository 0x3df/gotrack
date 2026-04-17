package tui

import "testing"

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
