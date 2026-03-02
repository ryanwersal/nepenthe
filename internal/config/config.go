package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Roots               []string             `toml:"roots"`
	EnabledCategories   []string             `toml:"enabledCategories"`
	CustomSentinelRules []CustomSentinelRule `toml:"customSentinelRules"`
	CustomFixedPaths    []CustomFixedPath    `toml:"customFixedPaths"`
	Schedule            Schedule             `toml:"schedule"`
	Consent             Consent              `toml:"consent"`
}

type CustomSentinelRule struct {
	Directory string   `toml:"directory"`
	Sentinels []string `toml:"sentinels"`
	Ecosystem string   `toml:"ecosystem"`
}

type CustomFixedPath struct {
	Path        string `toml:"path"`
	Ecosystem   string `toml:"ecosystem"`
	Description string `toml:"description"`
}

type Schedule struct {
	IntervalSeconds int `toml:"intervalSeconds"`
}

type Consent struct {
	Categories map[string]bool `toml:"categories"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "nepenthe"), nil
}

func ConfigPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func DefaultConfig() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("resolving home directory: %w", err)
	}
	return Config{
		Roots:             []string{home},
		EnabledCategories: []string{"dependencies", "dev-caches", "containers"},
		Schedule:          Schedule{IntervalSeconds: 86400},
	}, nil
}

func Load() (Config, error) {
	cfg, err := DefaultConfig()
	if err != nil {
		return Config{}, err
	}

	cfgPath, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}

	var fileCfg Config
	if _, err := toml.Decode(string(data), &fileCfg); err != nil {
		return cfg, err
	}

	// Merge: file values override defaults where present
	if len(fileCfg.Roots) > 0 {
		cfg.Roots = fileCfg.Roots
	}
	if len(fileCfg.EnabledCategories) > 0 {
		cfg.EnabledCategories = fileCfg.EnabledCategories
	}
	if len(fileCfg.CustomSentinelRules) > 0 {
		cfg.CustomSentinelRules = fileCfg.CustomSentinelRules
	}
	if len(fileCfg.CustomFixedPaths) > 0 {
		cfg.CustomFixedPaths = fileCfg.CustomFixedPaths
	}
	if fileCfg.Schedule.IntervalSeconds > 0 {
		cfg.Schedule.IntervalSeconds = fileCfg.Schedule.IntervalSeconds
	}
	if len(fileCfg.Consent.Categories) > 0 {
		cfg.Consent.Categories = fileCfg.Consent.Categories
	}

	return cfg, nil
}

func SaveConsent(category string, value bool) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	if cfg.Consent.Categories == nil {
		cfg.Consent.Categories = make(map[string]bool)
	}
	cfg.Consent.Categories[category] = value

	return saveConfig(cfg)
}

func saveConfig(cfg Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	cfgPath, err := ConfigPath()
	if err != nil {
		return err
	}

	tmp := cfgPath + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, cfgPath)
}

func CategoryEnabled(cfg Config, category string) bool {
	for _, c := range cfg.EnabledCategories {
		if c == category {
			return true
		}
	}
	return false
}
