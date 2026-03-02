package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

		// Pick cursor-bg-aware styles when this is the cursor row
		dim := dimStyle
		sel := selectedStyle
		exc := excludedStyle
		eco := ecosystemStyle
		accent := cursorAccentStyle
		text := lipgloss.NewStyle()
		if isCursor {
			dim = withCursorBg(dim)
			sel = withCursorBg(sel)
			exc = withCursorBg(exc)
			eco = withCursorBg(eco)
			accent = withCursorBg(accent)
			text = withCursorBg(text)
		}

		// Indicator
		var indicator string
		switch {
		case r.IsExcluded:
			indicator = exc.Render("✓")
		case selected[i]:
			indicator = sel.Render("●")
		default:
			indicator = dim.Render("○")
		}

		// Ecosystem label
		ecoLabel := eco.Render(r.Ecosystem)

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

		var line string
		if isCursor {
			content := fmt.Sprintf(" %s %s %s %s%s",
				accent.Render("▎"), indicator, ecoLabel, text.Render(path), dim.Render(sizeSuffix))
			line = padRow(content, width, true)
		} else {
			content := fmt.Sprintf("   %s %s %s%s",
				indicator, ecoLabel, path, dim.Render(sizeSuffix))
			line = padRow(content, width, false)
		}
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

		// Pick cursor-bg-aware styles when this is the cursor row
		dim := dimStyle
		sel := selectedStyle
		exc := excludedStyle
		eco := ecosystemStyle
		accent := cursorAccentStyle
		text := lipgloss.NewStyle()
		if isCursor {
			dim = withCursorBg(dim)
			sel = withCursorBg(sel)
			exc = withCursorBg(exc)
			eco = withCursorBg(eco)
			accent = withCursorBg(accent)
			text = withCursorBg(text)
		}

		// Tree prefix with box-drawing characters
		treePrefix := dim.Render(row.Prefix)

		if row.IsLeaf {
			// Leaf node: show indicator + ecosystem label + path basename + size
			r := results[node.ResultIdx]
			var indicator string
			switch {
			case r.IsExcluded:
				indicator = exc.Render("✓")
			case selected[node.ResultIdx]:
				indicator = sel.Render("●")
			default:
				indicator = dim.Render("○")
			}

			ecoLabel := eco.Render(r.Ecosystem)

			// Show full path, truncated if needed
			path := r.Path
			indentWidth := lipgloss.Width(row.Prefix) + 4 + 1 + 24 + 1 // prefix + accent/space + indicator + eco width + space
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

			var line string
			if isCursor {
				content := fmt.Sprintf(" %s %s%s %s %s%s",
					accent.Render("▎"), treePrefix, indicator, ecoLabel, text.Render(path), dim.Render(sizeSuffix))
				line = padRow(content, width, true)
			} else {
				content := fmt.Sprintf("   %s%s %s %s%s",
					treePrefix, indicator, ecoLabel, path, dim.Render(sizeSuffix))
				line = padRow(content, width, false)
			}
			b.WriteString(line)
		} else {
			// Internal node: show expand arrow + label + rollup
			var arrow string
			if node.Expanded {
				arrow = accent.Render("▾")
			} else {
				arrow = dim.Render("▸")
			}

			rollup := ""
			count := leafCount(node)
			if node.SizeBytes > 0 {
				rollup = fmt.Sprintf(" (%s, %d items)", format.Bytes(node.SizeBytes), count)
			} else if count > 0 {
				rollup = fmt.Sprintf(" (%d items)", count)
			}

			var line string
			if isCursor {
				content := fmt.Sprintf(" %s %s%s %s%s",
					accent.Render("▎"), treePrefix, arrow, text.Render(node.Label), dim.Render(rollup))
				line = padRow(content, width, true)
			} else {
				content := fmt.Sprintf("   %s%s %s%s",
					treePrefix, arrow, node.Label, dim.Render(rollup))
				line = padRow(content, width, false)
			}
			b.WriteString(line)
		}

		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
