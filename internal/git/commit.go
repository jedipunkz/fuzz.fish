package git

import (
	"os/exec"
	"strconv"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// commitLimit caps how many commits are read from the log. Large repositories
// would otherwise stall the TUI while the whole history is parsed.
const commitLimit = 2000

// Commit represents a single commit reachable from HEAD
type Commit struct {
	Hash    string // short hash
	Subject string
	When    int64 // author timestamp (unix seconds)
}

// Commits lists commits reachable from HEAD, newest first. The git binary is
// invoked with an argument list (no shell), so no value is interpreted by a
// shell.
func (r *Repository) Commits() ([]Commit, error) {
	cmd := exec.Command("git", "log", "--no-color",
		"--pretty=format:%h%x00%s%x00%ct", "-n", strconv.Itoa(commitLimit))
	cmd.Dir = r.Path
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return parseCommitLog(string(out)), nil
}

// parseCommitLog parses NUL-separated `git log` records, one per line.
func parseCommitLog(out string) []Commit {
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		hash, rest, ok := strings.Cut(line, "\x00")
		if !ok {
			continue
		}
		subject, ts, _ := strings.Cut(rest, "\x00")
		when, _ := strconv.ParseInt(ts, 10, 64)
		commits = append(commits, Commit{Hash: hash, Subject: subject, When: when})
	}
	return commits
}

// RecentCommits walks the history reachable from a commit hash and returns
// up to limit commits, newest first, in `git log --oneline` style. Pure
// in-process: the go-git walker reads the object database without spawning
// git. Shallow clones, unborn branches and unreadable objects surface as an
// empty list, and the caller falls back to a metadata-only preview.
func (r *Repository) RecentCommits(hash string, limit int) []Commit {
	if hash == "" || limit <= 0 {
		return nil
	}

	repo, err := gogit.PlainOpenWithOptions(r.Path, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return nil
	}

	walker, err := repo.Log(&gogit.LogOptions{
		From:  plumbing.NewHash(hash),
		Order: gogit.LogOrderCommitterTime,
	})
	if err != nil {
		return nil
	}
	defer walker.Close() //nolint:errcheck

	commits := make([]Commit, 0, limit)
	for i := 0; i < limit; i++ {
		c, err := walker.Next()
		if err != nil {
			break
		}
		commits = append(commits, Commit{
			Hash:    shortHash(c.Hash.String()),
			Subject: subjectOf(c.Message),
			When:    c.Committer.When.Unix(),
		})
	}
	return commits
}

// shortHash trims a commit hash hash to the 7-char oneline form.
func shortHash(hash string) string {
	if len(hash) > 7 {
		return hash[:7]
	}
	return hash
}

// subjectOf returns the first line of a commit message.
func subjectOf(message string) string {
	subject, _, _ := strings.Cut(message, "\n")
	return subject
}
