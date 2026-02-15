package main

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("161")).
		Width(60).
		Align(lipgloss.Center)

	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("111"))

	cursorLineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("38"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("141"))

	vaultDir string
)
