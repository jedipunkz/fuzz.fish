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

