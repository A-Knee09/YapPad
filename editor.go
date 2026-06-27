// NOTE: This file is for the inbuilt text area component

package main

import (
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

var knownEditors = []string{"inbuilt", "nvim", "vim", "vi", "nano", "helix", "micro", "emacs", "gedit", "kate", "code"}

func availableEditors() []string {
	available := []string{"inbuilt"}
	for _, e := range knownEditors[1:] { // skip inbuilt
		if _, err := exec.LookPath(e); err == nil {
			available = append(available, e)
		}
	}
	return available
}

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

func getEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	return "nvim"
}

func openInEditor(path, editor string) tea.Cmd {
	e := editor
	if e == "" {
		e = getEditor()
	}

	cmd := exec.Command(e, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return fileEditedMsg{err: err}
	})
}
