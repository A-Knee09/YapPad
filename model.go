package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	list         list.Model
	input        textinput.Model
	descInput    textinput.Model
	inputStep    int
	viewport     viewport.Model
	keys         *keyMap
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
	listKeys := newListKeyMap()

	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		log.Fatal(err)
	}

	defaultMode := defaultYapMode
	items := listFiles(sortModifiedDesc, defaultMode)

	delegate := list.NewDefaultDelegate()
	delegate.Styles = listItemStyles
	l := list.New(items, delegate, 0, 0)
	l.Title = "All Yaps Here"
	l.SetShowTitle(true)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.New,
			listKeys.Rename,
			listKeys.Delete,
			listKeys.TogglePreview,

			listKeys.ToggleHelpMenu,
			listKeys.CycleSort,
			listKeys.YapMode,
			listKeys.Quit,
		}
	}

	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("%s/%s (default)", defaultMode.defaultNoteDir(), defaultMode.defaultNoteName())
	ti.CharLimit = 128
	ti.Width = 40

	di := textinput.New()
	di.Placeholder = "Description (optional, press enter to skip)"
	di.CharLimit = 128
	di.Width = 40

	return model{
		list:        l,
		input:       ti,
		descInput:   di,
		keys:        listKeys,
		viewport:    viewport.New(0, 0),
		showPreview: true,
		sortMode:    sortModifiedDesc,
		yapMode:     defaultMode,
	}
}

// loadFileOrImage determines if a file is an image or text and dispatches
// to the appropriate handler.
func (m model) loadFileOrImage(path string) tea.Cmd {
	if isImageFile(path) {
		// Account for: left panel width + separator + border
		listWidth := m.width / 2
		xOffset := listWidth + 3 + 1

		// Account for: header height + border
		yOffset := 4 + 1

		// Shrink cols/rows to fit inside the border
		cols := m.viewport.Width - 2
		rows := m.viewport.Height - 2

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
	m.list.Title = m.yapMode.String() + " Yaps"
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
