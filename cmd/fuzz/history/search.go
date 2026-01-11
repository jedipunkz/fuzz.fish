package history

import (
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
)

// RunSearch runs the interactive history search
func RunSearch() {
	// Parse history
	entries := Parse()
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		os.Exit(1)
	}

	// Use go-fuzzyfinder
	idx, err := fuzzyfinder.Find(
		entries,
		func(i int) string {
			return FormatEntry(entries[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return GeneratePreview(entries[i], entries, i, w, h)
		}),
	)

	if err != nil {
		// User cancelled (Ctrl+C, ESC)
		os.Exit(0)
	}

	// Output selected command
	fmt.Print(entries[idx].Cmd)
}
