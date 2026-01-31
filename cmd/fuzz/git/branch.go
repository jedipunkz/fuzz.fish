package git

import (
	"sort"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// CollectBranches collects all git branches (local and remote)
// Lightweight version: does not fetch commit objects for performance
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

	// Collect local branches first, then remote branches
	var localBranches []Branch
	var remoteBranches []Branch

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

		// Get short hash only (no commit object fetch)
		shortHash := ref.Hash().String()[:7]

		branch := Branch{
			Name:              name,
			IsCurrent:         name == currentBranch,
			IsRemote:          isRemote,
			LastCommit:        shortHash,
			LastCommitMessage: "",
			CommitDate:        "",
		}

		if isRemote {
			remoteBranches = append(remoteBranches, branch)
		} else {
			localBranches = append(localBranches, branch)
		}

		return nil
	})

	if err != nil {
		return branches
	}

	// Sort alphabetically
	sort.Slice(localBranches, func(i, j int) bool {
		return localBranches[i].Name < localBranches[j].Name
	})
	sort.Slice(remoteBranches, func(i, j int) bool {
		return remoteBranches[i].Name < remoteBranches[j].Name
	})

	// Local branches first, then remote
	branches = append(branches, localBranches...)
	branches = append(branches, remoteBranches...)

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

