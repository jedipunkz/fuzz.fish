package app

import (
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/jedipunkz/fuzz.fish/internal/files"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
)

func TestRenderItem_TruncatesOnRuneBoundary(t *testing.T) {
	item := Item{Text: "echo 日本語のとても長いコマンドライン引数"}

	// Widths around a multibyte boundary: byte slicing produced a replacement
	// character for the widths that fell mid-rune.
	for _, width := range []int{20, 21, 22, 23} {
		m := model{mode: ModeHistory, listWidth: width, cursor: -1}
		var sb strings.Builder
		m.renderItem(&sb, 0, item)

		rendered := sb.String()
		if strings.ContainsRune(rendered, '�') {
			t.Errorf("renderItem(listWidth=%d) split a multibyte character: %q", width, rendered)
		}
		if got := lipgloss.Width(rendered); got > width {
			t.Errorf("renderItem(listWidth=%d) rendered %d cells, want <= %d", width, got, width)
		}
	}
}

// highlightedRunes returns the characters rendered with the match colour, so a
// test can assert which part of the line was highlighted.
func highlightedRunes(rendered string) string {
	const matchColor = "247;118;142"

	var out strings.Builder
	active := false
	for len(rendered) > 0 {
		if strings.HasPrefix(rendered, "\x1b[") {
			end := strings.IndexByte(rendered, 'm')
			if end < 0 {
				break
			}
			seq := rendered[:end+1]
			active = strings.Contains(seq, matchColor)
			rendered = rendered[end+1:]
			continue
		}
		r, size := utf8.DecodeRuneInString(rendered)
		if active {
			out.WriteRune(r)
		}
		rendered = rendered[size:]
	}
	return out.String()
}

func TestRenderItem_HighlightsMatchedText(t *testing.T) {
	tests := []struct {
		name string
		mode SearchMode
		item Item
		want string
	}{
		{
			name: "files mode skips the icon",
			mode: ModeFiles,
			item: Item{
				Text:           "internal/app.go",
				Original:       files.Entry{Path: "internal/app.go"},
				MatchedIndexes: []int{0, 1, 2},
			},
			want: "int",
		},
		{
			name: "git mode skips the marker",
			mode: ModeGitBranch,
			item: Item{
				Text:           "main",
				Original:       git.Branch{Name: "main"},
				MatchedIndexes: []int{0, 1},
			},
			want: "ma",
		},
		{
			name: "multibyte text uses byte offsets",
			mode: ModeHistory,
			item: Item{
				Text:           "echo 日本語",
				Original:       history.Entry{Cmd: "echo 日本語"},
				MatchedIndexes: []int{5, 8}, // 日 and 本 start at bytes 5 and 8
			},
			want: "日本",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{mode: tt.mode, listWidth: 40, cursor: -1}
			var sb strings.Builder
			m.renderItem(&sb, 0, tt.item)

			if got := highlightedRunes(sb.String()); got != tt.want {
				t.Errorf("highlighted %q, want %q", got, tt.want)
			}
		})
	}
}
