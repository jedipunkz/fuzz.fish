package app

import (
	"fmt"
	"testing"
)

// TestUpdateFilter_AllocCount asserts the allocation count of the filter hot
// path: CLAUDE.md requires updateFilter to avoid per-keystroke allocations on
// a warmed-up model, so a regression must fail go test. Unlike benchmark
// timings, allocation counts are deterministic and CI-safe. Thresholds are the
// measured values plus ~10% headroom; counts scale with the match set, so each
// query/size pair carries its own cap.
func TestUpdateFilter_AllocCount(t *testing.T) {
	queries := []struct {
		name  string
		query string
		// maxAllocs is per benchmark size, indexed by the position of the
		// size in sizes below.
		maxAllocs [2]float64
	}{
		{name: "empty", query: "", maxAllocs: [2]float64{2, 2}},
		{name: "single", query: "gpl", maxAllocs: [2]float64{310, 2800}},
		{name: "multi", query: "git pull", maxAllocs: [2]float64{1300, 12500}},
		{name: "glob", query: "nvim *.go", maxAllocs: [2]float64{155, 1400}},
	}
	sizes := []int{1000, 10000}
	for i, size := range sizes {
		m := historyModel(benchCommands(size))
		for _, q := range queries {
			// Warm up so only the steady-state updateFilter cost counts.
			m.updateFilter(q.query)
			t.Run(fmt.Sprintf("%s/n=%d", q.name, size), func(t *testing.T) {
				allocs := testing.AllocsPerRun(10, func() {
					m.updateFilter(q.query)
				})
				if allocs > q.maxAllocs[i] {
					t.Errorf("updateFilter(%q) n=%d: %g allocs/run, want <= %g",
						q.query, size, allocs, q.maxAllocs[i])
				}
			})
		}
	}
}
