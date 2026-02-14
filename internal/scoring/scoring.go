package scoring

import (
	"time"
	"unicode"
)

// Config holds configuration for the unified scoring algorithm
type Config struct {
	// Word boundary bonus (match after /, -, _, ., space)
	WordBoundaryBonus float64
	// Consecutive match bonus
	ConsecutiveBonus float64
	// Prefix match bonus (match at start)
	PrefixBonus float64
	// CamelCase bonus (match at uppercase letter)
	CamelCaseBonus float64
	// Recency bonus settings
	MaxRecencyBonus float64
	// Current branch bonus (git mode only)
	CurrentBranchBonus float64
}

// DefaultConfig returns the default scoring configuration
// inspired by fzy and fzf algorithms
func DefaultConfig() Config {
	return Config{
		WordBoundaryBonus:  50.0,  // Bonus for matching after word boundaries
		ConsecutiveBonus:   30.0,  // Bonus for each consecutive match
		PrefixBonus:        100.0, // Bonus for matching at the start
		CamelCaseBonus:     40.0,  // Bonus for matching uppercase in CamelCase
		MaxRecencyBonus:    3000.0,
		CurrentBranchBonus: 500.0, // Bonus for current git branch
	}
}

// isWordBoundary checks if the character at idx is a word boundary
func isWordBoundary(text string, idx int) bool {
	if idx == 0 {
		return true // Start of string is always a boundary
	}
	prev := rune(text[idx-1])
	return prev == '/' || prev == '-' || prev == '_' || prev == '.' || prev == ' '
}

// isCamelCaseBoundary checks if the character at idx is an uppercase letter
// following a lowercase letter (CamelCase boundary)
func isCamelCaseBoundary(text string, idx int) bool {
	if idx == 0 || idx >= len(text) {
		return false
	}
	curr := rune(text[idx])
	prev := rune(text[idx-1])
	return unicode.IsUpper(curr) && unicode.IsLower(prev)
}

// MatchBonus calculates bonus score based on match positions
// using fzy/fzf-inspired algorithm
func (c Config) MatchBonus(text string, matchedIndexes []int) float64 {
	if len(matchedIndexes) == 0 {
		return 0
	}

	var bonus float64

	// Prefix bonus: matching at the start is highly valuable
	if matchedIndexes[0] == 0 {
		bonus += c.PrefixBonus
	}

	prevIdx := -2 // Initialize to impossible value
	for _, idx := range matchedIndexes {
		// Word boundary bonus
		if isWordBoundary(text, idx) {
			bonus += c.WordBoundaryBonus
		}

		// CamelCase bonus
		if isCamelCaseBoundary(text, idx) {
			bonus += c.CamelCaseBonus
		}

		// Consecutive match bonus (affine gap penalty concept from fzy)
		if idx == prevIdx+1 {
			bonus += c.ConsecutiveBonus
		}

		prevIdx = idx
	}

	return bonus
}

// RecencyBonus calculates recency bonus using hyperbolic decay
// Formula: bonus = maxBonus / (1 + ageInHours)
// This strongly prioritizes recent items:
//
//	0h ago -> maxBonus, 1h -> maxBonus/2, 2h -> maxBonus/3, 24h -> maxBonus/25
func (c Config) RecencyBonus(timestamp, now int64) float64 {
	if timestamp <= 0 {
		return 0
	}

	ageHours := float64(now-timestamp) / 3600.0
	if ageHours < 0 {
		ageHours = 0
	}

	return c.MaxRecencyBonus / (1.0 + ageHours)
}

// ItemScore calculates the total score for an item
// combining fuzzy match score, position bonuses, and recency
func (c Config) ItemScore(text string, fuzzyScore int, matchedIndexes []int, timestamp int64, isCurrent bool, now int64) float64 {
	score := float64(fuzzyScore)

	// Add match position bonuses (fzy/fzf-inspired)
	score += c.MatchBonus(text, matchedIndexes)

	// Recency bonus
	if timestamp > 0 {
		score += c.RecencyBonus(timestamp, now)
	}

	// Current branch bonus
	if isCurrent {
		score += c.CurrentBranchBonus
	}

	return score
}

// CurrentTimestamp returns the current Unix timestamp
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}
