package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Styling definitions
var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	focusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	// Header style for the table
	headerStyle = table.DefaultStyles().Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(false)

	// Selected row style
	selectedStyle = table.DefaultStyles().Selected.
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)

type HistoryEntry struct {
	Cmd     string
	When    int64
	Paths   []string
	CmdLine int
}

type model struct {
	table       table.Model
	input       textinput.Model
	entries     []HistoryEntry
	filtered    []HistoryEntry
	width       int
	height      int
	selectedCmd string // The command selected by the user to execute
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate split sizes
		// Left panel (Table): 60% width
		// Right panel (Preview): 40% width (minus borders/padding)
		tableWidth := int(float64(m.width) * 0.6)

		// Adjust table dimensions
		// Reserve space for input (3 lines) and help (2 lines) and borders
		tableHeight := m.height - 7
		if tableHeight < 5 {
			tableHeight = 5
		}

		// Update table columns width based on available space
		// Time: 15, Dir: 25, Cmd: Rest
		timeWidth := 15
		dirWidth := 25
		cmdWidth := tableWidth - timeWidth - dirWidth - 10 // safety margin
		if cmdWidth < 10 {
			cmdWidth = 10
		}

		cols := []table.Column{
			{Title: "Time", Width: timeWidth},
			{Title: "Directory", Width: dirWidth},
			{Title: "Command", Width: cmdWidth},
		}
		m.table.SetColumns(cols)
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(tableWidth)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			// Select the current row
			selectedRow := m.table.SelectedRow()
			if selectedRow != nil && len(selectedRow) >= 3 {
				// The command is in the 3rd column
				// But wait, the table row data is formatted string.
				// We need the original full command, not truncated or formatted.
				// Best way is to map back via index.
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.filtered) {
					m.selectedCmd = m.filtered[idx].Cmd
				}
			}
			return m, tea.Quit
		case "down", "ctrl+n":
			m.table.MoveDown(1)
			return m, nil
		case "up", "ctrl+p":
			m.table.MoveUp(1)
			return m, nil
		}
	}

	// Handle Input and Filtering
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)

	// Filter logic
	// Only re-filter if input changed or init
	filterText := m.input.Value()

	// We need to re-filter every time input changes or just use simple logic
	// For now, let's just re-filter on every update if input changed
	// Ideally we would check if input actually changed but bubbletea doesn't give easy diff.
	// Let's just do it.

	var newFiltered []HistoryEntry
	var rows []table.Row

	lowerFilter := strings.ToLower(filterText)

	for _, e := range m.entries {
		// Search in Command and Directory
		match := true
		if filterText != "" {
			match = strings.Contains(strings.ToLower(e.Cmd), lowerFilter)
			if !match && len(e.Paths) > 0 {
				match = strings.Contains(strings.ToLower(e.Paths[0]), lowerFilter)
			}
		}

		if match {
			newFiltered = append(newFiltered, e)

			// Format row
			timeStr := formatTime(e.When)
			dirStr := ""
			if len(e.Paths) > 0 {
				dirStr = formatDir(e.Paths[0])
			}

			// Truncate dirStr if too long for table column (optional, table handles it)
			// But good to clean up home path

			rows = append(rows, table.Row{timeStr, dirStr, e.Cmd})
		}
	}
	m.filtered = newFiltered
	m.table.SetRows(rows)

	// If filter changed significantly (length diff), reset cursor?
	// Or just ensure cursor is valid
	if m.table.Cursor() >= len(rows) {
		m.table.SetCursor(len(rows) - 1)
	}

	cmd = tea.Batch(cmd, inputCmd)
	return m, cmd
}

func (m model) View() string {
	// 1. Calculate Layout
	// Top: Search Bar
	// Middle: Split Pane (Left: Table, Right: Preview)
	// Bottom: Help/Status

	// --- Search Bar ---
	searchView := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(m.width - 2). // Full width minus border
		Render(m.input.View())

	// --- Main Content (Split) ---

	// Left: Table
	tableHeight := m.height - 7
	if tableHeight < 0 {
		tableHeight = 0
	}

	// Apply styles to table
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	m.table.SetStyles(s)

	tableView := baseStyle.
		Width(m.table.Width()).
		Height(tableHeight).
		Render(m.table.View())

	// Right: Preview
	// Get selected entry
	var previewContent string

	// Calculate available width for preview
	availWidth := m.width - m.table.Width() - 6
	if availWidth < 10 {
		availWidth = 10 // Minimum width safeguard, though generatePreview will return empty if small
	}

	idx := m.table.Cursor()
	if idx >= 0 && idx < len(m.filtered) {
		entry := m.filtered[idx]
		previewContent = generatePreview(entry, m.filtered, idx, availWidth)
	} else {
		previewContent = "No selection"
	}

	previewWidth := m.width - m.table.Width() - 4 // borders/margin
	if previewWidth < 10 {
		previewWidth = 10
	}

	previewView := baseStyle.
		Width(previewWidth).
		Height(tableHeight).
		Padding(1).
		Render(previewContent)

	// Join Left and Right
	mainView := lipgloss.JoinHorizontal(lipgloss.Top, tableView, previewView)

	// --- Footer ---
	helpView := helpStyle.Render("Use arrows to move • Type to filter • Enter to select • Esc to quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		searchView,
		mainView,
		helpView,
	)
}

func generatePreview(entry HistoryEntry, all []HistoryEntry, idx int, maxWidth int) string {
	// Safety check for width
	if maxWidth < 10 {
		return "" // Not enough space for preview
	}

	// Format details
	sb := strings.Builder{}

	// Command
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Render("COMMAND"))
	sb.WriteString("\n")
	// Wrap long commands
	sb.WriteString(lipgloss.NewStyle().Width(maxWidth).Render(entry.Cmd))
	sb.WriteString("\n\n")

	// Time & Dir
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Time: "))
	sb.WriteString(formatTime(entry.When))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Dir:  "))
	if len(entry.Paths) > 0 {
		sb.WriteString(formatDir(entry.Paths[0]))
	}
	sb.WriteString("\n\n")

	// Context
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).Render("CONTEXT"))
	sb.WriteString("\n")

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
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
		if i == idx {
			cursor = "→ "
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		}

		// Truncate cmd for preview context
		cmd := e.Cmd
		if maxWidth > 8 && len(cmd) > maxWidth-5 {
			cmd = cmd[:maxWidth-8] + "..."
		}

		line := fmt.Sprintf("%s %s", cursor, cmd)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// --- Main & Helpers ---

func main() {
	// Check if stdin is a terminal
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// If not a terminal, we might be in a pipe, but bubbletea needs a TTY.
		// Usually when run from fish as `fh`, stdin is inherited.
		// However, if fish does something weird with pipes, it might fail.
		// Let's try to open /dev/tty explicitly if needed, but bubbletea usually handles it.
		// For now, assume it works or fail gracefully.
	}

	entries := parseHistory()

	// Initialize Table
	columns := []table.Column{
		{Title: "Time", Width: 15},
		{Title: "Directory", Width: 25},
		{Title: "Command", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(10), // Initial, updated in resize
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	// Initialize Input
	ti := textinput.New()
	ti.Placeholder = "Search history..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	m := model{
		table:    t,
		input:    ti,
		entries:  entries,
		filtered: entries, // Initially all
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}

	// On exit, print the selected command if any
	if m, ok := finalModel.(model); ok && m.selectedCmd != "" {
		fmt.Print(m.selectedCmd)
	}
}

// ... (Use existing parseHistory, getHistoryPath, formatTime logic but copied here) ...

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
	// Truncate logic can be added here if needed, but table handles width clipping
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
