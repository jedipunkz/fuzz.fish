package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// Test that Parse() doesn't panic
	// It reads from the actual history file, so we can't make
	// strong assertions about the result
	p := NewParser()
	entries := p.Parse()

	// Just verify it returns a slice (could be empty if no history)
	if entries == nil {
		t.Error("Parse() returned nil, expected a slice")
	}
}

func TestParseEmptyPath(t *testing.T) {
	p := &Parser{Path: ""}
	entries := p.Parse()
	if entries == nil {
		t.Error("Parse() returned nil for empty path")
	}
	if len(entries) != 0 {
		t.Errorf("Parse() returned %d entries for empty path, want 0", len(entries))
	}
}

func TestParse_UsesCacheWhenMetadataMatches(t *testing.T) {
	dir := t.TempDir()
	historyPath := dir + "/fish_history"
	cacheDir := dir + "/cache"
	input := "- cmd: original\n  when: 1000\n"
	if err := os.WriteFile(historyPath, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	p := &Parser{Path: historyPath, CacheDir: cacheDir}
	if entries := p.Parse(); len(entries) != 1 || entries[0].Cmd != "original" {
		t.Fatalf("initial Parse() = %#v, want original entry", entries)
	}

	info, err := os.Stat(historyPath)
	if err != nil {
		t.Fatal(err)
	}
	fakeCache := cacheFile{
		Version: cacheVersion,
		Meta:    p.cacheMeta(info),
		Entries: []Entry{{Cmd: "from cache", When: 2000}},
	}
	file, err := os.Create(p.cachePath())
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(file).Encode(fakeCache); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	entries := p.Parse()
	if len(entries) != 1 || entries[0].Cmd != "from cache" {
		t.Fatalf("Parse() = %#v, want cached entry", entries)
	}
}

func TestParse_InvalidatesCacheWhenHistoryChanges(t *testing.T) {
	dir := t.TempDir()
	historyPath := dir + "/fish_history"
	cacheDir := dir + "/cache"
	if err := os.WriteFile(historyPath, []byte("- cmd: first\n  when: 1000\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := &Parser{Path: historyPath, CacheDir: cacheDir}
	if entries := p.Parse(); len(entries) != 1 || entries[0].Cmd != "first" {
		t.Fatalf("initial Parse() = %#v, want first entry", entries)
	}

	if err := os.WriteFile(historyPath, []byte("- cmd: second command\n  when: 2000\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	newTime := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(historyPath, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	entries := p.Parse()
	if len(entries) != 1 || entries[0].Cmd != "second command" {
		t.Fatalf("Parse() = %#v, want reparsed second entry", entries)
	}
}

func TestParse_WritesPrivateCacheFile(t *testing.T) {
	dir := t.TempDir()
	historyPath := dir + "/fish_history"
	cacheDir := dir + "/cache"
	if err := os.WriteFile(historyPath, []byte("- cmd: secret-ish command\n  when: 1000\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	p := &Parser{Path: historyPath, CacheDir: cacheDir}
	p.Parse()

	info, err := os.Stat(p.cachePath())
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("cache permissions = %o, want 600", got)
	}

	dirInfo, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("cache dir permissions = %o, want 700", got)
	}
}

func TestCachePath_UsesXDGCacheHome(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	p := &Parser{Path: "/tmp/fish_history"}
	want := filepath.Join(cacheHome, "fuzz.fish", "history-cache.json")
	if got := p.cachePath(); got != want {
		t.Fatalf("cachePath() = %q, want %q", got, want)
	}
}

func TestParseReader_Empty(t *testing.T) {
	entries := parseReader(strings.NewReader(""))
	if len(entries) != 0 {
		t.Errorf("parseReader with empty input returned %d entries, want 0", len(entries))
	}
}

func TestParseReader_SingleCommand(t *testing.T) {
	input := "- cmd: ls -la\n  when: 1700000000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if entries[0].Cmd != "ls -la" {
		t.Errorf("Cmd = %q, want %q", entries[0].Cmd, "ls -la")
	}
	if entries[0].When != 1700000000 {
		t.Errorf("When = %d, want %d", entries[0].When, 1700000000)
	}
}

func TestParseReader_MissingWhen(t *testing.T) {
	input := "- cmd: echo hello\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if entries[0].When != 0 {
		t.Errorf("When = %d, want 0 for missing when field", entries[0].When)
	}
}

func TestParseReader_InvalidWhen(t *testing.T) {
	input := "- cmd: ls\n  when: notanumber\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if entries[0].When != 0 {
		t.Errorf("When = %d, want 0 for invalid when value", entries[0].When)
	}
}

func TestParseReader_CommandWithPaths(t *testing.T) {
	input := "- cmd: make build\n  when: 1700000000\n  paths:\n    - /home/user/project\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if len(entries[0].Paths) != 1 {
		t.Fatalf("Paths length = %d, want 1", len(entries[0].Paths))
	}
	if entries[0].Paths[0] != "/home/user/project" {
		t.Errorf("Path = %q, want %q", entries[0].Paths[0], "/home/user/project")
	}
}

func TestParseReader_MultiplePaths(t *testing.T) {
	input := "- cmd: make\n  when: 1700000000\n  paths:\n    - /home/user/project\n    - /tmp/build\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if len(entries[0].Paths) != 2 {
		t.Fatalf("Paths length = %d, want 2", len(entries[0].Paths))
	}
}

func TestParseReader_MultipleCommands_NewestFirst(t *testing.T) {
	input := "- cmd: first command\n  when: 1000\n- cmd: second command\n  when: 2000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 2 {
		t.Fatalf("parseReader returned %d entries, want 2", len(entries))
	}
	// Should be reversed (newest first)
	if entries[0].Cmd != "second command" {
		t.Errorf("entries[0].Cmd = %q, want %q", entries[0].Cmd, "second command")
	}
	if entries[1].Cmd != "first command" {
		t.Errorf("entries[1].Cmd = %q, want %q", entries[1].Cmd, "first command")
	}
}

func TestParseReader_Deduplication(t *testing.T) {
	// Same command appears twice; the newer occurrence should be kept
	input := "- cmd: git status\n  when: 1000\n- cmd: git status\n  when: 2000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1 (deduplication)", len(entries))
	}
	// After reversal the newest (when=2000) appears first, so it is kept
	if entries[0].When != 2000 {
		t.Errorf("After dedup, When = %d, want 2000 (newest entry kept)", entries[0].When)
	}
}

func TestParseReader_SpecialCharacters(t *testing.T) {
	input := "- cmd: echo \"hello world\" | grep -E 'test.*pattern'\n  when: 1700000000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	expected := "echo \"hello world\" | grep -E 'test.*pattern'"
	if entries[0].Cmd != expected {
		t.Errorf("Cmd = %q, want %q", entries[0].Cmd, expected)
	}
}

func TestParseReader_LongCommand(t *testing.T) {
	longCmd := "git commit -m '" + strings.Repeat("a", 1000) + "'"
	input := "- cmd: " + longCmd + "\n  when: 1700000000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 1 {
		t.Fatalf("parseReader returned %d entries, want 1", len(entries))
	}
	if entries[0].Cmd != longCmd {
		t.Errorf("Long command not parsed correctly: got %d chars, want %d", len(entries[0].Cmd), len(longCmd))
	}
}

func TestParseReader_BlankLines(t *testing.T) {
	input := "\n- cmd: ls\n  when: 1000\n\n- cmd: pwd\n  when: 2000\n\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 2 {
		t.Fatalf("parseReader returned %d entries, want 2", len(entries))
	}
}

func TestParseReader_CmdLineTracking(t *testing.T) {
	input := "- cmd: first\n  when: 1000\n- cmd: second\n  when: 2000\n"
	entries := parseReader(strings.NewReader(input))
	if len(entries) != 2 {
		t.Fatalf("parseReader returned %d entries, want 2", len(entries))
	}
	// After reversal: entries[0] = second (line 3), entries[1] = first (line 1)
	if entries[0].CmdLine != 3 {
		t.Errorf("entries[0].CmdLine = %d, want 3", entries[0].CmdLine)
	}
	if entries[1].CmdLine != 1 {
		t.Errorf("entries[1].CmdLine = %d, want 1", entries[1].CmdLine)
	}
}
