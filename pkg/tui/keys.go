package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Quit        key.Binding
	ForceQuit   key.Binding
	Help        key.Binding
	NextPanel   key.Binding
	PrevPanel   key.Binding
	Panel1      key.Binding
	Panel2      key.Binding
	Panel3      key.Binding
	Panel4      key.Binding
	Panel5      key.Binding
	Disconnect  key.Binding
	Add         key.Binding
	Edit        key.Binding
	Delete      key.Binding
	Select      key.Binding
	Describe    key.Binding
	Execute     key.Binding
	HistoryPrev key.Binding
	HistoryNext key.Binding
	Escape      key.Binding
	Confirm     key.Binding
	Deny        key.Binding
	Filter      key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		ForceQuit:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "force quit")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		NextPanel:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Panel1:      key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "connections")),
		Panel2:      key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "tables")),
		Panel3:      key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "structure")),
		Panel4:      key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "query")),
		Panel5:      key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "results")),
		Disconnect:  key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "disconnect")),
		Add:         key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:        key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit")),
		Delete:      key.NewBinding(key.WithKeys("d", "delete"), key.WithHelp("d", "delete")),
		Select:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Describe:    key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "structure")),
		Execute:     key.NewBinding(key.WithKeys("ctrl+e", "f5"), key.WithHelp("ctrl+e", "execute")),
		HistoryPrev: key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "prev query")),
		HistoryNext: key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "next query")),
		Escape:      key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close/blur")),
		Confirm:     key.NewBinding(key.WithKeys("y"), key.WithHelp("y", "yes")),
		Deny:        key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "no")),
		Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	}
}
