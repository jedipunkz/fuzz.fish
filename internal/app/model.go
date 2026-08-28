package app

import (
	"os"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/files"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
)

// Async load completion messages
type historyLoadedMsg struct{ entries []history.Entry }
type branchesLoadedMsg struct{ branches []git.Branch }
type filesLoadedMsg struct{ entries []files.Entry }
type worktreesLoadedMsg struct{ worktrees []git.Worktree }
type commitsLoadedMsg struct{ commits []git.Commit }

// Filter debounce message
type filterTickMsg struct{ query string }

// SearchMode represents the current search mode
type SearchMode int

const (
	ModeHistory SearchMode = iota
	ModeGitBranch
	ModeFiles
	ModeWorktree
	ModeCommit
)

// commitActions are the commands offered after selecting a commit. The
// selected template is joined with the short hash to build the command line;
// an empty template inserts the bare hash.
var commitActions = []struct {
	Label    string
	Template string
}{
	{"git show", "git show"},
	{"git diff", "git diff"},
	{"git revert", "git revert"},
	{"git cherry-pick", "git cherry-pick"},
	{"git rebase --onto", "git rebase --onto"},
	{"hash only", ""},
}

// Item represents a search result item
type Item struct {
	Text           string
	SearchText     string      // Fuzzy match target; falls back to Text when empty (e.g. worktree path + branch)
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
	worktrees      []git.Worktree
	commits        []git.Commit

	// Items state
	allItems       []Item           // All items for current mode (sorted newest/priority first)
	allItemsStr    []string         // Pre-built search strings for fuzzy matching (avoids per-keystroke allocation)
	filtered       []Item           // Filtered items

	cursor      int
	offset      int
	choice      *string // Result string to print
	choiceIsDir bool    // For files mode: whether the choice is a directory
	fetchBranch bool    // True when ctrl+g selects current branch for git pull

	// Commit action picker: non-empty pendingCommit means the picker is open
	// and the list keys drive the action list instead of the commit list.
	pendingCommit string
	actionCursor  int
	commitIsCmd   bool // True when the picked action produced a full command line
	quitting    bool
	statusMsg   string  // Transient status message (e.g., warning)
	loading     bool   // True while async data loading is in progress

	pendingQuery string // For filter debounce

	width      int
	height     int
	ready      bool
	listWidth  int
	mainHeight int

	// Preview cache
	previewCache     map[string]string // Cache for file previews
	lastPreviewKey   string            // Identifies the item the preview was rendered for
}

// Init initializes the model
func (m model) Init() tea.Cmd {
	return loadHistoryCmd()
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

func loadCommitsCmd() tea.Cmd {
	return func() tea.Msg {
		r := git.NewRepository(".")
		commits, _ := r.Commits()
		return commitsLoadedMsg{commits: commits}
	}
}

func loadWorktreesCmd() tea.Cmd {
	return func() tea.Msg {
		r := git.NewRepository(".")
		worktrees, _ := r.Worktrees()
		return worktreesLoadedMsg{worktrees: worktrees}
	}
}
