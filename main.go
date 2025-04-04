package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"go-browser-inventory/inventory"

	"github.com/olekukonko/tablewriter"
)

func main() {
	// Define flags
	browserFlag := flag.String("browser", "", "Comma-separated list of browsers to scan (e.g., chrome,firefox,brave,edge). Default: all")
	shortBrowserFlag := flag.String("b", "", "Shorthand for --browser")
	jsonFlag := flag.Bool("json", false, "Output inventory as JSON only")
	shortJsonFlag := flag.Bool("j", false, "Shorthand for --json")
	debugFlag := flag.Bool("debug", false, "Enable debug logging to stderr")
	outputFlag := flag.String("o", "out.json", "Output file for JSON inventory")

	flag.Parse()

	// Handle shorthand flags
	browser := *browserFlag
	if *shortBrowserFlag != "" {
		browser = *shortBrowserFlag
	}
	useJSON := *jsonFlag || *shortJsonFlag

	// Set up debug logging if enabled
	if *debugFlag && !useJSON {
		inventory.SetDebug(true)
	} else {
		inventory.SetDebug(false) // Suppress debug unless explicitly enabled
	}

	// Parse browser filter
	var browserFilter []string
	if browser != "" {
		browserFilter = strings.Split(strings.ToLower(browser), ",")
		for i, b := range browserFilter {
			browserFilter[i] = strings.TrimSpace(b)
		}
	}

	// Generate inventory
	err := inventory.Generate(*outputFlag, browserFilter, useJSON)
	if err != nil {
		if useJSON {
			fmt.Fprintf(os.Stderr, `{"error": "%v"}`+"\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error generating inventory: %v\n", err)
		}
		os.Exit(1)
	}

	// Read the generated JSON file
	data, err := os.ReadFile(*outputFlag)
	if err != nil {
		if useJSON {
			fmt.Fprintf(os.Stderr, `{"error": "failed to read output file: %v"}`+"\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error reading output file: %v\n", err)
		}
		os.Exit(1)
	}

	var inv inventory.Inventory
	if err := json.Unmarshal(data, &inv); err != nil {
		if useJSON {
			fmt.Fprintf(os.Stderr, `{"error": "failed to parse inventory: %v"}`+"\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error parsing inventory: %v\n", err)
		}
		os.Exit(1)
	}

	// Output based on flags
	if useJSON {
		fmt.Println(string(data))
	} else {
		// Console-friendly table output
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Name", "Version", "Enabled", "Browser"})
		table.SetAutoWrapText(false)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT})

		for _, ext := range inv.Extensions {
			enabled := "Yes"
			if !ext.Enabled {
				enabled = "No"
			}
			table.Append([]string{ext.ID, ext.Name, ext.Version, enabled, ext.Browser})
		}

		fmt.Printf("Browser Extension Inventory (Total: %d)\n", inv.Total)
		table.Render()
		fmt.Printf("Inventory written to %s\n", *outputFlag)
	}
}
