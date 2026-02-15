package main

import (
	"github.com/charmbracelet/bubbles/key"
)

type keyMap struct {
	Quit key.Binding
	New  key.Binding
	List key.Binding
	Save key.Binding
	Back key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.New, k.List, k.Save, k.Back, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.List, k.Save},
		{k.Back, k.Quit},
	}
}
