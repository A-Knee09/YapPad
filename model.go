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

var (
	vaultDir string

	themeOptions = []string{
		"default", "gruvbox", "nord", "tokyonight", "forest",
		"solarized", "catppuccin", "dracula", "dusk", "tide",
		"moss", "glacier", "plum",
	}
	editorOptions = []string{}
)

type model struct {
	// UI components
	list      list.Model
	viewport  viewport.Model
	input     textinput.Model
	descInput textinput.Model
	spinner   spinner.Model

	// Editor
	editorMode    bool
	editorFile    string
	editorContent textarea.Model
	editorPreview bool

	// State flags
	inputMode    bool
	inputStep    int
	renameMode   bool
	renameTarget string
	deleting     bool
	settingsOpen bool
	settingsTab  int
	settingsSave bool
	fullPreview  bool

	// Preview
	selectedFile         string
	previewSelectionPath string
	previewSelectionID   uint64
	previewPending       bool
	showPreview          bool
	manualHidePreview    bool
	showingImage         bool
	loadingFile          bool

	// Settings
	themeIndex  int
	editorIndex int
	tempTheme   string
	tempEditor  string

	// Terminal size
	width  int
	height int

	// Folder navigation state
	currentFolder      string
	folderCreationMode bool
	moveMode           bool
	moveTarget         string

	// Config
	theme    Theme
	editor   string
	sortMode sortMode
	keys     *keyMap
	ready    bool
}

func (m model) Init() tea.Cmd { return nil }

func initialModel(editor string, themeName string) model {
	themeIdx := 0
	for i, t := range themeOptions {
		if t == themeName {
			themeIdx = i
			break
		}
	}
	editorIdx := 0
	for i, e := range editorOptions {
		if e == editor {
			editorIdx = i
			break
		}
	}

	listKeys := newListKeyMap()

	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		log.Fatal(err)
	}

	items := listFolderContents(vaultDir, sortModifiedDesc)

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Contents of " + filepath.Base(vaultDir)
	l.SetShowTitle(true)

	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			listKeys.New,
			listKeys.NewFolder,
			listKeys.Rename,
			listKeys.Move,
			listKeys.Delete,
			listKeys.TogglePreview,
			listKeys.FullPreview,
			listKeys.ToggleHelpMenu,
			listKeys.CycleSort,
			listKeys.ThemesEditorSettings,
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
	editorOptions = availableEditors()

	return model{
		list:          l,
		input:         ti,
		descInput:     di,
		spinner:       s,
		keys:          listKeys,
		viewport:      viewport.New(0, 0),
		showPreview:   true,
		sortMode:      sortModifiedDesc,
		editor:        editor,
		theme:         t,
		themeIndex:    themeIdx,
		editorIndex:   editorIdx,
		currentFolder: vaultDir,
	}
}

func (m model) loadFileOrImage(path string) tea.Cmd {
	requestID, ctx := newImageRequest()
	if isImageFile(path) {
		listWidth := m.width / 2
		xOffset := listWidth + 6

		cols := m.viewport.Width - 1
		rows := m.height - 5

		yOffset := 4

		return renderImage(path, cols, rows, xOffset, yOffset, requestID, ctx)
	}
	return func() tea.Msg {
		if !writeImageIfCurrent(requestID, []byte("\x1b_Ga=d,d=a,q=2\x1b\\")) {
			return nil
		}
		msg := readFile(path)()
		if loaded, ok := msg.(fileLoadedMsg); ok {
			loaded.requestID = requestID
			return loaded
		}
		return msg
	}
}

func (m model) resolveFilePath(title string) string {
	return filepath.Join(vaultDir, title)
}

func (m model) previewHeader() string {
	title := m.previewHeaderStyle().Render(m.selectedFile)
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		repeatRune('─', max(0, m.viewport.Width-lipgloss.Width(title))),
	)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) previewFooter() string {
	info := m.previewFooterStyle().Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		repeatRune('─', max(0, m.viewport.Width-lipgloss.Width(info))),
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
