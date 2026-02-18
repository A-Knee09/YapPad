package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	list         list.Model
	input        textinput.Model
	viewport     viewport.Model
	keys         keyMap
	help         help.Model
	inputMode    bool
	renameMode   bool
	renameTarget string
	ready        bool
	selectedFile string
	showPreview  bool
	showingImage bool
	width        int
	height       int
	sortMode     sortMode
	deleting     bool
	yapMode      yapMode
}

func (m model) Init() tea.Cmd { return nil }

func initialModel() model {
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		log.Fatal(err)
	}

	for _, sub := range []string{"daily", "weekly", "monthly", "yearly", ".templates"} {
		os.MkdirAll(filepath.Join(vaultDir, sub), 0o755)
	}

	defaultMode := defaultYapMode
	items := listFiles(sortModifiedDesc, defaultMode)

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)

	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("%s/%s (default)", defaultMode.defaultNoteDir(), defaultMode.defaultNoteName())
	ti.CharLimit = 128
	ti.Width = 40

	keys := keyMap{
		New:           key.NewBinding(key.WithKeys("ctrl+n"), key.WithHelp("ctrl+n", "new")),
		Rename:        key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "rename")),
		Delete:        key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		TogglePreview: key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "preview")),
		CycleSort:     key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "sort")),
		YapMode:       key.NewBinding(key.WithKeys("0", "1", "2", "3", "4"), key.WithHelp("0-4", "yap mode")),
		TabMode:       key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "cycle mode (input)")),
		Quit:          key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	}

	return model{
		list:        l,
		input:       ti,
		keys:        keys,
		help:        help.New(),
		showPreview: true,
		sortMode:    sortModifiedDesc,
		yapMode:     defaultMode,
	}
}

// loadFileOrImage determines if a file is an image or text and dispatches
// to the appropriate handler.
func (m model) loadFileOrImage(path string) tea.Cmd {
	if isImageFile(path) {
		xOffset := m.width/3 + 3
		yOffset := 4
		cols := m.viewport.Width
		rows := m.viewport.Height
		return tea.Sequence(
			clearKittyGraphics(),
			func() tea.Msg { return clearViewportMsg{} },
			renderImage(path, cols, rows, xOffset, yOffset),
		)
	}
	m.showingImage = false
	return tea.Sequence(
		clearKittyGraphics(),
		readFile(path),
	)
}

// switchYapMode changes the yap mode, refreshes the list, and loads the
// first item's preview (or clears the viewport if the list is empty).
func (m model) switchYapMode(mode yapMode) (tea.Model, tea.Cmd) {
	m.yapMode = mode
	m.list.SetItems(listFiles(m.sortMode, m.yapMode))
	m.selectedFile = ""

	if m.list.SelectedItem() != nil {
		i := m.list.SelectedItem().(item)
		m.selectedFile = i.title
		return m, m.loadFileOrImage(m.resolveFilePath(i.title))
	}
	m.viewport.SetContent("")
	return m, clearKittyGraphics()
}

// resolveFilePath resolves the full path for a file given its display title.
func (m model) resolveFilePath(title string) string {
	if m.yapMode == yapAll {
		return filepath.Join(vaultDir, title)
	}
	return filepath.Join(vaultDir, m.yapMode.subdir(), title)
}
