/*
 NOTE:
all state transitions and event handling. The entire Update function — keyboard input, mouse, window resize, file loading messages, editor mode, input mode, delete confirmation, sort, preview toggle.
*/

package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

func (m model) refreshList() []list.Item {
	return listFolderContents(m.currentFolder, m.sortMode)
}

func waitForPreview(selectionID uint64, path string) tea.Cmd {
	return func() tea.Msg {
		timer := time.NewTimer(250 * time.Millisecond)
		defer timer.Stop()
		<-timer.C
		return previewDelayElapsedMsg{selectionID: selectionID, path: path}
	}
}

func (m model) updateSelectedPreview() (model, tea.Cmd) {
	selected, ok := m.list.SelectedItem().(item)
	if !ok {
		hadPendingPreview := m.previewSelectionPath != "" || m.selectedFile != "" || m.loadingFile || m.showingImage
		m.previewSelectionPath = ""
		m.previewSelectionID++
		m.previewPending = false
		m.selectedFile = ""
		m.loadingFile = false
		m.showingImage = false
		m.viewport.SetContent(folderTreePreview(m.currentFolder))
		if hadPendingPreview {
			return m, clearKittyGraphics()
		}
		return m, nil
	}
	if selected.path == m.previewSelectionPath {
		return m, nil
	}

	m.previewSelectionPath = selected.path
	m.previewSelectionID++
	m.previewPending = false
	if selected.isDir {
		m.selectedFile = ""
		m.loadingFile = false
		m.showingImage = false
		m.viewport.SetContent(folderTreePreview(selected.path))
		m.viewport.GotoTop()
		return m, clearKittyGraphics()
	}

	relPath, _ := filepath.Rel(vaultDir, selected.path)
	m.selectedFile = relPath
	m.showingImage = false
	if !m.showPreview {
		invalidateImageRequest()
		m.loadingFile = false
		m.showingImage = false
		return m, nil
	}
	m.loadingFile = true
	m.previewPending = true
	return m, tea.Batch(m.spinner.Tick, clearKittyGraphics(), waitForPreview(m.previewSelectionID, selected.path))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		var clearCmd tea.Cmd
		if m.showingImage {
			m.showingImage = false
			clearCmd = clearKittyGraphics()
		}

		wasFullPreview := m.fullPreview
		m.fullPreview = false

		const minWidthForPreview = 90

		if msg.Width < minWidthForPreview {
			m.showPreview = false
		} else if !m.manualHidePreview {
			m.showPreview = true
		}

		var listWidth, viewportWidth int
		if m.showPreview {
			listWidth = msg.Width / 2
			viewportWidth = msg.Width - listWidth - 4 - 2
		} else {
			listWidth = msg.Width - 2
			viewportWidth = 0
		}
		m.list.SetSize(listWidth, msg.Height-5)

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, msg.Height-10)
			m.ready = true
		}

		// NOTE: These two if conditions basically toggle fullPreview when window is being resized either on menu or inbuilt editor
		if wasFullPreview && m.selectedFile != "" {
			m.fullPreview = true
			m.viewport.Width = m.width - 4
			m.viewport.Height = m.height - 6
			m.loadingFile = true
			return m, tea.Batch(clearCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
		}
		if m.editorPreview {
			m.viewport.Width = m.width - 4
			m.viewport.Height = m.height - 6
			content := m.editorContent.Value()
			rendered := renderMarkdown(content)
			wrapped := wordwrap.String(rendered, m.viewport.Width)
			m.viewport.SetContent(wrapped)
			m.viewport.GotoTop()
			return m, clearCmd
		}

		m.viewport.Width = viewportWidth
		m.viewport.Height = msg.Height - 10

		if m.showPreview && m.selectedFile != "" {
			if m.previewPending {
				return m, tea.Batch(clearCmd, m.spinner.Tick)
			}
			m.loadingFile = true
			return m, tea.Batch(clearCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
		}
		if m.previewSelectionPath == "" {
			var previewCmd tea.Cmd
			m, previewCmd = m.updateSelectedPreview()
			return m, tea.Batch(clearCmd, previewCmd)
		}

		return m, clearCmd
	case spinner.TickMsg:
		if m.loadingFile {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case previewDelayElapsedMsg:
		if msg.selectionID != m.previewSelectionID || msg.path != m.previewSelectionPath {
			return m, nil
		}
		if !m.showPreview || m.settingsOpen || m.editorMode || m.inputMode || m.deleting {
			m.previewPending = false
			m.loadingFile = false
			return m, nil
		}
		m.previewPending = false
		return m, m.loadFileOrImage(msg.path)

	case fileLoadedMsg:
		if !isCurrentImageRequest(msg.requestID) {
			return m, nil
		}
		m.loadingFile = false
		m.previewPending = false
		m.showingImage = false
		wrapped := wordwrap.String(msg.content, m.viewport.Width)
		m.viewport.SetContent(wrapped)
		m.viewport.GotoTop()

	case editorSavedMsg:
		m.list.SetItems(m.refreshList())
		return m, m.list.NewStatusMessage("Saved!")

	case imageRenderedMsg:
		if !isCurrentImageRequest(msg.requestID) {
			return m, nil
		}
		m.loadingFile = false
		m.previewPending = false
		if msg.err != nil {
			m.showingImage = false
			m.viewport.SetContent("Image preview failed: " + msg.err.Error())
		} else {
			m.showingImage = true
			m.viewport.SetContent("")
		}

	case tea.MouseMsg:
		if m.fullPreview {
			var vpCmd tea.Cmd
			m.viewport, vpCmd = m.viewport.Update(msg)
			return m, vpCmd
		}
		if msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown {
			return m, nil
		}
		if m.showPreview {
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.viewport.ScrollUp(1)
			case tea.MouseButtonWheelDown:
				m.viewport.ScrollDown(1)
			}
			var cmdViewport tea.Cmd
			m.viewport, cmdViewport = m.viewport.Update(msg)
			return m, cmdViewport
		}
		return m, nil

	case fileEditedMsg:
		m.list.SetItems(m.refreshList())
		for i, it := range m.list.Items() {
			relPath, _ := filepath.Rel(vaultDir, it.(item).path)
			if relPath == m.selectedFile {
				m.list.Select(i)
				break
			}
		}
		if m.selectedFile != "" && m.showPreview {
			m.loadingFile = true
			return m, tea.Batch(tea.EnableMouseAllMotion, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
		}
		return m, tea.EnableMouseAllMotion

	case tea.KeyMsg:
		if m.settingsOpen {
			switch msg.String() {
			case "esc":
				// cancel, restore original theme/editor
				m.theme = getTheme(m.tempTheme)
				m.editor = m.tempEditor
				m.settingsOpen = false
				if m.selectedFile != "" && m.showPreview {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
			case "enter":
				// confirm
				if m.settingsSave {
					cfg := loadConfig()
					cfg.Theme = themeOptions[m.themeIndex]
					cfg.Editor = editorOptions[m.editorIndex]
					writeConfig(cfg)
				}
				m.settingsOpen = false
				if m.selectedFile != "" && m.showPreview {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
			case "tab":
				m.settingsTab = (m.settingsTab + 1) % 2
			case "s":
				m.settingsSave = !m.settingsSave
			case "up", "k":
				if m.settingsTab == 0 {
					m.themeIndex = (m.themeIndex - 1 + len(themeOptions)) % len(themeOptions)
					m.theme = getTheme(themeOptions[m.themeIndex])
				} else {
					m.editorIndex = (m.editorIndex - 1 + len(editorOptions)) % len(editorOptions)
					m.editor = editorOptions[m.editorIndex]
				}
			case "down", "j":
				if m.settingsTab == 0 {
					m.themeIndex = (m.themeIndex + 1) % len(themeOptions)
					m.theme = getTheme(themeOptions[m.themeIndex])
				} else {
					m.editorIndex = (m.editorIndex + 1) % len(editorOptions)
					m.editor = editorOptions[m.editorIndex]
				}
			}
			return m, nil
		}
		if m.fullPreview {
			if msg.String() == "f" {
				m.fullPreview = false
				listWidth := m.width / 2
				m.viewport.Width = m.width - listWidth - 6
				m.viewport.Height = m.height - 10
				if m.selectedFile != "" {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
				return m, nil
			}
			// only allow viewport scrolling
			var vpCmd tea.Cmd
			m.viewport, vpCmd = m.viewport.Update(msg)
			return m, vpCmd
		}
		// EDITOR MODE
		if m.editorMode {

			switch msg.String() {
			case "ctrl+s":
				return m, saveEditorContent(m.editorFile, m.editorContent.Value())
			case "ctrl+q":
				m.editorMode = false
				m.editorPreview = false
				m.editorContent.Blur()
				listWidth := m.width / 2
				m.viewport.Width = m.width - listWidth - 6
				m.viewport.Height = m.height - 10
				m.list.SetItems(m.refreshList())
				if m.showPreview && m.selectedFile != "" {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
				return m, nil

			// NOTE: Claude generated this solution, I wanted the preview to be done on the line where inbuilt editor cursor is, the cursor is present on is on the center of the preview screen
			case "ctrl+f":
				m.editorPreview = !m.editorPreview
				if m.editorPreview {
					m.viewport.Width = m.width - 4
					m.viewport.Height = m.height - 6
					content := m.editorContent.Value()
					rendered := renderMarkdown(content)
					wrapped := wordwrap.String(rendered, m.viewport.Width)
					m.viewport.SetContent(wrapped)
					m.viewport.GotoTop()

					// count rendered lines up to cursor position
					cursorPos := m.editorContent.LineInfo().CharOffset
					lineNum := m.editorContent.Line()
					rawLines := strings.Split(content, "\n")
					textUpToCursor := strings.Join(rawLines[:min(lineNum, len(rawLines))], "\n")
					renderedUpToCursor := renderMarkdown(textUpToCursor)
					wrappedUpToCursor := wordwrap.String(renderedUpToCursor, m.viewport.Width)
					renderedLine := strings.Count(wrappedUpToCursor, "\n")
					_ = cursorPos
					m.viewport.SetYOffset(max(0, renderedLine-(m.viewport.Height/2)))
				} else {
					listWidth := m.width / 2
					m.viewport.Width = m.width - listWidth - 6
					m.viewport.Height = m.height - 10
				}
				return m, nil
			}
			var editorCmd tea.Cmd
			m.editorContent, editorCmd = m.editorContent.Update(msg)
			return m, editorCmd
		}

		// DELETE CONFIRMATION MODE
		if m.deleting {
			switch msg.String() {
			case "y", "Y":
				if it, ok := m.list.SelectedItem().(item); ok {
					path := it.path
					os.Remove(path)
					deleteMetaDesc(path)
					m.list.SetItems(m.refreshList())
					m.deleting = false
					m.selectedFile = ""
					m.showingImage = false
					m.viewport.SetContent("")
					statusCmd := m.list.NewStatusMessage("Deleted " + it.title)
					return m, tea.Batch(statusCmd, clearKittyGraphics())
				}
			case "n", "N", "esc":
				m.deleting = false
				return m, nil
			default:
				return m, nil
			}
		}

		// INPUT MODE
		if m.inputMode {
			switch msg.String() {

			case "enter":
				if m.folderCreationMode {
					name := m.input.Value()
					if name != "" {
						newPath := filepath.Join(m.currentFolder, name)
						os.MkdirAll(newPath, 0o755)
					}
					m.folderCreationMode = false
					m.inputMode = false
					m.input.SetValue("")
					m.input.Focus()
					m.list.SetItems(m.refreshList())
					return m, nil
				}

				if m.moveMode {
					destDirRel := m.input.Value()
					destDir := filepath.Clean(filepath.Join(vaultDir, destDirRel))
					os.MkdirAll(destDir, 0o755)
					oldPath := m.moveTarget
					newPath := filepath.Join(destDir, filepath.Base(oldPath))
					if _, err := os.Stat(oldPath); err == nil {
						desc := readMetaDesc(oldPath)
						deleteMetaDesc(oldPath)
						os.Rename(oldPath, newPath)
						if desc != "" {
							writeMetaDesc(newPath, desc)
						}
					}
					m.moveMode = false
					m.inputMode = false
					m.input.SetValue("")
					m.input.Focus()
					m.list.SetItems(m.refreshList())
					m.selectedFile = ""
					m.viewport.SetContent("")
					return m, m.list.NewStatusMessage("Moved file to " + destDirRel)
				}

				if m.renameMode {
					if m.inputStep == 0 {
						name := m.input.Value()
						if name == "" {
							break
						}
						m.inputStep = 1
						m.descInput.Placeholder = "(optional, enter to skip)"
						m.input.Blur()
						m.descInput.Focus()
						return m, nil
					}

					name := m.input.Value()
					desc := m.descInput.Value()

					oldPath := m.resolveFilePath(m.renameTarget)
					originalExt := filepath.Ext(m.renameTarget)
					if filepath.Ext(name) == "" {
						name += originalExt
					}
					newPath := filepath.Join(filepath.Dir(oldPath), name)

					oldDesc := readMetaDesc(oldPath)
					finalDesc := desc
					if finalDesc == "" {
						finalDesc = oldDesc
					}

					os.MkdirAll(filepath.Dir(newPath), 0o755)
					os.Rename(oldPath, newPath)
					deleteMetaDesc(oldPath)
					writeMetaDesc(newPath, finalDesc)

					rel, _ := filepath.Rel(vaultDir, newPath)
					m.selectedFile = rel

					m.renameMode = false
					m.inputMode = false
					m.inputStep = 0
					m.input.SetValue("")
					m.descInput.SetValue("")
					m.input.Focus()
					m.list.SetItems(m.refreshList())
					return m, nil
				}

				// NEW FILE
				if m.inputStep == 0 {
					m.inputStep = 1
					m.input.Blur()
					m.descInput.SetValue("")
					m.descInput.Placeholder = "(optional, press enter to skip)"
					m.descInput.Focus()
					return m, nil
				}

				// create the file
				name := m.input.Value()
				desc := m.descInput.Value()

				var path string
				if name == "" {
					path = filepath.Join(m.currentFolder, "note.md")
				} else {
					if filepath.Ext(name) == "" {
						name += ".md"
					}
					path = filepath.Join(m.currentFolder, name)
				}

				os.MkdirAll(filepath.Dir(path), 0o755)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					os.WriteFile(path, []byte{}, 0o644)
				}

				writeMetaDesc(path, desc)

				rel, _ := filepath.Rel(vaultDir, path)
				m.selectedFile = rel

				m.inputMode = false
				m.inputStep = 0
				m.input.SetValue("")
				m.descInput.SetValue("")
				m.input.Focus()
				m.list.SetItems(m.refreshList())
				m.list.Title = "Contents of " + filepath.Base(m.currentFolder)

				if m.editor == "inbuilt" {
					var editorCmd tea.Cmd
					m, editorCmd = openInbuiltEditor(path, m)
					return m, editorCmd
				}
				return m, openInEditor(path, m.editor)

			case "esc":
				m.inputMode = false
				m.renameMode = false
				m.folderCreationMode = false
				m.moveMode = false
				m.inputStep = 0
				m.input.SetValue("")
				m.descInput.SetValue("")
				m.input.Focus()
				m.list.SetItems(m.refreshList())
				return m, nil
			}

			if m.inputStep == 0 {
				m.input, cmd = m.input.Update(msg)
				val := m.input.Value()
				if val != "" {
					allItems := m.refreshList()
					var filtered []list.Item
					lowerVal := strings.ToLower(val)
					for _, it := range allItems {
						if strings.Contains(strings.ToLower(it.(item).title), lowerVal) {
							filtered = append(filtered, it)
						}
					}
					m.list.SetItems(filtered)
				} else {
					m.list.SetItems(m.refreshList())
				}
			} else {
				m.descInput, cmd = m.descInput.Update(msg)
			}
			return m, cmd
		}

		// NORMAL MODE
		switch {

		case key.Matches(msg, m.keys.New):
			m.inputMode = true
			m.input.Placeholder = "filename.md (enter for default)"
			m.input.Focus()
			return m, nil

		case key.Matches(msg, m.keys.NewFolder):
			m.inputMode = true
			m.folderCreationMode = true
			m.input.Placeholder = "New folder name"
			m.input.SetValue("")
			m.input.Focus()
			return m, nil

		case key.Matches(msg, m.keys.Delete):
			if m.list.SelectedItem() != nil {
				if it, ok := m.list.SelectedItem().(item); ok && !it.isDir {
					m.deleting = true
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Rename):
			if it, ok := m.list.SelectedItem().(item); ok && !it.isDir {
				m.renameMode = true
				relPath, _ := filepath.Rel(vaultDir, it.path)
				m.renameTarget = relPath
				m.inputMode = true
				m.inputStep = 0
				m.input.SetValue(it.title)
				m.input.Focus()
				existingDesc := readMetaDesc(it.path)
				m.descInput.SetValue(existingDesc)
			}
			return m, nil

		case key.Matches(msg, m.keys.Move):
			if m.list.SelectedItem() != nil {
				if it, ok := m.list.SelectedItem().(item); ok && !it.isDir {
					m.inputMode = true
					m.moveMode = true
					m.moveTarget = it.path
					m.input.Placeholder = "Move to: path/relative/to/vault/"
					m.input.SetValue("")
					m.input.Focus()
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 6
			m.list.SetItems(m.refreshList())
			m.selectedFile = ""
			if m.list.SelectedItem() != nil && m.showPreview {
				i := m.list.SelectedItem().(item)
				if !i.isDir {
					relPath, _ := filepath.Rel(vaultDir, i.path)
					m.selectedFile = relPath
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(i.path))
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.ToggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.showPreview = !m.showPreview
			m.manualHidePreview = !m.showPreview

			newM, resizeCmd := m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
			m = newM.(model)

			if !m.showPreview {
				m.previewSelectionID++
				m.previewPending = false
				m.loadingFile = false
				m.showingImage = false
				return m, tea.Batch(resizeCmd, clearKittyGraphics())
			}

			if m.selectedFile != "" {
				if isImageFile(m.resolveFilePath(m.selectedFile)) {
					m.showingImage = true
				}
				m.loadingFile = true
				return m, tea.Batch(resizeCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
			}
			return m, resizeCmd

		case key.Matches(msg, m.keys.ThemesEditorSettings):
			m.settingsOpen = true
			m.settingsTab = 0
			m.settingsSave = false
			// save originals for cancel
			m.tempTheme = themeOptions[m.themeIndex]
			m.tempEditor = m.editor
			if m.showingImage || m.loadingFile {
				m.showingImage = false
				m.loadingFile = false
				return m, clearKittyGraphics()
			}
			// set current index
			for i, t := range themeOptions {
				if t == themeOptions[m.themeIndex] {
					m.themeIndex = i
					break
				}
			}
			for i, e := range editorOptions {
				if e == m.editor {
					m.editorIndex = i
					break
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.FullPreview):
			if m.selectedFile != "" && isImageFile(m.resolveFilePath(m.selectedFile)) {
				return m, nil
			}
			m.fullPreview = !m.fullPreview
			if m.fullPreview {
				m.viewport.Width = m.width - 4
				m.viewport.Height = m.height - 6
				if m.selectedFile != "" {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
			} else {
				listWidth := m.width / 2
				m.viewport.Width = m.width - listWidth - 6
				m.viewport.Height = m.height - 10
				if m.selectedFile != "" {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
				}
			}
			return m, nil

		case msg.String() == "esc":
			if m.currentFolder != vaultDir {
				m.currentFolder = filepath.Dir(m.currentFolder)
				m.list.SetItems(m.refreshList())
				m.list.Title = "Contents of " + filepath.Base(m.currentFolder)
				m.list.ResetSelected()
				m.previewSelectionPath = ""
				m.fullPreview = false
				var previewCmd tea.Cmd
				m, previewCmd = m.updateSelectedPreview()
				return m, previewCmd
			}

		case msg.String() == "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			if it, ok := m.list.SelectedItem().(item); ok {
				if it.isDir {
					m.currentFolder = it.path
					m.list.SetItems(m.refreshList())
					m.list.Title = "Contents of " + filepath.Base(m.currentFolder)
					m.list.ResetSelected()
					m.previewSelectionPath = ""
					var previewCmd tea.Cmd
					m, previewCmd = m.updateSelectedPreview()
					return m, previewCmd
				} else {
					path := it.path
					if isImageFile(path) {
						return m, openImageViewer(path)
					}
					if m.editor == "inbuilt" {
						var editorCmd tea.Cmd
						m, editorCmd = openInbuiltEditor(path, m)
						return m, editorCmd
					}
					return m, openInEditor(path, m.editor)
				}
			}
		}
	}

	var cmdList tea.Cmd
	m.list, cmdList = m.list.Update(msg)

	var cmdRead tea.Cmd
	m, cmdRead = m.updateSelectedPreview()
	var cmdViewport tea.Cmd
	m.viewport, cmdViewport = m.viewport.Update(msg)

	return m, tea.Batch(cmd, cmdList, cmdViewport, cmdRead)
}
