package git

import (
	"strings"
	"testing"
)

func TestParseWorktreePorcelain(t *testing.T) {
	out := "worktree /repo/main\n" +
		"HEAD 7a2ca8c612543f4be2e3e0f56a895c123d77c8cd\n" +
		"branch refs/heads/main\n" +
		"\n" +
		"worktree /repo/wt-detached\n" +
		"HEAD ce5e84bf40af34b3204673ae5c77241e1378bef9\n" +
		"detached\n" +
		"\n" +
		"worktree /repo/bare\n" +
		"bare\n" +
		"\n"

	got := parseWorktreePorcelain(out)

	want := []Worktree{
		{Path: "/repo/main", Branch: "main", Hash: "7a2ca8c612543f4be2e3e0f56a895c123d77c8cd", Head: "7a2ca8c"},
		{Path: "/repo/wt-detached", Branch: "(detached)", Hash: "ce5e84bf40af34b3204673ae5c77241e1378bef9", Head: "ce5e84b"},
		{Path: "/repo/bare", Branch: "(bare)", Head: ""},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d worktrees, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("worktree %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

// TestWorktreePreviewRecentCommits verifies a worktree preview lists the
// worktree head's last commits.
func TestWorktreePreviewRecentCommits(t *testing.T) {
	dir := initTestRepo(t)
	runGit(t, dir, "worktree", "add", dir+"/wt1")

	worktrees, err := NewRepository(dir).Worktrees()
	if err != nil {
		t.Fatalf("Worktrees() returned unexpected error: %v", err)
	}
	var wt *Worktree
	for i := range worktrees {
		if strings.HasSuffix(worktrees[i].Path, "wt1") {
			wt = &worktrees[i]
			break
		}
	}
	if wt == nil {
		t.Fatalf("worktree wt1 missing from %v", worktrees)
	}

	preview := wt.GeneratePreview(dir, 80, 30)
	if !strings.Contains(preview, "Recent commits") {
		t.Fatalf("preview missing the Recent commits section:\n%s", preview)
	}
	if !strings.Contains(preview, "initial commit") {
		t.Errorf("preview missing the commit subject:\n%s", preview)
	}
}
