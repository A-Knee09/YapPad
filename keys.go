// NOTE : Pretty self explainatory what this file does :D
package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	New                  key.Binding
	Rename               key.Binding
	Delete               key.Binding
	TogglePreview        key.Binding
	CycleSort            key.Binding
	ToggleHelpMenu       key.Binding
	FullPreview          key.Binding
	ThemesEditorSettings key.Binding
}

func newListKeyMap() *keyMap {
	return &keyMap{
		New:                  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "New")),
		Rename:               key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "Rename")),
		Delete:               key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "Delete")),
		TogglePreview:        key.NewBinding(key.WithKeys("p"), key.WithHelp("p", "Notes preview")),
		CycleSort:            key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "Sort")),
		ToggleHelpMenu:       key.NewBinding(key.WithKeys("h"), key.WithHelp("h", "(Un)Hide Help")),
		FullPreview:          key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "Full Reader View")),
		ThemesEditorSettings: key.NewBinding(key.WithKeys(","), key.WithHelp(",", "Settings")),
	}
}
