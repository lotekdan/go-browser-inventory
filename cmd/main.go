package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	browserPtr := flag.String("browser", "", "Specify a browser to scan (chrome, edge, firefox). Leave empty for all.")
	helpPtr := flag.Bool("help", false, "Display help information")
	debugPtr := flag.Bool("debug", false, "Enable debug output")
	jsonPtr := flag.Bool("json", false, "Output in JSON format (default is console-friendly)")

	flag.Parse()

	if *helpPtr {
		fmt.Println("Browser Extension Inventory Utility")
		fmt.Println("==================================")
		fmt.Println("This utility scans for browser extensions and outputs them in either JSON or console-friendly format.")
		fmt.Println("\nUsage:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  Scan all (console): go run .")
		fmt.Println("  Scan Chrome (JSON): go run . -browser chrome -json")
		fmt.Println("  Enable debug:       go run . -debug")
		fmt.Println("  Show help:          go run . -help")
		os.Exit(0)
	}

	selectedBrowser := *browserPtr
	if selectedBrowser != "" {
		validBrowsers := map[string]bool{
			"chrome":  true,
			"edge":    true,
			"firefox": true,
		}
		if !validBrowsers[strings.ToLower(selectedBrowser)] {
			fmt.Printf("Error: Invalid browser '%s'. Use 'chrome', 'edge', or 'firefox'.\n", selectedBrowser)
			fmt.Println("Usage:")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	inventory := NewBrowserInventory()
	extensions, err := inventory.GetExtensions(selectedBrowser, *debugPtr)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonPtr {
		output := InventoryOutput{
			Extensions: extensions,
			Total:      len(extensions),
		}
		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		printConsoleFriendly(extensions)
	}
}

// printConsoleFriendly prints extensions in a human-readable format
func printConsoleFriendly(extensions []Extension) {
	if len(extensions) == 0 {
		fmt.Println("No extensions found.")
		return
	}

	fmt.Println("Browser Extensions:")
	fmt.Println("==================")
	for i, ext := range extensions {
		fmt.Printf("%d. %s\n", i+1, ext.Name)
		fmt.Printf("   Browser: %s\n", ext.Browser)
		fmt.Printf("   Version: %s\n", ext.Version)
		fmt.Printf("   ID: %s\n", ext.ID)
		fmt.Printf("   Enabled: %v\n", ext.Enabled)
		fmt.Println("------------------")
	}
	fmt.Printf("Total extensions: %d\n", len(extensions))
}
