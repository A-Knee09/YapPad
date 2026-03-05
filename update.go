package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
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

		//  NOTE: Hard coded for now I'll need a better approach
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

		//  HACK: Not the best way, will fix later
		if !m.ready {
			m.viewport = viewport.New(viewportWidth, msg.Height-10)
			m.ready = true
			if m.list.SelectedItem() != nil {
				i := m.list.SelectedItem().(item)
				m.selectedFile = i.title
				if m.showPreview {
					return m, tea.Batch(clearCmd, m.loadFileOrImage(m.resolveFilePath(i.title)))
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
					return m, tea.Batch(clearCmd, m.loadFileOrImage(m.resolveFilePath(i.title)))
				}
			} else if m.showPreview && m.selectedFile != "" {
				if isImageFile(m.resolveFilePath(m.selectedFile)) {
					m.showingImage = true
				}
				return m, tea.Batch(clearCmd, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
			}
		}
		return m, clearCmd

	case fileLoadedMsg:
		m.showingImage = false
		wrapped := wordwrap.String(msg.content, m.viewport.Width)
		m.viewport.SetContent(wrapped)
		m.viewport.GotoTop()

	case editorSavedMsg:
		return m, m.list.NewStatusMessage("Saved!")

	case clearViewportMsg:
		// Blank the viewport so old text doesn't bleed under image overlay
		m.viewport.SetContent(strings.Repeat("\n", m.viewport.Height))

	case imageRenderedMsg:
		// Image was drawn directly to stdout as overlay.
		m.showingImage = true

	case tea.MouseMsg:
		if msg.Button != tea.MouseButtonWheelUp && msg.Button != tea.MouseButtonWheelDown {
			return m, nil
		}

		var listWidth int
		if m.showPreview {
			listWidth = m.width / 3
		} else {
			listWidth = m.width - 2
		}

		if msg.X < listWidth {
			// Scroll List
			switch msg.Button {
			case tea.MouseButtonWheelUp:
				m.list.CursorUp()
			case tea.MouseButtonWheelDown:
				m.list.CursorDown()
			}
			// Update selection immediately after scrolling
			if m.list.SelectedItem() != nil {
				i := m.list.SelectedItem().(item)
				if i.title != m.selectedFile {
					m.selectedFile = i.title
					path := m.resolveFilePath(i.title)
					return m, m.loadFileOrImage(path)
				}
			}
		} else if m.showPreview && msg.X > listWidth {
			// Scroll Viewport
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
		m.list.SetItems(listFiles(m.sortMode, m.yapMode))
		if m.selectedFile != "" && m.showPreview {
			return m, tea.Batch(tea.EnableMouseAllMotion, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
		}
		return m, tea.EnableMouseAllMotion

	case tea.KeyMsg:

		// NOTE: Inbuilt textarea
		if m.editorMode {
			switch msg.String() {
			case "ctrl+s":
				return m, saveEditorContent(m.editorFile, m.editorContent.Value())
			case "ctrl+q":
				m.editorMode = false
				m.editorContent.Blur()
				if m.showPreview {
					return m, m.loadFileOrImage(m.resolveFilePath(m.selectedFile))
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
					m.list.SetItems(listFiles(m.sortMode, m.yapMode))
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
						m.descInput.Placeholder = "New description (optional, enter to skip)"
						m.input.Blur()
						m.descInput.Focus()
						return m, nil
					}

					// rename + update desc
					name := m.input.Value()
					desc := m.descInput.Value()

					oldPath := m.resolveFilePath(m.renameTarget)
					if !strings.HasSuffix(name, ".md") {
						name += ".md"
					}
					newPath := filepath.Join(vaultDir, name)

					oldDesc := readMetaDesc(oldPath)
					finalDesc := desc
					if finalDesc == "" {
						finalDesc = oldDesc
					}

					if err := os.MkdirAll(filepath.Dir(newPath), 0o755); err != nil {
					}
					os.Rename(oldPath, newPath)
					deleteMetaDesc(oldPath)
					writeMetaDesc(newPath, finalDesc)

					// update selected file to new name
					rel, _ := filepath.Rel(vaultDir, newPath)
					if m.yapMode != yapAll {
						m.selectedFile = filepath.Base(newPath)
					} else {
						m.selectedFile = rel
					}

					m.renameMode = false
					m.inputMode = false
					m.inputStep = 0
					m.input.SetValue("")
					m.descInput.SetValue("")
					m.input.Focus()
					m.list.SetItems(listFiles(m.sortMode, m.yapMode))
					return m, nil
				}

				// NEW FILE
				if m.inputStep == 0 {
					m.inputStep = 1
					m.input.Blur()
					m.descInput.SetValue("")
					m.descInput.Placeholder = "Description (optional, press enter to skip)"
					m.descInput.Focus()
					return m, nil
				}

				// create the file
				name := m.input.Value()
				desc := m.descInput.Value()

				var path string
				if name == "" {
					subdir := m.yapMode.defaultNoteDir()
					defaultName := m.yapMode.defaultNoteName()
					path = filepath.Join(vaultDir, subdir, defaultName)
				} else {
					if !strings.HasSuffix(name, ".md") {
						name += ".md"
					}
					path = filepath.Join(vaultDir, name)
				}

				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				}

				if _, err := os.Stat(path); os.IsNotExist(err) {
					var content []byte
					if m.input.Value() == "" {
						tplPath := filepath.Join(vaultDir, ".templates", m.yapMode.defaultNoteDir()+".md")
						if tplData, err := os.ReadFile(tplPath); err == nil {
							content = tplData
						}
					}
					os.WriteFile(path, content, 0o644)
				}

				writeMetaDesc(path, desc)

				m.inputMode = false
				m.inputStep = 0
				m.input.SetValue("")
				m.descInput.SetValue("")
				m.input.Focus()

				rel, _ := filepath.Rel(vaultDir, path)
				if m.yapMode != yapAll {
					m.selectedFile = filepath.Base(path)
				} else {
					m.selectedFile = rel
				}
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
				m.list.SetItems(listFiles(m.sortMode, m.yapMode))
				return m, nil

			case "tab":
				if m.inputStep == 0 {
					switch m.yapMode {
					case yapAll, yapDaily:
						m.yapMode = yapWeekly
					case yapWeekly:
						m.yapMode = yapMonthly
					case yapMonthly:
						m.yapMode = yapYearly
					case yapYearly:
						m.yapMode = yapDaily
					}
					m.input.Placeholder = fmt.Sprintf("%s/%s (default)", m.yapMode.defaultNoteDir(), m.yapMode.defaultNoteName())
					m.list.SetItems(listFiles(m.sortMode, m.yapMode))
				}
				return m, nil
			}

			if m.inputStep == 0 {
				m.input, cmd = m.input.Update(msg)

				// Live-filter list based on typed input
				val := m.input.Value()
				if val != "" {
					allItems := listFiles(m.sortMode, yapAll)
					var filtered []list.Item
					lowerVal := strings.ToLower(val)
					for _, it := range allItems {
						if strings.Contains(strings.ToLower(it.(item).title), lowerVal) {
							filtered = append(filtered, it)
						}
					}
					m.list.SetItems(filtered)
				} else {
					m.list.SetItems(listFiles(m.sortMode, m.yapMode))
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
			m.input.Placeholder = fmt.Sprintf("%s/%s (default)", m.yapMode.defaultNoteDir(), m.yapMode.defaultNoteName())
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
				// Pre-fill existing description
				existingDesc := readMetaDesc(m.resolveFilePath(it.title))
				m.descInput.SetValue(existingDesc)
			}
			return m, nil

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 6
			m.list.SetItems(listFiles(m.sortMode, m.yapMode))
			m.selectedFile = ""
			if m.list.SelectedItem() != nil && m.showPreview {
				i := m.list.SelectedItem().(item)
				m.selectedFile = i.title
				return m, m.loadFileOrImage(m.resolveFilePath(i.title))
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
				return m, tea.Batch(resizeCmd, m.loadFileOrImage(m.resolveFilePath(m.selectedFile)))
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

		case msg.String() == "0" && m.list.FilterState() != list.Filtering:
			return m.switchYapMode(yapAll)

		case msg.String() == "1" && m.list.FilterState() != list.Filtering:
			return m.switchYapMode(yapDaily)

		case msg.String() == "2" && m.list.FilterState() != list.Filtering:
			return m.switchYapMode(yapWeekly)

		case msg.String() == "3" && m.list.FilterState() != list.Filtering:
			return m.switchYapMode(yapMonthly)

		case msg.String() == "4" && m.list.FilterState() != list.Filtering:
			return m.switchYapMode(yapYearly)
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
				cmdRead = m.loadFileOrImage(m.resolveFilePath(i.title))
			}
		}
	}
	var cmdViewport tea.Cmd
	m.viewport, cmdViewport = m.viewport.Update(msg)

	return m, tea.Batch(cmd, cmdList, cmdViewport, cmdRead)
}
