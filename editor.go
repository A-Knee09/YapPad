// NOTE: For the inbuilt text area component

package main

import (
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

func openInbuiltEditor(path string, m model) (model, tea.Cmd) {
	content, err := os.ReadFile(path)
	if err != nil {
		content = []byte{}
	}

	ta := textarea.New()
	ta.ShowLineNumbers = true
	ta.SetWidth(m.width)
	ta.SetHeight(m.height - 4)
	ta.SetValue(string(content))
	ta.Focus()

	m.editorMode = true
	m.editorFile = path
	m.editorContent = ta
	return m, nil
}

func saveEditorContent(path, content string) tea.Cmd {
	return func() tea.Msg {
		err := os.WriteFile(path, []byte(content), 0o644)
		if err != nil {
			return editorSavedMsg{}
		}
		return editorSavedMsg{}
	}
}
