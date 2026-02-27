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
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		const minWidthForPreview = 80

		if msg.Width < minWidthForPreview {
			m.showPreview = false
		} else {
			m.showPreview = true
		}

		var listWidth, viewportWidth int

		if m.showPreview {
			listWidth = msg.Width / 2
			viewportWidth = msg.Width - listWidth - 4
		} else {
			listWidth = msg.Width - 2
			viewportWidth = 0
		}

		m.viewport.Width = viewportWidth
		m.viewport.Height = msg.Height - 6

		m.list.SetSize(listWidth, msg.Height-6)

		if !m.ready {
			m.viewport = viewport.New(viewportWidth, msg.Height-6)
			m.viewport.HighPerformanceRendering = false
			m.ready = true

			// Initial load if something is selected
			if m.list.SelectedItem() != nil {
				i := m.list.SelectedItem().(item)
				m.selectedFile = i.title
				return m, m.loadFileOrImage(m.resolveFilePath(i.title))
			}

		} else {
			m.viewport.Width = viewportWidth
			m.viewport.Height = msg.Height - 6
		}

	case fileLoadedMsg:
		m.showingImage = false
		m.viewport.SetContent(msg.content)
		m.viewport.GotoTop()

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
		return m, tea.EnableMouseAllMotion

	case tea.KeyMsg:
		// DELETE CONFIRMATION MODE
		if m.deleting {
			switch msg.String() {
			case "y", "Y":
				if it, ok := m.list.SelectedItem().(item); ok {
					os.Remove(m.resolveFilePath(it.title))
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
				return m, openInEditor(path)

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
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.showPreview = !m.showPreview
			return m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

		case key.Matches(msg, m.keys.ToggleHelpMenu):
			m.list.SetShowHelp(!m.list.ShowHelp())

			if m.list.ShowHelp() {
				m.list.Help.Width = m.width // Use full width for help
			} else {
				m.list.Help.Width = m.width / 3 // Back to list width
			}
			return m, nil

		case msg.String() == "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			if it, ok := m.list.SelectedItem().(item); ok {
				path := m.resolveFilePath(it.title)
				return m, openInEditor(path)
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
			cmdRead = m.loadFileOrImage(m.resolveFilePath(i.title))
		}
	}

	var cmdViewport tea.Cmd
	m.viewport, cmdViewport = m.viewport.Update(msg)

	return m, tea.Batch(cmd, cmdList, cmdViewport, cmdRead)
}
