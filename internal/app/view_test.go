package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
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

func TestUpdatePreview_FollowsSelectedItem(t *testing.T) {
	m := model{
		mode:         ModeHistory,
		viewport:     viewport.New(),
		previewCache: map[string]string{},
		mainHeight:   10,
		historyEntries: []history.Entry{
			{Cmd: "alpha unique", When: 1000},
			{Cmd: "beta unique", When: 900},
		},
	}
	m.viewport.SetWidth(40)
	m.viewport.SetHeight(10)
	m.loadItemsForMode()

	// Both queries leave a single result, so the cursor sits at index 0 twice;
	// the preview must still follow the item, not the cursor position.
	m.updateFilter("alpha")
	first := m.viewport.View()
	m.updateFilter("beta")
	second := m.viewport.View()

	if first == second {
		t.Fatalf("preview did not change when the selected item changed:\n%s", first)
	}
	if selected := m.filtered[m.cursor].Text; !strings.Contains(markerLine(second), selected) {
		t.Errorf("preview marks %q, want the selected item %q", markerLine(second), selected)
	}
}

// markerLine returns the context line the preview marks as selected.
func markerLine(view string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, "→") {
			return line
		}
	}
	return ""
}

func TestUpdate_LoadingStaysWithTheActiveMode(t *testing.T) {
	m := model{mode: ModeFiles, viewport: viewport.New(), loading: true}

	// History finishes loading while the user is already waiting on files mode.
	updated, _ := m.Update(historyLoadedMsg{entries: []history.Entry{{Cmd: "ls", When: 1, Count: 1}}})
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("Update() returned %T, want model", updated)
	}

	if !got.loading {
		t.Error("loading cleared by another mode's result while files mode is still loading")
	}
	if len(got.historyEntries) != 1 {
		t.Errorf("historyEntries = %d, want the loaded entry to be kept", len(got.historyEntries))
	}
}

func TestSwitchToCachedMode_ClearsLoading(t *testing.T) {
	m := model{
		mode:           ModeHistory,
		input:          textinput.New(),
		viewport:       viewport.New(),
		previewCache:   map[string]string{},
		ready:          true,
		width:          80,
		listWidth:      40,
		mainHeight:     10,
		historyEntries: []history.Entry{{Cmd: "ls", When: 1, Count: 1}},
	}
	m.loadItemsForMode()
	m.updateFilter("")

	// Files mode has no cached data, so the walk starts and loading is set.
	if cmd := m.switchToFilesMode(); cmd == nil {
		t.Fatal("switchToFilesMode() returned no load command")
	}
	// Back to history before the walk finishes: the cached path runs.
	m.switchToHistoryMode()

	// The walk completes for a mode the user already left.
	updated, _ := m.Update(filesLoadedMsg{entries: []files.Entry{{Path: "a.go"}}})
	got, ok := updated.(model)
	if !ok {
		t.Fatalf("Update() returned %T, want model", updated)
	}

	got.updateFilter("zzz-no-such-entry")
	if len(got.filtered) != 0 {
		t.Fatalf("filtered = %d items, want an empty list", len(got.filtered))
	}
	if strings.Contains(got.View().Content, "Loading...") {
		t.Error("view renders Loading... after switching to a mode with cached data")
	}
}

func TestUpdate_ResizeRegeneratesPreview(t *testing.T) {
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go")}
	for i, path := range paths {
		body := "// " + strings.Repeat(string(rune('a'+i)), 300) + "\n"
		if err := os.WriteFile(path, []byte(body+body), 0o600); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}

	m := model{
		mode:         ModeFiles,
		viewport:     viewport.New(),
		previewCache: map[string]string{},
		fileEntries:  []files.Entry{{Path: paths[0]}, {Path: paths[1]}},
	}
	m.loadItemsForMode()
	m.updateFilter("")

	resize := func(m model, width, height int) model {
		updated, _ := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
		got, ok := updated.(model)
		if !ok {
			t.Fatalf("Update() returned %T, want model", updated)
		}
		return got
	}
	// wantView renders the selected entry for the pane size the model
	// currently has, independently of the model's own preview state.
	wantView := func(m model) string {
		entry := m.filtered[m.cursor].Original.(files.Entry)
		vp := viewport.New()
		vp.SetWidth(m.viewport.Width())
		vp.SetHeight(m.viewport.Height())
		vp.SetContent(entry.GeneratePreview(m.viewport.Width(), m.viewport.Height()))
		return vp.View()
	}

	m = resize(m, 200, 30)
	m = resize(m, 60, 30)

	if got, want := m.viewport.View(), wantView(m); got != want {
		t.Errorf("preview after resize was not re-rendered for the new pane size:\n%s\nwant:\n%s", got, want)
	}

	// Moving away and back must not restore the render cached at the old size.
	selected := m.cursor
	m.cursor = (selected + 1) % len(m.filtered)
	m.updatePreview()
	m.cursor = selected
	m.updatePreview()

	if got, want := m.viewport.View(), wantView(m); got != want {
		t.Errorf("preview served a cached render from the old pane size:\n%s\nwant:\n%s", got, want)
	}
}
