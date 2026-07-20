package main

import (
	"flag"

	"github.com/jedipunkz/fuzz.fish/internal/app"
)

func main() {
	// --query pre-fills the search box (e.g. with the current Fish command line).
	query := flag.String("query", "", "initial search query")
	flag.Parse()

	app.Run(*query)
}
