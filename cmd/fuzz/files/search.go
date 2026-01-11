package files

import (
	"fmt"
	"os"

	"github.com/ktr0731/go-fuzzyfinder"
)

// RunSearch runs the interactive file search
func RunSearch() {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Collect files
	files := Collect(cwd)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No files found")
		os.Exit(1)
	}

	// Use go-fuzzyfinder
	idx, err := fuzzyfinder.Find(
		files,
		func(i int) string {
			return FormatEntry(files[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return GeneratePreview(files[i], w, h)
		}),
	)

	if err != nil {
		// User cancelled (Ctrl+C, ESC)
		os.Exit(0)
	}

	// Output selected file/dir
	selected := files[idx]
	if selected.IsDir {
		fmt.Printf("DIR:%s", selected.Path)
	} else {
		fmt.Printf("FILE:%s", selected.Path)
	}
}
