package history

import (
	"fmt"
	"os"
	"syscall"

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

	// Save original stdout fd to output the result later
	origStdoutFd, err := syscall.Dup(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to dup stdout: %v\n", err)
		os.Exit(1)
	}
	defer syscall.Close(origStdoutFd)

	// Open /dev/tty for interactive TUI
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open /dev/tty: %v\n", err)
		os.Exit(1)
	}
	defer tty.Close()

	ttyFd := int(tty.Fd())

	// Redirect stdin, stdout, stderr to /dev/tty at fd level
	syscall.Dup2(ttyFd, int(os.Stdin.Fd()))
	syscall.Dup2(ttyFd, int(os.Stdout.Fd()))
	syscall.Dup2(ttyFd, int(os.Stderr.Fd()))

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

	// Restore stdout and output selected command
	syscall.Dup2(origStdoutFd, int(os.Stdout.Fd()))
	fmt.Print(entries[idx].Cmd)
}
