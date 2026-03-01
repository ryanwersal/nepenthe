package tui

import (
	"fmt"
	"strings"

	"github.com/ryanwersal/nepenthe/internal/format"
	"github.com/ryanwersal/nepenthe/internal/scanner"
)

func renderScanView(results []scanner.ScanResult, selected map[int]bool, cursor int, width, height int) string {
	if len(results) == 0 {
		return dimStyle.Render("  No results yet...")
	}

	// Virtual scroll: compute visible window
	listHeight := height - 6 // title + status + dividers + help
	if listHeight < 1 {
		listHeight = 1
	}

	start := 0
	if cursor >= listHeight {
		start = cursor - listHeight + 1
	}
	end := start + listHeight
	if end > len(results) {
		end = len(results)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		r := results[i]
		isCursor := i == cursor

		// Indicator
		var indicator string
		if r.IsExcluded {
			indicator = excludedStyle.Render("✓")
		} else if selected[i] {
			indicator = selectedStyle.Render("●")
		} else {
			indicator = dimStyle.Render("○")
		}

		// Cursor prefix
		prefix := "  "
		if isCursor {
			prefix = cursorStyle.Render("> ")
		}

		// Ecosystem label
		eco := ecosystemStyle.Render(r.Ecosystem)

		// Path (truncate if needed)
		path := r.Path
		maxPath := width - 32
		if maxPath < 20 {
			maxPath = 20
		}
		if len(path) > maxPath {
			path = "…" + path[len(path)-maxPath+1:]
		}

		// Size suffix
		sizeSuffix := ""
		if r.SizeBytes > 0 {
			sizeSuffix = fmt.Sprintf(" (%s", format.Bytes(r.SizeBytes))
			if r.FileCount > 0 {
				sizeSuffix += fmt.Sprintf(", %s files", format.Count(r.FileCount))
			}
			sizeSuffix += ")"
		}

		line := fmt.Sprintf("%s%s %s %s%s", prefix, indicator, eco, path, dimStyle.Render(sizeSuffix))
		b.WriteString(line)
		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func renderTreeView(flatRows []FlatRow, results []scanner.ScanResult, selected map[int]bool, cursor int, width, height int) string {
	if len(flatRows) == 0 {
		return dimStyle.Render("  No results...")
	}

	// Virtual scroll: compute visible window
	listHeight := height - 7 // title + status + dividers + help (2 lines)
	if listHeight < 1 {
		listHeight = 1
	}

	start := 0
	if cursor >= listHeight {
		start = cursor - listHeight + 1
	}
	end := start + listHeight
	if end > len(flatRows) {
		end = len(flatRows)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		row := flatRows[i]
		node := row.Node
		isCursor := i == cursor

		// Cursor prefix
		prefix := "  "
		if isCursor {
			prefix = cursorStyle.Render("> ")
		}

		// Indentation
		indent := strings.Repeat("  ", row.Depth)

		if row.IsLeaf {
			// Leaf node: show indicator + ecosystem label + path basename + size
			r := results[node.ResultIdx]
			var indicator string
			if r.IsExcluded {
				indicator = excludedStyle.Render("✓")
			} else if selected[node.ResultIdx] {
				indicator = selectedStyle.Render("●")
			} else {
				indicator = dimStyle.Render("○")
			}

			eco := ecosystemStyle.Render(r.Ecosystem)

			// Show full path, truncated if needed
			path := r.Path
			indentWidth := row.Depth*2 + 2 + 1 + 1 + 24 + 1 // indent + prefix + indicator + space + eco width + space
			maxPath := width - indentWidth - 1
			if maxPath < 20 {
				maxPath = 20
			}
			if len(path) > maxPath {
				path = "…" + path[len(path)-maxPath+1:]
			}

			sizeSuffix := ""
			if r.SizeBytes > 0 {
				sizeSuffix = fmt.Sprintf(" (%s", format.Bytes(r.SizeBytes))
				if r.FileCount > 0 {
					sizeSuffix += fmt.Sprintf(", %s files", format.Count(r.FileCount))
				}
				sizeSuffix += ")"
			}

			line := fmt.Sprintf("%s%s%s %s %s%s", prefix, indent, indicator, eco, path, dimStyle.Render(sizeSuffix))
			b.WriteString(line)
		} else {
			// Internal node: show expand arrow + label + rollup
			arrow := "▸"
			if node.Expanded {
				arrow = "▾"
			}

			rollup := ""
			count := leafCount(node)
			if node.SizeBytes > 0 {
				rollup = fmt.Sprintf(" (%s, %d items)", format.Bytes(node.SizeBytes), count)
			} else if count > 0 {
				rollup = fmt.Sprintf(" (%d items)", count)
			}

			line := fmt.Sprintf("%s%s%s %s%s", prefix, indent, dimStyle.Render(arrow), node.Label, dimStyle.Render(rollup))
			b.WriteString(line)
		}

		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
