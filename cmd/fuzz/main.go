package main

import (
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/files"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/history"
)

func main() {
	// Check for subcommand
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "files":
			files.RunSearch()
			return
		case "history":
			// Explicit history mode (default)
			history.RunSearch()
			return
		case "git":
			// Git subcommand - check for git-specific action
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "branch":
					git.RunBranchSearch()
					return
				default:
					fmt.Fprintf(os.Stderr, "Unknown git subcommand: %s\n", os.Args[2])
					fmt.Fprintf(os.Stderr, "Usage: fuzz git [branch]\n")
					os.Exit(1)
				}
			}
			// Default git action: branch search
			git.RunBranchSearch()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: fuzz [history|files|git]\n")
			os.Exit(1)
		}
	}

	// Default: history search
	history.RunSearch()
}
