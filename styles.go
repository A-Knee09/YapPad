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
		MoreMuted: lipgloss.Color("235"),
		Text:      lipgloss.Color("252"),
		SubText:   lipgloss.Color("244"),
	},
	"gruvbox": {
		Primary:   lipgloss.Color("214"),
		Secondary: lipgloss.Color("243"),
		Border:    lipgloss.Color("239"),
		Accent:    lipgloss.Color("167"),
		Muted:     lipgloss.Color("244"),
		MoreMuted: lipgloss.Color("236"),
		Text:      lipgloss.Color("223"),
		SubText:   lipgloss.Color("244"),
	},
	"nord": {
		Primary:   lipgloss.Color("110"),
		Secondary: lipgloss.Color("109"),
		Border:    lipgloss.Color("103"),
		Accent:    lipgloss.Color("111"),
		Muted:     lipgloss.Color("67"),
		MoreMuted: lipgloss.Color("103"),
		Text:      lipgloss.Color("189"),
		SubText:   lipgloss.Color("103"),
	},
	"tokyonight": {
		Primary:   lipgloss.Color("111"),
		Secondary: lipgloss.Color("110"),
		Border:    lipgloss.Color("111"),
		Accent:    lipgloss.Color("141"),
		Muted:     lipgloss.Color("103"),
		MoreMuted: lipgloss.Color("236"),
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
		MoreMuted: lipgloss.Color("238"),
		Text:      lipgloss.Color("254"),
		SubText:   lipgloss.Color("246"),
	},
	"catppuccin": {
		Primary:   lipgloss.Color("110"), // lavender
		Secondary: lipgloss.Color("180"), // peach
		Border:    lipgloss.Color("238"),
		Accent:    lipgloss.Color("150"), // green
		Muted:     lipgloss.Color("245"),
		MoreMuted: lipgloss.Color("236"),
		Text:      lipgloss.Color("255"),
		SubText:   lipgloss.Color("248"),
	},
	"dracula": {
		Primary:   lipgloss.Color("212"), // pink
		Secondary: lipgloss.Color("141"), // purple
		Border:    lipgloss.Color("239"),
		Accent:    lipgloss.Color("81"), // cyan
		Muted:     lipgloss.Color("245"),
		MoreMuted: lipgloss.Color("236"),
		Text:      lipgloss.Color("255"),
		SubText:   lipgloss.Color("246"),
	},
	"dusk": {
		Primary:   lipgloss.Color("97"),  // deep violet
		Secondary: lipgloss.Color("181"), // dusty rose
		Border:    lipgloss.Color("60"),  // dark violet
		Accent:    lipgloss.Color("175"), // warm pink
		Muted:     lipgloss.Color("245"), // muted purple
		MoreMuted: lipgloss.Color("60"),  // very dark violet
		Text:      lipgloss.Color("254"), // near white
		SubText:   lipgloss.Color("249"), // light grey
	},
	"tide": {
		Primary:   lipgloss.Color("67"),  // slate blue
		Secondary: lipgloss.Color("115"), // seafoam
		Border:    lipgloss.Color("67"),  // ocean blue
		Accent:    lipgloss.Color("216"), // coral
		Muted:     lipgloss.Color("109"), // muted teal
		MoreMuted: lipgloss.Color("60"),  // deep ocean
		Text:      lipgloss.Color("255"), // bright white
		SubText:   lipgloss.Color("152"), // light seafoam
	},
	"moss": {
		Primary:   lipgloss.Color("66"),  // dark green
		Secondary: lipgloss.Color("145"), // sage
		Border:    lipgloss.Color("65"),  // forest border
		Accent:    lipgloss.Color("179"), // amber
		Muted:     lipgloss.Color("108"), // muted green
		MoreMuted: lipgloss.Color("235"), // very dark green
		Text:      lipgloss.Color("254"), // near white
		SubText:   lipgloss.Color("151"), // light sage
	},
	"glacier": {
		Primary:   lipgloss.Color("109"), // glacier blue
		Secondary: lipgloss.Color("152"), // frost
		Border:    lipgloss.Color("67"),  // ice blue
		Accent:    lipgloss.Color("109"), // soft teal
		Muted:     lipgloss.Color("110"), // muted glacier
		MoreMuted: lipgloss.Color("66"),  // deep ice
		Text:      lipgloss.Color("255"), // bright white
		SubText:   lipgloss.Color("152"), // light frost
	},
	"plum": {
		Primary:   lipgloss.Color("96"),  // deep plum
		Secondary: lipgloss.Color("146"), // lavender
		Border:    lipgloss.Color("96"),  // plum border
		Accent:    lipgloss.Color("109"), // mint
		Muted:     lipgloss.Color("139"), // muted purple
		MoreMuted: lipgloss.Color("236"), // very dark plum
		Text:      lipgloss.Color("255"), // bright white
		SubText:   lipgloss.Color("182"), // light lavender
	},
	"algae": {
		Primary:   lipgloss.Color("107"), // #628141 green
		Secondary: lipgloss.Color("252"), // #E5D9B6 cream
		Border:    lipgloss.Color("238"), // #40513B dark green
		Accent:    lipgloss.Color("209"), // #E67E22 orange
		Muted:     lipgloss.Color("107"), // green muted
		MoreMuted: lipgloss.Color("238"), // dark green
		Text:      lipgloss.Color("252"), // cream
		SubText:   lipgloss.Color("245"), // muted cream
	},
}

func getTheme(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}
