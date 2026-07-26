package history

import (
	"strings"
	"testing"
)

func TestGeneratePreview_NarrowPane(t *testing.T) {
	all := []Entry{
		{Cmd: "git status --short", When: 1000},
		{Cmd: "go test ./...", When: 900},
	}

	// A pane only a few cells wide leaves almost no room for the context
	// lines; truncation must stay in range instead of panicking.
	for _, width := range []int{0, 1, 2, 3, 4} {
		if got := all[0].GeneratePreview(all, 0, width, 10); got == "" {
			t.Errorf("GeneratePreview(width=%d) returned empty output", width)
		}
	}
}

func TestGeneratePreview_MultibyteContext(t *testing.T) {
	all := []Entry{
		{Cmd: "echo 日本語のとても長いコマンドライン", When: 1000},
		{Cmd: "echo 日本語のとても長いコマンドライン引数付き", When: 900},
	}

	got := all[0].GeneratePreview(all, 0, 20, 10)
	if strings.ContainsRune(got, '�') {
		t.Errorf("GeneratePreview() split a multibyte character: %q", got)
	}
}
