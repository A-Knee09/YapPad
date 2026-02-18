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

		var listWidth, viewportWidth int
		if m.showPreview {
			listWidth = msg.Width / 3
			viewportWidth = msg.Width - listWidth - 4
		} else {
			listWidth = msg.Width - 2
			viewportWidth = 0
		}

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
				m.viewport.LineUp(1)
			case tea.MouseButtonWheelDown:
				m.viewport.LineDown(1)
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
				name := m.input.Value()

				if m.renameMode {
					if name == "" {
						break
					}
					oldPath := m.resolveFilePath(m.renameTarget)
					newPath := filepath.Join(vaultDir, name)

					if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
						// handle error or ignore
					}
					os.Rename(oldPath, newPath)

					m.renameMode = false
					m.inputMode = false
					m.input.SetValue("")
					m.list.SetItems(listFiles(m.sortMode, m.yapMode))
					return m, nil
				}

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

				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					// Handle error
				}

				// Create the file only if it doesn't already exist
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

				m.inputMode = false
				m.input.SetValue("")
				return m, openInEditor(path)

			case "tab":
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
				return m, nil

			case "esc":
				m.inputMode = false
				m.renameMode = false
				m.input.SetValue("")
				m.list.SetItems(listFiles(m.sortMode, m.yapMode))
				return m, nil
			}

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

			return m, cmd
		}

		// NORMAL MODE
		switch {

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Sequence(clearKittyGraphics(), tea.Quit)

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
				m.input.SetValue(it.title)
				m.input.Focus()
			}
			return m, nil

		case key.Matches(msg, m.keys.CycleSort):
			m.sortMode = (m.sortMode + 1) % 4
			m.list.SetItems(listFiles(m.sortMode, m.yapMode))
			return m, nil

		case key.Matches(msg, m.keys.TogglePreview):
			m.showPreview = !m.showPreview
			return m.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})

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
