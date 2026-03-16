package main

import (
	"time"

	"github.com/charmbracelet/bubbles/list"
)

// NOTE: Inbuilt editor using textarea component

type (
	editorSavedMsg  struct{}
	editorClosedMsg struct{}
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
	sortNameDesc
	sortNameAsc
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
	case sortNameDesc:
		return "Alphabetic (Descending)"
	case sortNameAsc:
		return "Alphabetic (Ascending)"
	default:
		return "Unknown"
	}
}

// Ensure item satisfies list.Item at compile time.
var _ list.Item = item{}
