package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/koki-develop/go-fzf"
)

type HistoryEntry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
}

const (
	// Tokyo Night color palette
	colorCyan           = "#7dcfff"
	colorPurple         = "#bb9af7"
	colorForeground     = "#c0caf5"
	colorYellow         = "#e0af68"
	colorOrange         = "#ff9e64"
	colorComment        = "#565f89"
	colorBlue           = "#7aa2f7"
	colorSelectionBg    = "#543970" // Darker muted purple for selection
)

var (
	// Styles for preview window
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPurple))

	contentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorForeground))

	contextHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorYellow)).
				Bold(true)

	activeContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorOrange)).
				Bold(true)

	inactiveContextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(colorComment))
)

func main() {
	// Parse history
	entries := parseHistory()
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		os.Exit(1)
	}

	// Use go-fzf with Tokyo Night theme
	f, err := fzf.New(
		fzf.WithStyles(
			fzf.WithStylePrompt(fzf.Style{ForegroundColor: colorBlue}),
			fzf.WithStyleInputText(fzf.Style{ForegroundColor: colorForeground}),
			fzf.WithStyleCursor(fzf.Style{ForegroundColor: colorBlue}),
			// Selection background set to brighter purple
			fzf.WithStyleCursorLine(fzf.Style{ForegroundColor: colorForeground, BackgroundColor: colorSelectionBg, Bold: true}),
			fzf.WithStyleMatches(fzf.Style{ForegroundColor: colorOrange}),
			fzf.WithStyleSelectedPrefix(fzf.Style{ForegroundColor: colorBlue}),
			fzf.WithStyleUnselectedPrefix(fzf.Style{ForegroundColor: colorComment}),
		),
		fzf.WithInputPosition(fzf.InputPositionBottom),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize fzf: %v\n", err)
		os.Exit(1)
	}

	idxs, err := f.Find(
		entries,
		func(i int) string {
			return formatEntry(entries[i])
		},
		fzf.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(entries) {
				return "No selection"
			}
			return generatePreview(entries[i], entries, i, w, h)
		}),
	)

	if err != nil {
		if errors.Is(err, fzf.ErrAbort) {
			// User cancelled
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "fzf error: %v\n", err)
		os.Exit(1)
	}

	// Output selected command
	if len(idxs) > 0 {
		fmt.Print(entries[idxs[0]].Cmd)
	}
}

func formatEntry(e HistoryEntry) string {
	var timeStr string
	if e.When == 0 {
		timeStr = "Unknown                   "
	} else {
		t := time.Unix(e.When, 0)
		// Format with day of week: "Fri 2026-01-10 15:30:45"
		timeStr = t.Format("Mon 2006-01-02 15:04:05")
	}
	// Replace escaped newlines and physical newlines with spaces to prevent display corruption
	cmd := strings.ReplaceAll(e.Cmd, "\\n", " ")
	cmd = strings.ReplaceAll(cmd, "\n", " ")
	// Format: "Fri 2026-01-10 15:30:45 | git status"
	return fmt.Sprintf("%-23s | %s", timeStr, cmd)
}

func generatePreview(entry HistoryEntry, all []HistoryEntry, idx, width, height int) string {
	var sb strings.Builder

	// Header
	// sb.WriteString(headerStyle.Render("Command") + "\n")
	// Wrap command to fit width
	// sb.WriteString(contentStyle.Copy().Width(width).Render(entry.Cmd))
	// sb.WriteString("\n\n")

	// Metadata
	// Time
	sb.WriteString(labelStyle.Render("Time") + "\n")
	sb.WriteString(contentStyle.Render(formatTime(entry.When)))
	sb.WriteString("\n")
	sb.WriteString(contentStyle.Faint(true).Render(formatRelativeTime(entry.When)))
	sb.WriteString("\n\n")

	// Dir
	if len(entry.Paths) > 0 {
		sb.WriteString(labelStyle.Render("Directory") + "\n")
		sb.WriteString(contentStyle.Render(formatDir(entry.Paths[0])))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Context (commands before/after)
	sb.WriteString(contextHeaderStyle.Render("Context") + "\n")
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
		cmd := e.Cmd

		if i == idx {
			cursor := "→ "
			// Wrap active context line
			line := activeContextStyle.Width(width).Render(cursor + cmd)
			sb.WriteString(line + "\n")
		} else {
			cursor := "  "
			// Truncate inactive lines to keep context compact
			maxWidth := width - lipgloss.Width(cursor)
			if maxWidth > 0 && len(cmd) > maxWidth {
				cmd = cmd[:maxWidth-3] + "..."
			}
			line := inactiveContextStyle.Render(cursor + cmd)
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}

func getHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share", "fish", "fish_history")
}

func parseHistory() []HistoryEntry {
	histPath := getHistoryPath()
	file, err := os.Open(histPath)
	if err != nil {
		return []HistoryEntry{}
	}
	defer file.Close() //nolint:errcheck

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

	return entries
}

func formatDir(path string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		path = strings.Replace(path, home, "~", 1)
	}
	return path
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "0000-00-00 00:00:00"
	}

	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

func formatRelativeTime(timestamp int64) string {
	if timestamp == 0 {
		return "unknown"
	}

	now := time.Now()
	t := time.Unix(timestamp, 0)
	diff := now.Sub(t)

	seconds := int(diff.Seconds())
	minutes := int(diff.Minutes())
	hours := int(diff.Hours())
	days := int(diff.Hours() / 24)
	weeks := days / 7
	months := days / 30
	years := days / 365

	switch {
	case seconds < 60:
		if seconds == 1 {
			return "1 second ago"
		}
		return fmt.Sprintf("%d seconds ago", seconds)
	case minutes < 60:
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	case hours < 24:
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	case days < 7:
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	case weeks < 4:
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	case months < 12:
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	default:
		if years == 1 {
			return "1 year ago"
		}
		return fmt.Sprintf("%d years ago", years)
	}
}
