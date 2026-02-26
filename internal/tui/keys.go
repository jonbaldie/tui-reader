package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	NextPage   key.Binding
	PrevPage   key.Binding
	FirstPage  key.Binding
	LastPage   key.Binding
	NextLink   key.Binding
	PrevLink   key.Binding
	FollowLink key.Binding
	GoBack     key.Binding
	Quit       key.Binding
}

var keys = keyMap{
	NextPage: key.NewBinding(
		key.WithKeys("right", "l", " ", "pgdown"),
		key.WithHelp("→/l/space", "next page"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h", "pgup"),
		key.WithHelp("←/h", "previous page"),
	),
	FirstPage: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("home/g", "first page"),
	),
	LastPage: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("end/G", "last page"),
	),
	NextLink: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next link"),
	),
	PrevLink: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous link"),
	),
	FollowLink: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "follow link"),
	),
	GoBack: key.NewBinding(
		key.WithKeys("b", "backspace"),
		key.WithHelp("b", "go back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c", "esc"),
		key.WithHelp("q", "quit"),
	),
}
