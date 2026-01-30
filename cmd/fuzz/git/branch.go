package git

import (
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// CollectBranches collects all git branches (local and remote)
func CollectBranches() []Branch {
	var branches []Branch

	repo, err := git.PlainOpen(".")
	if err != nil {
		return branches
	}

	// Get current branch
	currentBranch := getCurrentBranch()

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return branches
	}

	// Collect branches with commit info
	type branchInfo struct {
		branch     Branch
		commitTime time.Time
	}
	var branchInfos []branchInfo

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		refName := ref.Name().String()

		// Skip HEAD and other non-branch references
		if strings.Contains(refName, "HEAD") {
			return nil
		}

		// Only process branches (local and remote)
		if !strings.HasPrefix(refName, "refs/heads/") && !strings.HasPrefix(refName, "refs/remotes/") {
			return nil
		}

		// Determine if remote
		isRemote := strings.HasPrefix(refName, "refs/remotes/")

		// Get short name
		var name string
		if isRemote {
			name = strings.TrimPrefix(refName, "refs/remotes/")
		} else {
			name = strings.TrimPrefix(refName, "refs/heads/")
		}

		// Get commit
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return nil
		}

		// Get short hash (7 characters like git)
		shortHash := ref.Hash().String()[:7]

		// Format commit message (first line only)
		message := strings.Split(commit.Message, "\n")[0]

		// Format date
		commitDate := formatDate(commit.Committer.When.Format("2006-01-02 15:04:05 -0700"))

		branchInfos = append(branchInfos, branchInfo{
			branch: Branch{
				Name:              name,
				IsCurrent:         name == currentBranch,
				IsRemote:          isRemote,
				LastCommit:        shortHash,
				LastCommitMessage: message,
				CommitDate:        commitDate,
				CommitTimestamp:   commit.Committer.When.Unix(),
			},
			commitTime: commit.Committer.When,
		})

		return nil
	})

	if err != nil {
		return branches
	}

	// Sort by commit date (newest first)
	sort.Slice(branchInfos, func(i, j int) bool {
		return branchInfos[i].commitTime.After(branchInfos[j].commitTime)
	})

	// Extract branches from branchInfos
	for _, info := range branchInfos {
		branches = append(branches, info.branch)
	}

	return branches
}

// getCurrentBranch returns the current git branch name
func getCurrentBranch() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}

	head, err := repo.Head()
	if err != nil {
		return ""
	}

	// Get branch name from reference
	if head.Name().IsBranch() {
		return head.Name().Short()
	}

	return ""
}

// formatDate formats ISO8601 date to a readable format
func formatDate(dateStr string) string {
	t, err := time.Parse("2006-01-02 15:04:05 -0700", dateStr)
	if err != nil {
		return dateStr
	}
	return t.Format("2006-01-02 15:04")
}
