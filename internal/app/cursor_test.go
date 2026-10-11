package app

import (
	"testing"
	"time"
)

func TestValidateCursor(t *testing.T) {
	tests := []struct {
		name           string
		filtered       int
		cursor, offset int
		mainHeight     int
		wantCursor     int
		wantOffset     int
	}{
		{"empty list", 0, 7, 3, 5, 0, 0},
		{"cursor past end", 3, 9, 0, 5, 2, 0},
		{"negative cursor clamps to zero", 3, -1, 0, 5, 0, 0},
		{"negative offset clamps to zero", 3, 0, -4, 5, 0, 0},
		{"cursor above offset pulls offset down", 10, 2, 5, 5, 2, 2},
		{"cursor below window scrolls down", 10, 8, 0, 5, 8, 4},
		{"cursor within window stays", 10, 6, 2, 5, 6, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := model{
				cursor:     tt.cursor,
				offset:     tt.offset,
				mainHeight: tt.mainHeight,
			}
			m.filtered = make([]Item, tt.filtered)
			m.validateCursor()

			if m.cursor != tt.wantCursor || m.offset != tt.wantOffset {
				t.Errorf("(cursor, offset) = (%d, %d), want (%d, %d)",
					m.cursor, m.offset, tt.wantCursor, tt.wantOffset)
			}
		})
	}
}

func TestFormatTimeAgoView(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name string
		ago  time.Duration
		want string
	}{
		{"seconds", 45 * time.Second, "45s ago"},
		{"minutes", 5 * time.Minute, "5m ago"},
		{"hours", 3 * time.Hour, "3h ago"},
		{"days", 2 * 24 * time.Hour, "2d ago"},
	}
	for _, tt := range tests {
		if got := formatTimeAgo(now.Add(-tt.ago).Unix()); got != tt.want {
			t.Errorf("%s: formatTimeAgo = %q, want %q", tt.name, got, tt.want)
		}
	}
}
