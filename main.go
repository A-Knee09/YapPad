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

var vaultDir string
var defaultYapMode yapMode = yapAll

func main() {
	modeFlag := flag.String("mode", "all", "")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `YapPad — a terminal journal & note-taking app

Usage:
  yap [options] [vault-dir]

Options:
  --mode <mode>  Set default yap mode (default: all)
                 Modes: all, daily, weekly, monthly, yearly

Vault Directory:
  Optional path to the notes directory.
  Defaults to ~/.YapPad

Keybindings:
  ctrl+n       Create new note
  ctrl+r       Rename selected note
  ctrl+d       Delete selected note
  ctrl+p       Toggle preview pane
  ctrl+s       Cycle sort mode
  ctrl+c       Quit

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

	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseAllMotion())

	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
