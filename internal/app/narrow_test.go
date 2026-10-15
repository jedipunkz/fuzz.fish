package app

import (
	"slices"
	"testing"
)

// TestUpdateFilter_NarrowingMatchesFullScan types queries one after another,
// so each filter may narrow to the previous matches, and checks every result
// against a fresh model that has to scan all items.
func TestUpdateFilter_NarrowingMatchesFullScan(t *testing.T) {
	tests := []struct {
		name    string
		cmds    []string
		queries []string
	}{
		{
			name: "typing",
			cmds: append(benchCommands(2000), "Git Pull --rebase", "GIT_TRACE=1 git pull"),
			queries: []string{
				"g", "gi", "git", "git ", "git p", "git pu", "git pull", "git pull o",
				"git pull or",                                     // narrowed
				"git p",                                           // backspace: full scan
				"git p*", "git pu*", "git pull *", "git pull *ra", // glob, narrowed
				"git pull",             // glob to fuzzy: full scan
				"zzz", "zzzz", "zzzz ", // no matches, narrowed to nothing
				"", "nvim", "nvim *", "nvim *.go", "nvim *.go 1",
				"   ", "  n",
			},
		},
		{
			// "a" ranks these in reverse allItems order but "ab" scores them
			// equally, so the tie order shows whether the narrowed
			// candidates were fed back in allItems order.
			name:    "tie order",
			cmds:    []string{"a__bb", "AA_bA"},
			queries: []string{"a", "ab"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := historyModel(tt.cmds)
			for _, q := range tt.queries {
				m.updateFilter(q)
				want := historyModel(tt.cmds)
				want.updateFilter(q)
				if !sameItems(m.filtered, want.filtered) {
					t.Fatalf("updateFilter(%q) after narrowing differs from a full scan", q)
				}
			}
		})
	}
}

// TestUpdateFilter_NoNarrowingAfterItemsCleared checks that a refresh, which
// drops allItems while the reload is in flight, also drops the recorded
// matches: they index the old items.
func TestUpdateFilter_NoNarrowingAfterItemsCleared(t *testing.T) {
	m := historyModel(benchCommands(100))
	m.updateFilter("git")
	m.refreshMode()
	m.updateFilter("git p") // must not index the cleared items
	if len(m.filtered) != 0 {
		t.Fatalf("got %d items with no items loaded, want 0", len(m.filtered))
	}
}

func sameItems(a, b []Item) bool {
	return slices.EqualFunc(a, b, func(x, y Item) bool {
		return x.Index == y.Index && slices.Equal(x.MatchedIndexes, y.MatchedIndexes)
	})
}
