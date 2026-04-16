package git

import (
	"os"
	"testing"
)

func TestNewRepository(t *testing.T) {
	r := NewRepository("/some/path")
	if r == nil {
		t.Fatal("NewRepository() returned nil")
	}
	if r.Path != "/some/path" {
		t.Errorf("Path = %q, want %q", r.Path, "/some/path")
	}
}

func TestIsRepo(t *testing.T) {
	// This test depends on the environment
	// In a git repository, it should return true
	// We can't make assumptions about the test environment,
	// so we just check that the function doesn't panic
	r := NewRepository(".")
	result := r.IsRepo()
	_ = result // Just ensure it runs without error
}

func TestIsRepo_NotGitDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "not-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	r := NewRepository(tmpDir)
	if r.IsRepo() {
		t.Errorf("IsRepo() = true for non-git directory, want false")
	}
}

func TestIsRepo_NonExistentDir(t *testing.T) {
	r := NewRepository("/nonexistent/path/that/does/not/exist")
	if r.IsRepo() {
		t.Errorf("IsRepo() = true for non-existent directory, want false")
	}
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

func TestBranches_NotGitDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "not-git-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	r := NewRepository(tmpDir)
	branches, err := r.Branches()
	if err == nil {
		t.Error("Branches() on non-git directory should return an error")
	}
	if len(branches) != 0 {
		t.Errorf("Branches() on non-git directory returned %d branches, want 0", len(branches))
	}
}

func TestBranches_InGitRepo(t *testing.T) {
	// Run in the actual git repo (working directory)
	r := NewRepository(".")
	if !r.IsRepo() {
		t.Skip("not running in a git repository")
	}

	branches, err := r.Branches()
	if err != nil {
		t.Fatalf("Branches() returned unexpected error: %v", err)
	}
	if len(branches) == 0 {
		t.Error("Branches() returned no branches in a valid git repo")
	}

	// Verify branch structure
	for _, b := range branches {
		if b.Name == "" {
			t.Error("found branch with empty name")
		}
		if b.LastCommit == "" {
			t.Error("found branch with empty commit hash")
		}
		if len(b.LastCommit) > 7 {
			t.Errorf("LastCommit %q is longer than 7 chars", b.LastCommit)
		}
	}

	// At most one branch should be marked as current
	// (0 in detached HEAD / CI environments, 1 in normal branch checkout)
	currentCount := 0
	for _, b := range branches {
		if b.IsCurrent {
			currentCount++
		}
	}
	if currentCount > 1 {
		t.Errorf("expected at most 1 current branch, got %d", currentCount)
	}
}
