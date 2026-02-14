package git

import (
	"sort"
	"strings"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// Branch represents a git branch
type Branch struct {
	Name              string
	IsCurrent         bool
	IsRemote          bool
	LastCommit        string
	LastCommitMessage string
	CommitDate        string
	CommitTimestamp   int64 // Unix timestamp for recency scoring
}

// Repository provides git operations for a working directory
type Repository struct {
	Path string
}

// NewRepository creates a Repository for the given path
func NewRepository(path string) *Repository {
	return &Repository{Path: path}
}

// IsRepo checks if the path is a git repository
func (r *Repository) IsRepo() bool {
	_, err := gogit.PlainOpen(r.Path)
	return err == nil
}

// Branches collects all git branches (local and remote)
// Lightweight version: does not fetch commit objects for performance
func (r *Repository) Branches() ([]Branch, error) {
	var branches []Branch

	repo, err := gogit.PlainOpen(r.Path)
	if err != nil {
		return branches, err
	}

	// Get current branch (reusing the repo object)
	currentBranch := r.currentBranch(repo)

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return branches, err
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
		return branches, err
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

	return branches, nil
}

// currentBranch returns the current git branch name using existing repo
func (r *Repository) currentBranch(repo *gogit.Repository) string {
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
