package app

import (
	"time"
	"unicode"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
)

// ScoringConfig holds configuration for the unified scoring algorithm
type ScoringConfig struct {
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

// DefaultScoringConfig returns the default scoring configuration
// inspired by fzy and fzf algorithms
func DefaultScoringConfig() ScoringConfig {
	return ScoringConfig{
		WordBoundaryBonus:  50.0,  // Bonus for matching after word boundaries
		ConsecutiveBonus:   30.0,  // Bonus for each consecutive match
		PrefixBonus:        100.0, // Bonus for matching at the start
		CamelCaseBonus:     40.0,  // Bonus for matching uppercase in CamelCase
		MaxRecencyBonus:    3000.0,
		CurrentBranchBonus: 500.0, // Bonus for current git branch
	}
}

// isWordBoundary checks if the character at prevIdx is a word boundary
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

// CalculateMatchBonus calculates bonus score based on match positions
// using fzy/fzf-inspired algorithm
func CalculateMatchBonus(text string, matchedIndexes []int, config ScoringConfig) float64 {
	if len(matchedIndexes) == 0 {
		return 0
	}

	var bonus float64

	// Prefix bonus: matching at the start is highly valuable
	if matchedIndexes[0] == 0 {
		bonus += config.PrefixBonus
	}

	prevIdx := -2 // Initialize to impossible value
	for _, idx := range matchedIndexes {
		// Word boundary bonus
		if isWordBoundary(text, idx) {
			bonus += config.WordBoundaryBonus
		}

		// CamelCase bonus
		if isCamelCaseBoundary(text, idx) {
			bonus += config.CamelCaseBonus
		}

		// Consecutive match bonus (affine gap penalty concept from fzy)
		if idx == prevIdx+1 {
			bonus += config.ConsecutiveBonus
		}

		prevIdx = idx
	}

	return bonus
}

// CalculateRecencyBonus calculates recency bonus using hyperbolic decay
// Formula: bonus = maxBonus / (1 + ageInHours)
// This strongly prioritizes recent items:
//
//	0h ago -> maxBonus, 1h -> maxBonus/2, 2h -> maxBonus/3, 24h -> maxBonus/25
func CalculateRecencyBonus(timestamp int64, now int64, maxBonus float64) float64 {
	if timestamp <= 0 {
		return 0
	}

	ageHours := float64(now-timestamp) / 3600.0
	if ageHours < 0 {
		ageHours = 0
	}

	return maxBonus / (1.0 + ageHours)
}

// CalculateItemScore calculates the total score for an item
// combining fuzzy match score, position bonuses, and recency
func CalculateItemScore(item Item, fuzzyScore int, matchedIndexes []int, mode SearchMode, config ScoringConfig, now int64) float64 {
	score := float64(fuzzyScore)

	// Add match position bonuses (fzy/fzf-inspired)
	score += CalculateMatchBonus(item.Text, matchedIndexes, config)

	// Mode-specific scoring
	switch mode {
	case ModeHistory:
		if entry, ok := item.Original.(history.Entry); ok && entry.When > 0 {
			score += CalculateRecencyBonus(entry.When, now, config.MaxRecencyBonus)
		}

	case ModeGitBranch:
		if branch, ok := item.Original.(git.Branch); ok {
			// Recency bonus based on commit time
			if branch.CommitTimestamp > 0 {
				score += CalculateRecencyBonus(branch.CommitTimestamp, now, config.MaxRecencyBonus)
			}
			// Current branch bonus
			if branch.IsCurrent {
				score += config.CurrentBranchBonus
			}
		}

	case ModeFiles:
		// Files use only match quality scoring (no recency)
		// Could add frecency tracking in the future
	}

	return score
}

// GetCurrentTimestamp returns the current Unix timestamp
func GetCurrentTimestamp() int64 {
	return time.Now().Unix()
}
