package app

import (
	"sort"
	"strings"

	"github.com/jedipunkz/fuzz.fish/internal/git"
	"github.com/jedipunkz/fuzz.fish/internal/history"
	"github.com/jedipunkz/fuzz.fish/internal/scoring"
)

// queryHasGlob reports whether the query should be matched with glob semantics.
// A '*' anywhere switches the whole query from fuzzy to glob matching, so a
// query like "nvim *.go" surfaces only commands that literally contain "nvim"
// and ".go" (atuin-style) instead of scattering those characters fuzzily.
func queryHasGlob(query string) bool {
	return strings.Contains(query, "*")
}

// globMatch matches a single whitespace-delimited glob token against text.
//
// The token is split on '*' into literal segments that must each appear
// contiguously and in order; '*' allows an arbitrary run (including empty)
// between them. Matching is unanchored on both ends, so "nvim" matches any
// text containing "nvim" and "*.go" matches any text containing ".go".
//
// Both arguments must already be lowercased by the caller so matching is
// case-insensitive; the returned indexes are byte offsets into text, aligned
// with the original string for ASCII (matching the existing highlight code).
func globMatch(token, text string) (matched []int, ok bool) {
	pos := 0
	for _, seg := range strings.Split(token, "*") {
		if seg == "" {
			continue
		}
		i := strings.Index(text[pos:], seg)
		if i < 0 {
			return nil, false
		}
		start := pos + i
		for k := 0; k < len(seg); k++ {
			matched = append(matched, start+k)
		}
		pos = start + len(seg)
	}
	return matched, true
}

// globFilter populates m.filtered using glob matching. Every token must match
// (AND); the union of matched indexes feeds the same scoring and highlighting
// pipeline as fuzzy matching, so frecency and match-quality ordering behave
// consistently across both search modes.
func (m *model) globFilter(tokens []string) {
	lowerTokens := make([]string, len(tokens))
	for i, t := range tokens {
		lowerTokens[i] = strings.ToLower(t)
	}

	config := scoring.DefaultConfig()
	now := scoring.CurrentTimestamp()

	type hit struct {
		itemIdx int
		idx     []int
		score   float64
	}
	hits := make([]hit, 0, len(m.allItems))

	for i := range m.allItems {
		text := strings.ToLower(m.allItemsStr[i])
		var idx []int
		matchedLen := 0
		ok := true
		for _, token := range lowerTokens {
			mIdx, matched := globMatch(token, text)
			if !matched {
				ok = false
				break
			}
			idx = append(idx, mIdx...)
			matchedLen += len(mIdx)
		}
		if !ok {
			continue
		}
		idx = sortDedupe(idx)

		item := m.allItems[i]
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
		score := config.ItemScore(item.Text, matchedLen, idx, timestamp, frequency, isCurrent, now)
		hits = append(hits, hit{itemIdx: i, idx: idx, score: score})
	}

	// Higher score at the bottom (higher priority), matching the fuzzy path.
	sort.SliceStable(hits, func(i, j int) bool {
		return hits[i].score < hits[j].score
	})

	m.filtered = make([]Item, len(hits))
	for rank, h := range hits {
		item := m.allItems[h.itemIdx]
		item.MatchedIndexes = h.idx
		m.filtered[rank] = item
	}
}
