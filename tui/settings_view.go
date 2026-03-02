package tui

import (
	"fmt"
	"strings"

	"github.com/ryanwersal/nepenthe/internal/config"
	"github.com/ryanwersal/nepenthe/internal/scanner"
)

type settingsItemKind int

const (
	settingsHeader settingsItemKind = iota
	settingsToggle
	settingsPath
	settingsAddButton
	settingsValue
)

type settingsEditField int

const (
	editNone settingsEditField = iota
	editRoot
	editCustomPath
	editCustomEcosystem
	editSchedule
)

type settingsItem struct {
	Kind      settingsItemKind
	Label     string
	Value     string // category name, path, or formatted value
	Enabled   bool   // for toggle items
	Deletable bool   // for path items
	EditField settingsEditField
}

func buildSettingsItems(cfg config.Config) []settingsItem {
	var items []settingsItem

	// Categories section
	items = append(items, settingsItem{Kind: settingsHeader, Label: "Categories"})
	for _, cat := range scanner.AllCategories {
		if cat == scanner.CategoryCustom {
			continue
		}
		enabled := config.CategoryEnabled(cfg, string(cat))
		items = append(items, settingsItem{
			Kind:    settingsToggle,
			Label:   scanner.CategoryLabel[cat],
			Value:   string(cat),
			Enabled: enabled,
		})
	}

	// Scan Roots section
	items = append(items, settingsItem{Kind: settingsHeader, Label: "Scan Roots"})
	for _, root := range cfg.Roots {
		items = append(items, settingsItem{
			Kind:      settingsPath,
			Label:     root,
			Value:     root,
			Deletable: len(cfg.Roots) > 1,
		})
	}
	items = append(items, settingsItem{
		Kind:      settingsAddButton,
		Label:     "[+ add root]",
		EditField: editRoot,
	})

	// Custom Paths section
	items = append(items, settingsItem{Kind: settingsHeader, Label: "Custom Paths"})
	if len(cfg.CustomFixedPaths) == 0 {
		items = append(items, settingsItem{Kind: settingsHeader, Label: "(none)"})
	} else {
		for _, cf := range cfg.CustomFixedPaths {
			label := cf.Path
			if cf.Ecosystem != "" {
				label += " (" + cf.Ecosystem + ")"
			}
			items = append(items, settingsItem{
				Kind:      settingsPath,
				Label:     label,
				Value:     cf.Path,
				Deletable: true,
			})
		}
	}
	items = append(items, settingsItem{
		Kind:      settingsAddButton,
		Label:     "[+ add custom path]",
		EditField: editCustomPath,
	})

	// Schedule section
	items = append(items, settingsItem{Kind: settingsHeader, Label: "Schedule"})
	items = append(items, settingsItem{
		Kind:      settingsValue,
		Label:     "Interval",
		Value:     fmt.Sprintf("%ds", cfg.Schedule.IntervalSeconds),
		EditField: editSchedule,
	})

	return items
}

// firstSelectableIndex returns the index of the first non-header item.
func firstSelectableIndex(items []settingsItem) int {
	for i, item := range items {
		if item.Kind != settingsHeader {
			return i
		}
	}
	return 0
}

func renderSettingsView(items []settingsItem, cursor int, width, height int) string {
	if len(items) == 0 {
		return dimStyle.Render("  No settings...")
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
	if end > len(items) {
		end = len(items)
	}

	var b strings.Builder
	for i := start; i < end; i++ {
		item := items[i]
		isCursor := i == cursor

		switch item.Kind {
		case settingsHeader:
			b.WriteString("  ")
			b.WriteString(dimStyle.Render(item.Label))

		case settingsToggle:
			prefix := "  "
			if isCursor {
				prefix = cursorStyle.Render("> ")
			}
			check := dimStyle.Render("[ ]")
			if item.Enabled {
				check = excludedStyle.Render("[x]")
			}
			label := fmt.Sprintf("%s — %s", scanner.Category(item.Value), item.Label)
			fmt.Fprintf(&b, "%s  %s %s", prefix, check, label)

		case settingsPath:
			prefix := "  "
			if isCursor {
				prefix = cursorStyle.Render("> ")
			}
			fmt.Fprintf(&b, "%s  %s", prefix, item.Label)

		case settingsAddButton:
			prefix := "  "
			if isCursor {
				prefix = cursorStyle.Render("> ")
			}
			fmt.Fprintf(&b, "%s  %s", prefix, dimStyle.Render(item.Label))

		case settingsValue:
			prefix := "  "
			if isCursor {
				prefix = cursorStyle.Render("> ")
			}
			fmt.Fprintf(&b, "%s  %s: %s", prefix, item.Label, item.Value)
		}

		if i < end-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
