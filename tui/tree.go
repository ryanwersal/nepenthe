package tui

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/ryanwersal/nepenthe/internal/scanner"
)

// GroupMode determines how scan results are grouped in the tree view.
type GroupMode int

const (
	GroupByDirectory GroupMode = iota
	GroupByEcosystem
)

// TreeNode represents a node in the tree hierarchy.
type TreeNode struct {
	Label     string
	FullPath  string
	Result    *scanner.ScanResult
	ResultIdx int // -1 for internal nodes
	Children  []*TreeNode
	Parent    *TreeNode
	SizeBytes int64
	FileCount int64
	Expanded  bool
	Depth     int
}

// FlatRow is a flattened tree node for virtual scrolling.
type FlatRow struct {
	Node   *TreeNode
	Depth  int
	IsLeaf bool
}

// buildDirectoryTree builds a trie from path components, compresses single-child
// chains, computes rollups, and sorts children alphabetically.
func buildDirectoryTree(results []scanner.ScanResult) (*TreeNode, map[int]*TreeNode) {
	root := &TreeNode{
		Label:     "/",
		FullPath:  "/",
		ResultIdx: -1,
		Expanded:  true,
	}
	resultToNode := make(map[int]*TreeNode)

	for i := range results {
		r := &results[i]
		parts := splitPath(r.Path)

		current := root
		for j, part := range parts {
			child := findChild(current, part)
			if child == nil {
				child = &TreeNode{
					Label:     part,
					FullPath:  "/" + strings.Join(parts[:j+1], "/"),
					ResultIdx: -1,
					Parent:    current,
					Expanded:  true,
					Depth:     j + 1,
				}
				current.Children = append(current.Children, child)
			}
			current = child
		}
		// Mark the leaf with the result
		current.Result = r
		current.ResultIdx = i
		current.SizeBytes = r.SizeBytes
		current.FileCount = r.FileCount
		resultToNode[i] = current
	}

	// Sort children alphabetically at each level
	sortChildren(root)

	// Compress single-child internal chains
	compressTree(root)

	// Compute rollups bottom-up
	computeRollups(root)

	// Collapse all, only root expanded
	collapseAll(root)
	root.Expanded = true

	return root, resultToNode
}

// buildEcosystemTree groups results by Ecosystem string.
func buildEcosystemTree(results []scanner.ScanResult) (*TreeNode, map[int]*TreeNode) {
	root := &TreeNode{
		Label:     "All",
		FullPath:  "",
		ResultIdx: -1,
		Expanded:  true,
	}
	resultToNode := make(map[int]*TreeNode)

	groups := make(map[string]*TreeNode)
	for i := range results {
		r := &results[i]
		eco := r.Ecosystem
		if eco == "" {
			eco = "Other"
		}

		group, ok := groups[eco]
		if !ok {
			group = &TreeNode{
				Label:     eco,
				FullPath:  eco,
				ResultIdx: -1,
				Parent:    root,
				Expanded:  false,
				Depth:     1,
			}
			groups[eco] = group
			root.Children = append(root.Children, group)
		}

		leaf := &TreeNode{
			Label:     filepath.Base(r.Path),
			FullPath:  r.Path,
			Result:    r,
			ResultIdx: i,
			Parent:    group,
			SizeBytes: r.SizeBytes,
			FileCount: r.FileCount,
			Depth:     2,
		}
		group.Children = append(group.Children, leaf)
		resultToNode[i] = leaf
	}

	// Sort ecosystem groups alphabetically
	sort.Slice(root.Children, func(i, j int) bool {
		return root.Children[i].Label < root.Children[j].Label
	})

	// Sort leaves within each group by path
	for _, group := range root.Children {
		sort.Slice(group.Children, func(i, j int) bool {
			return group.Children[i].FullPath < group.Children[j].FullPath
		})
	}

	computeRollups(root)
	return root, resultToNode
}

// flattenTree walks expanded nodes depth-first into []FlatRow for virtual scrolling.
func flattenTree(root *TreeNode) []FlatRow {
	var rows []FlatRow
	flattenNode(root, 0, &rows)
	return rows
}

func flattenNode(node *TreeNode, depth int, rows *[]FlatRow) {
	// Skip the virtual root node itself
	if node.Parent != nil {
		isLeaf := len(node.Children) == 0
		*rows = append(*rows, FlatRow{
			Node:   node,
			Depth:  depth,
			IsLeaf: isLeaf,
		})
	}

	if !node.Expanded {
		return
	}

	childDepth := depth + 1
	if node.Parent == nil {
		childDepth = 0 // root's children start at depth 0
	}
	for _, child := range node.Children {
		flattenNode(child, childDepth, rows)
	}
}

// computeRollups performs a bottom-up sum of SizeBytes, FileCount, and child count.
func computeRollups(node *TreeNode) {
	if len(node.Children) == 0 {
		return
	}
	node.SizeBytes = 0
	node.FileCount = 0
	for _, child := range node.Children {
		computeRollups(child)
		node.SizeBytes += child.SizeBytes
		node.FileCount += child.FileCount
	}
}

// updateNodeSize updates a leaf's size and walks parent pointers to update rollups.
func updateNodeSize(resultToNode map[int]*TreeNode, idx int, sizeBytes, fileCount int64) {
	node, ok := resultToNode[idx]
	if !ok {
		return
	}

	oldSize := node.SizeBytes
	oldCount := node.FileCount
	node.SizeBytes = sizeBytes
	node.FileCount = fileCount

	diffSize := sizeBytes - oldSize
	diffCount := fileCount - oldCount

	// Walk up parent chain to update rollups incrementally
	for p := node.Parent; p != nil; p = p.Parent {
		p.SizeBytes += diffSize
		p.FileCount += diffCount
	}
}

// leafCount returns the number of leaf nodes (results) under a node.
func leafCount(node *TreeNode) int {
	if len(node.Children) == 0 {
		return 1
	}
	count := 0
	for _, child := range node.Children {
		count += leafCount(child)
	}
	return count
}

// leafIndices collects all ResultIdx values from leaf nodes under a node.
func leafIndices(node *TreeNode) []int {
	var indices []int
	collectLeafIndices(node, &indices)
	return indices
}

func collectLeafIndices(node *TreeNode, indices *[]int) {
	if len(node.Children) == 0 && node.ResultIdx >= 0 {
		*indices = append(*indices, node.ResultIdx)
		return
	}
	for _, child := range node.Children {
		collectLeafIndices(child, indices)
	}
}

// Helper functions

func splitPath(path string) []string {
	path = filepath.Clean(path)
	if path == "/" {
		return nil
	}
	// Remove leading /
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	return strings.Split(path, "/")
}

func findChild(node *TreeNode, label string) *TreeNode {
	for _, child := range node.Children {
		if child.Label == label {
			return child
		}
	}
	return nil
}

func sortChildren(node *TreeNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Label < node.Children[j].Label
	})
	for _, child := range node.Children {
		sortChildren(child)
	}
}

// compressTree collapses single-child internal chains into combined labels.
func compressTree(node *TreeNode) {
	for _, child := range node.Children {
		compressTree(child)
	}

	// Compress: if this is an internal node with exactly one child,
	// and that child is also internal, merge them.
	for len(node.Children) == 1 && node.Children[0].ResultIdx < 0 && len(node.Children[0].Children) > 0 {
		child := node.Children[0]
		node.Label = node.Label + "/" + child.Label
		node.FullPath = child.FullPath
		node.Children = child.Children
		for _, grandchild := range node.Children {
			grandchild.Parent = node
		}
	}
}

func collapseAll(node *TreeNode) {
	node.Expanded = false
	for _, child := range node.Children {
		collapseAll(child)
	}
}
