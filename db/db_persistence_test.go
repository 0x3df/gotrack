package db

import "testing"

func TestInitDB_DoesNotWipeExistingEntries(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	workspace := t.TempDir()
	if err := SetWorkspacePath(workspace); err != nil {
		t.Fatalf("SetWorkspacePath() error = %v", err)
	}
	if err := InitDB(); err != nil {
		t.Fatalf("InitDB() error = %v", err)
	}
	initial := makeEntry("2026-04-20", map[string]interface{}{"persist": true})
	if err := UpsertEntry(&initial); err != nil {
		t.Fatalf("UpsertEntry() error = %v", err)
	}

	if err := InitDB(); err != nil {
		t.Fatalf("second InitDB() error = %v", err)
	}
	stored, err := GetEntryForDate("2026-04-20")
	if err != nil {
		t.Fatalf("GetEntryForDate() error = %v", err)
	}
	if stored == nil {
		t.Fatal("entry should persist across InitDB calls, got nil")
	}
}
