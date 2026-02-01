package app

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// loadItemsForMode loads all items for the current mode
func (m *model) loadItemsForMode() {
	m.allItems = []Item{}

	switch m.mode {
	case ModeHistory:
		// History: entries are Newest -> Oldest
		// We want Newest at Bottom.
		// Item[0] should be Oldest, Item[N] should be Newest.
		for i := range m.historyEntries {
			e := m.historyEntries[len(m.historyEntries)-1-i]
			m.allItems = append(m.allItems, Item{
				Text:     e.Cmd,
				Index:    len(m.historyEntries) - 1 - i,
				Original: e,
			})
		}
	case ModeGitBranch:
		// Git: branches are collected.
		// Sort? CollectBranches usually returns some order.
		// We want Default/Current at bottom?
		// Let's assume input branches are standard.
		// We reverse them to put first item at bottom.
		for i := range m.gitBranches {
			b := m.gitBranches[len(m.gitBranches)-1-i]
			m.allItems = append(m.allItems, Item{
				Text:      b.Name,
				Index:     len(m.gitBranches) - 1 - i,
				Original:  b,
				IsCurrent: b.IsCurrent,
				IsRemote:  b.IsRemote,
			})
		}
	case ModeFiles:
		// Files: entries are in directory order
		// We reverse them to put first item at bottom.
		for i := range m.fileEntries {
			f := m.fileEntries[len(m.fileEntries)-1-i]
			m.allItems = append(m.allItems, Item{
				Text:     f.Path,
				Index:    len(m.fileEntries) - 1 - i,
				Original: f,
				IsDir:    f.IsDir,
			})
		}
	}

	// Pre-build search strings to avoid per-keystroke allocation
	m.allItemsStr = make([]string, len(m.allItems))
	for i, item := range m.allItems {
		m.allItemsStr[i] = item.Text
	}
}

// updateFilter updates the filtered items based on the query
func (m *model) updateFilter(query string) {
	if query == "" {
		// Return all items (which are already in display order)
		m.filtered = make([]Item, len(m.allItems))
		copy(m.filtered, m.allItems)
	} else {
		// Fuzzy search using pre-built search strings (avoids per-keystroke allocation)
		tokens := strings.Fields(query)
		if len(tokens) > 0 {
			matches := fuzzy.Find(tokens[0], m.allItemsStr)

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
					newMatches[i] = matches[sm.Index]
				}
				matches = newMatches
			}

			// Pre-calculate scores for all matches (O(n) instead of O(n log n) in comparator)
			config := DefaultScoringConfig()
			now := GetCurrentTimestamp()
			scores := make([]float64, len(matches))
			for i, mat := range matches {
				item := m.allItems[mat.Index]
				scores[i] = CalculateItemScore(item, mat.Score, mat.MatchedIndexes, m.mode, config, now)
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
				item.MatchedIndexes = mat.MatchedIndexes
				m.filtered[rank] = item
			}
		} else {
			// Query is just whitespace, treat as empty
			m.filtered = make([]Item, len(m.allItems))
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
