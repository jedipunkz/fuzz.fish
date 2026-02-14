package scoring

import (
	"testing"
)

func TestIsWordBoundary(t *testing.T) {
	tests := []struct {
		text     string
		idx      int
		expected bool
	}{
		{"hello", 0, true},           // Start of string
		{"hello/world", 6, true},     // After /
		{"hello-world", 6, true},     // After -
		{"hello_world", 6, true},     // After _
		{"hello.world", 6, true},     // After .
		{"hello world", 6, true},     // After space
		{"helloworld", 5, false},     // Middle of word
		{"feature/branch", 8, true},  // After / in branch name
		{"fix-bug", 4, true},         // After - in branch name
	}

	for _, tt := range tests {
		result := isWordBoundary(tt.text, tt.idx)
		if result != tt.expected {
			t.Errorf("isWordBoundary(%q, %d) = %v, want %v", tt.text, tt.idx, result, tt.expected)
		}
	}
}

func TestIsCamelCaseBoundary(t *testing.T) {
	tests := []struct {
		text     string
		idx      int
		expected bool
	}{
		{"camelCase", 5, true},       // C in camelCase
		{"HelloWorld", 5, true},      // W in HelloWorld
		{"helloworld", 5, false},     // All lowercase
		{"HELLOWORLD", 5, false},     // All uppercase
		{"hello", 0, false},          // Start of string
	}

	for _, tt := range tests {
		result := isCamelCaseBoundary(tt.text, tt.idx)
		if result != tt.expected {
			t.Errorf("isCamelCaseBoundary(%q, %d) = %v, want %v", tt.text, tt.idx, result, tt.expected)
		}
	}
}

func TestMatchBonus(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name           string
		text           string
		matchedIndexes []int
		minBonus       float64
		description    string
	}{
		{
			name:           "Prefix match",
			text:           "feature/branch",
			matchedIndexes: []int{0, 1, 2},
			minBonus:       config.PrefixBonus,
			description:    "Should have prefix bonus",
		},
		{
			name:           "Word boundary match after slash",
			text:           "feature/branch",
			matchedIndexes: []int{8, 9, 10},
			minBonus:       config.WordBoundaryBonus,
			description:    "Should have word boundary bonus",
		},
		{
			name:           "Consecutive matches",
			text:           "hello",
			matchedIndexes: []int{0, 1, 2},
			minBonus:       config.PrefixBonus + config.ConsecutiveBonus*2,
			description:    "Should have prefix + consecutive bonuses",
		},
		{
			name:           "Empty matches",
			text:           "hello",
			matchedIndexes: []int{},
			minBonus:       0,
			description:    "No bonus for empty matches",
		},
	}

	for _, tt := range tests {
		bonus := config.MatchBonus(tt.text, tt.matchedIndexes)
		if bonus < tt.minBonus {
			t.Errorf("%s: MatchBonus() = %v, want >= %v (%s)", tt.name, bonus, tt.minBonus, tt.description)
		}
	}
}

func TestRecencyBonus(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	tests := []struct {
		name      string
		timestamp int64
		minBonus  float64
		maxBonus  float64
	}{
		{"Just now", now, 2900, 3100},
		{"1 hour ago", now - 3600, 1400, 1600},
		{"2 hours ago", now - 7200, 900, 1100},
		{"24 hours ago", now - 86400, 100, 150},
		{"Zero timestamp", 0, 0, 0},
		{"Future timestamp", now + 3600, 2900, 3100}, // Treated as "just now"
	}

	for _, tt := range tests {
		bonus := config.RecencyBonus(tt.timestamp, now)
		if bonus < tt.minBonus || bonus > tt.maxBonus {
			t.Errorf("%s: RecencyBonus() = %v, want in range [%v, %v]", tt.name, bonus, tt.minBonus, tt.maxBonus)
		}
	}
}

func TestItemScoreHistory(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	score := config.ItemScore("git commit -m 'test'", 100, []int{0, 1, 2}, now-3600, false, now)

	// Should have base score + prefix bonus + consecutive bonus + recency bonus
	minExpected := 100.0 + config.PrefixBonus + config.ConsecutiveBonus*2 + 1400 // ~1500 recency
	if score < minExpected {
		t.Errorf("History item score = %v, want >= %v", score, minExpected)
	}
}

func TestItemScoreGitBranch(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// Test current branch bonus
	currentScore := config.ItemScore("main", 100, []int{0}, now-3600, true, now)
	otherScore := config.ItemScore("feature/test", 100, []int{0}, now-3600, false, now)

	// Current branch should have higher score due to current branch bonus
	if currentScore <= otherScore {
		t.Errorf("Current branch score (%v) should be > other branch score (%v)", currentScore, otherScore)
	}

	// Difference should be approximately the current branch bonus
	diff := currentScore - otherScore
	if diff < config.CurrentBranchBonus-50 || diff > config.CurrentBranchBonus+50 {
		t.Errorf("Score difference = %v, want ~%v (current branch bonus)", diff, config.CurrentBranchBonus)
	}
}

func TestItemScoreFiles(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	score := config.ItemScore("src/components/Button.tsx", 100, []int{4, 5, 6}, 0, false, now)

	// Files should only have match quality bonuses, no recency
	// Match at position 4 (after src/) should have word boundary bonus
	minExpected := 100.0 + config.WordBoundaryBonus + config.ConsecutiveBonus*2
	if score < minExpected {
		t.Errorf("File item score = %v, want >= %v", score, minExpected)
	}
}
