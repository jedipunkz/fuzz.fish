package files

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jedipunkz/fuzz.fish/internal/ui"
)

func TestDirectoryListing_ListsEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	listing := Entry{Path: dir, IsDir: true}.DirectoryListing()
	if !strings.Contains(listing, ui.IconFile+" a.txt") {
		t.Errorf("listing = %q, want a.txt with file icon", listing)
	}
	if !strings.Contains(listing, ui.IconDir+" sub") {
		t.Errorf("listing = %q, want sub with dir icon", listing)
	}
}

func TestDirectoryListing_TruncatesToLimit(t *testing.T) {
	dir := t.TempDir()
	for i := 0; i < ui.MaxDirectoryEntries+3; i++ {
		if err := os.WriteFile(filepath.Join(dir, "f"+string(rune('a'+i))+".txt"), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}

	listing := Entry{Path: dir, IsDir: true}.DirectoryListing()
	n := strings.Count(listing, "\n")
	if n != ui.MaxDirectoryEntries+1 {
		t.Errorf("listing has %d lines, want %d entries plus the overflow notice", n, ui.MaxDirectoryEntries+1)
	}
	if !strings.Contains(listing, "and 3 more") {
		t.Errorf("listing = %q, want the remaining count notice", listing)
	}
}

func TestDirectoryListing_MissingDir(t *testing.T) {
	e := Entry{Path: filepath.Join(t.TempDir(), "nope"), IsDir: true}
	if got := e.DirectoryListing(); got != "" {
		t.Errorf("DirectoryListing = %q, want empty for a missing directory", got)
	}
}

func TestGeneratePreview_Directory(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := Entry{Path: dir, IsDir: true}.GeneratePreview(60, 20)
	if !strings.Contains(got, "Contents") {
		t.Errorf("preview = %q, want the Contents header", got)
	}
	if !strings.Contains(got, "a.txt") {
		t.Errorf("preview = %q, want the entry name", got)
	}
}

func TestGeneratePreview_EmptyDirectory(t *testing.T) {
	got := Entry{Path: t.TempDir(), IsDir: true}.GeneratePreview(60, 20)
	if !strings.Contains(got, "(empty)") {
		t.Errorf("preview = %q, want the empty notice", got)
	}
}

func TestGeneratePreview_TextFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a.txt")
	if err := os.WriteFile(path, []byte("line one\nline two\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := Entry{Path: path}.GeneratePreview(60, 20)
	if !strings.Contains(got, "Preview") {
		t.Errorf("preview = %q, want the Preview header", got)
	}
	if !strings.Contains(got, "line one") {
		t.Errorf("preview = %q, want the file content", got)
	}
}

func TestGeneratePreview_BinaryFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blob.bin")
	if err := os.WriteFile(path, []byte("ELF\x00\x00\x00payload"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := Entry{Path: path}.GeneratePreview(60, 20)
	if !strings.Contains(got, "(binary or empty file)") {
		t.Errorf("preview = %q, want the binary notice", got)
	}
}
