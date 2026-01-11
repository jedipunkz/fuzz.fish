package history

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// FormatEntry formats a history entry for display in the TUI list
func FormatEntry(e Entry) string {
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

// FormatDir abbreviates a directory path by replacing the home directory with ~
func FormatDir(path string) string {
	home, err := os.UserHomeDir()
	if err == nil {
		path = strings.Replace(path, home, "~", 1)
	}
	return path
}
