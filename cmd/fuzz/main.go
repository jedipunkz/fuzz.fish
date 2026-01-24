package main

import (
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/cmd/fuzz/app"
)

func main() {
	// Check for subcommand
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "history", "git":
			// Both history and git modes use the unified app
			// Use Ctrl+R to toggle between them
			// Use Ctrl+F to switch to files mode
			app.Run()
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
			fmt.Fprintf(os.Stderr, "Usage: fuzz [history|git]\n")
			os.Exit(1)
		}
	}

	// Default: unified app
	app.Run()
}
