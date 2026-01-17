package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
)

// SearchMode represents the current search mode
type SearchMode int

const (
	ModeHistory SearchMode = iota
	ModeGitBranch
)

// Item represents a search result item
type Item struct {
	Text           string
	Index          int         // Index in the original source slice
	Original       interface{} // The original object (history.Entry or git.Branch)
	IsCurrent      bool        // For git branch (icon logic)
	IsRemote       bool        // For git branch (icon logic)
	MatchedIndexes []int       // Indexes of matched characters for highlighting
}

// model represents the application state
type model struct {
	mode     SearchMode
	input    textinput.Model
	viewport viewport.Model

	// Data sources
	historyEntries []history.Entry
	gitBranches    []git.Branch

	// Items state
	allItems []Item // All items for current mode (sorted newest/priority first)
	filtered []Item // Filtered items

	cursor   int
	offset   int
	choice   *string // Result string to print
	quitting bool

	width      int
	height     int
	ready      bool
	listWidth  int
	mainHeight int
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return textinput.Blink
}
