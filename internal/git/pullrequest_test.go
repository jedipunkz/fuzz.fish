package git

import "testing"

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
	if prs[1].Branch != "bob:main" || prs[1].Worktree != "" {
		t.Errorf("fork PR: branch=%q worktree=%q, want bob:main and no worktree", prs[1].Branch, prs[1].Worktree)
	}
}
