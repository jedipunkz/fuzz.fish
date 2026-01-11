package git

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
)

// GeneratePreview generates a preview of the branch for the TUI preview window
func GeneratePreview(branch Branch, width, height int) string {
	var sb strings.Builder

	// Branch info
	sb.WriteString(ui.LabelStyle.Render("Branch") + "\n")
	sb.WriteString(ui.ContentStyle.Render(branch.Name) + "\n\n")

	// Last commit
	sb.WriteString(ui.LabelStyle.Render("Last Commit") + "\n")
	sb.WriteString(ui.ContentStyle.Render(branch.LastCommit) + "\n")
	sb.WriteString(ui.ContentStyle.Render(branch.LastCommitMessage) + "\n")
	sb.WriteString(ui.ContentStyle.Faint(true).Render(branch.CommitDate) + "\n\n")

	// Recent commits
	sb.WriteString(ui.ContextHeaderStyle.Render("Recent Commits") + "\n")
	commits := getRecentCommits(branch.Name, 5)
	sb.WriteString(commits + "\n")

	// Changed files (if not current branch)
	if !branch.IsCurrent {
		sb.WriteString(ui.ContextHeaderStyle.Render("Changes from Current Branch") + "\n")
		diff := getBranchDiff(branch.Name)
		sb.WriteString(diff)
	}

	return sb.String()
}

// getRecentCommits returns the recent commits for a branch
func getRecentCommits(branchName string, count int) string {
	cmd := exec.Command("git", "log", branchName,
		"--color=always",
		fmt.Sprintf("--max-count=%d", count),
		"--pretty=format:%C(yellow)%h%C(reset) %C(dim)%cd%C(reset) %s",
		"--date=relative")
	output, err := cmd.Output()
	if err != nil {
		return ui.InactiveContextStyle.Render("  No commits found")
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var sb strings.Builder
	for _, line := range lines {
		sb.WriteString(ui.InactiveContextStyle.Render("  "+line) + "\n")
	}
	return sb.String()
}

// getBranchDiff returns the diff summary between current branch and target branch
func getBranchDiff(branchName string) string {
	currentBranch := getCurrentBranch()
	if currentBranch == "" {
		return ui.InactiveContextStyle.Render("  Could not determine current branch")
	}

	// Get files changed
	cmd := exec.Command("git", "diff", "--name-status", currentBranch+"..."+branchName)
	output, err := cmd.Output()
	if err != nil {
		return ui.InactiveContextStyle.Render("  No differences")
	}

	diffStr := strings.TrimSpace(string(output))
	if diffStr == "" {
		return ui.InactiveContextStyle.Render("  No differences")
	}

	lines := strings.Split(diffStr, "\n")
	var sb strings.Builder
	count := 0
	maxLines := 15
	for _, line := range lines {
		if count >= maxLines {
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  ... and %d more files", len(lines)-maxLines)) + "\n")
			break
		}
		sb.WriteString(ui.InactiveContextStyle.Render("  "+line) + "\n")
		count++
	}

	return sb.String()
}
