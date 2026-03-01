package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().Bold(true)
	dimStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("226")) // yellow
	excludedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // green
	otherStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))  // blue
	cursorStyle       = lipgloss.NewStyle().Bold(true)
	ecosystemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Width(24)
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusBarStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
	dividerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	progressFillStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // green
	progressBgStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	progressPulse     = lipgloss.NewStyle().Foreground(lipgloss.Color("75"))  // light blue
)

// renderProgressBar renders a determinate progress bar: [████████░░░░]
func renderProgressBar(width, current, total int) string {
	if total <= 0 {
		return progressBgStyle.Render("[" + strings.Repeat("░", width) + "]")
	}
	filled := width * current / total
	if filled > width {
		filled = width
	}
	empty := width - filled
	return progressBgStyle.Render("[") +
		progressFillStyle.Render(strings.Repeat("█", filled)) +
		progressBgStyle.Render(strings.Repeat("░", empty)) +
		progressBgStyle.Render("]")
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
	b.WriteString(progressBgStyle.Render("["))
	b.WriteString(progressBgStyle.Render(strings.Repeat("─", pos)))
	b.WriteString(progressPulse.Render(strings.Repeat("━", pulseWidth)))
	b.WriteString(progressBgStyle.Render(strings.Repeat("─", span-pos)))
	b.WriteString(progressBgStyle.Render("]"))
	return b.String()
}
