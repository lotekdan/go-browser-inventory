package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"go-browser-inventory/db"
	"go-browser-inventory/internal/browsers"
)

type output struct {
	Extensions []browsers.Extension `json:"extensions"`
	Total      int                  `json:"total"`
}

func main() {
	browser := flag.String("browser", "", "Browser to list extensions for (Chrome, Edge, Firefox). Leave empty for all.")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	debug := flag.Bool("debug", false, "Enable debug output for troubleshooting")
	updateCache := flag.Bool("update-cache", false, "Force update of database records, bypassing cache")
	flag.Parse()

	// Initialize SQLite DB (fatal error if fails)
	dbConn, err := db.NewDB("./browser_inventory.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing DB: %v\n", err)
		os.Exit(1)
	}
	defer dbConn.Close()

	// List of browsers to query
	browserList := []string{"Chrome", "Edge", "Firefox"}
	if *browser != "" {
		browserList = []string{*browser}
	}

	// Collect extensions for all relevant browsers
	var allExtensions []browsers.Extension
	var fetchError bool // Track if any non-fatal errors occur
	bi := browsers.NewBrowserInventory()
	for _, b := range browserList {
		var extensions []browsers.Extension
		if !*updateCache {
			extensions, err = dbConn.GetExtensions(b)
			if err != nil {
				if *debug {
					fmt.Fprintf(os.Stderr, "Error retrieving cached extensions for %s: %v\n", b, err)
				}
				// Proceed to fetch fresh extensions
			} else if extensions != nil {
				allExtensions = append(allExtensions, extensions...)
				continue
			}
		}

		// Fetch fresh extensions if cache is stale, empty, or -update-cache is set
		if extensions == nil || *updateCache {
			extensions, err = bi.GetExtensions(b, *debug)
			if err != nil {
				if *debug {
					fmt.Fprintf(os.Stderr, "Error fetching extensions for %s: %v\n", b, err)
				}
				fetchError = true
				continue
			}

			// Update cache
			if err := dbConn.UpdateExtensions(b, extensions); err != nil {
				if *debug {
					fmt.Fprintf(os.Stderr, "Error updating cache for %s: %v\n", b, err)
				}
				// Still use the fetched extensions even if cache update fails
			}
			allExtensions = append(allExtensions, extensions...)
		}
	}

	// Output logic
	if *jsonOutput {
		if fetchError {
			// Return empty JSON if any errors occurred
			fmt.Println(`{"extensions": [], "total": 0}`)
		} else {
			out := output{
				Extensions: allExtensions,
				Total:      len(allExtensions),
			}
			jsonData, err := json.MarshalIndent(out, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		}
	} else {
		if len(allExtensions) == 0 {
			fmt.Println("No extensions found.")
			return
		}

		fmt.Println("Browser Extensions:")
		fmt.Println("===================")
		for i, ext := range allExtensions {
			fmt.Printf("%d. %s\n", i+1, ext.Name)
			fmt.Printf("   Browser: %s\n", ext.Browser)
			fmt.Printf("   Version: %s\n", ext.Version)
			fmt.Printf("   ID: %s\n", ext.ID)
			fmt.Printf("   Enabled: %v\n", ext.Enabled)
			if ext.Profile != "" {
				fmt.Printf("   Profile: %s\n", ext.Profile)
			}
			fmt.Println("------------------")
		}
		fmt.Printf("Total extensions: %d\n", len(allExtensions))
	}
}
