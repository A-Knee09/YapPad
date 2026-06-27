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

func (m model) settingsOverlay() string {
	// Theme list
	var themeSb strings.Builder
	for i, t := range themeOptions {
		if i == m.themeIndex && m.settingsTab == 0 {
			themeSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render("● " + t))
			themeSb.WriteString("\n")
		} else if i == m.themeIndex {
			themeSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).PaddingLeft(2).Render(t))
			themeSb.WriteString("\n")
		} else {
			themeSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.SubText).PaddingLeft(2).Render(t))
			themeSb.WriteString("\n")
		}
	}

	// Palette swatches
	var paletteSb strings.Builder
	swatches := []struct {
		name  string
		color lipgloss.Color
	}{
		{"Primary  ", m.theme.Primary},
		{"Secondary", m.theme.Secondary},
		{"Accent   ", m.theme.Accent},
		{"Border   ", m.theme.Border},
		{"Muted    ", m.theme.Muted},
		{"Text     ", m.theme.Text},
		{"SubText  ", m.theme.SubText},
	}
	for _, s := range swatches {
		swatch := lipgloss.NewStyle().Background(s.color).Foreground(s.color).Render("   ")
		label := lipgloss.NewStyle().Foreground(m.theme.SubText).Render(" " + s.name)
		paletteSb.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, swatch, label))
		paletteSb.WriteString("\n")
	}

	themeTitle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Muted).Render("Theme")
	paletteTitle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Muted).Render("Palette")
	if m.settingsTab == 0 {
		themeTitle = lipgloss.NewStyle().Bold(true).Foreground(m.theme.Primary).Render("Theme")
	}

	themeCol := lipgloss.JoinVertical(lipgloss.Left, themeTitle, "", themeSb.String())
	paletteCol := lipgloss.JoinVertical(lipgloss.Left, paletteTitle, "", paletteSb.String())
	themeAndPalette := lipgloss.JoinHorizontal(lipgloss.Top, themeCol, "    ", paletteCol)

	themeBorderColor := m.theme.Border
	if m.settingsTab == 0 {
		themeBorderColor = m.theme.Primary
	}
	themeBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(themeBorderColor).
		Padding(1, 2).
		Render(themeAndPalette)

	// Editor box
	var editorSb strings.Builder
	for i, e := range editorOptions {
		if i == m.editorIndex && m.settingsTab == 1 {
			editorSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Primary).Bold(true).Render("● " + e))
			editorSb.WriteString("\n")
		} else if i == m.editorIndex {
			editorSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.Accent).Bold(true).PaddingLeft(2).Render(e))
			editorSb.WriteString("\n")
		} else {
			editorSb.WriteString(lipgloss.NewStyle().Foreground(m.theme.SubText).PaddingLeft(2).Render(e))
			editorSb.WriteString("\n")
		}
	}

	editorTitle := lipgloss.NewStyle().Bold(true).Foreground(m.theme.Muted).Render("Editor")
	if m.settingsTab == 1 {
		editorTitle = lipgloss.NewStyle().Bold(true).Foreground(m.theme.Primary).Render("Editor")
	}

	editorBorderColor := m.theme.Border
	if m.settingsTab == 1 {
		editorBorderColor = m.theme.Primary
	}
	editorBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(editorBorderColor).
		Padding(1, 2).
		Render(editorTitle + "\n\n" + editorSb.String())

	// Save + footer
	saveStr := "[ ] Save to config"
	if m.settingsSave {
		saveStr = "[●] Save to config"
	}
	save := lipgloss.NewStyle().Foreground(m.theme.Secondary).Render(saveStr)
	footer := lipgloss.NewStyle().Foreground(m.theme.Muted).Render("tab: switch  s: save  enter: confirm  esc: cancel")

	boxes := lipgloss.JoinHorizontal(lipgloss.Top, themeBox, "  ", editorBox)
	content := lipgloss.JoinVertical(lipgloss.Left, boxes, "\n"+save, footer)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
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
	sortStatus := m.statusStyle().Render(fmt.Sprintf("Sort: %s", m.sortMode))
	header := lipgloss.JoinHorizontal(lipgloss.Center, title, sortStatus)

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
		if m.editorPreview {
			previewView := fmt.Sprintf("%s\n%s\n%s", m.previewHeader(), m.viewport.View(), m.previewFooter())
			status := m.statusStyle().Render("ctrl+f: back to editing  ctrl+q: close")
			return fmt.Sprintf("\n%s\n\n%s",
				lipgloss.JoinHorizontal(lipgloss.Center, title, status),
				previewView,
			)
		}
		editorStatus := m.statusStyle().Render("ctrl+s: save  ctrl+f: preview  ctrl+q: close")
		return fmt.Sprintf("\n%s\n\n%s",
			lipgloss.JoinHorizontal(lipgloss.Center, title, editorStatus),
			m.editorContent.View(),
		)
	}

	if m.settingsOpen {
		return m.settingsOverlay()
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
	if m.fullPreview {
		m.viewport.Width = m.width - 4
		m.viewport.Height = m.height - 6
		previewView := fmt.Sprintf("%s\n%s\n%s", m.previewHeader(), m.viewport.View(), m.previewFooter())
		return fmt.Sprintf("\n%s\n\n%s", header, previewView)
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
