package app

import (
	"sort"
	"strings"

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
		// Build frequency map for frecency scoring
		m.historyFreqMap = make(map[string]int, n)
		for _, e := range m.historyEntries {
			m.historyFreqMap[e.Cmd]++
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
	default:
		m.allItems = m.allItems[:0]
	}

	// Pre-build search strings to avoid per-keystroke allocation
	m.allItemsStr = make([]string, len(m.allItems))
	for i, item := range m.allItems {
		if item.SearchText != "" {
			m.allItemsStr[i] = item.SearchText
		} else {
			m.allItemsStr[i] = item.Text
		}
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

// updateFilter updates the filtered items based on the query
func (m *model) updateFilter(query string) {
	if query == "" {
		// Return all items (which are already in display order)
		// Reuse existing slice if capacity allows
		if cap(m.filtered) >= len(m.allItems) {
			m.filtered = m.filtered[:len(m.allItems)]
		} else {
			m.filtered = make([]Item, len(m.allItems))
		}
		copy(m.filtered, m.allItems)
	} else {
		// Fuzzy search using pre-built search strings (avoids per-keystroke allocation)
		tokens := strings.Fields(query)
		if len(tokens) > 0 && queryHasGlob(query) {
			// Glob matching: a '*' in the query switches to literal, ordered
			// substring matching (e.g. "nvim *.go") instead of fuzzy scatter.
			m.globFilter(tokens)
		} else if len(tokens) > 0 {
			matches := fuzzy.Find(tokens[0], m.allItemsStr)

			// Aggregate per-item match quality across every token. Multi-token
			// queries ("git pull") AND each token, but the combined score must
			// reflect all tokens: summed fuzzy score and the union of matched
			// indexes. Keeping only the first token's data hides where later
			// tokens matched, so a contiguous match ("git pull origin main")
			// could not be distinguished from a scattered one
			// ("git config pull.rebase true").
			aggScore := make(map[int]int, len(matches))
			aggIdx := make(map[int][]int, len(matches))
			for _, mat := range matches {
				aggScore[mat.Index] = mat.Score
				aggIdx[mat.Index] = append([]int(nil), mat.MatchedIndexes...)
			}

			for _, token := range tokens[1:] {
				if len(matches) == 0 {
					break
				}
				subset := make([]string, len(matches))
				for i, mat := range matches {
					subset[i] = m.allItemsStr[mat.Index]
				}
				subMatches := fuzzy.Find(token, subset)
				newMatches := make(fuzzy.Matches, len(subMatches))
				for i, sm := range subMatches {
					orig := matches[sm.Index]
					aggScore[orig.Index] += sm.Score
					aggIdx[orig.Index] = append(aggIdx[orig.Index], sm.MatchedIndexes...)
					newMatches[i] = orig
				}
				matches = newMatches
			}

			// Sort and dedupe each item's matched indexes so gap/boundary
			// bonuses and highlighting see the full, ordered match set.
			for idx, ids := range aggIdx {
				aggIdx[idx] = sortDedupe(ids)
			}

			// Pre-calculate scores for all matches (O(n) instead of O(n log n) in comparator)
			config := scoring.DefaultConfig()
			now := scoring.CurrentTimestamp()
			scores := make([]float64, len(matches))
			for i, mat := range matches {
				item := m.allItems[mat.Index]
				var timestamp int64
				var frequency int
				var isCurrent bool
				switch m.mode {
				case ModeHistory:
					if entry, ok := item.Original.(history.Entry); ok {
						timestamp = entry.When
					}
					frequency = m.historyFreqMap[item.Text]
				case ModeGitBranch:
					if branch, ok := item.Original.(git.Branch); ok {
						timestamp = branch.CommitTimestamp
						isCurrent = branch.IsCurrent
					}
				}
				scores[i] = config.ItemScore(item.Text, aggScore[mat.Index], aggIdx[mat.Index], timestamp, frequency, isCurrent, now)
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
				item.MatchedIndexes = aggIdx[mat.Index]
				m.filtered[rank] = item
			}
		} else {
			// Query is just whitespace, treat as empty
			if cap(m.filtered) >= len(m.allItems) {
				m.filtered = m.filtered[:len(m.allItems)]
			} else {
				m.filtered = make([]Item, len(m.allItems))
			}
			copy(m.filtered, m.allItems)
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
	m.updatePreview()
}
