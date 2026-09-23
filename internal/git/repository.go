package git

import (
	"os/exec"
	"sort"
	"strconv"
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
	_, err := gogit.PlainOpenWithOptions(r.Path, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	return err == nil
}

// Branches collects all git branches (local and remote)
// Lightweight version: does not fetch commit objects for performance
func (r *Repository) Branches() ([]Branch, error) {
	var branches []Branch

	repo, err := gogit.PlainOpenWithOptions(r.Path, &gogit.PlainOpenOptions{DetectDotGit: true, EnableDotGitCommonDir: true})
	if err != nil {
		return branches, err
	}

	// Get current branch (reusing the repo object)
	currentBranch := r.currentBranch(repo)

	// Read all branch commit timestamps in one git invocation
	timestamps := r.branchTimestamps()

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
		hashStr := ref.Hash().String()
		shortHash := hashStr
		if len(hashStr) > 7 {
			shortHash = hashStr[:7]
		}

		branch := Branch{
			Name:              name,
			IsCurrent:         name == currentBranch,
			IsRemote:          isRemote,
			LastCommit:        shortHash,
			LastCommitMessage: "",
			CommitDate:        "",
			CommitTimestamp:   timestamps[refName],
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

// branchTimestamps reads the committer timestamp of every branch ref in one
// `git for-each-ref` invocation. The git binary is invoked with an argument
// list (no shell), so ref names are never interpreted by a shell. A missing
// git binary or a failed command yields a nil map, leaving timestamps at 0.
func (r *Repository) branchTimestamps() map[string]int64 {
	cmd := exec.Command("git", "for-each-ref",
		"--format=%(refname)%00%(committerdate:unix)", "refs/heads", "refs/remotes")
	cmd.Dir = r.Path
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	return parseRefTimestamps(string(out))
}

// parseRefTimestamps parses NUL-separated `git for-each-ref` records, one per
// line, into a map from full ref name to unix timestamp.
func parseRefTimestamps(out string) map[string]int64 {
	timestamps := make(map[string]int64)
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		refName, ts, ok := strings.Cut(line, "\x00")
		if !ok {
			continue
		}
		when, _ := strconv.ParseInt(ts, 10, 64)
		timestamps[refName] = when
	}
	return timestamps
}
