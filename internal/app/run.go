package app

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

// Run starts the application. initialQuery pre-fills the search box (e.g. with
// the current Fish command line) so results are already filtered on startup.
// keys maps a key string to a config.Action* name.
func Run(initialQuery string, keys map[string]string) {
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
		keys:         keys,
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
				if dir := checkoutPullRequest(m.choicePR, tty); dir != "" {
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

// checkoutPullRequest fetches the pull request's head branch with
// `gh pr checkout` and returns the directory to cd into: the worktree that
// already has the branch checked out, else the current worktree's root.
// A failed checkout is reported but still returns the directory.
func checkoutPullRequest(pr git.PullRequest, tty *os.File) string {
	dir := pr.Worktree
	if dir == "" {
		out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "git rev-parse failed: %v\n", err)
			return ""
		}
		dir = strings.TrimSpace(string(out))
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
