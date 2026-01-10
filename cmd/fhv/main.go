package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/koki-develop/go-fzf"
)

type HistoryEntry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
}

func main() {
	// Parse history
	entries := parseHistory()
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		os.Exit(1)
	}

	// Create display items
	items := make([]string, len(entries))
	for i, e := range entries {
		items[i] = formatEntry(e)
	}

	// Configure go-fzf with styling
	f, err := fzf.New(
		fzf.WithPrompt("🔍 "),
		fzf.WithInputPlaceholder("Type to search history..."),
		fzf.WithLimit(1), // Single selection only
		fzf.WithInputPosition(fzf.InputPositionTop),
		fzf.WithStyles(
			fzf.WithStylePrompt(fzf.Style{
				ForegroundColor: "205", // Pink
				Bold:            true,
			}),
			fzf.WithStyleCursorLine(fzf.Style{
				BackgroundColor: "57",  // Purple background for selected
				ForegroundColor: "229", // Light yellow text
			}),
			fzf.WithStyleMatches(fzf.Style{
				ForegroundColor: "205", // Pink for matched chars
				Bold:            true,
			}),
			fzf.WithStyleInputText(fzf.Style{
				ForegroundColor: "255", // White
				Bold:            true,
			}),
			fzf.WithStyleInputPlaceholder(fzf.Style{
				ForegroundColor: "240", // Gray
			}),
		),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing fzf: %v\n", err)
		os.Exit(1)
	}

	// Run fuzzy finder with preview
	idxs, err := f.Find(
		items,
		func(i int) string { return items[i] },
		fzf.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(entries) {
				return "Invalid selection"
			}
			return generatePreview(entries[i], entries, i, w, h)
		}),
	)
	if err != nil {
		// User cancelled or error occurred
		os.Exit(0)
	}

	// Output selected command to stdout
	if len(idxs) > 0 && idxs[0] < len(entries) {
		fmt.Print(entries[idxs[0]].Cmd)
	}
}

func formatEntry(e HistoryEntry) string {
	timeStr := formatTime(e.When)
	// Format: "5m ago | git status"
	return fmt.Sprintf("%-12s | %s", timeStr, e.Cmd)
}

func generatePreview(entry HistoryEntry, all []HistoryEntry, idx, width, height int) string {
	var sb strings.Builder

	// Header
	sb.WriteString("COMMAND\n")
	sb.WriteString(entry.Cmd)
	sb.WriteString("\n\n")

	// Metadata
	sb.WriteString(fmt.Sprintf("Time: %s\n", formatTime(entry.When)))
	if len(entry.Paths) > 0 {
		sb.WriteString(fmt.Sprintf("Dir:  %s\n", formatDir(entry.Paths[0])))
	}
	sb.WriteString("\n")

	// Context (commands before/after)
	sb.WriteString("CONTEXT\n")
	start := idx - 3
	if start < 0 {
		start = 0
	}
	end := idx + 4
	if end > len(all) {
		end = len(all)
	}

	for i := start; i < end; i++ {
		e := all[i]
		cursor := "  "
		if i == idx {
			cursor = "→ "
		}

		cmd := e.Cmd
		// Truncate if needed
		maxWidth := width - 5
		if maxWidth > 0 && len(cmd) > maxWidth {
			cmd = cmd[:maxWidth-3] + "..."
		}

		sb.WriteString(fmt.Sprintf("%s%s\n", cursor, cmd))
	}

	return sb.String()
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

func formatDir(path string) string {
	home, _ := os.UserHomeDir()
	path = strings.Replace(path, home, "~", 1)
	return path
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
