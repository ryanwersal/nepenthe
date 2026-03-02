package state

import "testing"

func TestAddExclusion(t *testing.T) {
	s := &State{Version: 1}
	AddExclusion(s, "/a", "dependencies", "sticky", "Node.js")
	if len(s.Exclusions) != 1 {
		t.Fatalf("expected 1 exclusion, got %d", len(s.Exclusions))
	}
	if s.Exclusions[0].Path != "/a" {
		t.Errorf("expected path /a, got %s", s.Exclusions[0].Path)
	}

	// Dedup: adding same path again should not add a second entry
	AddExclusion(s, "/a", "dependencies", "sticky", "Node.js")
	if len(s.Exclusions) != 1 {
		t.Errorf("expected 1 exclusion after dedup, got %d", len(s.Exclusions))
	}

	// Different path should be added
	AddExclusion(s, "/b", "dev-caches", "sticky", "Xcode")
	if len(s.Exclusions) != 2 {
		t.Errorf("expected 2 exclusions, got %d", len(s.Exclusions))
	}
}

func TestRemoveExclusion(t *testing.T) {
	s := &State{Version: 1}
	AddExclusion(s, "/a", "dependencies", "sticky", "Node.js")
	AddExclusion(s, "/b", "dev-caches", "sticky", "Xcode")

	RemoveExclusion(s, "/a")
	if len(s.Exclusions) != 1 {
		t.Fatalf("expected 1 exclusion after removal, got %d", len(s.Exclusions))
	}
	if s.Exclusions[0].Path != "/b" {
		t.Errorf("expected remaining path /b, got %s", s.Exclusions[0].Path)
	}

	// Removing non-existent path should be a no-op
	RemoveExclusion(s, "/nonexistent")
	if len(s.Exclusions) != 1 {
		t.Errorf("expected 1 exclusion after no-op removal, got %d", len(s.Exclusions))
	}
}

func TestIsTracked(t *testing.T) {
	s := &State{Version: 1}
	AddExclusion(s, "/a", "dependencies", "sticky", "Node.js")

	if !IsTracked(s, "/a") {
		t.Error("expected /a to be tracked")
	}
	if IsTracked(s, "/b") {
		t.Error("expected /b to not be tracked")
	}
}

func TestClearAll(t *testing.T) {
	s := &State{Version: 1}
	AddExclusion(s, "/a", "dependencies", "sticky", "Node.js")
	AddExclusion(s, "/b", "dev-caches", "sticky", "Xcode")

	removed := ClearAll(s)
	if len(removed) != 2 {
		t.Errorf("expected 2 removed, got %d", len(removed))
	}
	if len(s.Exclusions) != 0 {
		t.Errorf("expected 0 exclusions after clear, got %d", len(s.Exclusions))
	}
}
