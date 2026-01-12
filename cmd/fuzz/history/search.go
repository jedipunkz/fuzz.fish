package history

import (
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/utils"
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

	// Setup TTY for interactive TUI
	origStdoutFd, err := utils.SetupTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
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
		utils.RestoreTTY(origStdoutFd)
		os.Exit(0)
	}

	// Restore stdout BEFORE outputting the command
	utils.RestoreTTY(origStdoutFd)

	// Output selected command
	fmt.Print(entries[idx].Cmd)
}
