package ui

import (
	"fmt"
	"os"
	"strings"
)

// IsBinary checks if the given content appears to be binary
func IsBinary(content []byte) bool {
	// Simple binary detection: check for null bytes in first BinaryDetectionBytes
	checkSize := BinaryDetectionBytes
	if len(content) < checkSize {
		checkSize = len(content)
	}
	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

// GetFilePreview returns a preview of the file contents with syntax highlighting
func GetFilePreview(path string, maxLines int) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Check if binary
	if IsBinary(content) {
		return ""
	}

	// Try syntax highlighting with chroma
	highlighted, err := HighlightCode(string(content), path)
	if err == nil && highlighted != "" {
		// Limit lines using SplitN to avoid splitting the entire file
		lines := strings.SplitN(highlighted, "\n", maxLines+1)
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}

		var sb strings.Builder
		sb.Grow(len(lines) * (MaxLineLength + 4))
		for _, line := range lines {
			// Truncate long lines
			if len(line) > MaxLineLength {
				line = line[:MaxLineLength] + "..."
			}
			sb.WriteString("  ")
			sb.WriteString(line)
			sb.WriteByte('\n')
		}

		return sb.String()
	}

	// Fallback to plain text using SplitN to avoid splitting the entire file
	lines := strings.SplitN(string(content), "\n", maxLines+1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var sb strings.Builder
	sb.Grow(len(lines) * (MaxLineLength + 4))
	for _, line := range lines {
		// Truncate long lines
		if len(line) > MaxLineLength {
			line = line[:MaxLineLength] + "..."
		}
		sb.WriteString(InactiveContextStyle.Render(fmt.Sprintf("  %s", line)))
		sb.WriteByte('\n')
	}

	return sb.String()
}
