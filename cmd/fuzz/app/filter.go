package app

import (
	"sort"
	"strings"

	"github.com/sahilm/fuzzy"
)

// loadItemsForMode loads all items for the current mode
func (m *model) loadItemsForMode() {
	m.allItems = []Item{}

	if m.mode == ModeHistory {
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
	} else {
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
	}
}

// updateFilter updates the filtered items based on the query
func (m *model) updateFilter(query string) {
	if query == "" {
		// Return all items (which are already in display order)
		m.filtered = make([]Item, len(m.allItems))
		copy(m.filtered, m.allItems)
	} else {
		// Fuzzy search
		src := make([]string, len(m.allItems))
		// We need search against original list order?
		// m.allItems is already reversed for display.
		// Usually we search against the "source of truth".
		// Let's search against m.allItems text.
		for i, item := range m.allItems {
			src[i] = item.Text
		}

		tokens := strings.Fields(query)
		if len(tokens) > 0 {
			matches := fuzzy.Find(tokens[0], src)

			for _, token := range tokens[1:] {
				if len(matches) == 0 {
					break
				}
				subset := make([]string, len(matches))
				for i, mat := range matches {
					subset[i] = src[mat.Index]
				}
				subMatches := fuzzy.Find(token, subset)
				newMatches := make(fuzzy.Matches, len(subMatches))
				for i, sm := range subMatches {
					newMatches[i] = matches[sm.Index]
				}
				matches = newMatches
			}

			// Sort logic
			sort.SliceStable(matches, func(i, j int) bool {
				scoreI := float64(matches[i].Score)
				scoreJ := float64(matches[j].Score)

				if m.mode == ModeHistory {
					total := float64(len(m.allItems))
					// Increase bonus for recency to ensure newer commands appear at the bottom
					maxBonus := 10000.0
					recencyI := float64(matches[i].Index) / total
					recencyJ := float64(matches[j].Index) / total
					scoreI += recencyI * maxBonus
					scoreJ += recencyJ * maxBonus
				}
				return scoreI < scoreJ
			})

			m.filtered = make([]Item, len(matches))
			for i, mat := range matches {
				item := m.allItems[mat.Index]
				item.MatchedIndexes = mat.MatchedIndexes
				m.filtered[i] = item
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
