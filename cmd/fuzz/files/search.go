package files

import (
	"errors"
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/ui"
	"github.com/koki-develop/go-fzf"
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

	// Use go-fzf with Tokyo Night theme
	f, err := ui.NewFinder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize fzf: %v\n", err)
		os.Exit(1)
	}

	idxs, err := f.Find(
		files,
		func(i int) string {
			return FormatEntry(files[i])
		},
		fzf.WithPreviewWindow(func(i, w, h int) string {
			if i < 0 || i >= len(files) {
				return "No selection"
			}
			return GeneratePreview(files[i], w, h)
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

	// Output selected file/dir
	if len(idxs) > 0 {
		selected := files[idxs[0]]
		if selected.IsDir {
			fmt.Printf("DIR:%s", selected.Path)
		} else {
			fmt.Printf("FILE:%s", selected.Path)
		}
	}
}
