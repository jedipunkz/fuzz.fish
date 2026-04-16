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

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.WordBoundaryBonus <= 0 {
		t.Errorf("WordBoundaryBonus should be positive, got %v", config.WordBoundaryBonus)
	}
	if config.ConsecutiveBonus <= 0 {
		t.Errorf("ConsecutiveBonus should be positive, got %v", config.ConsecutiveBonus)
	}
	if config.PrefixBonus <= 0 {
		t.Errorf("PrefixBonus should be positive, got %v", config.PrefixBonus)
	}
	if config.MaxRecencyBonus <= 0 {
		t.Errorf("MaxRecencyBonus should be positive, got %v", config.MaxRecencyBonus)
	}
	if config.CurrentBranchBonus <= 0 {
		t.Errorf("CurrentBranchBonus should be positive, got %v", config.CurrentBranchBonus)
	}
}

func TestMatchBonus_SingleMiddleMatch(t *testing.T) {
	config := DefaultConfig()
	// Match in the middle of a word - no special bonus
	bonus := config.MatchBonus("hello", []int{2})
	if bonus != 0 {
		t.Errorf("MatchBonus for middle match = %v, want 0", bonus)
	}
}

func TestMatchBonus_WordBoundaryAfterSlash(t *testing.T) {
	config := DefaultConfig()
	bonus := config.MatchBonus("hello/world", []int{6})
	if bonus < config.WordBoundaryBonus {
		t.Errorf("MatchBonus for word boundary = %v, want >= %v", bonus, config.WordBoundaryBonus)
	}
}

func TestIsWordBoundary_OutOfRange(t *testing.T) {
	// idx > len(text) should return false, not panic
	result := isWordBoundary("hello", 100)
	if result {
		t.Errorf("isWordBoundary with out-of-range idx = true, want false")
	}
}

func TestIsCamelCaseBoundary_OutOfRange(t *testing.T) {
	// idx >= len(text) should return false
	result := isCamelCaseBoundary("hello", 10)
	if result {
		t.Errorf("isCamelCaseBoundary with out-of-range idx = true, want false")
	}
}

func TestRecencyBonus_ZeroTimestamp(t *testing.T) {
	config := DefaultConfig()
	bonus := config.RecencyBonus(0, 1000000)
	if bonus != 0 {
		t.Errorf("RecencyBonus(0, ...) = %v, want 0", bonus)
	}
}

func TestRecencyBonus_NegativeTimestamp(t *testing.T) {
	config := DefaultConfig()
	bonus := config.RecencyBonus(-1, 1000000)
	if bonus != 0 {
		t.Errorf("RecencyBonus(-1, ...) = %v, want 0", bonus)
	}
}

func TestItemScore_NoRecency(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	// timestamp=0 means no recency bonus
	score := config.ItemScore("test", 50, []int{0}, 0, false, now)
	// PrefixBonus + WordBoundaryBonus (idx=0 is always boundary) + base score
	expected := 50.0 + config.PrefixBonus + config.WordBoundaryBonus
	if score != expected {
		t.Errorf("ItemScore with no recency = %v, want %v", score, expected)
	}
}

func TestItemScore_CurrentBranchBonus(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	scoreWith := config.ItemScore("main", 50, []int{}, 0, true, now)
	scoreWithout := config.ItemScore("main", 50, []int{}, 0, false, now)
	if scoreWith-scoreWithout != config.CurrentBranchBonus {
		t.Errorf("CurrentBranchBonus diff = %v, want %v", scoreWith-scoreWithout, config.CurrentBranchBonus)
	}
}

func TestItemScore_ZeroFuzzyScore(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	score := config.ItemScore("test", 0, []int{}, 0, false, now)
	// No bonuses at all: score should be 0
	if score != 0 {
		t.Errorf("ItemScore with zero fuzzy and no bonuses = %v, want 0", score)
	}
}

func TestCurrentTimestamp(t *testing.T) {
	ts := CurrentTimestamp()
	if ts <= 0 {
		t.Errorf("CurrentTimestamp() = %d, want positive", ts)
	}
}
