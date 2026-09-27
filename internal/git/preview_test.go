package git

import (
	"strings"
	"testing"
)

func TestRecentCommits(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "commit", "--allow-empty", "-m", "second")
	runGit(t, dir, "commit", "--allow-empty", "-m", "third")

	branches, err := NewRepository(dir).Branches()
	if err != nil {
		t.Fatalf("Branches() returned unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("got %d branches, want 1", len(branches))
	}

	recent := NewRepository(dir).RecentCommits(branches[0].Hash, 5)
	if len(recent) != 3 {
		t.Fatalf("got %d commits, want 3", len(recent))
	}
	if recent[0].Subject != "third" {
		t.Errorf("newest commit subject = %q, want %q (newest first)", recent[0].Subject, "third")
	}
	if recent[2].Subject != "initial commit" {
		t.Errorf("oldest commit subject = %q, want %q", recent[2].Subject, "initial commit")
	}
	if got, want := recent[0].Hash, branches[0].LastCommit; got != want {
		t.Errorf("recent hash %q, want the short branch commit hash %q", got, want)
	}
}

func TestRecentCommitsLimit(t *testing.T) {
	dir := initTestRepo(t)
	for i := 0; i < 4; i++ {
		runGit(t, dir, "commit", "--allow-empty", "-m", "commit")
	}
	if recent := NewRepository(dir).RecentCommits(branchHash(t, dir), 2); len(recent) != 2 {
		t.Fatalf("got %d commits, want 2", len(recent))
	}
}

func TestRecentCommitsEmptyHash(t *testing.T) {
	if recent := NewRepository(".").RecentCommits("", 5); recent != nil {
		t.Errorf("RecentCommits(\"\") = %v, want nil", recent)
	}
}

func TestBranchPreviewRecentCommits(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "commit", "--allow-empty", "-m", "second")

	branches, err := NewRepository(dir).Branches()
	if err != nil {
		t.Fatalf("Branches() returned unexpected error: %v", err)
	}

	preview := branches[0].GeneratePreview(dir, 80, 30)
	if !strings.Contains(preview, "Recent commits") {
		t.Fatalf("preview missing the Recent commits section:\n%s", preview)
	}
	if !strings.Contains(preview, "second") {
		t.Errorf("preview missing the newest commit subject:\n%s", preview)
	}
}

// TestBranchPreviewWithoutLog verifies the fallback for branches whose commit
// cannot be read: the metadata sections still render, without a Recent
// commits section.
func TestBranchPreviewWithoutLog(t *testing.T) {
	b := Branch{Name: "main", LastCommit: "a1b2c3d"}

	preview := b.GeneratePreview("/nonexistent", 80, 30)
	if strings.Contains(preview, "Recent commits") {
		t.Errorf("unwalkable branch shows a Recent commits section:\n%s", preview)
	}
	if !strings.Contains(preview, "a1b2c3d") {
		t.Errorf("preview lost the commit hash:\n%s", preview)
	}
}

// branchHash returns the full commit hash of the repo's first branch.
func branchHash(t *testing.T, dir string) string {
	t.Helper()
	branches, err := NewRepository(dir).Branches()
	if err != nil || len(branches) == 0 {
		t.Fatalf("Branches() = %d branches, %v — want one branch", len(branches), err)
	}
	return branches[0].Hash
}
