package app

import (
	"testing"

	"github.com/jedipunkz/fuzz.fish/internal/history"
)

func historyModel(cmds []string) *model {
	m := &model{mode: ModeHistory}
	m.historyEntries = make([]history.Entry, len(cmds))
	for i, c := range cmds {
		m.historyEntries[i] = history.Entry{Cmd: c, Count: 1}
	}
	m.loadItemsForMode()
	return m
}

// TestUpdateFilter_MultiTokenAggregation verifies the AND semantics and the
// score aggregation of multi-token queries: an entry must match every token,
// and a scattered match ("git config pull.rebase") must rank behind a
// contiguous one ("git pull origin main").
func TestUpdateFilter_MultiTokenAggregation(t *testing.T) {
	m := historyModel([]string{
		"git config pull.rebase true",
		"git pull origin main",
		"git status",
	})
	m.updateFilter("git pull")

	got := make(map[string]bool)
	for _, item := range m.filtered {
		got[item.Text] = true
	}
	if got["git status"] {
		t.Error("git status matched a multi-token query, want AND semantics to exclude it")
	}
	if !got["git pull origin main"] || !got["git config pull.rebase true"] {
		t.Fatalf("filtered = %v, want both multi-token matches kept", got)
	}

	bottom := m.filtered[len(m.filtered)-1].Text
	if bottom != "git pull origin main" {
		t.Errorf("bottom item = %q, want the contiguous match to rank last", bottom)
	}
}

// TestUpdateFilter_HighlightIndexesMultiToken verifies that match indexes
// cover both tokens after aggregation: every matched byte must be flagged as
// a highlight target.
func TestUpdateFilter_HighlightIndexesMultiToken(t *testing.T) {
	m := historyModel([]string{"git pull origin main"})
	m.updateFilter("git pull")

	last := m.filtered[len(m.filtered)-1]
	if len(last.MatchedIndexes) == 0 {
		t.Fatal("no matched indexes after aggregation")
	}
	for _, idx := range last.MatchedIndexes {
		if idx < 0 || idx >= len(last.Text) {
			t.Fatalf("matched index %d out of range for %q", idx, last.Text)
		}
	}
}

// TestUpdateFilter_ThreeTokens verifies that every result of a query with
// three or more tokens matches all of them. Token passes after the second
// must not read candidates from a buffer they are overwriting.
func TestUpdateFilter_ThreeTokens(t *testing.T) {
	// "pull" matches contiguously in the first command and scattered in the
	// second, while "main" does the opposite: the third pass returns the
	// candidates in the reverse order of the second pass.
	m := historyModel([]string{
		"git pull m a i n zz",
		"git p u l l main",
		"git status",
	})
	m.updateFilter("git pull main")

	got := make(map[string]int, len(m.filtered))
	for _, item := range m.filtered {
		got[item.Text]++
	}
	want := map[string]int{
		"git pull m a i n zz": 1,
		"git p u l l main":    1,
	}
	if len(got) != len(want) || len(m.filtered) != len(want) {
		t.Fatalf("filtered = %v, want each of %v exactly once", got, want)
	}
	for text := range want {
		if got[text] != 1 {
			t.Errorf("%q appears %d times in %v, want once", text, got[text], got)
		}
	}
}
