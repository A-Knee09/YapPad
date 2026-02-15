package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("161")).
		Width(60).
		Align(lipgloss.Center)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))

	cursorLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("38"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))
)

type model struct {
	keys                   keyMap
	help                   help.Model
	newFileInput           textinput.Model
	createFileInputVisible bool
}

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

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.New):
			m.createFileInputVisible = true
			return m, nil
		}
	}

	if m.createFileInputVisible {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	welcome := style.Render("Welcome to Note Maker twin :D")
	helpView := m.help.View(m.keys)
	view := ""
	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, helpView)
}

func initialModel() model {
	// Keybinds
	keys := keyMap{
		New: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Create New File 🗒"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "Quit ⏻"),
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

	return model{
		keys:                   keys,
		newFileInput:           ti,
		createFileInputVisible: false,
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
