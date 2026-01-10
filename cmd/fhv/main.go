package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// HistoryEntry represents a single fish history entry
type HistoryEntry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "preview" {
		if len(os.Args) < 3 {
			os.Exit(1)
		}
		lineNum, err := strconv.Atoi(os.Args[2])
		if err != nil {
			os.Exit(1)
		}
		showPreview(lineNum)
		return
	}

	entries := parseHistory()
	displayEntries(entries)
}

func getHistoryPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "fish", "fish_history")
}

func parseHistory() []HistoryEntry {
	histPath := getHistoryPath()
	file, err := os.Open(histPath)
	if err != nil {
		return []HistoryEntry{}
	}
	defer file.Close()

	var entries []HistoryEntry
	var current *HistoryEntry
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if strings.HasPrefix(line, "- cmd: ") {
			if current != nil {
				entries = append(entries, *current)
			}
			current = &HistoryEntry{
				Cmd:     strings.TrimPrefix(line, "- cmd: "),
				CmdLine: lineNum,
			}
		} else if current != nil {
			if strings.HasPrefix(line, "  when: ") {
				whenStr := strings.TrimPrefix(line, "  when: ")
				when, _ := strconv.ParseInt(whenStr, 10, 64)
				current.When = when
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

	return entries
}

func displayEntries(entries []HistoryEntry) {
	for i, entry := range entries {
		timeStr := formatTime(entry.When)
		dir := "~"
		if len(entry.Paths) > 0 {
			dir = entry.Paths[0]
			home, _ := os.UserHomeDir()
			dir = strings.Replace(dir, home, "~", 1)
		}

		// Truncate directory path to keep alignment
		formattedDir := truncatePath(dir, 30)

		// Format: [id] [time (fixed width)] [dir (fixed width)] [command]
		// %-12s: Left-align time, 12 chars
		// %30s:  Right-align dir, 30 chars
		fmt.Printf("%d\t%-12s\t%30s\t%s\n", i, timeStr, formattedDir, entry.Cmd)
	}
}

// truncatePath ensures the path fits within maxLen characters.
// If it's longer, it keeps the end of the path and prefixes with ".."
func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	// Keep the last (maxLen - 2) characters and add ".."
	// e.g. if maxLen=10, "/usr/local/bin" -> "..cal/bin"
	return ".." + path[len(path)-(maxLen-2):]
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "unknown"
	}

	t := time.Unix(timestamp, 0)
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	} else {
		return t.Format("2006-01-02")
	}
}

func showPreview(entryIndex int) {
	entries := parseHistory()
	if entryIndex < 0 || entryIndex >= len(entries) {
		return
	}

	// Only show context, no header details
	fmt.Println("┌─ Context ─────────────────────────────────────────────┐")

	start := entryIndex - 5
	if start < 0 {
		start = 0
	}
	end := entryIndex + 6
	if end > len(entries) {
		end = len(entries)
	}

	for i := start; i < end; i++ {
		e := entries[i]
		prefix := "  "
		if i == entryIndex {
			prefix = "→ "
		}
		// Also align context view slightly for better readability
		fmt.Printf("%s [%-10s] %s\n", prefix, formatTime(e.When), e.Cmd)
	}
	fmt.Println("└───────────────────────────────────────────────────────┘")
}
