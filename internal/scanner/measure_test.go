package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMeasureFileCount(t *testing.T) {
	dir := t.TempDir()

	// Create known files
	for _, name := range []string{"a.txt", "b.txt", "c.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("hello"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a subdirectory with one file
	subdir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "d.txt"), []byte("world"), 0o644); err != nil {
		t.Fatal(err)
	}

	count, err := measureFileCount(dir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 4 {
		t.Errorf("expected 4 files, got %d", count)
	}
}

func TestMeasureFileCountEmptyDir(t *testing.T) {
	dir := t.TempDir()

	count, err := measureFileCount(dir)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 files in empty dir, got %d", count)
	}
}
