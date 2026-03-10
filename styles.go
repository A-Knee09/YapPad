// NOTE: File for setting styling. Will make changes here for themes

package main

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Border    lipgloss.Color
	Accent    lipgloss.Color
	Muted     lipgloss.Color
	MoreMuted lipgloss.Color
	Text      lipgloss.Color
	SubText   lipgloss.Color
}

var themes = map[string]Theme{
	"default": {
		Primary:   lipgloss.Color("62"),
		Secondary: lipgloss.Color("241"),
		Border:    lipgloss.Color("237"),
		Accent:    lipgloss.Color("135"),
		Muted:     lipgloss.Color("240"),
		Text:      lipgloss.Color("252"),
		SubText:   lipgloss.Color("244"),
	},
	"gruvbox": {
		Primary:   lipgloss.Color("214"),
		Secondary: lipgloss.Color("243"),
		Border:    lipgloss.Color("239"),
		Accent:    lipgloss.Color("167"),
		Muted:     lipgloss.Color("244"),
		Text:      lipgloss.Color("223"),
		SubText:   lipgloss.Color("244"),
	},
	"nord": {
		Primary:   lipgloss.Color("110"),
		Secondary: lipgloss.Color("109"),
		Border:    lipgloss.Color("103"),
		Accent:    lipgloss.Color("111"),
		Muted:     lipgloss.Color("103"),
		Text:      lipgloss.Color("189"),
		SubText:   lipgloss.Color("103"),
	},
	"tokyonight": {
		Primary:   lipgloss.Color("111"),
		Secondary: lipgloss.Color("110"),
		Border:    lipgloss.Color("237"),
		Accent:    lipgloss.Color("141"),
		Muted:     lipgloss.Color("103"),
		Text:      lipgloss.Color("189"),
		SubText:   lipgloss.Color("103"),
	},

	"forest": {
		Primary:   lipgloss.Color("71"),  // green
		Secondary: lipgloss.Color("108"), // moss
		Border:    lipgloss.Color("58"),
		Accent:    lipgloss.Color("179"), // amber
		Muted:     lipgloss.Color("94"),
		MoreMuted: lipgloss.Color("58"),
		Text:      lipgloss.Color("253"),
		SubText:   lipgloss.Color("246"),
	},
	"solarized": {
		Primary:   lipgloss.Color("136"), // yellow
		Secondary: lipgloss.Color("37"),  // cyan
		Border:    lipgloss.Color("240"), // gray
		Accent:    lipgloss.Color("166"), // orange
		Muted:     lipgloss.Color("244"),
		Text:      lipgloss.Color("254"),
		SubText:   lipgloss.Color("246"),
	},
	"catppuccin": {
		Primary:   lipgloss.Color("110"), // lavender
		Secondary: lipgloss.Color("180"), // peach
		Border:    lipgloss.Color("238"),
		Accent:    lipgloss.Color("150"), // green
		Muted:     lipgloss.Color("245"),
		Text:      lipgloss.Color("255"),
		SubText:   lipgloss.Color("248"),
	},
	"dracula": {
		Primary:   lipgloss.Color("212"), // pink
		Secondary: lipgloss.Color("141"), // purple
		Border:    lipgloss.Color("239"),
		Accent:    lipgloss.Color("81"), // cyan
		Muted:     lipgloss.Color("245"),
		Text:      lipgloss.Color("255"),
		SubText:   lipgloss.Color("246"),
	},
}

func getTheme(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}
