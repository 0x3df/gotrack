package db

import (
	"os"
	"testing"
)

func TestLoadConfig_NoWorkspaceConfigured(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}
	if cfg != nil {
		t.Fatalf("LoadConfig() cfg = %#v, want nil", cfg)
	}
}

func TestLoadConfig_NormalizesLegacyTrackerUnits(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	workspace := t.TempDir()
	if err := SetWorkspacePath(workspace); err != nil {
		t.Fatalf("SetWorkspacePath() error = %v", err)
	}

	path, err := GetConfigPath()
	if err != nil {
		t.Fatalf("GetConfigPath() error = %v", err)
	}

	legacy := `{
	  "setup_complete": true,
	  "categories": [
	    {
	      "id": "cat-1",
	      "name": "Health",
	      "color": "#10B981",
	      "order": 0,
	      "trackers": [
	        {"id": "t1", "name": "Deep Work", "type": "duration", "order": 0},
	        {"id": "t2", "name": "Pushups", "type": "count", "order": 1},
	        {"id": "t3", "name": "Weight", "type": "numeric", "order": 2},
	        {"id": "t4", "name": "Body Fat", "type": "numeric", "order": 3},
	        {"id": "t5", "name": "Journal", "type": "text", "order": 4}
	      ]
	    }
	  ]
	}`
	if err := os.WriteFile(path, []byte(legacy), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v, want nil", err)
	}

	got := map[string]string{}
	for _, cat := range cfg.Categories {
		for _, tracker := range cat.Trackers {
			got[tracker.Name] = tracker.Unit
		}
	}

	want := map[string]string{
		"Deep Work": "minutes",
		"Pushups":   "count",
		"Weight":    "lb",
		"Body Fat":  "value",
		"Journal":   "",
	}

	for name, unit := range want {
		if got[name] != unit {
			t.Fatalf("tracker %q unit = %q, want %q", name, got[name], unit)
		}
	}
}
