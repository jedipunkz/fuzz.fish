package app

import (
	"sort"
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
	"github.com/jedipunkz/fuzz.fish/internal/scoring"
	"github.com/sahilm/fuzzy"
)

// loadItemsForMode loads all items for the current mode
func (m *model) loadItemsForMode() {
	switch m.mode {
	case ModeHistory:
		// History: entries are Newest -> Oldest
		// We want Newest at Bottom.
		// Item[0] should be Oldest, Item[N] should be Newest.
		n := len(m.historyEntries)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.historyEntries {
			e := m.historyEntries[n-1-i]
			m.allItems[i] = Item{
				Text:     e.Cmd,
				Index:    n - 1 - i,
				Original: e,
			}
		}
	case ModeGitBranch:
		// Git: branches are collected.
		// We reverse them to put first item at bottom.
		n := len(m.gitBranches)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.gitBranches {
			b := m.gitBranches[n-1-i]
			m.allItems[i] = Item{
				Text:      b.Name,
				Index:     n - 1 - i,
				Original:  b,
				IsCurrent: b.IsCurrent,
				IsRemote:  b.IsRemote,
			}
		}
	case ModeFiles:
		// Files: entries are in directory order
		// We reverse them to put first item at bottom.
		n := len(m.fileEntries)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.fileEntries {
			f := m.fileEntries[n-1-i]
			m.allItems[i] = Item{
				Text:     f.Path,
				Index:    n - 1 - i,
				Original: f,
				IsDir:    f.IsDir,
			}
		}
	case ModeWorktree:
		// Worktrees: keep listing order, reverse so first item sits at bottom.
		n := len(m.worktrees)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.worktrees {
			w := m.worktrees[n-1-i]
			m.allItems[i] = Item{
				Text: w.Path,
				// Search path and branch together so both are fuzzy-matchable.
				// Mirrors the display layout (path + " [branch]") minus the icon.
				SearchText: w.Path + " [" + w.Branch + "]",
				Index:      n - 1 - i,
				Original:   w,
				IsCurrent:  w.IsCurrent,
				IsDir:      true,
			}
		}
	case ModeCommit:
		// Commits: git log is newest first, reverse so the newest sits at the
		// bottom next to the cursor.
		n := len(m.commits)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.commits {
			c := m.commits[n-1-i]
			m.allItems[i] = Item{
				Text: c.Hash,
				// Search hash and subject together so the message is
				// incrementally searchable. Mirrors the display layout.
				SearchText: c.Hash + " " + c.Subject,
				Index:      n - 1 - i,
				Original:   c,
			}
		}
	case ModePullRequest:
		// gh lists newest first, reverse so the newest sits at the bottom.
		n := len(m.pullRequests)
		if cap(m.allItems) >= n {
			m.allItems = m.allItems[:n]
		} else {
			m.allItems = make([]Item, n)
		}
		for i := range m.pullRequests {
			pr := m.pullRequests[n-1-i]
			m.allItems[i] = Item{
				Text:     "#" + strconv.Itoa(pr.Number) + " " + pr.Title,
				Index:    n - 1 - i,
				Original: pr,
			}
		}
	default:
		m.allItems = m.allItems[:0]
	}

	// Pre-build the search strings and their lowercase view: fuzzy matching
	// consumes allItemsStr, and glob matching would otherwise re-lowercase
	// every item on every keystroke.
	m.allItemsStr = make([]string, len(m.allItems))
	m.allItemsStrLower = make([]string, len(m.allItems))
	for i := range m.allItems {
		s := m.allItems[i].SearchText
		if s == "" {
			s = m.allItems[i].Text
		}
		m.allItemsStr[i] = s
		m.allItemsStrLower[i] = strings.ToLower(s)
	}
}

// sortDedupe returns the indexes sorted ascending with duplicates removed.
// Tokens may match overlapping or out-of-order positions, so the combined
// match set is normalized before scoring and highlighting.
func sortDedupe(ids []int) []int {
	if len(ids) < 2 {
		return ids
	}
	sort.Ints(ids)
	out := ids[:1]
	for _, id := range ids[1:] {
		if id != out[len(out)-1] {
			out = append(out, id)
		}
	}
	return out
}

// scoringSignals extracts an item's secondary ranking inputs: the recency
// timestamp, the usage frequency behind frecency, and whether the item is the
// current branch. Modes that carry none of them score on match quality alone.
func (m *model) scoringSignals(item Item) (timestamp int64, frequency int, isCurrent bool) {
	switch m.mode {
	case ModeHistory:
		if entry, ok := item.Original.(history.Entry); ok {
			timestamp, frequency = entry.When, entry.Count
		}
	case ModeGitBranch:
		if branch, ok := item.Original.(git.Branch); ok {
			timestamp, isCurrent = branch.CommitTimestamp, branch.IsCurrent
		}
	case ModeCommit:
		if c, ok := item.Original.(git.Commit); ok {
			timestamp = c.When
		}
	}
	return timestamp, frequency, isCurrent
}

// resizeFiltered sets filtered to length n, reusing the slice when capacity
// allows so typing does not allocate a new result list per keystroke.
func (m *model) resizeFiltered(n int) {
	if cap(m.filtered) >= n {
		m.filtered = m.filtered[:n]
	} else {
		m.filtered = make([]Item, n)
	}
}

// filterAllItems fills filtered with a copy of allItems.
func (m *model) filterAllItems() {
	m.resizeFiltered(len(m.allItems))
	copy(m.filtered, m.allItems)
}

// rankedItem is a matched candidate awaiting ordering: its allItems index,
// the matched indexes for highlighting, and its combined score.
type rankedItem struct {
	itemIdx int
	idx     []int
	score   float64
}

// setFilteredRanked fills filtered with the ranked candidates. Higher combined
// score should appear at bottom (higher priority), so they are sorted
// ascending; the sort is stable to keep the matcher's order on ties.
func (m *model) setFilteredRanked(ranked []rankedItem) {
	sort.SliceStable(ranked, func(i, j int) bool {
		return ranked[i].score < ranked[j].score
	})

	m.resizeFiltered(len(ranked))
	for rank, r := range ranked {
		item := m.allItems[r.itemIdx]
		item.MatchedIndexes = r.idx
		m.filtered[rank] = item
	}
}

// updateFilter updates the filtered items based on the query and returns the
// tea.Cmd generating the preview of the newly selected item (nil when the
// render is synchronous).
func (m *model) updateFilter(query string) tea.Cmd {
	if query == "" {
		// Return all items (which are already in display order)
		m.filterAllItems()
	} else {
		// Fuzzy search using pre-built search strings (avoids per-keystroke allocation)
		tokens := strings.Fields(query)
		if len(tokens) > 0 && queryHasGlob(query) {
			// Glob matching: a '*' in the query switches to literal, ordered
			// substring matching (e.g. "nvim *.go") instead of fuzzy scatter.
			m.globFilter(tokens)
		} else if len(tokens) > 0 {
			matches := fuzzy.Find(tokens[0], m.allItemsStr)

			// Per-candidate allItems index, aggregated fuzzy score and matched
			// indexes, kept aligned with each other through every token pass.
			items := make([]int, len(matches))
			aggScores := make([]int, len(matches))
			aggIdx := make([][]int, len(matches))
			for i, mat := range matches {
				items[i] = mat.Index
				aggScores[i] = mat.Score
				aggIdx[i] = mat.MatchedIndexes
			}

			// Multi-token queries ("git pull") AND each token, but the
			// combined score must reflect all tokens: summed fuzzy score and
			// the union of matched indexes. Keeping only the first token's
			// data hides where later tokens matched, so a contiguous match
			// ("git pull origin main") could not be distinguished from a
			// scattered one ("git config pull.rebase true").
			if len(tokens) > 1 {
				// subset is reused across token passes within this filter run:
				// capacity stabilizes after the first pass.
				subset := make([]string, len(items))
				for _, token := range tokens[1:] {
					if len(items) == 0 {
						break
					}
					subset = subset[:len(items)]
					for i, itemIdx := range items {
						subset[i] = m.allItemsStr[itemIdx]
					}
					subMatches := fuzzy.Find(token, subset)
					// Fresh slices per pass: subMatches is ordered by score, so
					// writing into the slices being read would overwrite
					// candidates before they are read.
					nextItems := make([]int, len(subMatches))
					nextScores := make([]int, len(subMatches))
					nextIdx := make([][]int, len(subMatches))
					for j, sm := range subMatches {
						nextItems[j] = items[sm.Index]
						nextScores[j] = aggScores[sm.Index] + sm.Score
						// fuzzy.Find allocates MatchedIndexes per match, so each
						// candidate owns its slice and it can be extended in place.
						nextIdx[j] = append(aggIdx[sm.Index], sm.MatchedIndexes...)
					}
					items, aggScores, aggIdx = nextItems, nextScores, nextIdx
				}
			}

			// Pre-calculate scores for all matches (O(n) instead of O(n log n) in comparator)
			config := scoring.DefaultConfig()
			now := scoring.CurrentTimestamp()
			ranked := make([]rankedItem, len(items))
			for i, itemIdx := range items {
				idx := sortDedupe(aggIdx[i])
				timestamp, frequency, isCurrent := m.scoringSignals(m.allItems[itemIdx])
				// Score against the string the indexes were matched in, not the
				// display text: they differ in worktree mode, where the branch
				// suffix is part of the search string.
				score := config.ItemScore(m.allItemsStr[itemIdx], aggScores[i], idx, timestamp, frequency, isCurrent, now)
				ranked[i] = rankedItem{itemIdx: itemIdx, idx: idx, score: score}
			}
			m.setFilteredRanked(ranked)
		} else {
			// Query is just whitespace, treat as empty
			m.filterAllItems()
		}
	}

	if len(m.filtered) > 0 {
		m.cursor = len(m.filtered) - 1
		m.offset = m.cursor - m.mainHeight + 1
		if m.offset < 0 {
			m.offset = 0
		}
	} else {
		m.cursor = 0
		m.offset = 0
	}
	return m.updatePreview()
}
