package utils

import (
	"fmt"
	"os"
	"syscall"
)

// SetupTTY redirects stdin, stdout, stderr to /dev/tty for TUI interaction
// and returns the original stdout file descriptor for later restoration.
// The caller must call RestoreTTY with the returned fd when done.
func SetupTTY() (origStdoutFd int, err error) {
	// Save original stdout fd to output the result later
	origStdoutFd, err = syscall.Dup(int(os.Stdout.Fd()))
	if err != nil {
		return 0, fmt.Errorf("failed to dup stdout: %w", err)
	}

	// Open /dev/tty for interactive TUI
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		syscall.Close(origStdoutFd)
		return 0, fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer tty.Close()

	ttyFd := int(tty.Fd())

	// Redirect stdin, stdout, stderr to /dev/tty at fd level
	syscall.Dup2(ttyFd, int(os.Stdin.Fd()))
	syscall.Dup2(ttyFd, int(os.Stdout.Fd()))
	syscall.Dup2(ttyFd, int(os.Stderr.Fd()))

	return origStdoutFd, nil
}

// RestoreTTY restores stdout to the original file descriptor
func RestoreTTY(origStdoutFd int) {
	syscall.Dup2(origStdoutFd, int(os.Stdout.Fd()))
	syscall.Close(origStdoutFd)
}
