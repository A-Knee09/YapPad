/*
NOTE:
Defines the model struct with all state, initialModel constructor, Init, loadFileOrImage, switchYapMode, and resolveFilePath
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	list              list.Model
	input             textinput.Model
	descInput         textinput.Model
	inputStep         int
	viewport          viewport.Model
	keys              *keyMap
	inputMode         bool
	renameMode        bool
	renameTarget      string
	ready             bool
	selectedFile      string
	showPreview       bool
	manualHidePreview bool
	showingImage      bool
	width             int
	height            int
	sortMode          sortMode
	deleting          bool
	yapMode           yapMode
	editor            string
	editorMode        bool
	editorFile        string
	editorContent     textarea.Model
	spinner           spinner.Model
	loadingFile       bool
	theme             Theme
}

func (m model) Init() tea.Cmd { return nil }

func initialModel(editor string, themeName string) model {
	listKeys := newListKeyMap()

	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		log.Fatal(err)
	}

	defaultMode := defaultYapMode
	items := listFiles(sortModifiedDesc, defaultMode)

	delegate := list.NewDefaultDelegate()
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
		}
	}

	t := getTheme(themeName)
	ti := textinput.New()
	ti.Placeholder = fmt.Sprintf("%s/%s (default)", defaultMode.defaultNoteDir(), defaultMode.defaultNoteName())
	ti.CharLimit = 128
	ti.Width = 40

	di := textinput.New()
	di.Placeholder = "Description (optional, press enter to skip)"
	di.CharLimit = 128
	di.Width = 40

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("62"))

	return model{
		list:        l,
		input:       ti,
		descInput:   di,
		spinner:     s,
		keys:        listKeys,
		viewport:    viewport.New(0, 0),
		showPreview: true,
		sortMode:    sortModifiedDesc,
		yapMode:     defaultMode,
		editor:      editor,
		theme:       t,
	}
}

// NOTE: loadFileOrImage determines if a file is an image or text and dispatches to the appropriate handler.
func (m model) loadFileOrImage(path string) tea.Cmd {
	if isImageFile(path) {
		listWidth := m.width / 2
		xOffset := listWidth + 6

		cols := m.viewport.Width - 1
		rows := m.viewport.Height

		// Adjust yOffset based on aspect ratio
		yOffset := 4 + 4
		ratio, err := getImageAspectRatio(path)
		if err == nil && ratio > 0.8 {
			xOffset = listWidth + 15
			yOffset = 4 + 3
		}

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

// NOTE: switchYapMode changes the yap mode, refreshes the list, and loads the first item's preview (or clears the viewport if the list is empty).
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

// NOTE: resolveFilePath resolves the full path for a file given its display title.
func (m model) resolveFilePath(title string) string {
	if m.yapMode == yapAll {
		return filepath.Join(vaultDir, title)
	}
	return filepath.Join(vaultDir, m.yapMode.subdir(), title)
}
