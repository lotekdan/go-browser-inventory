package main

import (
	"flag"
	"os"
	"strings"

	"go-browser-inventory/debug"
	"go-browser-inventory/inventory"
)

func main() {
	outputPath := flag.String("o", "out.json", "Output file path for the JSON inventory")
	debugMode := flag.Bool("debug", false, "Enable debug logging to stderr")
	browser := flag.String("browser", "", "Specify browsers to scan (e.g., chrome, firefox, or chrome,firefox); leave empty for all")
	jsonOutput := flag.Bool("json", true, "Output results in JSON format (default true)")
	flag.Parse()

	if *debugMode {
		debug.Enable()
	} else {
		debug.Disable()
	}

	// Split browser flag into a list
	var browsers []string
	if *browser != "" {
		browsers = strings.Split(*browser, ",")
		for i, b := range browsers {
			browsers[i] = strings.TrimSpace(b)
		}
	}

	err := inventory.Generate(*outputPath, browsers, *jsonOutput)
	if err != nil {
		debug.Printf("Error generating inventory: %v", err)
		os.Exit(1)
	}
}
