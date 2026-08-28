package git

import (
	"os/exec"
	"strconv"
	"strings"
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
