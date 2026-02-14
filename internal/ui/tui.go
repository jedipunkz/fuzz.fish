package ui

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
		_ = syscall.Close(origStdoutFd)
		return 0, fmt.Errorf("failed to open /dev/tty: %w", err)
	}
	defer func() {
		_ = tty.Close()
	}()

	ttyFd := int(tty.Fd())

	// Redirect stdin, stdout, stderr to /dev/tty at fd level
	if err = syscall.Dup2(ttyFd, int(os.Stdin.Fd())); err != nil {
		return 0, fmt.Errorf("failed to dup2 stdin: %w", err)
	}
	if err = syscall.Dup2(ttyFd, int(os.Stdout.Fd())); err != nil {
		return 0, fmt.Errorf("failed to dup2 stdout: %w", err)
	}
	if err = syscall.Dup2(ttyFd, int(os.Stderr.Fd())); err != nil {
		return 0, fmt.Errorf("failed to dup2 stderr: %w", err)
	}

	return origStdoutFd, nil
}

// RestoreTTY restores stdout to the original file descriptor
func RestoreTTY(origStdoutFd int) {
	_ = syscall.Dup2(origStdoutFd, int(os.Stdout.Fd()))
	_ = syscall.Close(origStdoutFd)
}
