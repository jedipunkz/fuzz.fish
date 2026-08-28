package git

import "testing"

func TestParseCommitLog(t *testing.T) {
	out := "a1b2c3d\x00fix: handle empty subject\x001700000000\n" +
		"e4f5a6b\x00feat: add commit search\x001699999999\n" +
		"\n" +
		"broken-line-without-separator\n"

	commits := parseCommitLog(out)
	if len(commits) != 2 {
		t.Fatalf("got %d commits, want 2", len(commits))
	}
	if commits[0].Hash != "a1b2c3d" || commits[0].Subject != "fix: handle empty subject" || commits[0].When != 1700000000 {
		t.Errorf("first commit parsed as %+v", commits[0])
	}
	if commits[1].Hash != "e4f5a6b" || commits[1].Subject != "feat: add commit search" {
		t.Errorf("second commit parsed as %+v", commits[1])
	}
}
