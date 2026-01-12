package git

import "fmt"

// FormatBranch formats a branch entry for display in the TUI list
func FormatBranch(b Branch) string {
	var icon string
	if b.IsCurrent {
		icon = "✓"
	} else if b.IsRemote {
		icon = "🌐"
	} else {
		icon = "⎇"
	}

	return fmt.Sprintf("%s %s", icon, b.Name)
}
