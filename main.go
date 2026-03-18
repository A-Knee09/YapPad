/*
NOTE:
This is the entry point and deals with
CLI flag parsing (--mode, --editor, --theme, --version),
sets up vault directory
and launches the Bubble Tea program.
*/
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

var Version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println(Version)
		return
	}

	cfgFile := configPath()
	var cfg Config
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		cfg = runSetup()
	} else {
		cfg = loadConfig()
	}

	// Parse flags — override config values if explicitly provided
	themeFlag := flag.String("theme", "", "")
	editorFlag := flag.String("editor", "", "")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `YapPad %s — a terminal note-taking app

Usage:
  yap [flags] [.]

  yap          open your configured vault
  yap .        open current directory as vault (session only)

Flags:
  --theme <name>    override config theme for this session
  --editor <name>   override config editor for this session
  --version         print version and exit
  --help            show this help

Themes:
  default, gruvbox, nord, tokyonight, forest, solarized,
  dracula, dusk, tide, moss, glacier, plum

Editors:
  inbuilt, nano, nvim, vim, hx

Config:
  %s

Keybindings:
  n          new file
  r          rename file
  d          delete file
  enter      open in editor
  ctrl+p     toggle preview
  ctrl+s     cycle sort
  /          filter
  ?          toggle help
  q          quit
`, Version, configPath())
	}
	flag.Parse()

	if *themeFlag != "" {
		cfg.Theme = *themeFlag
	}
	if *editorFlag != "" {
		cfg.Editor = *editorFlag
	}

	// Override vault with current directory if `yap .` is used
	if flag.NArg() > 0 && flag.Arg(0) == "." {
		cwd, err := os.Getwd()
		if err == nil {
			cfg.Vault = cwd
		}
	}

	vaultDir = cfg.Vault

	p := tea.NewProgram(
		initialModel(cfg.Editor, cfg.Theme),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
