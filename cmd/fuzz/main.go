package main

import (
	"flag"
	"fmt"

	"github.com/jedipunkz/fuzz.fish/internal/app"
)

// version is stamped by the release workflow with
// -ldflags "-X main.version=<tag>". Builds from source report "dev".
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
