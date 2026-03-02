package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCategoryEnabled(t *testing.T) {
	cfg := Config{
		EnabledCategories: []string{"dependencies", "dev-caches"},
	}

	if !CategoryEnabled(cfg, "dependencies") {
		t.Error("expected dependencies to be enabled")
	}
	if !CategoryEnabled(cfg, "dev-caches") {
		t.Error("expected dev-caches to be enabled")
	}
	if CategoryEnabled(cfg, "vms") {
		t.Error("expected vms to not be enabled")
	}
	if CategoryEnabled(cfg, "") {
		t.Error("expected empty string to not be enabled")
	}
}

func TestLoadRoundTrip(t *testing.T) {
	// Set up a temp home directory so Load finds our config
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfgDir := filepath.Join(tmpHome, ".config", "nepenthe")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	tomlContent := `
roots = ["/tmp/test"]
enabledCategories = ["dependencies", "vms"]

[schedule]
intervalSeconds = 3600
`
	if err := os.WriteFile(filepath.Join(cfgDir, "config.toml"), []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// File values should override defaults
	if len(cfg.Roots) != 1 || cfg.Roots[0] != "/tmp/test" {
		t.Errorf("expected roots [/tmp/test], got %v", cfg.Roots)
	}
	if len(cfg.EnabledCategories) != 2 {
		t.Errorf("expected 2 enabled categories, got %d", len(cfg.EnabledCategories))
	}
	if !CategoryEnabled(cfg, "vms") {
		t.Error("expected vms to be enabled from file")
	}
	if cfg.Schedule.IntervalSeconds != 3600 {
		t.Errorf("expected interval 3600, got %d", cfg.Schedule.IntervalSeconds)
	}
}

func TestLoadMissingFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	// Should return defaults
	if len(cfg.Roots) != 1 || cfg.Roots[0] != tmpHome {
		t.Errorf("expected default root %s, got %v", tmpHome, cfg.Roots)
	}
	if cfg.Schedule.IntervalSeconds != 86400 {
		t.Errorf("expected default interval 86400, got %d", cfg.Schedule.IntervalSeconds)
	}
}
