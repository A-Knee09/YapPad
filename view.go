package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	title := titleStyle.Render("YapPad")
	modeStatus := statusStyle.Render(fmt.Sprintf("Mode: %s", m.yapMode))
	sortStatus := statusStyle.Render(fmt.Sprintf("Sort: %s", m.sortMode))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, modeStatus, sortStatus)

	if m.deleting {
		if m.showPreview {
			var previewView string
			if m.showingImage {
				previewView = m.viewport.View()
			} else {
				previewView = fmt.Sprintf("%s\n%s\n%s", m.previewHeader(), m.viewport.View(), m.previewFooter())
			}
			return fmt.Sprintf(
				"\n%s\n\n  Are you sure you want to delete this file? (y/n)\n\n%s",
				header,
				lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", previewView),
			)
		}
		return fmt.Sprintf(
			"\n%s\n\n  Are you sure you want to delete this file? (y/n)\n\n%s\n",
			header,
			m.list.View(),
		)
	}

	if m.editorMode {
		editorStatus := statusStyle.Render("ctrl+s: save  ctrl+q: close")
		return fmt.Sprintf(
			"\n%s\n\n%s",
			lipgloss.JoinHorizontal(lipgloss.Center, title, editorStatus),
			m.editorContent.View(),
		)
	}

	if m.inputMode {
		if m.inputStep == 0 {
			return fmt.Sprintf(
				"\n%s\n\n  File Name %s\n\n%s",
				header,
				m.input.View(),
				m.list.View(),
			)
		}
		return fmt.Sprintf(
			"\n%s\n\n File Name %s\n Description %s\n\n%s",
			header,
			m.input.View(),
			m.descInput.View(),
			m.list.View(),
		)
	}

	if m.showPreview {
		listWidth := m.width / 2
		spacer := strings.Repeat(" ", max(0, listWidth-lipgloss.Width(m.list.View())))
		var previewView string
		if m.loadingFile {
			previewView = fmt.Sprintf("%s\n\n  %s Loading...", m.previewHeader(), m.spinner.View())
		} else if m.showingImage {
			previewView = m.viewport.View()
		} else {
			previewView = fmt.Sprintf("%s\n%s\n%s", m.previewHeader(), m.viewport.View(), m.previewFooter())
		}
		return fmt.Sprintf(
			"\n%s\n\n%s",
			header,
			lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), spacer, previewView),
		)
	}

	return fmt.Sprintf(
		"\n%s\n\n%s",
		header,
		m.list.View(),
	)
}
