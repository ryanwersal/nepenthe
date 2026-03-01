package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Roots                []string             `toml:"roots"`
	EnabledTiers         []int                `toml:"enabledTiers"`
	CustomSentinelRules  []CustomSentinelRule `toml:"customSentinelRules"`
	CustomFixedPaths     []CustomFixedPath    `toml:"customFixedPaths"`
	Schedule             Schedule             `toml:"schedule"`
	Consent              Consent              `toml:"consent"`
	Concurrency          Concurrency          `toml:"concurrency"`
}

type Concurrency struct {
	ScanWorkers    int `toml:"scanWorkers"`
	MeasureWorkers int `toml:"measureWorkers"`
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
	Tier2 *bool `toml:"tier2"`
	Tier3 *bool `toml:"tier3"`
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
		Roots:        []string{home},
		EnabledTiers: []int{1, 2, 3},
		Schedule:     Schedule{IntervalSeconds: 86400},
		Concurrency:  Concurrency{ScanWorkers: 8, MeasureWorkers: 4},
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
	if len(fileCfg.EnabledTiers) > 0 {
		cfg.EnabledTiers = fileCfg.EnabledTiers
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
	cfg.Consent = fileCfg.Consent
	if fileCfg.Concurrency.ScanWorkers > 0 {
		cfg.Concurrency.ScanWorkers = fileCfg.Concurrency.ScanWorkers
	}
	if fileCfg.Concurrency.MeasureWorkers > 0 {
		cfg.Concurrency.MeasureWorkers = fileCfg.Concurrency.MeasureWorkers
	}

	return cfg, nil
}

func SaveConsent(tier int, value bool) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	switch tier {
	case 2:
		cfg.Consent.Tier2 = &value
	case 3:
		cfg.Consent.Tier3 = &value
	}

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

func TierEnabled(cfg Config, tier int) bool {
	for _, t := range cfg.EnabledTiers {
		if t == tier {
			return true
		}
	}
	return false
}
