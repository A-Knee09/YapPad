package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	New           key.Binding
	Rename        key.Binding
	Delete        key.Binding
	TogglePreview key.Binding
	CycleSort     key.Binding
	YapMode       key.Binding
	TabMode       key.Binding
	Quit          key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.New, k.YapMode, k.Rename, k.Delete, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.Rename, k.Delete},
		{k.YapMode, k.TabMode, k.TogglePreview, k.CycleSort, k.Quit},
	}
}
