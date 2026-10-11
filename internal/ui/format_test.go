package ui

import (
	"strings"
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	ts := time.Date(2023, 3, 14, 9, 26, 53, 0, time.Local).Unix()
	if got := FormatTime(ts); got != "2023-03-14 09:26:53" {
		t.Errorf("FormatTime = %q, want 2023-03-14 09:26:53", got)
	}
}

func TestFormatTime_Zero(t *testing.T) {
	if got := FormatTime(0); got != "0000-00-00 00:00:00" {
		t.Errorf("FormatTime(0) = %q, want 0000-00-00 00:00:00", got)
	}
}

func TestFormatTimeAgo(t *testing.T) {
	for _, tt := range []struct {
		value int
		unit  string
		want  string
	}{
		{1, "second", "1 second ago"},
		{5, "second", "5 seconds ago"},
		{1, "year", "1 year ago"},
		{3, "month", "3 months ago"},
	} {
		if got := formatTimeAgo(tt.value, tt.unit); got != tt.want {
			t.Errorf("formatTimeAgo(%d, %q) = %q, want %q", tt.value, tt.unit, got, tt.want)
		}
	}
}

func TestFormatRelativeTime(t *testing.T) {
	if got := FormatRelativeTime(0); got != "unknown" {
		t.Errorf("FormatRelativeTime(0) = %q, want unknown", got)
	}

	now := time.Now()
	for _, tt := range []struct {
		name string
		ago  time.Duration
		want string
	}{
		{"seconds", 30 * time.Second, "30 seconds ago"},
		{"one second", 1 * time.Second, "1 second ago"},
		{"minutes", 150 * time.Second, "2 minutes ago"},
		{"one minute", 61 * time.Second, "1 minute ago"},
		{"hours", 3*time.Hour + 10*time.Minute, "3 hours ago"},
		{"one hour", 61 * time.Minute, "1 hour ago"},
		{"days", 3*24*time.Hour + 12*time.Hour, "3 days ago"},
		{"weeks", 20 * 24 * time.Hour, "2 weeks ago"},
		{"months", 90 * 24 * time.Hour, "3 months ago"},
		{"years", 400 * 24 * time.Hour, "1 year ago"},
	} {
		if got := FormatRelativeTime(now.Add(-tt.ago).Unix()); got != tt.want {
			t.Errorf("%s: FormatRelativeTime = %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestFormatRelativeTime_Future(t *testing.T) {
	future := time.Now().Add(2 * time.Minute).Unix()
	got := FormatRelativeTime(future)
	if !strings.Contains(got, "second") {
		t.Errorf("FormatRelativeTime(future) = %q, want seconds unit", got)
	}
}
