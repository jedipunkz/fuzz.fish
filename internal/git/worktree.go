package git

import (
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree entry
type Worktree struct {
	Path      string
	Branch    string // short branch name, or "(detached)" / "(bare)"
	Head      string // short commit hash
	IsCurrent bool   // whether this worktree is the one we are running in
}

// Worktrees lists all worktrees of the repository using
// `git worktree list --porcelain`. The git binary is invoked with an
// argument list (no shell), so worktree paths are never interpreted by a shell.
func (r *Repository) Worktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = r.Path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	worktrees := parseWorktreePorcelain(string(out))

	// Mark the worktree containing the current directory as current.
	if cwd, err := filepath.Abs(r.Path); err == nil {
		for i := range worktrees {
			if wtPath, err := filepath.Abs(worktrees[i].Path); err == nil && wtPath == cwd {
				worktrees[i].IsCurrent = true
				break
			}
		}
	}

	return worktrees, nil
}

// parseWorktreePorcelain parses the output of `git worktree list --porcelain`.
// Each record is a block of lines separated by a blank line:
//
//	worktree /path/to/wt
//	HEAD <full-hash>
//	branch refs/heads/<name>   (or "detached" / "bare")
func parseWorktreePorcelain(out string) []Worktree {
	var worktrees []Worktree
	var cur Worktree
	hasCur := false

	flush := func() {
		if hasCur {
			worktrees = append(worktrees, cur)
		}
		cur = Worktree{}
		hasCur = false
	}

	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			flush()
			continue
		}

		key, value, _ := strings.Cut(line, " ")
		switch key {
		case "worktree":
			cur.Path = value
			cur.Branch = "(detached)"
			hasCur = true
		case "HEAD":
			if len(value) > 7 {
				cur.Head = value[:7]
			} else {
				cur.Head = value
			}
		case "branch":
			cur.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "detached":
			cur.Branch = "(detached)"
		case "bare":
			cur.Branch = "(bare)"
		}
	}
	flush()

	return worktrees
}
