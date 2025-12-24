package main

import (
	"flag"
	"fmt"
	"os"

	"asciigis/internal/tui"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [geojson path]\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), "\nExamples:\n  %s /path/to/data.geojson\n  %s   # start then enter path interactively\n", os.Args[0], os.Args[0])
	}

	var mapWidth int
	var mapHeight int
	flag.IntVar(&mapWidth, "W", 0, "Fixed canvas width (cells). 0 = auto")
	flag.IntVar(&mapWidth, "width", 0, "Fixed canvas width (cells). 0 = auto")
	flag.IntVar(&mapHeight, "H", 0, "Fixed canvas height (cells). 0 = auto")
	flag.IntVar(&mapHeight, "height", 0, "Fixed canvas height (cells). 0 = auto")

	flag.Parse()
	geoPath := ""
	if flag.NArg() >= 1 {
		geoPath = flag.Arg(0)
	}
	if err := tui.RunWithOptions(geoPath, tui.Options{MapWidth: mapWidth, MapHeight: mapHeight}); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
