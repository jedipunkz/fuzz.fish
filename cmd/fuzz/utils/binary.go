package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
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

// GetFilePreview returns a preview of the file contents
func GetFilePreview(path string, maxLines int) string {
	// Try to use bat for syntax highlighting
	if hasBat() {
		content := runBat(path, maxLines)
		if content != "" {
			return content
		}
	}

	// Fallback to plain cat
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Check if binary
	if IsBinary(content) {
		return ""
	}

	// Limit lines
	lines := strings.Split(string(content), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var sb strings.Builder
	for _, line := range lines {
		// Truncate long lines
		if len(line) > MaxLineLength {
			line = line[:MaxLineLength] + "..."
		}
		sb.WriteString(ui.InactiveContextStyle.Render(fmt.Sprintf("  %s", line)) + "\n")
	}

	return sb.String()
}

func hasBat() bool {
	_, err := exec.LookPath("bat")
	return err == nil
}

func runBat(path string, maxLines int) string {
	cmd := exec.Command("bat", "--color=always", "--style=numbers", fmt.Sprintf("--line-range=:%d", maxLines), path)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Add indentation
	lines := strings.Split(string(output), "\n")
	var sb strings.Builder
	for _, line := range lines {
		if line != "" {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}
