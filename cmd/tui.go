package cmd

import (
	"github.com/ryanwersal/nepenthe/tui"
	tea "github.com/charmbracelet/bubbletea"
)

func runTUI() error {
	m, err := tui.New()
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := finalModel.(tui.Model); ok && fm.Err() != nil {
		return fm.Err()
	}
	return nil
}
