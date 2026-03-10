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
	"log"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	vaultDir       string
	defaultYapMode yapMode = yapAll
	Version                = "v1.0.0-dev"
)

func main() {
	modeFlag := flag.String("mode", "all", "")
	editorFlag := flag.String("editor", "", "editor to use: nano, nvim, or inbuilt")
	versionFlag := flag.Bool("version", false, "Print version")
	themeFlag := flag.String("theme", "default", "theme: default, algae, gruvbox, nord, tokyonight")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `YapPad — a terminal journal & note-taking app

Usage:
  yap [options] [vault-dir]

Options:
  --mode <mode>  Set default yap mode (default: all)
                 Modes: all, daily, weekly, monthly, yearly

  --editor <editor name> Run with nvim or nano
  --version      Print version information

Vault Directory:
  Optional path to the notes directory.
  Defaults to ~/.YapPad

Keybindings:
  ctrl+n       Create new note
  ctrl+r       Rename selected note
  ctrl+d       Delete selected note
  ctrl+p       Toggle preview pane
  ctrl+s       Cycle sort mode

  0-4          Switch yap mode (0=all, 1=daily, 2=weekly, 3=monthly, 4=yearly)
  tab          Cycle yap mode while creating a note
  enter        Open selected note in editor
  /            Filter notes

Examples:
  yap                        Open default vault in "all" mode
  yap --mode daily           Open in daily journal mode
  yap --mode weekly ~/notes  Open ~/notes in weekly mode
`)
	}

	flag.Parse()

	if *versionFlag {
		fmt.Printf("YapPad version %s\n", Version)
		os.Exit(0)
	}

	switch strings.ToLower(*editorFlag) {
	case "", "nano", "nvim", "inbuilt":
	default:
		log.Fatalf("unknown editor: %s (use nano, nvim, or inbuilt)", *editorFlag)
	}

	switch strings.ToLower(*modeFlag) {
	case "all", "0":
		defaultYapMode = yapAll
	case "daily", "1":
		defaultYapMode = yapDaily
	case "weekly", "2":
		defaultYapMode = yapWeekly
	case "monthly", "3":
		defaultYapMode = yapMonthly
	case "yearly", "4":
		defaultYapMode = yapYearly
	default:
		log.Fatalf("unknown mode: %s (use all, daily, weekly, monthly, yearly)", *modeFlag)
	}

	if flag.NArg() > 0 {
		vaultDir = flag.Arg(0)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		vaultDir = filepath.Join(home, ".YapPad")
	}

	p := tea.NewProgram(initialModel(*editorFlag, *themeFlag), tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
