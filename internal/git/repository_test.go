package git

import (
	"os"
	"os/exec"
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

func TestParseRefTimestamps(t *testing.T) {
	out := "refs/heads/main\x001700000000\n" +
		"refs/heads/feat/branch with space\x001699999999\n" +
		"refs/remotes/origin/main\x001700000001\n" +
		"\n" +
		"refs/heads/no-separator\n"

	got := parseRefTimestamps(out)

	want := map[string]int64{
		"refs/heads/main":                   1700000000,
		"refs/heads/feat/branch with space": 1699999999,
		"refs/remotes/origin/main":          1700000001,
	}

	if len(got) != len(want) {
		t.Fatalf("got %d refs, want %d", len(got), len(want))
	}
	for ref, ts := range want {
		if got[ref] != ts {
			t.Errorf("timestamp for %q = %d, want %d", ref, got[ref], ts)
		}
	}
}

// initTestRepo creates a git repository in a temp directory with a single
// commit on branch "main" and returns its path.
func initTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
			"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	run("init", "--initial-branch=main")
	run("commit", "--allow-empty", "-m", "initial commit")

	return dir
}

func TestBranches_CommitTimestamp(t *testing.T) {
	dir := initTestRepo(t)

	branches, err := NewRepository(dir).Branches()
	if err != nil {
		t.Fatalf("Branches() returned unexpected error: %v", err)
	}
	if len(branches) != 1 {
		t.Fatalf("got %d branches, want 1", len(branches))
	}
	if branches[0].Name != "main" {
		t.Errorf("branch name = %q, want %q", branches[0].Name, "main")
	}
	if branches[0].CommitTimestamp <= 0 {
		t.Errorf("CommitTimestamp = %d, want a positive unix timestamp", branches[0].CommitTimestamp)
	}
}
