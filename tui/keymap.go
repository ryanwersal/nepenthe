package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	Toggle   key.Binding
	All      key.Binding
	None     key.Binding
	Apply    key.Binding
	Remove   key.Binding
	SwitchV  key.Binding
	Settings key.Binding
	Group    key.Binding
	Expand   key.Binding
	Collapse key.Binding
	Delete   key.Binding
	Quit     key.Binding
}

// ShortHelp returns key bindings for the compact help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Apply, k.Quit}
}

// FullHelp returns key bindings grouped for the expanded help view.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Expand, k.Collapse},
		{k.Toggle, k.All, k.None, k.Group},
		{k.Apply, k.Remove, k.SwitchV, k.Settings},
		{k.Delete, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("k/↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("j/↓", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	All: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "all"),
	),
	None: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "none"),
	),
	Apply: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "apply"),
	),
	Remove: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "remove"),
	),
	SwitchV: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "exclusions"),
	),
	Settings: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "settings"),
	),
	Group: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "group"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "delete"),
	),
	Expand: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("l/→", "expand"),
	),
	Collapse: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("h/←", "collapse"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

// Per-view keymap variants for help display.

type scanKeyMap struct{ keyMap }

func (k scanKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Apply, k.SwitchV, k.Settings, k.Quit}
}

func (k scanKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Expand, k.Collapse},
		{k.Toggle, k.All, k.None, k.Group},
		{k.Apply, k.Remove, k.SwitchV, k.Settings},
		{k.Quit},
	}
}

type exclusionsKeyMap struct{ keyMap }

func (k exclusionsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.SwitchV, k.Settings, k.Quit}
}

func (k exclusionsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.SwitchV, k.Settings, k.Quit},
	}
}

type settingsKeyMap struct{ keyMap }

func (k settingsKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Toggle, k.Apply, k.Delete, k.Settings, k.Quit}
}

func (k settingsKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.Toggle, k.Apply, k.Delete},
		{k.Settings, k.Quit},
	}
}

type settingsEditKeyMap struct{}

func (k settingsEditKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithHelp("enter", "confirm")),
		key.NewBinding(key.WithHelp("esc", "cancel")),
	}
}

func (k settingsEditKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}
