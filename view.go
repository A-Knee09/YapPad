package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("62")).
	Padding(0, 1)

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	MarginLeft(2)

func (m model) View() string {

	title := titleStyle.Render("YapPad")
	modeStatus := statusStyle.Render(fmt.Sprintf("Mode: %s", m.yapMode))
	sortStatus := statusStyle.Render(fmt.Sprintf("Sort: %s", m.sortMode))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, modeStatus, sortStatus)

	if m.deleting {
		return fmt.Sprintf(
			"\n%s\n\n  Are you sure you want to delete this file? (y/n)\n",
			header,
		)
	}

	if m.inputMode {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			header,
			m.input.View(),
			m.list.View(),
		)
	}

	if m.showPreview {
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s",
			header,
			lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", m.viewport.View()),
			m.help.View(m.keys),
		)
	}

	return fmt.Sprintf(
		"\n%s\n\n%s\n\n%s",
		header,
		m.list.View(),
		m.help.View(m.keys),
	)
}
