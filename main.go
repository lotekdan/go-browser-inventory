package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"go-browser-inventory/internal/browsers"
)

type output struct {
	Extensions []browsers.Extension `json:"extensions"`
	Total      int                  `json:"total"`
}

func main() {
	browser := flag.String("browser", "", "Browser to list extensions for (Chrome, Edge, Firefox)")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	debug := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	bi := browsers.NewBrowserInventory()
	extensions, err := bi.GetExtensions(*browser, *debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		out := output{
			Extensions: extensions,
			Total:      len(extensions),
		}
		jsonData, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonData))
	} else {
		if len(extensions) == 0 {
			fmt.Println("No extensions found.")
			return
		}

		fmt.Println("Browser Extensions:")
		fmt.Println("===================")
		for i, ext := range extensions {
			fmt.Printf("%d. %s\n", i+1, ext.Name)
			fmt.Printf("   Browser: %s\n", ext.Browser)
			fmt.Printf("   Version: %s\n", ext.Version)
			fmt.Printf("   ID: %s\n", ext.ID)
			fmt.Printf("   Enabled: %v\n", ext.Enabled)
			if ext.Profile != "" { // Add Profile to output
				fmt.Printf("   Profile: %s\n", ext.Profile)
			}
			fmt.Println("------------------")
		}
		fmt.Printf("Total extensions: %d\n", len(extensions))
	}
}
