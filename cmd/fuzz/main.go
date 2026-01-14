package main

import (
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/app"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/files"
	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/git"
)

func main() {
	// Check for subcommand
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "files":
			files.RunSearch()
			return
		case "history":
			// Explicit history mode (using unified app)
			app.Run()
			return
		case "git":
			// Git subcommand - check for git-specific action
			if len(os.Args) > 2 {
				switch os.Args[2] {
				case "branch":
					// Using unified app but could force git mode if implemented
					// For now, let's keep git.RunBranchSearch for backward compatibility
					// OR better: use app.Run() and switch to git mode initially?
					// The user requested 'ctrl-r' to switch.
					// If they invoke 'git branch' directly, maybe they want just that.
					// Let's stick to user request: "ctrl-r" launches history, then toggle.
					git.RunBranchSearch()
					return
				default:
					fmt.Fprintf(os.Stderr, "Unknown git subcommand: %s\n", os.Args[2])
					fmt.Fprintf(os.Stderr, "Usage: fuzz git [branch]\n")
					os.Exit(1)
				}
			}
			git.RunBranchSearch()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: fuzz [history|files|git]\n")
			os.Exit(1)
		}
	}

	// Default: unified app
	app.Run()
}
