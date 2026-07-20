package app

import (
	"reflect"
	"testing"

	"github.com/jedipunkz/fuzz.fish/internal/history"
)

func TestQueryHasGlob(t *testing.T) {
	cases := map[string]bool{
		"nvim *.go": true,
		"*.go":      true,
		"git pull":  false,
		"":          false,
	}
	for q, want := range cases {
		if got := queryHasGlob(q); got != want {
			t.Errorf("queryHasGlob(%q) = %v, want %v", q, got, want)
		}
	}
}

func TestGlobMatch(t *testing.T) {
	tests := []struct {
		token  string
		text   string
		want   []int
		wantOK bool
	}{
		{"*.go", "nvim main.go", []int{9, 10, 11}, true},    // ".go"
		{"nvim", "nvim main.go", []int{0, 1, 2, 3}, true},   // "nvim"
		{"*.go", "cargo build", nil, false},                 // no ".go"
		{"a*b", "xaxxb", []int{1, 4}, true},                 // a...b
		{"*", "anything", nil, true},                        // matches, no indexes
		{"*.go", "nvim FILTER.GO", []int{11, 12, 13}, true}, // case-insensitive via caller lowercasing
	}
	for _, tt := range tests {
		// globMatch expects lowercased inputs; lowercase the text to mirror the caller.
		got, ok := globMatch(tt.token, lower(tt.text))
		if ok != tt.wantOK {
			t.Errorf("globMatch(%q, %q) ok = %v, want %v", tt.token, tt.text, ok, tt.wantOK)
			continue
		}
		if ok && !reflect.DeepEqual(got, tt.want) {
			t.Errorf("globMatch(%q, %q) = %v, want %v", tt.token, tt.text, got, tt.want)
		}
	}
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if b[i] >= 'A' && b[i] <= 'Z' {
			b[i] += 'a' - 'A'
		}
	}
	return string(b)
}

// TestGlobFilterHistory verifies that "nvim *.go" surfaces only nvim commands
// that opened .go files, filtering out unrelated and wildcard-mismatched entries.
func TestGlobFilterHistory(t *testing.T) {
	cmds := []string{
		"nvim internal/app/filter.go",
		"nvim README.md",
		"git status",
		"vim main.go",
		"nvim cmd/fuzz/main.go",
	}
	m := &model{mode: ModeHistory}
	m.allItems = make([]Item, len(cmds))
	m.allItemsStr = make([]string, len(cmds))
	m.historyFreqMap = make(map[string]int)
	for i, c := range cmds {
		m.allItems[i] = Item{Text: c, Index: i, Original: history.Entry{Cmd: c}}
		m.allItemsStr[i] = c
		m.historyFreqMap[c]++
	}

	m.updateFilter("nvim *.go")

	got := make(map[string]bool)
	for _, item := range m.filtered {
		got[item.Text] = true
	}
	want := []string{
		"nvim internal/app/filter.go",
		"nvim cmd/fuzz/main.go",
	}
	if len(got) != len(want) {
		t.Fatalf("filtered = %v, want exactly %v", got, want)
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("expected %q in filtered results, got %v", w, got)
		}
	}
}
