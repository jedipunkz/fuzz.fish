package history

import (
	"testing"
)

func TestParse(t *testing.T) {
	// Test that Parse() doesn't panic
	// It reads from the actual history file, so we can't make
	// strong assertions about the result
	entries := Parse()

	// Just verify it returns a slice (could be empty if no history)
	if entries == nil {
		t.Error("Parse() returned nil, expected a slice")
	}
}
