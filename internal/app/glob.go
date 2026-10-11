package app

import (
	"slices"
	"strings"

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
// The token arrives pre-split on '*' (segs) into literal segments that must
// each appear contiguously and in order; '*' allows an arbitrary run
// (including empty) between them. Matching is unanchored on both ends, so
// "nvim" matches any text containing "nvim" and "*.go" matches any text
// containing ".go". Splitting is left to the caller so it runs once per
// keystroke rather than once per candidate.
//
// Both arguments must already be lowercased by the caller so matching is
// case-insensitive; the matched indexes are byte offsets into text, aligned
// with the original string for ASCII (matching the existing highlight code),
// appended to dst so the caller can reuse one buffer across candidates.
func globMatch(dst []int, segs []string, text string) (matched []int, ok bool) {
	pos := 0
	for _, seg := range segs {
		if seg == "" {
			continue
		}
		i := strings.Index(text[pos:], seg)
		if i < 0 {
			return dst, false
		}
		start := pos + i
		for k := 0; k < len(seg); k++ {
			dst = append(dst, start+k)
		}
		pos = start + len(seg)
	}
	return dst, true
}

// globFilter populates m.filtered using glob matching. Every token must match
// (AND); the union of matched indexes feeds the same scoring and highlighting
// pipeline as fuzzy matching, so frecency and match-quality ordering behave
// consistently across both search modes.
func (m *model) globFilter(tokens []string) {
	tokenSegs := make([][]string, len(tokens))
	for i, t := range tokens {
		tokenSegs[i] = strings.Split(strings.ToLower(t), "*")
	}

	config := scoring.DefaultConfig()
	now := scoring.CurrentTimestamp()

	hits := make([]rankedItem, 0, len(m.allItems))
	// buf collects one candidate's matched indexes and is reused across
	// candidates; only hits get their own copy.
	var buf []int

	for i := range m.allItems {
		// Lowercased at load time in loadItemsForMode: re-lowercasing every
		// candidate string here would allocate once per item per keystroke.
		text := m.allItemsStrLower[i]
		buf = buf[:0]
		ok := true
		for _, segs := range tokenSegs {
			buf, ok = globMatch(buf, segs, text)
			if !ok {
				break
			}
		}
		if !ok {
			continue
		}
		idx := slices.Clone(sortDedupe(buf))

		timestamp, frequency, isCurrent := m.scoringSignals(m.allItems[i])
		// Glob matches have no fuzzy score to pass through: matchedLen is not on
		// the same scale as one, and using it would shift the balance between
		// match quality and frecency compared with the fuzzy path. MatchBonus
		// already rewards contiguous, boundary-aligned matches, so it carries
		// the match quality alone here.
		score := config.ItemScore(m.allItemsStr[i], 0, idx, timestamp, frequency, isCurrent, now)
		hits = append(hits, rankedItem{itemIdx: i, idx: idx, score: score})
	}

	// Ordered like the fuzzy path: higher score at the bottom.
	m.setFilteredRanked(hits)
}
