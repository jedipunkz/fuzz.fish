package scoring

import "testing"

func BenchmarkItemScore(b *testing.B) {
	c := DefaultConfig()
	now := CurrentTimestamp()
	cases := []struct {
		name      string
		text      string
		indexes   []int
		frequency int
	}{
		{"contiguous", "git pull origin main", []int{0, 1, 2, 4, 5, 6, 7}, 3},
		{"scattered", "git config pull.rebase true", []int{0, 1, 2, 11, 12, 13, 14}, 3},
		{"recency", "feature/very-long-branch-name", []int{0, 8, 13, 18}, 0},
	}
	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				c.ItemScore(tc.text, 10, tc.indexes, now-3600, tc.frequency, false, now)
			}
		})
	}
}
