package git

import (
	"os/exec"
	"strings"
	"time"
)

// CollectBranches collects all git branches (local and remote)
func CollectBranches() []Branch {
	var branches []Branch

	// Get current branch
	currentBranch := getCurrentBranch()

	// Get all branches with their last commit info
	cmd := exec.Command("git", "for-each-ref", "--sort=-committerdate",
		"--format=%(refname:short)|%(objectname:short)|%(subject)|%(committerdate:iso8601)",
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

		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		name := parts[0]
		commit := parts[1]
		message := parts[2]
		dateStr := parts[3]

		// Parse date
		commitDate := formatDate(dateStr)

		// Determine if remote
		isRemote := strings.HasPrefix(name, "origin/")

		// Skip HEAD references
		if strings.Contains(name, "HEAD") {
			continue
		}

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
