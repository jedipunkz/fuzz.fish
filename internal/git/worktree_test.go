package git

import "testing"

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
		{Path: "/repo/main", Branch: "main", Head: "7a2ca8c"},
		{Path: "/repo/wt-detached", Branch: "(detached)", Head: "ce5e84b"},
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
