package git

import (
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// This test depends on the environment
	// In a git repository, it should return true
	// We can't make assumptions about the test environment,
	// so we just check that the function doesn't panic
	result := IsGitRepo()
	_ = result // Just ensure it runs without error
}

func TestGetCurrentBranch(t *testing.T) {
	// This test depends on the environment
	// We just check that the function doesn't panic
	result := getCurrentBranch()
	_ = result // Just ensure it runs without error
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid date",
			input:    "2024-01-15 10:30:45 +0900",
			expected: "2024-01-15 10:30",
		},
		{
			name:     "invalid date returns original",
			input:    "invalid",
			expected: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDate(tt.input)
			if result != tt.expected {
				t.Errorf("formatDate(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
