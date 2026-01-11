package history

import (
	"errors"
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/koki-develop/go-fzf"
)

// RunSearch runs the interactive history search
func RunSearch() {
	// Parse history
	entries := Parse()
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No history found")
		os.Exit(1)
	}

	// Use go-fzf with Tokyo Night theme
	f, err := ui.NewFinder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize fzf: %v\n", err)
		os.Exit(1)
	}

	idxs, err := f.Find(
		entries,
		func(i int) string {
			return FormatEntry(entries[i])
		},
		fzf.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(entries) {
				return "No selection"
			}
			return GeneratePreview(entries[i], entries, i, w, h)
		}),
	)

	if err != nil {
		if errors.Is(err, fzf.ErrAbort) {
			// User cancelled
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "fzf error: %v\n", err)
		os.Exit(1)
	}

	// Output selected command
	if len(idxs) > 0 {
		fmt.Print(entries[idxs[0]].Cmd)
	}
}
