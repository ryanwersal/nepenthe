package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

type State struct {
	Version    int                `json:"version"`
	Exclusions []TrackedExclusion `json:"exclusions"`
}

type TrackedExclusion struct {
	Path       string `json:"path"`
	Category   string `json:"category"`
	Type       string `json:"type"` // "sticky" or "fixed"
	ExcludedAt string `json:"excludedAt"`
	Ecosystem  string `json:"ecosystem"`
}

func statePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".config", "nepenthe", "state.json"), nil
}

func Load() (State, error) {
	s := State{Version: 1}
	path, err := statePath()
	if err != nil {
		return s, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return s, nil
		}
		return s, err
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return State{Version: 1}, err
	}
	return s, nil
}

func Save(s State) error {
	path, err := statePath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func AddExclusion(s *State, path string, category string, typ string, ecosystem string) {
	for _, e := range s.Exclusions {
		if e.Path == path {
			return
		}
	}
	s.Exclusions = append(s.Exclusions, TrackedExclusion{
		Path:       path,
		Category:   category,
		Type:       typ,
		ExcludedAt: time.Now().UTC().Format(time.RFC3339),
		Ecosystem:  ecosystem,
	})
}

func RemoveExclusion(s *State, path string) {
	filtered := make([]TrackedExclusion, 0, len(s.Exclusions))
	for _, e := range s.Exclusions {
		if e.Path != path {
			filtered = append(filtered, e)
		}
	}
	s.Exclusions = filtered
}

func ClearAll(s *State) []TrackedExclusion {
	removed := make([]TrackedExclusion, len(s.Exclusions))
	copy(removed, s.Exclusions)
	s.Exclusions = nil
	return removed
}

func IsTracked(s *State, path string) bool {
	return slices.ContainsFunc(s.Exclusions, func(e TrackedExclusion) bool {
		return e.Path == path
	})
}
