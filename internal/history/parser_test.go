package history

import (
	"strings"
	"testing"
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
