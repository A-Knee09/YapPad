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
	"github.com/charmbracelet/lipgloss"
)

var (
	style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("161")).
		Width(60).
		Align(lipgloss.Center)

	cursorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))
	cursorLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("38"))
	promptStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("141"))

	vaultDir string
	docStyle = lipgloss.NewStyle().Margin(1, 2)
)

type keyMap struct {
	Quit   key.Binding
	New    key.Binding
	List   key.Binding
	Save   key.Binding
	Back   key.Binding
	Delete key.Binding
	Rename key.Binding
	Help   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.New, k.List, k.Save, k.Delete, k.Rename, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.List, k.Save, k.Delete, k.Rename},
		{k.Back, k.Help, k.Quit},
	}
}

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
	renameMode             bool
	renameTarget           string
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

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-5)

	case tea.KeyMsg:

		switch {

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil

		case key.Matches(msg, m.keys.New):
			m.createFileInputVisible = true
			m.newFileInput.Focus()
			return m, nil

		case key.Matches(msg, m.keys.Save):
			if m.currentFile == nil {
				break
			}
			if err := m.currentFile.Truncate(0); err != nil {
				return m, nil
			}
			if _, err := m.currentFile.Seek(0, 0); err != nil {
				return m, nil
			}
			if _, err := m.currentFile.WriteString(m.noteTextArea.Value()); err != nil {
				return m, nil
			}
			m.currentFile.Close()
			m.currentFile = nil
			m.noteTextArea.SetValue("")
			return m, nil

		case key.Matches(msg, m.keys.List):
			m.list.SetItems(listFiles())
			m.showingList = true
			return m, nil

		case key.Matches(msg, m.keys.Back):
			m.createFileInputVisible = false
			m.renameMode = false
			m.newFileInput.Blur()
			m.newFileInput.SetValue("")
			m.showingList = false
			return m, nil

		case key.Matches(msg, m.keys.Delete):
			if m.showingList {
				selected, ok := m.list.SelectedItem().(item)
				if ok {
					path := filepath.Join(vaultDir, selected.title)
					if err := os.Remove(path); err == nil {
						m.list.SetItems(listFiles())
					}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Rename):
			if m.showingList {
				selected, ok := m.list.SelectedItem().(item)
				if ok {
					m.renameMode = true
					m.renameTarget = selected.title
					m.createFileInputVisible = true
					m.showingList = false
					m.newFileInput.SetValue(selected.title)
					m.newFileInput.Focus()
				}
			}
			return m, nil
		}

		if m.showingList {
			switch msg.String() {
			case "enter":
				selected, ok := m.list.SelectedItem().(item)
				if ok {
					path := filepath.Join(vaultDir, selected.title)
					content, err := os.ReadFile(path)
					if err != nil {
						return m, nil
					}
					m.noteTextArea.SetValue(string(content))
					f, err := os.OpenFile(path, os.O_RDWR, 0o644)
					if err != nil {
						return m, nil
					}
					m.currentFile = f
					m.showingList = false
					return m, nil
				}
			}
		}

		if m.createFileInputVisible {
			switch msg.String() {

			case "enter":
				filename := m.newFileInput.Value()
				if filename == "" {
					return m, nil
				}

				if m.renameMode {
					oldPath := filepath.Join(vaultDir, m.renameTarget)
					newPath := filepath.Join(vaultDir, filename)

					if _, err := os.Stat(newPath); err == nil {
						return m, nil
					}

					if err := os.Rename(oldPath, newPath); err == nil {
						m.list.SetItems(listFiles())
					}

					m.renameMode = false
					m.createFileInputVisible = false
					m.newFileInput.Blur()
					m.newFileInput.SetValue("")
					m.showingList = true
					return m, nil
				}

				path := fmt.Sprintf("%s/%s.md", vaultDir, filename)
				if _, err := os.Stat(path); err == nil {
					return m, nil
				}

				f, err := os.Create(path)
				if err != nil {
					return m, nil
				}

				m.currentFile = f
				m.createFileInputVisible = false
				m.newFileInput.Blur()
				m.newFileInput.SetValue("")
				return m, nil

			case "esc":
				m.renameMode = false
				m.createFileInputVisible = false
				m.newFileInput.Blur()
				m.newFileInput.SetValue("")
				m.showingList = true
				return m, nil
			}
		}
	}

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
	view := ""
	helpView := m.help.View(m.keys)

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

	keys := keyMap{
		New:    key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "new file 🗒")),
		Quit:   key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit ⏻")),
		List:   key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "list files ☰")),
		Save:   key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save ⎙")),
		Back:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back ➜]")),
		Delete: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete ✖")),
		Rename: key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "rename ✎")),
		Help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	}

	ti := textinput.New()
	ti.Placeholder = "What would you like to name the file"
	ti.CharLimit = 156
	ti.Width = 70
	ti.Cursor.Style = cursorStyle
	ti.PromptStyle = cursorLineStyle
	ti.TextStyle = promptStyle

	ta := textarea.New()
	ta.ShowLineNumbers = false
	ta.Placeholder = "Write your yap here"
	ta.Focus()

	noteList := listFiles()
	finalList := list.New(noteList, list.NewDefaultDelegate(), 0, 0)
	finalList.Title = "All Notes"

	return model{
		keys:         keys,
		newFileInput: ti,
		noteTextArea: ta,
		help:         help.New(),
		list:         finalList,
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
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
