/*
NOTE:
Defines the model struct with all state, initialModel constructor, Init, loadFileOrImage, and resolveFilePath
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

var vaultDir string

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

	items := listFiles(sortModifiedDesc)

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
		}
	}

	t := getTheme(themeName)

	ti := textinput.New()
	ti.Placeholder = "filename.md (enter for default)"
	ti.CharLimit = 128
	ti.Width = 40
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(t.Primary)

	di := textinput.New()
	di.Placeholder = "Description (optional, press enter to skip)"
	di.CharLimit = 128
	di.Width = 40
	di.Cursor.Style = lipgloss.NewStyle().Foreground(t.Primary)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(t.Primary)

	return model{
		list:        l,
		input:       ti,
		descInput:   di,
		spinner:     s,
		keys:        listKeys,
		viewport:    viewport.New(0, 0),
		showPreview: true,
		sortMode:    sortModifiedDesc,
		editor:      editor,
		theme:       t,
	}
}

func (m model) loadFileOrImage(path string) tea.Cmd {
	if isImageFile(path) {
		listWidth := m.width / 2
		xOffset := listWidth + 6

		cols := m.viewport.Width - 1
		rows := m.viewport.Height

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

func (m model) resolveFilePath(title string) string {
	return filepath.Join(vaultDir, title)
}

func (m model) previewHeader() string {
	title := m.previewHeaderStyle().Render(m.selectedFile)
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		fmt.Sprintf("%s", repeatRune('─', max(0, m.viewport.Width-lipgloss.Width(title)))),
	)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) previewFooter() string {
	info := m.previewFooterStyle().Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		fmt.Sprintf("%s", repeatRune('─', max(0, m.viewport.Width-lipgloss.Width(info)))),
	)
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func repeatRune(r rune, n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = r
	}
	return string(b)
}
