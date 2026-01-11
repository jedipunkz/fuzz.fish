package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
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

type FileEntry struct {
	Path  string
	IsDir bool
	Size  int64
	Mode  fs.FileMode
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
	// Check for subcommand
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "files":
			runFileSearch()
			return
		case "history":
			// Explicit history mode (default)
			runHistorySearch()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: fhv [history|files]\n")
			os.Exit(1)
		}
	}

	// Default: history search
	runHistorySearch()
}

func runHistorySearch() {
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

// File search functionality

func runFileSearch() {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Collect files
	files := collectFiles(cwd)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files found")
		os.Exit(1)
	}

	// Use go-fzf with Tokyo Night theme
	f, err := fzf.New(
		fzf.WithStyles(
			fzf.WithStylePrompt(fzf.Style{ForegroundColor: colorBlue}),
			fzf.WithStyleInputText(fzf.Style{ForegroundColor: colorForeground}),
			fzf.WithStyleCursor(fzf.Style{ForegroundColor: colorBlue}),
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
		files,
		func(i int) string {
			return formatFileEntry(files[i])
		},
		fzf.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(files) {
				return "No selection"
			}
			return generateFilePreview(files[i], w, h)
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

	// Output selected file path with special marker for directories
	if len(idxs) > 0 {
		selected := files[idxs[0]]
		if selected.IsDir {
			// Output: DIR:<path>
			fmt.Printf("DIR:%s", selected.Path)
		} else {
			// Output: FILE:<path>
			fmt.Printf("FILE:%s", selected.Path)
		}
	}
}

func collectFiles(root string) []FileEntry {
	var files []FileEntry

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip .git directory
		if d.IsDir() && d.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip hidden files/directories (except current dir)
		if path != root && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			relPath = path
		}

		// Skip current directory itself
		if relPath == "." {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		files = append(files, FileEntry{
			Path:  relPath,
			IsDir: d.IsDir(),
			Size:  info.Size(),
			Mode:  info.Mode(),
		})

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
	}

	return files
}

func formatFileEntry(e FileEntry) string {
	icon := "📄"
	if e.IsDir {
		icon = "📁"
	}
	return fmt.Sprintf("%s %s", icon, e.Path)
}

func generateFilePreview(entry FileEntry, width, height int) string {
	var sb strings.Builder

	// Header
	sb.WriteString(headerStyle.Render("File Information") + "\n\n")

	// Path
	sb.WriteString(labelStyle.Render("Path") + "\n")
	sb.WriteString(contentStyle.Render(entry.Path) + "\n\n")

	// Type
	sb.WriteString(labelStyle.Render("Type") + "\n")
	if entry.IsDir {
		sb.WriteString(contentStyle.Render("Directory") + "\n\n")

		// Directory listing
		sb.WriteString(contextHeaderStyle.Render("Contents") + "\n")
		listing := getDirectoryListing(entry.Path)
		if listing != "" {
			sb.WriteString(listing)
		} else {
			sb.WriteString(inactiveContextStyle.Render("  (empty)") + "\n")
		}
	} else {
		sb.WriteString(contentStyle.Render("File") + "\n\n")

		// Size
		sb.WriteString(labelStyle.Render("Size") + "\n")
		sb.WriteString(contentStyle.Render(formatFileSize(entry.Size)) + "\n\n")

		// Permissions
		sb.WriteString(labelStyle.Render("Permissions") + "\n")
		sb.WriteString(contentStyle.Render(entry.Mode.String()) + "\n\n")

		// File preview
		sb.WriteString(contextHeaderStyle.Render("Preview") + "\n")
		preview := getFilePreview(entry.Path, height-15)
		if preview != "" {
			sb.WriteString(preview)
		} else {
			sb.WriteString(inactiveContextStyle.Render("  (binary or empty file)") + "\n")
		}
	}

	return sb.String()
}

func getDirectoryListing(path string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}

	var sb strings.Builder
	count := 0
	maxEntries := 20

	for _, entry := range entries {
		if count >= maxEntries {
			sb.WriteString(inactiveContextStyle.Render(fmt.Sprintf("  ... and %d more", len(entries)-maxEntries)) + "\n")
			break
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}
		sb.WriteString(inactiveContextStyle.Render(fmt.Sprintf("  %s %s", icon, entry.Name())) + "\n")
		count++
	}

	return sb.String()
}

func getFilePreview(path string, maxLines int) string {
	// Try to use bat for syntax highlighting
	if hasBat() {
		content := runBat(path, maxLines)
		if content != "" {
			return content
		}
	}

	// Fallback to plain cat
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	// Check if binary
	if isBinary(content) {
		return ""
	}

	// Limit lines
	lines := strings.Split(string(content), "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	var sb strings.Builder
	for i, line := range lines {
		if i >= maxLines {
			break
		}
		// Truncate long lines
		if len(line) > 120 {
			line = line[:120] + "..."
		}
		sb.WriteString(inactiveContextStyle.Render(fmt.Sprintf("  %s", line)) + "\n")
	}

	return sb.String()
}

func hasBat() bool {
	_, err := exec.LookPath("bat")
	return err == nil
}

func runBat(path string, maxLines int) string {
	cmd := exec.Command("bat", "--color=always", "--style=numbers", fmt.Sprintf("--line-range=:%d", maxLines), path)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Add indentation
	lines := strings.Split(string(output), "\n")
	var sb strings.Builder
	for _, line := range lines {
		if line != "" {
			sb.WriteString("  " + line + "\n")
		}
	}

	return sb.String()
}

func isBinary(content []byte) bool {
	// Simple binary detection: check for null bytes in first 8KB
	checkSize := 8192
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

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d bytes", size)
	}
}
