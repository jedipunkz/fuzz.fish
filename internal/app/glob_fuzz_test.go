package app

import (
	"strings"
	"testing"
)

func FuzzGlobMatch(f *testing.F) {
	// Seeds taken from the hand-written cases in glob_test.go.
	f.Add("*.go", "nvim main.go")
	f.Add("nvim", "nvim main.go")
	f.Add("*.go", "cargo build")
	f.Add("a*b", "xaxxb")
	f.Add("*", "anything")
	f.Add("*.go", "nvim FILTER.GO")
	f.Add("", "")
	f.Add("***", "")
	f.Add("a*b*c", "abcabcabc")

	f.Fuzz(func(t *testing.T, token, text string) {
		lowerText := lower(text)
		segs := strings.Split(strings.ToLower(token), "*")
		got, ok := globMatch(nil, segs, lowerText)
		if !ok {
			return
		}
		// Matched indexes must point inside text and be strictly increasing,
		// since each segment is found at or after the previous one's end.
		prev := -1
		for _, idx := range got {
			if idx <= prev {
				t.Fatalf("globMatch(%q, %q): index %d not greater than %d", token, text, idx, prev)
			}
			if idx >= len(lowerText) {
				t.Fatalf("globMatch(%q, %q): index %d out of range (len %d)", token, text, idx, len(lowerText))
			}
			prev = idx
		}
	})
}
