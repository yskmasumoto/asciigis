package main

import (
	"flag"
	"fmt"
	"os"

	"asciigis/internal/tui"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [geojson path]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n  %s /path/to/data.geojson\n  %s   # start then enter path interactively\n", os.Args[0], os.Args[0])
	}

	flag.Parse()
	geoPath := ""
	if flag.NArg() >= 1 {
		geoPath = flag.Arg(0)
	}
	if err := tui.Run(geoPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
