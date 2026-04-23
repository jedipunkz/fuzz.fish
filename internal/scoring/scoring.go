package scoring

import (
	"math"
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
	// MatchWeight is the multiplier applied to match quality score.
	// Higher values make match quality the dominant ranking factor.
	MatchWeight float64
	// FrecencyWeight is the multiplier for frecency bonus (frequency × time decay).
	// Used in history mode. Frecency score = log1p(frequency) × timeMultiplier × FrecencyWeight.
	FrecencyWeight float64
	// MaxRecencyBonus is the max recency bonus for non-history modes (e.g. git branches).
	// Uses hyperbolic decay: bonus = MaxRecencyBonus / (1 + ageInHours).
	MaxRecencyBonus float64
	// Current branch bonus (git mode only)
	CurrentBranchBonus float64
}

// DefaultConfig returns the default scoring configuration.
// Match quality is weighted 10× to be the primary ranking factor.
// Frecency provides a secondary boost based on frequency × recency.
func DefaultConfig() Config {
	return Config{
		WordBoundaryBonus:  50.0,
		ConsecutiveBonus:   30.0,
		PrefixBonus:        100.0,
		CamelCaseBonus:     40.0,
		MatchWeight:        10.0,  // Match quality is primary
		FrecencyWeight:     50.0,  // log1p(freq) × multiplier × 50
		MaxRecencyBonus:    200.0, // For git branches (was 3000, reduced to same scale)
		CurrentBranchBonus: 500.0,
	}
}

// isWordBoundary checks if the character at idx is a word boundary
func isWordBoundary(text string, idx int) bool {
	if idx == 0 {
		return true // Start of string is always a boundary
	}
	if idx > len(text) {
		return false // Out-of-range index is not a valid boundary
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

// RecencyBonus calculates recency bonus using hyperbolic decay.
// Used for non-history modes (e.g. git branch commit timestamps).
// Formula: bonus = MaxRecencyBonus / (1 + ageInHours)
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

// FrecencyBonus calculates a frecency score combining frequency and recency.
// Used in history mode. Based on zoxide's step-wise time multiplier approach:
//
//	< 1 hour:  ×4  (very recent)
//	< 1 day:   ×2
//	< 1 week:  ×1
//	older:     ×0.5
//
// Score = log1p(frequency) × timeMultiplier × FrecencyWeight
// log1p smooths the frequency so that e.g. 100 uses ≠ 100× bonus over 1 use.
func (c Config) FrecencyBonus(timestamp int64, frequency int, now int64) float64 {
	if timestamp <= 0 || frequency <= 0 {
		return 0
	}

	ageHours := float64(now-timestamp) / 3600.0
	if ageHours < 0 {
		ageHours = 0
	}

	var multiplier float64
	switch {
	case ageHours < 1:
		multiplier = 4.0
	case ageHours < 24:
		multiplier = 2.0
	case ageHours < 168: // 1 week
		multiplier = 1.0
	default:
		multiplier = 0.5
	}

	return math.Log1p(float64(frequency)) * multiplier * c.FrecencyWeight
}

// ItemScore calculates the total score for an item.
//
// For history mode (frequency > 0):
//
//	score = (fuzzyScore + matchBonus) × MatchWeight + FrecencyBonus
//
// For other modes (frequency == 0, e.g. git branches, files):
//
//	score = (fuzzyScore + matchBonus) × MatchWeight + RecencyBonus
//
// This ensures match quality is the primary ranking factor (~10×),
// with frecency/recency as a secondary signal.
func (c Config) ItemScore(text string, fuzzyScore int, matchedIndexes []int, timestamp int64, frequency int, isCurrent bool, now int64) float64 {
	// Match quality is the primary signal, amplified by MatchWeight
	matchScore := (float64(fuzzyScore) + c.MatchBonus(text, matchedIndexes)) * c.MatchWeight

	var score float64
	if frequency > 0 {
		// History mode: frecency (frequency × time decay) as secondary signal
		score = matchScore + c.FrecencyBonus(timestamp, frequency, now)
	} else {
		// Non-history mode: simple recency as secondary signal
		score = matchScore + c.RecencyBonus(timestamp, now)
	}

	if isCurrent {
		score += c.CurrentBranchBonus
	}

	return score
}

// CurrentTimestamp returns the current Unix timestamp
func CurrentTimestamp() int64 {
	return time.Now().Unix()
}
