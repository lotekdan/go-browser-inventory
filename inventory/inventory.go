package inventory

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"go-browser-inventory/browsers"
	"go-browser-inventory/debug"
	"go-browser-inventory/models"
)

type Inventory struct {
	Extensions []models.Extension `json:"extensions"`
	Total      int                `json:"total"`
}

// SetDebug enables or disables debug logging
func SetDebug(enabled bool) {
	debug.SetEnabled(enabled)
}

func Generate(outputPath string, browserFilter []string, jsonOutput bool) error {
	browserList := []browsers.Browser{
		&browsers.Chrome{},
		&browsers.Firefox{},
		&browsers.Brave{},
		&browsers.Edge{},
	}

	filterMap := make(map[string]bool)
	for _, b := range browserFilter {
		filterMap[strings.ToLower(b)] = true
	}

	var allExtensions []models.Extension
	for _, browser := range browserList {
		if len(filterMap) > 0 && !filterMap[browser.Name()] {
			continue
		}

		if !browser.Detect() {
			debug.Printf("Browser not detected: %s", browser.Name())
			continue
		}
		debug.Printf("Detected browser: %s", browser.Name())

		exts, err := browser.Extensions()
		if err != nil {
			debug.Printf("Error getting extensions for %s: %v", browser.Name(), err)
			continue
		}
		allExtensions = append(allExtensions, exts...)
	}

	extMap := make(map[string]models.Extension)
	for _, ext := range allExtensions {
		key := ext.ID + "_" + ext.Browser
		if existing, ok := extMap[key]; ok {
			if compareVersions(ext.Version, existing.Version) > 0 {
				extMap[key] = ext
			}
		} else {
			extMap[key] = ext
		}
	}

	var deduped []models.Extension
	for _, ext := range extMap {
		deduped = append(deduped, ext)
	}

	inventory := Inventory{
		Extensions: deduped,
		Total:      len(deduped),
	}

	data, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return err
	}

	debug.Printf("Total extensions found: %d", inventory.Total)
	debug.Printf("JSON inventory written to %s", outputPath)
	return nil
}

func compareVersions(v1, v2 string) int {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		n1, _ := strconv.Atoi(v1Parts[i])
		n2, _ := strconv.Atoi(v2Parts[i])
		if n1 != n2 {
			return n1 - n2
		}
	}
	return len(v1Parts) - len(v2Parts)
}
