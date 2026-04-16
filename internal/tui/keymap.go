package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	Back     key.Binding
	Quit     key.Binding
	Help     key.Binding
	Setup    key.Binding
	Start    key.Binding
	Stop     key.Binding
	Restart  key.Binding
	Monitor  key.Binding
	Logs     key.Binding
	Clean    key.Binding
	CleanAll key.Binding
	Doctor   key.Binding
	Refresh  key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Back:     key.NewBinding(key.WithKeys("esc", "q"), key.WithHelp("esc/q", "back")),
		Quit:     key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Help:     key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "help")),
		Setup:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "setup")),
		Start:    key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "start")),
		Stop:     key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "stop")),
		Restart:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restart")),
		Monitor:  key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "monitor")),
		Logs:     key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "logs")),
		Clean:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "clean")),
		CleanAll: key.NewBinding(key.WithKeys("X"), key.WithHelp("X", "clean-all")),
		Doctor:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "doctor")),
		Refresh:  key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh")),
	}
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Back, k.Quit, k.Help},
		{k.Setup, k.Start, k.Stop, k.Restart, k.Monitor, k.Logs},
		{k.Clean, k.CleanAll, k.Refresh},
	}
}
