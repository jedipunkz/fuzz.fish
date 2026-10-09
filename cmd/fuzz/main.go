package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jedipunkz/fuzz.fish/internal/app"
	"github.com/jedipunkz/fuzz.fish/internal/config"
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

	keys, err := config.LoadKeymap()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fuzz: config: %v\n", err)
		os.Exit(1)
	}

	app.Run(*query, keys)
}
