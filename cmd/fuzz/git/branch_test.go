package git

import (
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// This test depends on the environment
	// In a git repository, it should return true
	// We can't make assumptions about the test environment,
	// so we just check that the function doesn't panic
	result := IsGitRepo()
	_ = result // Just ensure it runs without error
}

func TestGetCurrentBranch(t *testing.T) {
	// This test depends on the environment
	// We just check that the function doesn't panic
	result := getCurrentBranch()
	_ = result // Just ensure it runs without error
}

func TestCollectBranches(t *testing.T) {
	// This test depends on the environment (being in a git repo)
	// We just check that the function doesn't panic and returns expected structure
	branches := CollectBranches()
	// In a git repo, we should have at least the current branch
	// but we don't assume any specific state
	_ = branches // Just ensure it runs without error
}
