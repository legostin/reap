package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Enter      key.Binding
	Kill       key.Binding
	ForceK     key.Binding
	KillParent key.Binding
	Filter     key.Binding
	Sort       key.Binding
	SortRev    key.Binding
	System     key.Binding
	Tree       key.Binding
	Refresh    key.Binding
	Help       key.Binding
	Quit       key.Binding
	Escape     key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up"),
		key.WithHelp("↑", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "details"),
	),
	Kill: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "kill (SIGTERM)"),
	),
	ForceK: key.NewBinding(
		key.WithKeys("K"),
		key.WithHelp("K", "force kill (SIGKILL)"),
	),
	KillParent: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "kill parent"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	SortRev: key.NewBinding(
		key.WithKeys("S"),
		key.WithHelp("S", "reverse sort"),
	),
	System: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle system"),
	),
	Tree: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "toggle tree"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
}
