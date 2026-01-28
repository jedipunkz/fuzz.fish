package app

import (
	"sort"
	"strings"
	"time"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
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
			// Higher combined score should appear at bottom (higher priority)
			// So we sort ascending: lower scores first, higher scores last (at bottom)
			sort.SliceStable(matches, func(i, j int) bool {
				scoreI := float64(matches[i].Score)
				scoreJ := float64(matches[j].Score)

				if m.mode == ModeHistory {
					// Time-based recency bonus using actual timestamps
					// Hyperbolic decay: bonus = maxBonus / (1 + ageInHours)
					// This strongly prioritizes recent commands:
					//   0h ago -> 3000, 1h -> 1500, 2h -> 1000, 6h -> 429, 24h -> 120
					now := time.Now().Unix()
					maxBonus := 3000.0

					recencyBonus := func(idx int) float64 {
						item := m.allItems[idx]
						if entry, ok := item.Original.(history.Entry); ok && entry.When > 0 {
							ageHours := float64(now-entry.When) / 3600.0
							if ageHours < 0 {
								ageHours = 0
							}
							return maxBonus / (1.0 + ageHours)
						}
						// Fallback: position-based bonus
						total := float64(len(m.allItems))
						return float64(idx) / total * maxBonus * 0.1
					}

					scoreI += recencyBonus(matches[i].Index)
					scoreJ += recencyBonus(matches[j].Index)
				}
				// Ascending sort: lower scores at top, higher scores at bottom
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
