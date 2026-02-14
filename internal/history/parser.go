package history

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Parser reads and parses Fish shell history
type Parser struct {
	Path string // history file path
}

// NewParser returns a Parser with the default Fish history file path
func NewParser() *Parser {
	home, err := os.UserHomeDir()
	if err != nil {
		return &Parser{}
	}
	return &Parser{
		Path: filepath.Join(home, ".local", "share", "fish", "fish_history"),
	}
}

// Parse reads and parses the Fish shell history file
func (p *Parser) Parse() []Entry {
	if p.Path == "" {
		return []Entry{}
	}

	file, err := os.Open(p.Path)
	if err != nil {
		return []Entry{}
	}
	defer file.Close() //nolint:errcheck

	var entries []Entry
	var current *Entry
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if strings.HasPrefix(line, "- cmd: ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &Entry{
				Cmd:     strings.TrimPrefix(line, "- cmd: "),
				CmdLine: lineNum,
			}
		} else if current != nil {
			if strings.HasPrefix(line, "  when: ") {
				whenStr := strings.TrimPrefix(line, "  when: ")
				when, err := strconv.ParseInt(whenStr, 10, 64)
				if err == nil {
					current.When = when
				}
			} else if strings.HasPrefix(line, "    - ") {
				path := strings.TrimPrefix(line, "    - ")
				current.Paths = append(current.Paths, path)
			}
		}
	}

	if current != nil {
		entries = append(entries, *current)
	}

	// Reverse to show newest first
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}

	// Deduplicate commands - keep only the newest occurrence
	seen := make(map[string]bool)
	deduplicated := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if !seen[entry.Cmd] {
			seen[entry.Cmd] = true
			deduplicated = append(deduplicated, entry)
		}
	}

	return deduplicated
}
