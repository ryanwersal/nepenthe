package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Roots             []string          `toml:"roots"`
	EnabledCategories []string          `toml:"enabledCategories"`
	CustomFixedPaths  []CustomFixedPath `toml:"customFixedPaths"`
	Schedule          Schedule          `toml:"schedule"`
	Consent           Consent           `toml:"consent"`
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
		if errors.Is(err, fs.ErrNotExist) {
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
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, cfgPath)
}

func CategoryEnabled(cfg Config, category string) bool {
	return slices.Contains(cfg.EnabledCategories, category)
}

func ToggleCategory(category string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	found := false
	for i, c := range cfg.EnabledCategories {
		if c == category {
			cfg.EnabledCategories = append(cfg.EnabledCategories[:i], cfg.EnabledCategories[i+1:]...)
			found = true
			break
		}
	}
	if !found {
		cfg.EnabledCategories = append(cfg.EnabledCategories, category)
	}
	return cfg, saveConfig(cfg)
}

func AddRoot(root string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	for _, r := range cfg.Roots {
		if r == root {
			return cfg, nil
		}
	}
	cfg.Roots = append(cfg.Roots, root)
	return cfg, saveConfig(cfg)
}

func RemoveRoot(root string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	if len(cfg.Roots) <= 1 {
		return cfg, fmt.Errorf("cannot remove last root")
	}
	for i, r := range cfg.Roots {
		if r == root {
			cfg.Roots = append(cfg.Roots[:i], cfg.Roots[i+1:]...)
			break
		}
	}
	return cfg, saveConfig(cfg)
}

func AddCustomFixedPath(path, ecosystem string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	cfg.CustomFixedPaths = append(cfg.CustomFixedPaths, CustomFixedPath{
		Path:      path,
		Ecosystem: ecosystem,
	})
	return cfg, saveConfig(cfg)
}

func RemoveCustomFixedPath(path string) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	for i, cf := range cfg.CustomFixedPaths {
		if cf.Path == path {
			cfg.CustomFixedPaths = append(cfg.CustomFixedPaths[:i], cfg.CustomFixedPaths[i+1:]...)
			break
		}
	}
	return cfg, saveConfig(cfg)
}

func SetScheduleInterval(seconds int) (Config, error) {
	cfg, err := Load()
	if err != nil {
		return cfg, err
	}
	cfg.Schedule.IntervalSeconds = seconds
	return cfg, saveConfig(cfg)
}
