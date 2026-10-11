package history

import (
	"fmt"
	"strings"
	"testing"
)

// benchHistory renders n history records in Fish's on-disk format, with every
// fourth command repeated so deduplication is exercised.
func benchHistory(n int) string {
	var sb strings.Builder
	for i := range n {
		fmt.Fprintf(&sb, "- cmd: git commit -m \"change %d\"\n  when: %d\n", i%(n*3/4+1), 1700000000+i)
		if i%5 == 0 {
			fmt.Fprintf(&sb, "  paths:\n    - path/to/file_%d.go\n", i)
		}
	}
	return sb.String()
}

func BenchmarkParseReader(b *testing.B) {
	for _, size := range []int{1000, 10000, 100000} {
		data := benchHistory(size)
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for b.Loop() {
				parseReader(strings.NewReader(data))
			}
		})
	}
}
