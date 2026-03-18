/*
 NOTE:
all state transitions and event handling. The entire Update function — keyboard input, mouse, window resize, file loading messages, editor mode, input mode, delete confirmation, sort, preview toggle.
*/

package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

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
		m.viewport.Width = viewportWidth
		m.viewport.Height = msg.Height - 10
		m.list.SetSize(listWidth, msg.Height-5)

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, msg.Height-10)
			m.ready = true
			if m.list.SelectedItem() != nil {
				i := m.list.SelectedItem().(item)
				m.selectedFile = i.title
				if m.showPreview {
					m.loadingFile = true
					return m, tea.Batch(clearCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(i.title)))
				}
			}
		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = msg.Height - 10
			if m.showPreview && m.selectedFile == "" {
				if m.list.SelectedItem() != nil {
					i := m.list.SelectedItem().(item)
					m.selectedFile = i.title
					if isImageFile(m.resolveFilePath(i.title)) {
						m.showingImage = true
					}
					m.loadingFile = true
					return m, tea.Batch(clearCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(i.title)))
				}
			} else if m.showPreview && m.selectedFile != "" {
				if isImageFile(m.resolveFilePath(m.selectedFile)) {
					m.showingImage = true
				}
				m.loadingFile = true
				return m, tea.Batch(clearCmd, m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
			}
		}
		return m, clearCmd

	case spinner.TickMsg:
		if m.loadingFile {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case fileLoadedMsg:
		m.loadingFile = false
		m.showingImage = false
		wrapped := wordwrap.String(msg.content, m.viewport.Width)
		m.viewport.SetContent(wrapped)
		m.viewport.GotoTop()

	case editorSavedMsg:
		m.list.SetItems(listFiles(m.sortMode))
		return m, m.list.NewStatusMessage("Saved!")

	case clearViewportMsg:
		m.viewport.SetContent(strings.Repeat("\n", m.viewport.Height))

	case imageRenderedMsg:
		m.loadingFile = false
		m.showingImage = true

	case tea.MouseMsg:
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
		m.list.SetItems(listFiles(m.sortMode))
		m.viewport.SetContent("")
		if m.selectedFile != "" && m.showPreview {
			m.loadingFile = true
			return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
		}
		return m, nil

	case tea.KeyMsg:

		// EDITOR MODE
		if m.editorMode {
			switch msg.String() {
			case "ctrl+s":
				return m, saveEditorContent(m.editorFile, m.editorContent.Value())
			case "ctrl+q":
				m.editorMode = false
				m.editorContent.Blur()
				m.list.SetItems(listFiles(m.sortMode))
				if m.showPreview {
					m.loadingFile = true
					return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
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
					path := m.resolveFilePath(it.title)
					os.Remove(path)
					deleteMetaDesc(path)
					m.list.SetItems(listFiles(m.sortMode))
					statusCmd := m.list.NewStatusMessage("Deleted " + it.title)
					m.deleting = false
					return m, statusCmd
				}
				m.deleting = false
				return m, nil
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
					newPath := filepath.Join(vaultDir, name)

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
					m.list.SetItems(listFiles(m.sortMode))
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
					path = filepath.Join(vaultDir, "note.md")
				} else {
					if filepath.Ext(name) == "" {
						name += ".md"
					}
					path = filepath.Join(vaultDir, name)
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
				m.list.SetItems(listFiles(m.sortMode))

				if m.editor == "inbuilt" {
					var editorCmd tea.Cmd
					m, editorCmd = openInbuiltEditor(path, m)
					return m, editorCmd
				}
				return m, openInEditor(path, m.editor)

			case "esc":
				m.inputMode = false
				m.renameMode = false
				m.inputStep = 0
				m.input.SetValue("")
				m.descInput.SetValue("")
				m.input.Focus()
				m.list.SetItems(listFiles(m.sortMode))
				return m, nil
			}

			if m.inputStep == 0 {
				m.input, cmd = m.input.Update(msg)
				val := m.input.Value()
				if val != "" {
					allItems := listFiles(m.sortMode)
					var filtered []list.Item
					lowerVal := strings.ToLower(val)
					for _, it := range allItems {
						if strings.Contains(strings.ToLower(it.(item).title), lowerVal) {
							filtered = append(filtered, it)
						}
					}
					m.list.SetItems(filtered)
				} else {
					m.list.SetItems(listFiles(m.sortMode))
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

		case key.Matches(msg, m.keys.Delete):
			if m.list.SelectedItem() != nil {
				m.deleting = true
			}
			return m, nil

		case key.Matches(msg, m.keys.Rename):
			if it, ok := m.list.SelectedItem().(item); ok {
				m.renameMode = true
				m.renameTarget = it.title
				m.inputMode = true
				m.inputStep = 0
				m.input.SetValue(it.title)
				m.input.Focus()
				existingDesc := readMetaDesc(m.resolveFilePath(it.title))
				m.descInput.SetValue(existingDesc)
			}
			return m, nil

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 6
			m.list.SetItems(listFiles(m.sortMode))
			m.selectedFile = ""
			if m.list.SelectedItem() != nil && m.showPreview {
				i := m.list.SelectedItem().(item)
				m.selectedFile = i.title
				m.loadingFile = true
				return m, tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(i.title)))
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

		case msg.String() == "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			if it, ok := m.list.SelectedItem().(item); ok {
				path := m.resolveFilePath(it.title)
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

	var cmdList tea.Cmd
	m.list, cmdList = m.list.Update(msg)

	var cmdRead tea.Cmd
	if m.list.SelectedItem() != nil {
		i := m.list.SelectedItem().(item)
		if i.title != m.selectedFile {
			m.selectedFile = i.title
			if m.showPreview {
				m.loadingFile = true
				cmdRead = tea.Batch(m.spinner.Tick, m.loadFileOrImage(m.resolveFilePath(i.title)))
			}
		}
	}
	var cmdViewport tea.Cmd
	m.viewport, cmdViewport = m.viewport.Update(msg)

	return m, tea.Batch(cmd, cmdList, cmdViewport, cmdRead)
}
