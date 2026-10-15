package app

import (
	"fmt"
	"testing"
)

// benchCommands builds n distinct, realistic-looking history commands so the
// fuzzy matcher sees a mix of hits, misses and scattered matches. The CI
// benchmark workflow runs these benchmarks against the merge base of a PR in
// .github/workflows/bench.yml.
func benchCommands(n int) []string {
	templates := []string{
		"git pull origin feature/branch-%d",
		"git config pull.rebase true # %d",
		"git commit -m \"fix: handle case %d\"",
		"nvim internal/app/file_%d.go",
		"go test ./internal/... -run TestCase%d",
		"docker run --rm -it image:%d",
		"kubectl get pods -n namespace-%d",
		"cd ~/src/github.com/user/project-%d",
	}
	cmds := make([]string, n)
	for i := range cmds {
		cmds[i] = fmt.Sprintf(templates[i%len(templates)], i)
	}
	return cmds
}

func BenchmarkUpdateFilter(b *testing.B) {
	queries := []struct {
		name  string
		query string
	}{
		{"single", "gpl"},
		{"multi", "git pull"},
		{"glob", "nvim *.go"},
	}
	for _, size := range []int{1000, 10000, 50000} {
		m := historyModel(benchCommands(size))
		for _, q := range queries {
			b.Run(fmt.Sprintf("%s/n=%d", q.name, size), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					m.updateFilter(q.query)
				}
			})
		}
	}
}

func BenchmarkLoadItemsForMode(b *testing.B) {
	for _, size := range []int{1000, 10000, 50000} {
		m := historyModel(benchCommands(size))
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.loadItemsForMode()
			}
		})
	}
}

// BenchmarkUpdateFilterTyping types a query one character at a time, the way
// the debounced filter sees it, so each keystroke can narrow to the previous
// matches instead of rescanning every item.
func BenchmarkUpdateFilterTyping(b *testing.B) {
	const query = "git pull origin"
	for _, size := range []int{1000, 10000, 50000} {
		m := historyModel(benchCommands(size))
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				m.updateFilter("")
				for i := 1; i <= len(query); i++ {
					m.updateFilter(query[:i])
				}
			}
		})
	}
}
