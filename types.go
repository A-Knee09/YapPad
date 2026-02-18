package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// Tea messages

type fileEditedMsg struct {
	err error
}

type fileLoadedMsg struct {
	content string
}

type imageRenderedMsg struct{}

type clearViewportMsg struct{}

// List item

type item struct {
	title   string
	desc    string
	modTime time.Time
	creTime time.Time
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

// Sort modes

type sortMode int

const (
	sortModifiedDesc sortMode = iota
	sortModifiedAsc
	sortCreatedDesc
	sortCreatedAsc
)

func (s sortMode) String() string {
	switch s {
	case sortModifiedDesc:
		return "Modified (Newest)"
	case sortModifiedAsc:
		return "Modified (Oldest)"
	case sortCreatedDesc:
		return "Created (Newest)"
	case sortCreatedAsc:
		return "Created (Oldest)"
	default:
		return "Unknown"
	}
}

// Yap modes (journal types)

type yapMode int

const (
	yapAll     yapMode = iota // 0 — show everything
	yapDaily                  // 1 — default
	yapWeekly                 // 2
	yapMonthly                // 3
	yapYearly                 // 4
)

func (y yapMode) String() string {
	switch y {
	case yapAll:
		return "All"
	case yapDaily:
		return "Daily"
	case yapWeekly:
		return "Weekly"
	case yapMonthly:
		return "Monthly"
	case yapYearly:
		return "Yearly"
	default:
		return "Unknown"
	}
}

// yapSubdir returns the subdirectory for a yap mode.
func (y yapMode) subdir() string {
	switch y {
	case yapDaily:
		return "daily"
	case yapWeekly:
		return "weekly"
	case yapMonthly:
		return "monthly"
	case yapYearly:
		return "yearly"
	default:
		return ""
	}
}

// defaultNoteName returns the default journal filename for the current time.
func (y yapMode) defaultNoteName() string {
	now := time.Now()
	switch y {
	case yapDaily, yapAll:
		return now.Format("2006-01-02") + ".md"
	case yapWeekly:
		year, week := now.ISOWeek()
		return fmt.Sprintf("%d-W%02d.md", year, week)
	case yapMonthly:
		return now.Format("2006-01") + ".md"
	case yapYearly:
		return now.Format("2006") + ".md"
	default:
		return now.Format("2006-01-02") + ".md"
	}
}

// defaultNoteDir returns the subdirectory for the default note.
// For yapAll, defaults to daily.
func (y yapMode) defaultNoteDir() string {
	if y == yapAll {
		return "daily"
	}
	return y.subdir()
}

// Ensure item satisfies list.Item at compile time.
var _ list.Item = item{}
