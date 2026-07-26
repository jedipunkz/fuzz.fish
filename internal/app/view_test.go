package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
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
