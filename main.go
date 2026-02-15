package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory")
	}
	vaultDir = fmt.Sprintf("%s/.notemaker", homeDir)
}

type model struct {
	keys                   keyMap
	help                   help.Model
	newFileInput           textinput.Model
	createFileInputVisible bool
	currentFile            *os.File
	noteTextArea           textarea.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		// Global keybindings
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.New):
			m.createFileInputVisible = true
			m.newFileInput.Focus()
			return m, nil
		}

		// Contextual keys (only when input is visible)
		if m.createFileInputVisible {
			switch msg.String() {

			case "enter":
				filename := m.newFileInput.Value()
				if filename == "" {
					return m, nil
				}

				filepath := fmt.Sprintf("%s/%s.md", vaultDir, filename)

				// If file already exists, do nothing
				if _, err := os.Stat(filepath); err == nil {
					return m, nil
				}

				f, err := os.Create(filepath)
				if err != nil {
					// Don't crash the TUI
					return m, nil
				}

				m.currentFile = f
				m.createFileInputVisible = false
				m.newFileInput.Blur()
				m.newFileInput.SetValue("")
				return m, nil

			case "esc":
				m.createFileInputVisible = false
				m.newFileInput.Blur()
				m.newFileInput.SetValue("")
				return m, nil
			}
		}
	}

	// Let textinput handle typing when visible
	if m.createFileInputVisible {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
		return m, cmd
	}

	if m.currentFile != nil {
		m.noteTextArea, cmd = m.noteTextArea.Update(msg)
	}

	return m, nil
}

func (m model) View() string {
	welcome := style.Render("Welcome to Note Maker twin :D")
	helpView := m.help.View(m.keys)
	view := ""
	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}
	if m.currentFile != nil {
		view = m.noteTextArea.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, helpView)
}

func initialModel() model {
	err := os.MkdirAll(vaultDir, 0o750)
	if err != nil {
		log.Fatal(err)
	}

	// Keybinds
	keys := keyMap{
		New: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "new file 🗒"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit ⏻"),
		),
		List: key.NewBinding(
			key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "list files ☰"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save ⎙"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back ➜]"),
		),
	}

	// Init text input
	ti := textinput.New()
	ti.Placeholder = "What would you like to name the file"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 70
	ti.Cursor.Style = cursorStyle
	ti.PromptStyle = cursorLineStyle
	ti.TextStyle = promptStyle

	// Init text textarea
	ta := textarea.New()
	ta.Placeholder = "Write your yap here"
	ta.Focus()

	return model{
		keys:                   keys,
		newFileInput:           ti,
		createFileInputVisible: false,
		noteTextArea:           ta,
		help:                   help.New(),
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
