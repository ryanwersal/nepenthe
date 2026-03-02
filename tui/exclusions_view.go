package tui

import (
	"fmt"
	"strings"

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

		prefix := "  "
		if isCursor {
			prefix = cursorStyle.Render("> ")
		}

		var sourceLabel string
		if e.Source == "nepenthe" {
			sourceLabel = excludedStyle.Render(fmt.Sprintf("%-12s", "nepenthe"))
		} else {
			sourceLabel = otherStyle.Render(fmt.Sprintf("%-12s", "other"))
		}

		path := e.Path
		maxPath := width - 18
		if maxPath < 20 {
			maxPath = 20
		}
		if len(path) > maxPath {
			path = "…" + path[len(path)-maxPath+1:]
		}

		fmt.Fprintf(&b, "%s%s %s", prefix, sourceLabel, path)
		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
