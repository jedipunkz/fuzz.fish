package utils

import (
	"testing"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "bytes",
			input:    500,
			expected: "500 bytes",
		},
		{
			name:     "kilobytes",
			input:    1024,
			expected: "1.00 KB",
		},
		{
			name:     "megabytes",
			input:    1024 * 1024,
			expected: "1.00 MB",
		},
		{
			name:     "gigabytes",
			input:    1024 * 1024 * 1024,
			expected: "1.00 GB",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.input)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
