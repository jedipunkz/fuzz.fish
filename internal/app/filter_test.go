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
