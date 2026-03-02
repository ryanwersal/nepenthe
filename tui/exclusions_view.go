package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ryanwersal/nepenthe/internal/state"
)

type SystemExclusion struct {
	Path   string
	Source string // "nepenthe" or "other"
}

func categorizeExclusions(allPaths []string, st *state.State) []SystemExclusion {
	var result []SystemExclusion
	for _, p := range allPaths {
		source := "other"
		if state.IsTracked(st, p) {
			source = "nepenthe"
		}
		result = append(result, SystemExclusion{Path: p, Source: source})
	}
	return result
}

func renderExclusionsView(exclusions []SystemExclusion, cursor int, width, height int) string {
	if len(exclusions) == 0 {
		return dimStyle.Render("  No exclusions found...")
	}

	listHeight := height - 6
	if listHeight < 1 {
		listHeight = 1
	}

	start := 0
	if cursor >= listHeight {
		start = cursor - listHeight + 1
	}
	end := start + listHeight
	if end > len(exclusions) {
		end = len(exclusions)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		e := exclusions[i]
		isCursor := i == cursor

		accent := cursorAccentStyle
		text := lipgloss.NewStyle()
		if isCursor {
			accent = withCursorBg(accent)
			text = withCursorBg(text)
		}

		// Source tags keep their own distinct backgrounds, no cursor bg needed
		var sourceLabel string
		if e.Source == "nepenthe" {
			sourceLabel = nepentheTagStyle.Render("nepenthe")
		} else {
			sourceLabel = otherTagStyle.Render("other")
		}

		path := e.Path
		maxPath := width - 20
		if maxPath < 20 {
			maxPath = 20
		}
		if len(path) > maxPath {
			path = "…" + path[len(path)-maxPath+1:]
		}

		var line string
		if isCursor {
			content := fmt.Sprintf(" %s %s %s", accent.Render("▎"), sourceLabel, text.Render(path))
			line = padRow(content, width, true)
		} else {
			content := fmt.Sprintf("   %s %s", sourceLabel, path)
			line = padRow(content, width, false)
		}
		b.WriteString(line)
		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
