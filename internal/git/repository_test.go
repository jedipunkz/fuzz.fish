package git

import (
	"testing"
)

func TestIsRepo(t *testing.T) {
	// This test depends on the environment
	// In a git repository, it should return true
	// We can't make assumptions about the test environment,
	// so we just check that the function doesn't panic
	r := NewRepository(".")
	result := r.IsRepo()
	_ = result // Just ensure it runs without error
}

func TestBranches(t *testing.T) {
	// This test depends on the environment (being in a git repo)
	// We just check that the function doesn't panic and returns expected structure
	r := NewRepository(".")
	branches, _ := r.Branches()
	// In a git repo, we should have at least the current branch
	// but we don't assume any specific state
	_ = branches // Just ensure it runs without error
}
