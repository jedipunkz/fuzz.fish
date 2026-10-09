package git

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParsePullRequests(t *testing.T) {
	data := []byte(`[
		{"number":12,"title":"feat: x","url":"https://github.com/o/r/pull/12","author":{"login":"alice"},
		 "headRefName":"feat/x","headRepositoryOwner":{"login":"o"},"isCrossRepository":false},
		{"number":13,"title":"fix: y","url":"https://github.com/o/r/pull/13","author":{"login":"bob"},
		 "headRefName":"main","headRepositoryOwner":{"login":"bob"},"isCrossRepository":true}
	]`)
	worktrees := []Worktree{
		{Path: "/repo", Branch: "main"},
		{Path: "/wt/x", Branch: "feat/x"},
	}

	prs, err := parsePullRequests(data, worktrees)
	if err != nil {
		t.Fatal(err)
	}
	if len(prs) != 2 {
		t.Fatalf("len = %d, want 2", len(prs))
	}

	want0 := PullRequest{Number: 12, Title: "feat: x", URL: "https://github.com/o/r/pull/12",
		Repository: "o/r", Author: "alice", Branch: "feat/x", Worktree: "/wt/x"}
	if prs[0] != want0 {
		t.Errorf("prs[0] = %+v, want %+v", prs[0], want0)
	}

	// A fork PR from "main" must not be linked to the local main worktree.
	if prs[1].Fork != "bob" || prs[1].Worktree != "" {
		t.Errorf("fork PR: fork=%q worktree=%q, want bob and no worktree", prs[1].Fork, prs[1].Worktree)
	}
}

func TestPullRequestWorktreePath_WorktreeDir(t *testing.T) {
	r := NewRepository(".")
	pr := PullRequest{Number: 7, URL: "https://github.com/user/repo/pull/7", Branch: "feat/foo"}

	got, err := r.PullRequestWorktreePath(pr, "/home/u/gm/.worktrees")
	if err != nil {
		t.Fatal(err)
	}
	if want := "/home/u/gm/.worktrees/github.com/user/repo/feat/foo"; got != want {
		t.Errorf("path = %q, want %q", got, want)
	}

	pr.Branch = "../../../../etc"
	if _, err := r.PullRequestWorktreePath(pr, "/home/u/gm/.worktrees"); err == nil {
		t.Error("expected error for a branch escaping worktree_dir")
	}
}

func TestPullRequestWorktreePath_Default(t *testing.T) {
	dir := t.TempDir()
	repo := filepath.Join(dir, "myrepo")
	if out, err := exec.Command("git", "init", "-q", repo).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}

	got, err := NewRepository(repo).PullRequestWorktreePath(PullRequest{Number: 42}, "")
	if err != nil {
		t.Fatal(err)
	}
	// git reports the resolved path (macOS /var -> /private/var).
	if filepath.Base(got) != "myrepo-pr-42" || filepath.Base(filepath.Dir(got)) != filepath.Base(dir) {
		t.Errorf("path = %q, want <tmp>/myrepo-pr-42", got)
	}
}
