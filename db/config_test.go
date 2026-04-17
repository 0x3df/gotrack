package db

import "testing"

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
