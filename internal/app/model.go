package app

import (
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/jedipunkz/fuzz.fish/internal/files"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
)

// Async load completion messages
type historyLoadedMsg struct{ entries []history.Entry }
type branchesLoadedMsg struct{ branches []git.Branch }
type filesLoadedMsg struct{ entries []files.Entry }

// Filter debounce message
type filterTickMsg struct{ query string }

// SearchMode represents the current search mode
type SearchMode int

const (
	ModeHistory SearchMode = iota
	ModeGitBranch
	ModeFiles
)

// Item represents a search result item
type Item struct {
	Text           string
	Index          int         // Index in the original source slice
	Original       interface{} // The original object (history.Entry, git.Branch, or files.Entry)
	IsCurrent      bool        // For git branch (icon logic)
	IsRemote       bool        // For git branch (icon logic)
	IsDir          bool        // For files (directory indicator)
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
	fileEntries    []files.Entry

	// Items state
	allItems    []Item    // All items for current mode (sorted newest/priority first)
	allItemsStr []string  // Pre-built search strings for fuzzy matching (avoids per-keystroke allocation)
	filtered    []Item    // Filtered items

	cursor      int
	offset      int
	choice      *string // Result string to print
	choiceIsDir bool    // For files mode: whether the choice is a directory
	quitting    bool
	loading     bool   // True while async data loading is in progress

	pendingQuery string // For filter debounce

	width      int
	height     int
	ready      bool
	listWidth  int
	mainHeight int

	// Preview cache
	previewCache     map[string]string // Cache for file previews
	lastPreviewIndex int               // Last previewed item index to avoid re-rendering
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, loadHistoryCmd())
}

func loadHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		p := history.NewParser()
		return historyLoadedMsg{entries: p.Parse()}
	}
}

func loadBranchesCmd() tea.Cmd {
	return func() tea.Msg {
		r := git.NewRepository(".")
		branches, _ := r.Branches()
		return branchesLoadedMsg{branches: branches}
	}
}

func loadFilesCmd() tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return filesLoadedMsg{}
		}
		c := files.NewCollector(cwd)
		return filesLoadedMsg{entries: c.Collect()}
	}
}
