package git

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ui.InactiveContextStyle.Render("  No commits found")
	}

	// Resolve branch reference
	var hash plumbing.Hash
	ref, err := repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		// Try as remote branch
		ref, err = repo.Reference(plumbing.NewRemoteReferenceName("origin", branchName), true)
		if err != nil {
			// Try as full reference name (e.g., "origin/main")
			if strings.Contains(branchName, "/") {
				parts := strings.SplitN(branchName, "/", 2)
				if len(parts) == 2 {
					ref, err = repo.Reference(plumbing.NewRemoteReferenceName(parts[0], parts[1]), true)
				}
			}
			if err != nil {
				return ui.InactiveContextStyle.Render("  No commits found")
			}
		}
	}
	hash = ref.Hash()

	// Get commit log
	commits, err := repo.Log(&git.LogOptions{
		From: hash,
	})
	if err != nil {
		return ui.InactiveContextStyle.Render("  No commits found")
	}

	var sb strings.Builder
	commitCount := 0
	_ = commits.ForEach(func(c *object.Commit) error {
		if commitCount >= count {
			return fmt.Errorf("reached limit")
		}

		// Format: hash date message
		shortHash := c.Hash.String()[:7]
		relativeTime := getRelativeTime(c.Committer.When)
		message := strings.Split(c.Message, "\n")[0]

		// Mimic git log --color output with yellow hash and dim date
		line := fmt.Sprintf("\033[33m%s\033[0m \033[2m%s\033[0m %s",
			shortHash, relativeTime, message)

		sb.WriteString(ui.InactiveContextStyle.Render("  "+line) + "\n")
		commitCount++
		return nil
	})

	if sb.Len() == 0 {
		return ui.InactiveContextStyle.Render("  No commits found")
	}

	return sb.String()
}

// getRelativeTime returns a human-readable relative time string
func getRelativeTime(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	} else if duration < 30*24*time.Hour {
		weeks := int(duration.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	} else if duration < 365*24*time.Hour {
		months := int(duration.Hours() / 24 / 30)
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	} else {
		years := int(duration.Hours() / 24 / 365)
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}

// getBranchDiff returns the diff summary between current branch and target branch
func getBranchDiff(branchName string) string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not open repository")
	}

	currentBranch := getCurrentBranch()
	if currentBranch == "" {
		return ui.InactiveContextStyle.Render("  Could not determine current branch")
	}

	// Get current branch reference
	currentRef, err := repo.Reference(plumbing.NewBranchReferenceName(currentBranch), true)
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not get current branch reference")
	}

	// Get target branch reference
	var targetRef *plumbing.Reference
	targetRef, err = repo.Reference(plumbing.NewBranchReferenceName(branchName), true)
	if err != nil {
		// Try as remote branch
		if strings.Contains(branchName, "/") {
			parts := strings.SplitN(branchName, "/", 2)
			if len(parts) == 2 {
				targetRef, err = repo.Reference(plumbing.NewRemoteReferenceName(parts[0], parts[1]), true)
			}
		}
		if err != nil {
			return ui.InactiveContextStyle.Render("  Could not get target branch reference")
		}
	}

	// Get commits
	currentCommit, err := repo.CommitObject(currentRef.Hash())
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not get current commit")
	}

	targetCommit, err := repo.CommitObject(targetRef.Hash())
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not get target commit")
	}

	// Get trees
	currentTree, err := currentCommit.Tree()
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not get current tree")
	}

	targetTree, err := targetCommit.Tree()
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not get target tree")
	}

	// Get diff
	changes, err := currentTree.Diff(targetTree)
	if err != nil {
		return ui.InactiveContextStyle.Render("  Could not compute diff")
	}

	if len(changes) == 0 {
		return ui.InactiveContextStyle.Render("  No differences")
	}

	var sb strings.Builder
	maxLines := 15
	for i, change := range changes {
		if i >= maxLines {
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  ... and %d more files", len(changes)-maxLines)) + "\n")
			break
		}

		// Determine status (A=Added, M=Modified, D=Deleted)
		var status string
		from, to := change.From, change.To
		if from.Name == "" {
			status = "A"
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s\t%s", status, to.Name)) + "\n")
		} else if to.Name == "" {
			status = "D"
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s\t%s", status, from.Name)) + "\n")
		} else {
			status = "M"
			sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s\t%s", status, to.Name)) + "\n")
		}
	}

	return sb.String()
}
