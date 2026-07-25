package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// tabIndent is what a tab is replaced with in previews.
const tabIndent = "    "

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

// previewLine expands tabs and cuts a line to width display cells, leaving ANSI
// sequences intact. Tabs must be expanded first: the terminal renders them as
// several cells, so measuring them as one would let the line wrap anyway.
func previewLine(line string, width int) string {
	line = strings.ReplaceAll(line, "\t", tabIndent)
	if width <= 0 {
		return line
	}
	return ansi.Truncate(line, width, "…")
}

// GetFilePreview returns a preview of the file contents with syntax highlighting.
// maxWidth is the display width of the preview pane: lines wider than that are
// truncated so they do not wrap and grow the pane beyond its height.
func GetFilePreview(path string, maxLines, maxWidth int) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Preview lines are rendered with a two-space indent.
	contentWidth := maxWidth - 2

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
		sb.Grow(len(lines) * (maxWidth + 4))
		for _, line := range lines {
			sb.WriteString("  ")
			sb.WriteString(previewLine(line, contentWidth))
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
	sb.Grow(len(lines) * (maxWidth + 4))
	for _, line := range lines {
		line = previewLine(line, contentWidth)
		sb.WriteString(InactiveContextStyle.Render(fmt.Sprintf("  %s", line)))
		sb.WriteByte('\n')
	}

	return sb.String()
}
