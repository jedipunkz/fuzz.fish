package git

import (
	"sort"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// IsGitRepo checks if the current directory is a git repository
func IsGitRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// CollectBranches collects all git branches (local and remote)
// Optimized version: opens repo once and uses parallel commit fetching
func CollectBranches() []Branch {
	var branches []Branch

	repo, err := git.PlainOpen(".")
	if err != nil {
		return branches
	}

	// Get current branch (reusing the repo object)
	currentBranch := getCurrentBranchFromRepo(repo)

	// Get all references
	refs, err := repo.References()
	if err != nil {
		return branches
	}

	// First pass: collect all branch references without fetching commits
	type refInfo struct {
		name     string
		isRemote bool
		hash     plumbing.Hash
	}
	var refInfos []refInfo

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

		refInfos = append(refInfos, refInfo{
			name:     name,
			isRemote: isRemote,
			hash:     ref.Hash(),
		})

		return nil
	})

	if err != nil || len(refInfos) == 0 {
		return branches
	}

	// Second pass: fetch commit info in parallel using worker pool
	type branchResult struct {
		index  int
		branch Branch
		valid  bool
	}

	results := make([]branchResult, len(refInfos))
	var wg sync.WaitGroup

	// Use a worker pool to limit concurrent goroutines
	// This prevents too many concurrent file descriptor operations
	workerCount := 8
	if len(refInfos) < workerCount {
		workerCount = len(refInfos)
	}

	jobs := make(chan int, len(refInfos))

	// Start workers
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				info := refInfos[i]

				// Get commit
				commit, err := repo.CommitObject(info.hash)
				if err != nil {
					results[i] = branchResult{index: i, valid: false}
					continue
				}

				// Get short hash (7 characters like git)
				shortHash := info.hash.String()[:7]

				// Format commit message (first line only)
				message := strings.Split(commit.Message, "\n")[0]

				// Format date directly (no intermediate parsing)
				commitDate := commit.Committer.When.Format("2006-01-02 15:04")

				results[i] = branchResult{
					index: i,
					valid: true,
					branch: Branch{
						Name:              info.name,
						IsCurrent:         info.name == currentBranch,
						IsRemote:          info.isRemote,
						LastCommit:        shortHash,
						LastCommitMessage: message,
						CommitDate:        commitDate,
						CommitTimestamp:   commit.Committer.When.Unix(),
					},
				}
			}
		}()
	}

	// Send jobs
	for i := range refInfos {
		jobs <- i
	}
	close(jobs)

	// Wait for all workers to complete
	wg.Wait()

	// Collect valid results
	for _, result := range results {
		if result.valid {
			branches = append(branches, result.branch)
		}
	}

	// Sort by commit date (newest first)
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].CommitTimestamp > branches[j].CommitTimestamp
	})

	return branches
}

// getCurrentBranchFromRepo returns the current git branch name using existing repo
func getCurrentBranchFromRepo(repo *git.Repository) string {
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

// getCurrentBranch returns the current git branch name (opens repo internally)
func getCurrentBranch() string {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return ""
	}
	return getCurrentBranchFromRepo(repo)
}

// sortBranchesByDate sorts branches by commit timestamp (newest first)
func sortBranchesByDate(branches []Branch) {
	sort.Slice(branches, func(i, j int) bool {
		return branches[i].CommitTimestamp > branches[j].CommitTimestamp
	})
}
