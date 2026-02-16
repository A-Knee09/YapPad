package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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
	list                   list.Model
	showingList            bool
}

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	// terminal resize handle
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)

	case tea.KeyMsg:

		// Global keybindings
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.New):
			m.createFileInputVisible = true
			m.newFileInput.Focus()
			return m, nil

		case key.Matches(msg, m.keys.Save):
			if m.currentFile == nil {
				break
			}
			if err := m.currentFile.Truncate(0); err != nil {
				fmt.Println("Cannot save the file :(")
				return m, nil
			}

			if _, err := m.currentFile.Seek(0, 0); err != nil {
				fmt.Println("Cannot save the file :(")
				return m, nil
			}

			if _, err := m.currentFile.WriteString(m.noteTextArea.Value()); err != nil {
				fmt.Println("Cannot save the file :(")
				return m, nil
			}

			if err := m.currentFile.Close(); err != nil {
				fmt.Println("Cannot close the file")
			}
			m.currentFile = nil
			m.noteTextArea.SetValue("")
			return m, nil

		case key.Matches(msg, m.keys.List):
			noteList := listFiles()
			m.list.SetItems(noteList)
			m.showingList = true
			return m, nil
		}

		if m.showingList {
			switch msg.String() {
			case "enter":
				item, ok := m.list.SelectedItem().(item)
				if ok {
					filepath := filepath.Join(vaultDir, item.title)
					content, err := os.ReadFile(filepath)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						return m, nil
					}

					m.noteTextArea.SetValue(string(content))

					f, err := os.OpenFile(filepath, os.O_RDWR, 0o644)
					if err != nil {
						log.Printf("Error reading file: %v", err)
						return m, nil
					}
					m.currentFile = f
					m.showingList = false
					return m, nil
				}
			}
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

	if m.showingList {
		m.list, cmd = m.list.Update(msg)
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

	if m.showingList {
		view = m.list.View()
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
	ta.ShowLineNumbers = false
	ta.Placeholder = "Write your yap here"
	ta.Focus()

	// list
	noteList := listFiles()
	finalList := list.New(noteList, list.NewDefaultDelegate(), 0, 0)
	finalList.Title = "All Notes"

	return model{
		keys:                   keys,
		newFileInput:           ti,
		createFileInputVisible: false,
		noteTextArea:           ta,
		help:                   help.New(),
		list:                   finalList,
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func listFiles() []list.Item {
	items := make([]list.Item, 0)
	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		log.Fatal("Error reading notes")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			modTime := info.ModTime().Format("2006-01-02 15:04:05")
			items = append(items, item{
				title: entry.Name(),
				desc:  fmt.Sprintf("Modified: %s", modTime),
			})
		}
	}
	return items
}
