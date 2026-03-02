package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Adaptive color palette — auto-switches for dark/light terminals.
var (
	colorPrimary  = lipgloss.AdaptiveColor{Dark: "#7C3AED", Light: "#6D28D9"} // violet
	colorSelected = lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"} // amber
	colorExcluded = lipgloss.AdaptiveColor{Dark: "#34D399", Light: "#059669"} // emerald
	colorDim      = lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#9CA3AF"}
	colorCursorBg = lipgloss.AdaptiveColor{Dark: "#3A3A3A", Light: "#DCDCDC"}
	colorBorder   = lipgloss.AdaptiveColor{Dark: "#383838", Light: "#D0D0D0"}
	colorHeaderBg = lipgloss.AdaptiveColor{Dark: "#7C3AED", Light: "#6D28D9"}
	colorHeaderFg = lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"}
	colorBody     = lipgloss.AdaptiveColor{Dark: "#E5E7EB", Light: "#1F2937"}
	colorHelpKey  = lipgloss.AdaptiveColor{Dark: "#404040", Light: "#D4D4D4"}
)

// Shared text styles.
var (
	dimStyle       = lipgloss.NewStyle().Foreground(colorDim)
	selectedStyle  = lipgloss.NewStyle().Foreground(colorSelected)
	excludedStyle  = lipgloss.NewStyle().Foreground(colorExcluded)
	ecosystemStyle = lipgloss.NewStyle().Foreground(colorDim).Width(24)
	statusBarStyle = lipgloss.NewStyle().Foreground(colorBody)
)

// Header bar and tabs.
var (
	headerBarStyle = lipgloss.NewStyle().
			Background(colorHeaderBg).
			Foreground(colorHeaderFg).
			Bold(true).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Background(colorHeaderFg).
			Foreground(colorPrimary).
			Bold(true).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Dark: "#A78BFA", Light: "#C4B5FD"}).
				Padding(0, 1)
)

// Content border.
var contentBorderStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder)

// Cursor row highlight.
var (
	cursorAccentStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)
)

// Help key pills.
var (
	helpKeyStyle  = lipgloss.NewStyle().Background(colorHelpKey).Foreground(colorHeaderFg).Bold(true).Padding(0, 1)
	helpDescStyle = lipgloss.NewStyle().Foreground(colorDim)
)

// Settings section headers.
var sectionHeaderStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

// Exclusion source tags.
var (
	nepentheTagStyle = lipgloss.NewStyle().
				Background(colorExcluded).
				Foreground(lipgloss.AdaptiveColor{Dark: "#000000", Light: "#FFFFFF"}).
				Bold(true).
				Padding(0, 1)

	otherTagStyle = lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(lipgloss.AdaptiveColor{Dark: "#FFFFFF", Light: "#FFFFFF"}).
			Bold(true).
			Padding(0, 1)
)

// Progress bar styles — purple to match overall theme.
var (
	progressFillStyle = lipgloss.NewStyle().Foreground(colorPrimary)
	progressBgStyle   = lipgloss.NewStyle().Foreground(colorDim)
	progressPulse     = lipgloss.NewStyle().Foreground(colorPrimary)
)

// withCursorBg returns a copy of the style with the cursor background color added.
func withCursorBg(s lipgloss.Style) lipgloss.Style {
	return s.Background(colorCursorBg)
}

// padRow pads a row to full width. For cursor rows the content segments must
// already have been rendered with cursor-bg-aware styles; this only fills the
// remaining space with the cursor background.
func padRow(content string, width int, isCursor bool) string {
	cw := lipgloss.Width(content)
	pad := width - cw
	if pad <= 0 {
		return content
	}
	if isCursor {
		return content + lipgloss.NewStyle().Background(colorCursorBg).Render(strings.Repeat(" ", pad))
	}
	return content
}

// renderProgressBar renders a determinate progress bar: ████████░░░░
func renderProgressBar(width, current, total int) string {
	if total <= 0 {
		return progressBgStyle.Render(strings.Repeat("░", width))
	}
	filled := width * current / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return progressFillStyle.Render(strings.Repeat("█", filled)) +
		progressBgStyle.Render(strings.Repeat("░", empty))
}

// renderIndeterminateBar renders a pulsing bar that bounces back and forth.
func renderIndeterminateBar(width, tick int) string {
	pulseWidth := 4
	if pulseWidth > width {
		pulseWidth = width
	}
	// Bounce: tick moves from 0..range, then back
	span := width - pulseWidth
	if span <= 0 {
		return progressPulse.Render(strings.Repeat("━", width))
	}
	pos := tick % (span * 2)
	if pos >= span {
		pos = span*2 - pos
	}

	var b strings.Builder
	b.WriteString(progressBgStyle.Render(strings.Repeat("─", pos)))
	b.WriteString(progressPulse.Render(strings.Repeat("━", pulseWidth)))
	b.WriteString(progressBgStyle.Render(strings.Repeat("─", span-pos)))
	return b.String()
}
