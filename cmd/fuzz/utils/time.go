package utils

import (
	"fmt"
	"time"
)

// FormatTime formats a Unix timestamp as YYYY-MM-DD HH:MM:SS
func FormatTime(timestamp int64) string {
	if timestamp == 0 {
		return "0000-00-00 00:00:00"
	}

	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// FormatRelativeTime formats a Unix timestamp as a human-readable relative time
func FormatRelativeTime(timestamp int64) string {
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
