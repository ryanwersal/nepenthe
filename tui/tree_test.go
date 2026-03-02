package tui

import (
	"testing"

	"github.com/ryanwersal/nepenthe/internal/scanner"
)

func makeResults(paths ...string) []scanner.ScanResult {
	results := make([]scanner.ScanResult, len(paths))
	for i, p := range paths {
		results[i] = scanner.ScanResult{
			Path:      p,
			Ecosystem: "Test",
			Category:  scanner.CategoryDependencies,
		}
	}
	return results
}

func TestBuildDirectoryTree(t *testing.T) {
	results := makeResults(
		"/home/user/project1/node_modules",
		"/home/user/project2/node_modules",
	)

	root, resultToNode := buildDirectoryTree(results)
	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if len(resultToNode) != 2 {
		t.Errorf("expected 2 result-to-node mappings, got %d", len(resultToNode))
	}

	// Both results should be reachable
	for i := range results {
		if _, ok := resultToNode[i]; !ok {
			t.Errorf("expected result %d in resultToNode", i)
		}
	}
}

func TestBuildEcosystemTree(t *testing.T) {
	results := []scanner.ScanResult{
		{Path: "/a/node_modules", Ecosystem: "Node.js", Category: scanner.CategoryDependencies},
		{Path: "/b/node_modules", Ecosystem: "Node.js", Category: scanner.CategoryDependencies},
		{Path: "/c/target", Ecosystem: "Rust", Category: scanner.CategoryDependencies},
	}

	root, resultToNode := buildEcosystemTree(results)
	if root == nil {
		t.Fatal("expected non-nil root")
	}

	// Should have 2 ecosystem groups
	if len(root.Children) != 2 {
		t.Errorf("expected 2 ecosystem groups, got %d", len(root.Children))
	}
	if len(resultToNode) != 3 {
		t.Errorf("expected 3 result-to-node mappings, got %d", len(resultToNode))
	}

	// Check ecosystem groups are sorted
	if root.Children[0].Label != "Node.js" {
		t.Errorf("expected first group to be Node.js, got %s", root.Children[0].Label)
	}
	if root.Children[1].Label != "Rust" {
		t.Errorf("expected second group to be Rust, got %s", root.Children[1].Label)
	}
}

func TestFlattenTree(t *testing.T) {
	results := makeResults(
		"/home/user/project/node_modules",
	)

	root, _ := buildDirectoryTree(results)
	rows := flattenTree(root)
	if len(rows) == 0 {
		t.Fatal("expected non-empty rows")
	}

	// Should contain at least one leaf row
	hasLeaf := false
	for _, row := range rows {
		if row.IsLeaf {
			hasLeaf = true
			break
		}
	}
	if !hasLeaf {
		t.Error("expected at least one leaf row")
	}
}

func TestCompressTree(t *testing.T) {
	// Build a tree where single-child chains should be compressed
	results := makeResults(
		"/a/b/c/d/leaf",
	)

	root, _ := buildDirectoryTree(results)

	// After compression, the chain a/b/c/d should be compressed
	// Root should have fewer levels than 5
	rows := flattenTree(root)
	// Should have at most 2 rows (compressed chain + leaf)
	if len(rows) > 2 {
		t.Errorf("expected at most 2 rows after compression, got %d", len(rows))
	}
}

func TestLeafIndices(t *testing.T) {
	results := makeResults(
		"/home/a/node_modules",
		"/home/b/node_modules",
		"/home/c/node_modules",
	)

	root, _ := buildDirectoryTree(results)
	indices := leafIndices(root)

	if len(indices) != 3 {
		t.Errorf("expected 3 leaf indices, got %d", len(indices))
	}

	// All indices should be in range [0, 3)
	for _, idx := range indices {
		if idx < 0 || idx >= 3 {
			t.Errorf("leaf index %d out of range", idx)
		}
	}
}
