package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

// tabIndent is what a tab is replaced with in previews.
const tabIndent = "    "

// maxPreviewLineBytes caps how many bytes each previewed line may contribute to
// the read budget. Lines are truncated to the pane width anyway, so a minified
// bundle on one line must not pull the whole file into memory.
const maxPreviewLineBytes = 4096

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

// readHead reads at most maxLines lines from path. It stops early instead of
// loading the whole file: the preview pane only shows the first screenful, so
// reading (and later highlighting) a multi-megabyte file in full is wasted
// work that stalls the TUI. Reading is also capped by a byte budget so a file
// with one very long line cannot defeat the line limit.
//
// ok is false for unreadable and binary files, which both render as no preview.
func readHead(path string, maxLines int) (string, bool) {
	file, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer file.Close() //nolint:errcheck

	budget := int64(maxLines) * maxPreviewLineBytes
	reader := bufio.NewReaderSize(io.LimitReader(file, budget), BinaryDetectionBytes)

	head, err := reader.Peek(BinaryDetectionBytes)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", false
	}
	if IsBinary(head) {
		return "", false
	}

	var sb strings.Builder
	for i := 0; i < maxLines; i++ {
		line, err := reader.ReadString('\n')
		sb.WriteString(line)
		if err != nil {
			break
		}
	}
	return sb.String(), true
}

// GetFilePreview returns a preview of the file contents with syntax highlighting.
// maxWidth is the display width of the preview pane: lines wider than that are
// truncated so they do not wrap and grow the pane beyond its height.
//
// Only the first maxLines lines are read and highlighted. A file whose head is
// cut mid-string or mid-comment can colour differently from the full file, which
// is an acceptable trade for not tokenizing megabytes per cursor move.
func GetFilePreview(path string, maxLines, maxWidth int) string {
	if maxLines <= 0 {
		return ""
	}

	content, ok := readHead(path, maxLines)
	if !ok || content == "" {
		return ""
	}

	// Preview lines are rendered with a two-space indent.
	contentWidth := maxWidth - 2

	// Try syntax highlighting with chroma
	highlighted, err := HighlightCode(content, path)
	styled := err == nil && highlighted != ""
	if styled {
		content = highlighted
	}

	lines := strings.SplitN(content, "\n", maxLines+1)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var sb strings.Builder
	sb.Grow(len(lines) * (maxWidth + 4))
	for _, line := range lines {
		line = previewLine(line, contentWidth)
		if styled {
			// Chroma already emitted colours; a surrounding style would fight them.
			sb.WriteString("  ")
			sb.WriteString(line)
		} else {
			sb.WriteString(InactiveContextStyle.Render(fmt.Sprintf("  %s", line)))
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
