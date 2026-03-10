package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func (m model) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(m.theme.Primary).
		Padding(0, 1).MarginLeft(2)
}

func (m model) listTitleStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("230")).
		Background(m.theme.Primary).
		Padding(0, 1)
}

func (m model) statusStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(m.theme.Secondary).
		MarginLeft(2)
}

func (m model) previewHeaderStyle() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Right = "├"
	return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).BorderForeground(m.theme.Border)
}

func (m model) previewFooterStyle() lipgloss.Style {
	b := lipgloss.RoundedBorder()
	b.Left = "┤"
	return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1).BorderForeground(m.theme.Border)
}

func (m model) listItemStyles() list.DefaultItemStyles {
	s := list.NewDefaultItemStyles()

	s.NormalTitle = s.NormalTitle.
		Foreground(m.theme.Text).
		Padding(0, 0, 0, 2)

	s.NormalDesc = s.NormalDesc.
		Foreground(m.theme.SubText).
		Padding(0, 0, 0, 2)

	s.SelectedTitle = s.SelectedTitle.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(m.theme.Accent).
		Foreground(m.theme.Accent).
		Bold(true).
		Padding(0, 0, 0, 1)

	s.SelectedDesc = s.SelectedDesc.
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(m.theme.Accent).
		Foreground(m.theme.Secondary).
		Padding(0, 0, 0, 1)

	s.DimmedTitle = s.DimmedTitle.
		Foreground(m.theme.Muted).
		Padding(0, 0, 0, 2)

	s.DimmedDesc = s.DimmedDesc.
		Foreground(m.theme.MoreMuted).
		Padding(0, 0, 0, 2)

	s.FilterMatch = s.FilterMatch.
		Foreground(m.theme.Accent).
		Underline(true)

	return s
}

func (m model) previewHeader() string {
	title := m.previewHeaderStyle().Render(m.selectedFile)
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title))),
	)
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) previewFooter() string {
	info := m.previewFooterStyle().Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := lipgloss.NewStyle().Foreground(m.theme.Border).Render(
		strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info))),
	)
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func (m model) View() string {
	delegate := list.NewDefaultDelegate()
	delegate.Styles = m.listItemStyles()
	m.list.SetDelegate(delegate)
	m.list.Styles.Title = m.listTitleStyle()
	m.list.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(m.theme.Primary)
	m.list.Styles.FilterCursor = lipgloss.NewStyle().Foreground(m.theme.Accent)
	m.list.Styles.StatusBar = lipgloss.NewStyle().
		Foreground(m.theme.Muted).
		Padding(0, 0, 1, 2)
	m.list.Styles.StatusBarFilterCount = lipgloss.NewStyle().Foreground(m.theme.Accent)

	title := m.titleStyle().Render("YapPad")
	modeStatus := m.statusStyle().Render(fmt.Sprintf("Mode: %s", m.yapMode))
	sortStatus := m.statusStyle().Render(fmt.Sprintf("Sort: %s", m.sortMode))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, modeStatus, sortStatus)

	deletePrompt := lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).Render("  Are you sure you want to delete this file?") +
		lipgloss.NewStyle().Foreground(m.theme.Secondary).Render(" (y/n)")

	if m.deleting {
		if m.showPreview {
			var previewView string
			if m.showingImage {
				previewView = m.viewport.View()
			} else {
				previewView = fmt.Sprintf("%s\n%s\n%s", m.previewHeader(), m.viewport.View(), m.previewFooter())
			}
			return fmt.Sprintf(
				"\n%s\n\n%s\n\n%s",
				header,
				deletePrompt,
				lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", previewView),
			)
		}
		return fmt.Sprintf(
			"\n%s\n\n%s\n\n%s\n",
			header,
			deletePrompt,
			m.list.View(),
		)
	}

	if m.editorMode {
		editorStatus := m.statusStyle().Render("ctrl+s: save  ctrl+q: close")
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
