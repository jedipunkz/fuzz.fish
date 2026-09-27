package app

import (
	"sort"
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

// filterAllItems fills filtered with a copy of allItems, reusing the slice
// when capacity allows.
func (m *model) filterAllItems() {
	if cap(m.filtered) >= len(m.allItems) {
		m.filtered = m.filtered[:len(m.allItems)]
	} else {
		m.filtered = make([]Item, len(m.allItems))
	}
	copy(m.filtered, m.allItems)
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

			// Per-match aggregation aligned with the matches slice; built by
			// whichever path matches this query, then fed to the shared
			// scoring/sorting tail below.
			aggScores := make([]int, len(matches))
			aggIdx := make([][]int, len(matches))

			if len(tokens) == 1 {
				// Single-token queries (the common case) need no cross-token
				// aggregation: fuzzy.Find's score and match indexes feed
				// scoring and highlighting directly, avoiding map allocations.
				for i, mat := range matches {
					aggScores[i] = mat.Score
					aggIdx[i] = sortDedupe(mat.MatchedIndexes)
				}
			} else {
				// Multi-token queries ("git pull") AND each token, but the
				// combined score must reflect all tokens: summed fuzzy score
				// and the union of matched indexes. Keeping only the first
				// token's data hides where later tokens matched, so a
				// contiguous match ("git pull origin main") could not be
				// distinguished from a scattered one
				// ("git config pull.rebase true").
				//
				// Aggregation is keyed by the allItems index (matches change
				// order between token passes), then materialized aligned with
				// the final matches below.
				aggScore := make(map[int]int, len(matches))
				aggIdxByKey := make(map[int][]int, len(matches))
				for _, mat := range matches {
					aggScore[mat.Index] = mat.Score
					aggIdxByKey[mat.Index] = append([]int(nil), mat.MatchedIndexes...)
				}

				// subset and newMatches are reused across token passes within
				// this filter run: capacity stabilizes after the first pass.
				subset := make([]string, len(matches))
				newMatches := fuzzy.Matches{}

				for _, token := range tokens[1:] {
					if len(matches) == 0 {
						break
					}
					subset = subset[:len(matches)]
					for i, mat := range matches {
						subset[i] = m.allItemsStr[mat.Index]
					}
					subMatches := fuzzy.Find(token, subset)
					if cap(newMatches) < len(subMatches) {
						newMatches = make(fuzzy.Matches, 0, len(matches))
					} else {
						newMatches = newMatches[:0]
					}
					for _, sm := range subMatches {
						orig := matches[sm.Index]
						aggScore[orig.Index] += sm.Score
						aggIdxByKey[orig.Index] = append(aggIdxByKey[orig.Index], sm.MatchedIndexes...)
						newMatches = append(newMatches, orig)
					}
					matches = newMatches
				}

				for i, mat := range matches {
					aggScores[i] = aggScore[mat.Index]
					aggIdx[i] = sortDedupe(aggIdxByKey[mat.Index])
				}
			}

			// Pre-calculate scores for all matches (O(n) instead of O(n log n) in comparator)
			config := scoring.DefaultConfig()
			now := scoring.CurrentTimestamp()
			scores := make([]float64, len(matches))
			for i, mat := range matches {
				timestamp, frequency, isCurrent := m.scoringSignals(m.allItems[mat.Index])
				// Score against the string the indexes were matched in, not the
				// display text: they differ in worktree mode, where the branch
				// suffix is part of the search string.
				scores[i] = config.ItemScore(m.allItemsStr[mat.Index], aggScores[i], aggIdx[i], timestamp, frequency, isCurrent, now)
			}

			// Create index array for sorting (scores array must stay aligned with original matches)
			indices := make([]int, len(matches))
			for i := range indices {
				indices[i] = i
			}

			// Sort indices by pre-calculated scores
			// Higher combined score should appear at bottom (higher priority)
			// So we sort ascending: lower scores first, higher scores last (at bottom)
			sort.SliceStable(indices, func(i, j int) bool {
				return scores[indices[i]] < scores[indices[j]]
			})

			// Build filtered list using sorted indices
			m.filtered = make([]Item, len(indices))
			for rank, idx := range indices {
				mat := matches[idx]
				item := m.allItems[mat.Index]
				item.MatchedIndexes = aggIdx[idx]
				m.filtered[rank] = item
			}
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
