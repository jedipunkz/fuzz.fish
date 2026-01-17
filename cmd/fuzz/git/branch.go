package git

import (
	"os/exec"
	"strings"
	"time"
)

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// CollectBranches collects all git branches (local and remote)
func CollectBranches() []Branch {
	var branches []Branch

	// Get current branch
	currentBranch := getCurrentBranch()

	// Get all branches with their last commit info
	// Include full refname to distinguish between local and remote branches
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate",
		"--format=%(refname)|%(refname:short)|%(objectname:short)|%(subject)|%(committerdate:iso8601)",
		"refs/heads/", "refs/remotes/")
	output, err := cmd.Output()
	if err != nil {
		return branches
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 5)
		if len(parts) != 5 {
			continue
		}

		refname := parts[0]      // Full ref name (e.g., "refs/heads/main" or "refs/remotes/origin/main")
		name := parts[1]         // Short name (e.g., "main" or "origin/main")
		commit := parts[2]
		message := parts[3]
		dateStr := parts[4]

		// Parse date
		commitDate := formatDate(dateStr)

		// Skip HEAD references (check full refname, not short name)
		// e.g., "refs/remotes/origin/HEAD" has short name "origin"
		if strings.Contains(refname, "HEAD") {
			continue
		}

		// Determine if remote based on full refname
		isRemote := strings.HasPrefix(refname, "refs/remotes/")

		branches = append(branches, Branch{
			Name:              name,
			IsCurrent:         name == currentBranch,
			IsRemote:          isRemote,
			LastCommit:        commit,
			LastCommitMessage: message,
			CommitDate:        commitDate,
		})
	}

	return branches
}

// getCurrentBranch returns the current git branch name
func getCurrentBranch() string {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// formatDate formats ISO8601 date to a readable format
func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02 15:04")
}
