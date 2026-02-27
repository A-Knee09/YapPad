package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

var viewportStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("237"))

	// Padding(0, 2).
	// Margin(0, 10)

var titleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("230")).
	Background(lipgloss.Color("62")).
	Padding(0, 1).MarginLeft(2)

var statusStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("241")).
	MarginLeft(2)

var listItemStyles = func() (s list.DefaultItemStyles) {
	s = list.NewDefaultItemStyles()

	s.NormalTitle = s.NormalTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalDesc.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)

	s.SelectedTitle = s.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#EE6FF8", Dark: "#EE6FF8"}).
		Bold(true).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(lipgloss.AdaptiveColor{Light: "#F793FF", Dark: "#AD58B4"}).
		Foreground(lipgloss.AdaptiveColor{Light: "#AD58B4", Dark: "#AD58B4"}).
		Padding(0, 0, 0, 1)

	s.DimmedTitle = s.DimmedTitle.
		Foreground(lipgloss.AdaptiveColor{Light: "#A49FA5", Dark: "#777777"}).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedDesc.
		Foreground(lipgloss.AdaptiveColor{Light: "#C2B8C2", Dark: "#4D4D4D"}).
		Padding(0, 0, 0, 2)

	s.FilterMatch = s.FilterMatch.
		Foreground(lipgloss.Color("#ff00ff")).
		Underline(true)

	return s
}()

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
		if m.inputStep == 0 {
			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s",
				header,
				m.input.View(),
				m.list.View(),
			)
		}
		return fmt.Sprintf(
			"\n%s\n\nFilename: %s\nDescription: %s",
			header,
			m.input.Value(),
			m.descInput.View(),
		)
	}

	if m.showPreview {
		var previewView string
		if m.showingImage {
			previewView = m.viewport.View()
		} else {
			previewView = viewportStyle.Render(m.viewport.View())
		}

		return fmt.Sprintf(
			"\n%s\n\n%s",
			header,
			lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", previewView),
		)
	}
	return fmt.Sprintf(
		"\n%s\n\n%s",
		header,
		m.list.View(),
	)
}
