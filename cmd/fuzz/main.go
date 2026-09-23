package main

import (
	"flag"
	"fmt"

	"github.com/jedipunkz/fuzz.fish/internal/app"
)

// version is stamped by the release workflow with
// -ldflags "-X main.version=<tag>". conf.d/fuzz.fish compares this against the
// version it pins and reinstalls the binary when they differ.
var version = "dev"

func main() {
	// --query pre-fills the search box (e.g. with the current Fish command line).
	query := flag.String("query", "", "initial search query")
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	app.Run(*query)
}
