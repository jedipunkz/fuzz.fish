package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/utils"
	"github.com/ktr0731/go-fuzzyfinder"
)

// RunBranchSearch runs the interactive git branch search
func RunBranchSearch() {
	// Check if we're in a git repository
	if !isGitRepo() {
		fmt.Fprintln(os.Stderr, "Not a git repository")
		os.Exit(1)
	}

	// Collect branches
	branches := CollectBranches()
	if len(branches) == 0 {
		fmt.Fprintln(os.Stderr, "No branches found")
		os.Exit(1)
	}

	// Setup TTY for interactive TUI
	origStdoutFd, err := utils.SetupTTY()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	defer utils.RestoreTTY(origStdoutFd)

	// Use go-fuzzyfinder
	idx, err := fuzzyfinder.Find(
		branches,
		func(i int) string {
			return FormatBranch(branches[i])
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return GeneratePreview(branches[i], w, h)
		}),
	)

	if err != nil {
		// User cancelled (Ctrl+C, ESC)
		os.Exit(0)
	}

	// Output selected branch name (stdout will be restored by defer)
	selectedBranch := branches[idx]

	// If it's a remote branch, extract the branch name without remote prefix
	branchName := selectedBranch.Name
	if selectedBranch.IsRemote {
		// Remove "origin/" or other remote prefixes
		parts := strings.SplitN(branchName, "/", 2)
		if len(parts) == 2 {
			branchName = parts[1]
		}
	}

	fmt.Print(branchName)
}

// isGitRepo checks if the current directory is inside a git repository
func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	return err == nil
}
