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
	Cmd      string
	When     int64
	Paths    []string
	CmdLine  int
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

		// Format: [time] [dir] command
		fmt.Printf("%d\t%s\t%s\t%s\n", i, timeStr, dir, entry.Cmd)
	}
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

	entry := entries[entryIndex]

	fmt.Println("╭─ Command Details ─────────────────────────────────────╮")
	fmt.Printf("│ Command: %s\n", entry.Cmd)
	fmt.Printf("│ Time:    %s (%s)\n",
		formatTime(entry.When),
		time.Unix(entry.When, 0).Format("2006-01-02 15:04:05"))
	if len(entry.Paths) > 0 {
		fmt.Printf("│ Dir:     %s\n", entry.Paths[0])
	}
	fmt.Println("╰───────────────────────────────────────────────────────╯")
	fmt.Println()

	// Show context: previous and next commands
	fmt.Println("┌─ Command Context ─────────────────────────────────────┐")

	start := entryIndex - 3
	if start < 0 {
		start = 0
	}
	end := entryIndex + 4
	if end > len(entries) {
		end = len(entries)
	}

	for i := start; i < end; i++ {
		e := entries[i]
		prefix := "  "
		if i == entryIndex {
			prefix = "→ "
		}
		fmt.Printf("%s [%s] %s\n", prefix, formatTime(e.When), e.Cmd)
	}
	fmt.Println("└───────────────────────────────────────────────────────┘")
}
