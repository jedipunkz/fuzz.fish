package app

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strconv"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jedipunkz/fuzz.fish/internal/config"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// Run starts the application. initialQuery pre-fills the search box (e.g. with
// the current Fish command line) so results are already filtered on startup.
func Run(initialQuery string, cfg config.Config) {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.CharLimit = 156
	s := textinput.DefaultDarkStyles()
	s.Focused.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorCyan))
	s.Focused.Text = lipgloss.NewStyle().Foreground(lipgloss.Color(ui.ColorForeground))
	s.Cursor.Blink = false
	ti.SetStyles(s)
	ti.SetVirtualCursor(false)
	ti.Focus()
	if initialQuery != "" {
		ti.SetValue(initialQuery)
		ti.CursorEnd()
	}

	m := model{
		mode:         ModeHistory,
		keys:         cfg.Keys,
		input:        ti,
		viewport:     viewport.New(),
		previewCache: make(map[string]string),
		loading:      true,
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = tty.Close() }()

	prog := tea.NewProgram(m, tea.WithInput(tty), tea.WithOutput(tty))
	finalModel, err := prog.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}

	if m, ok := finalModel.(model); ok {
		if m.choice != nil {
			switch m.mode {
			case ModeHistory:
				fmt.Printf("CMD:%s", *m.choice)
			case ModeGitBranch:
				if m.fetchBranch {
					cmd := exec.Command("git", "pull")
					cmd.Stdin = tty
					cmd.Stdout = tty
					cmd.Stderr = tty
					if err := cmd.Run(); err != nil {
						fmt.Fprintf(os.Stderr, "git pull failed: %v\n", err)
					}
				} else {
					fmt.Printf("BRANCH:%s", *m.choice)
				}
			case ModeFiles:
				if m.choiceIsDir {
					fmt.Printf("DIR:%s", *m.choice)
				} else {
					fmt.Printf("FILE:%s", *m.choice)
				}
			case ModeWorktree:
				fmt.Printf("DIR:%s", *m.choice)
			case ModePullRequest:
				if dir := checkoutPullRequest(m.choicePR, cfg.WorktreeDir, tty); dir != "" {
					fmt.Printf("DIR:%s", dir)
				}
			case ModeCommit:
				if m.commitIsCmd {
					fmt.Printf("CMD:%s", *m.choice)
				} else {
					fmt.Printf("HASH:%s", *m.choice)
				}
			}
		}
	}
}

// checkoutPullRequest returns the directory to cd into for pr: the worktree
// that already has its head branch checked out, else a worktree created for
// it (see git.PullRequestWorktreePath). `gh pr checkout` then fetches the
// branch inside that worktree, which also covers fork PRs. A failed checkout
// is reported but still returns the directory.
func checkoutPullRequest(pr git.PullRequest, worktreeDir string, tty *os.File) string {
	dir := pr.Worktree
	if dir == "" {
		path, err := git.NewRepository(".").PullRequestWorktreePath(pr, worktreeDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "fuzz: %v\n", err)
			return ""
		}
		// An existing directory is a worktree created on an earlier pick
		// (e.g. a fork PR, which never matches pr.Worktree): reuse it.
		if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
			// --detach: gh pr checkout creates or switches to the branch.
			add := exec.Command("git", "worktree", "add", "--detach", path)
			add.Stdout = tty
			add.Stderr = tty
			if err := add.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "git worktree add failed: %v\n", err)
				return ""
			}
		}
		dir = path
	}

	cmd := exec.Command("gh", "pr", "checkout", strconv.Itoa(pr.Number))
	cmd.Dir = dir
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "gh pr checkout failed: %v\n", err)
	}
	return dir
}
