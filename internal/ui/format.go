package ui

import (
	"fmt"
	"time"
)

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(size int64) string {
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

// FormatTime formats a Unix timestamp as YYYY-MM-DD HH:MM:SS
func FormatTime(timestamp int64) string {
	if timestamp == 0 {
		return "0000-00-00 00:00:00"
	}

	t := time.Unix(timestamp, 0)
	return t.Format("2006-01-02 15:04:05")
}

// formatTimeAgo formats a time value with the appropriate unit and singular/plural form
func formatTimeAgo(value int, unit string) string {
	if value == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", value, unit)
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
		return formatTimeAgo(seconds, "second")
	case minutes < 60:
		return formatTimeAgo(minutes, "minute")
	case hours < 24:
		return formatTimeAgo(hours, "hour")
	case days < 7:
		return formatTimeAgo(days, "day")
	case weeks < 4:
		return formatTimeAgo(weeks, "week")
	case months < 12:
		return formatTimeAgo(months, "month")
	default:
		return formatTimeAgo(years, "year")
	}
}
