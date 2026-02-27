package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	New            key.Binding
	Rename         key.Binding
	Delete         key.Binding
	TogglePreview  key.Binding
	CycleSort      key.Binding
	YapMode        key.Binding
	TabMode        key.Binding
	ToggleHelpMenu key.Binding
}

func newListKeyMap() *keyMap {
	return &keyMap{
		New:            key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "new")),
		Rename:         key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "rename")),
		Delete:         key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		TogglePreview:  key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "preview")),
		CycleSort:      key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "sort")),
		YapMode:        key.NewBinding(key.WithKeys("0", "1", "2", "3", "4"), key.WithHelp("0-4", "yap mode")),
		TabMode:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle mode (input)")),
		ToggleHelpMenu: key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "Toggle Help")),
	}
}
