package scoring

import (
	"math"
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
		{"Just now", now, config.MaxRecencyBonus * 0.95, config.MaxRecencyBonus * 1.05},
		{"1 hour ago", now - 3600, config.MaxRecencyBonus/2 - 10, config.MaxRecencyBonus/2 + 10},
		{"24 hours ago", now - 86400, 0, config.MaxRecencyBonus / 10},
		{"Zero timestamp", 0, 0, 0},
		{"Future timestamp", now + 3600, config.MaxRecencyBonus * 0.95, config.MaxRecencyBonus * 1.05},
	}

	for _, tt := range tests {
		bonus := config.RecencyBonus(tt.timestamp, now)
		if bonus < tt.minBonus || bonus > tt.maxBonus {
			t.Errorf("%s: RecencyBonus() = %v, want in range [%v, %v]", tt.name, bonus, tt.minBonus, tt.maxBonus)
		}
	}
}

func TestFrecencyBonus(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	tests := []struct {
		name      string
		timestamp int64
		frequency int
		minBonus  float64
		maxBonus  float64
	}{
		// Zero/invalid cases
		{"Zero timestamp", 0, 5, 0, 0},
		{"Zero frequency", now - 1800, 0, 0, 0},
		// Multiplier tiers: log1p(freq) * mult * FrecencyWeight
		{"freq=1, <1h ago (×4)", now - 1800, 1, math.Log1p(1) * 4 * config.FrecencyWeight * 0.9, math.Log1p(1) * 4 * config.FrecencyWeight * 1.1},
		{"freq=1, <1day ago (×2)", now - 7200, 1, math.Log1p(1) * 2 * config.FrecencyWeight * 0.9, math.Log1p(1) * 2 * config.FrecencyWeight * 1.1},
		{"freq=1, <1week ago (×1)", now - 86400*2, 1, math.Log1p(1) * 1 * config.FrecencyWeight * 0.9, math.Log1p(1) * 1 * config.FrecencyWeight * 1.1},
		{"freq=1, old (×0.5)", now - 86400*10, 1, math.Log1p(1) * 0.5 * config.FrecencyWeight * 0.9, math.Log1p(1) * 0.5 * config.FrecencyWeight * 1.1},
		// Frequency scaling via log1p
		{"freq=10, <1h ago", now - 1800, 10, math.Log1p(10) * 4 * config.FrecencyWeight * 0.9, math.Log1p(10) * 4 * config.FrecencyWeight * 1.1},
		{"freq=100, <1h ago", now - 1800, 100, math.Log1p(100) * 4 * config.FrecencyWeight * 0.9, math.Log1p(100) * 4 * config.FrecencyWeight * 1.1},
	}

	for _, tt := range tests {
		bonus := config.FrecencyBonus(tt.timestamp, tt.frequency, now)
		if bonus < tt.minBonus || bonus > tt.maxBonus {
			t.Errorf("%s: FrecencyBonus() = %v, want in range [%v, %v]", tt.name, bonus, tt.minBonus, tt.maxBonus)
		}
	}
}

// TestFrecencyFrequencyOrdering verifies that a frequently used command
// ranks higher than a rarely used one when both match equally well.
func TestFrecencyFrequencyOrdering(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	oneDayAgo := now - 86400

	// Same command, same recency, different frequency
	rareScore := config.FrecencyBonus(oneDayAgo, 1, now)
	frequentScore := config.FrecencyBonus(oneDayAgo, 50, now)

	if frequentScore <= rareScore {
		t.Errorf("frequent command frecency (%v) should > rare command frecency (%v)", frequentScore, rareScore)
	}
}

// TestFrecencyVsRecency verifies that a frequently used older command
// can outrank a rarely used recent command.
func TestFrecencyVsRecency(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// Command used 50 times, 2 days ago (×1 multiplier)
	frequentOld := config.FrecencyBonus(now-86400*2, 50, now)
	// Command used once, 30 minutes ago (×4 multiplier)
	rareRecent := config.FrecencyBonus(now-1800, 1, now)

	if frequentOld <= rareRecent {
		t.Errorf("frequent+old frecency (%v) should > rare+recent frecency (%v)", frequentOld, rareRecent)
	}
}

func TestItemScoreHistory(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// History mode: frequency > 0 triggers frecency path
	score := config.ItemScore("git commit -m 'test'", 100, []int{0, 1, 2}, now-3600, 5, false, now)

	// matchScore = (100 + PrefixBonus + ConsecutiveBonus*2) * MatchWeight
	matchQuality := (100.0 + config.PrefixBonus + config.ConsecutiveBonus*2) * config.MatchWeight
	// frecency = log1p(5) * 2.0 * FrecencyWeight (1h ago → ×2 multiplier)
	frecency := math.Log1p(5) * 2.0 * config.FrecencyWeight

	minExpected := matchQuality + frecency*0.9
	if score < minExpected {
		t.Errorf("History item score = %v, want >= %v (matchScore=%v frecency=%v)", score, minExpected, matchQuality, frecency)
	}
}

// TestItemScoreMatchQualityDominates verifies that a high-quality match
// outranks a low-quality match even when the latter is more recent/frequent.
func TestItemScoreMatchQualityDominates(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// Poor match, very recent and frequent
	poorMatchScore := config.ItemScore("git stash pop", 10, []int{0}, now-60, 100, false, now)
	// Good match, older and rare
	goodMatchScore := config.ItemScore("git commit", 100, []int{0, 1, 2}, now-86400*3, 2, false, now)

	if goodMatchScore <= poorMatchScore {
		t.Errorf("good match score (%v) should > poor match score (%v) — match quality should dominate", goodMatchScore, poorMatchScore)
	}
}

func TestItemScoreGitBranch(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// Git mode: frequency=0 triggers RecencyBonus path
	currentScore := config.ItemScore("main", 100, []int{0}, now-3600, 0, true, now)
	otherScore := config.ItemScore("feature/test", 100, []int{0}, now-3600, 0, false, now)

	if currentScore <= otherScore {
		t.Errorf("Current branch score (%v) should be > other branch score (%v)", currentScore, otherScore)
	}

	diff := currentScore - otherScore
	if diff < config.CurrentBranchBonus-50 || diff > config.CurrentBranchBonus+50 {
		t.Errorf("Score difference = %v, want ~%v (current branch bonus)", diff, config.CurrentBranchBonus)
	}
}

func TestItemScoreFiles(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)

	// Files mode: frequency=0, timestamp=0 → only match quality
	score := config.ItemScore("src/components/Button.tsx", 100, []int{4, 5, 6}, 0, 0, false, now)

	// matchScore = (100 + WordBoundaryBonus + ConsecutiveBonus*2) * MatchWeight
	minExpected := (100.0 + config.WordBoundaryBonus + config.ConsecutiveBonus*2) * config.MatchWeight
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
	if config.MatchWeight <= 0 {
		t.Errorf("MatchWeight should be positive, got %v", config.MatchWeight)
	}
	if config.FrecencyWeight <= 0 {
		t.Errorf("FrecencyWeight should be positive, got %v", config.FrecencyWeight)
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
	result := isWordBoundary("hello", 100)
	if result {
		t.Errorf("isWordBoundary with out-of-range idx = true, want false")
	}
}

func TestIsCamelCaseBoundary_OutOfRange(t *testing.T) {
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
	// timestamp=0, frequency=0 → no recency/frecency bonus
	score := config.ItemScore("test", 50, []int{0}, 0, 0, false, now)
	// matchScore = (50 + PrefixBonus + WordBoundaryBonus) * MatchWeight
	expected := (50.0 + config.PrefixBonus + config.WordBoundaryBonus) * config.MatchWeight
	if score != expected {
		t.Errorf("ItemScore with no recency = %v, want %v", score, expected)
	}
}

func TestItemScore_CurrentBranchBonus(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	scoreWith := config.ItemScore("main", 50, []int{}, 0, 0, true, now)
	scoreWithout := config.ItemScore("main", 50, []int{}, 0, 0, false, now)
	if scoreWith-scoreWithout != config.CurrentBranchBonus {
		t.Errorf("CurrentBranchBonus diff = %v, want %v", scoreWith-scoreWithout, config.CurrentBranchBonus)
	}
}

func TestItemScore_ZeroFuzzyScore(t *testing.T) {
	config := DefaultConfig()
	now := int64(1000000)
	// fuzzyScore=0, no matched indexes, no timestamp → score should be 0
	score := config.ItemScore("test", 0, []int{}, 0, 0, false, now)
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
