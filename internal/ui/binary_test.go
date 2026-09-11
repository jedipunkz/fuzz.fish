package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGetFilePreview_StopsAtMaxLines(t *testing.T) {
	path := writeTemp(t, "many.txt", strings.Repeat("line\n", 1000))

	got := strings.Count(GetFilePreview(path, 30, 60), "\n")
	if got != 30 {
		t.Fatalf("preview has %d lines, want 30", got)
	}
}

func TestGetFilePreview_ShorterThanMaxLines(t *testing.T) {
	path := writeTemp(t, "few.txt", "alpha\nbeta\n")

	preview := GetFilePreview(path, 30, 60)
	if !strings.Contains(preview, "alpha") || !strings.Contains(preview, "beta") {
		t.Fatalf("preview = %q, want both lines", preview)
	}
}

func TestGetFilePreview_LongLineStaysWithinBudget(t *testing.T) {
	// One line far longer than the per-line budget must not be read whole.
	path := writeTemp(t, "minified.txt", strings.Repeat("x", 10*maxPreviewLineBytes)+"\n")

	preview := GetFilePreview(path, 4, 60)
	if len(preview) > 4*maxPreviewLineBytes {
		t.Fatalf("preview is %d bytes, want at most %d", len(preview), 4*maxPreviewLineBytes)
	}
	if preview == "" {
		t.Fatal("preview is empty, want the truncated line")
	}
}

func TestGetFilePreview_Binary(t *testing.T) {
	path := writeTemp(t, "blob.bin", "ELF\x00\x00\x00payload")

	if got := GetFilePreview(path, 30, 60); got != "" {
		t.Fatalf("preview = %q, want empty for binary file", got)
	}
}

func TestGetFilePreview_Missing(t *testing.T) {
	if got := GetFilePreview(filepath.Join(t.TempDir(), "nope.txt"), 30, 60); got != "" {
		t.Fatalf("preview = %q, want empty for missing file", got)
	}
}

func TestGetFilePreview_NoRoom(t *testing.T) {
	path := writeTemp(t, "any.txt", "content\n")

	if got := GetFilePreview(path, 0, 60); got != "" {
		t.Fatalf("preview = %q, want empty when no lines fit", got)
	}
}
