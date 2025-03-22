package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getChromiumExtensions handles Chromium-based browser extensions (Chrome, Edge)
func (bi *BrowserInventory) getChromiumExtensions(basePath string, config BrowserConfig, debug bool) ([]Extension, error) {
	// Base directory is one level up from the profile (e.g., ~/.config/google-chrome)
	profileBase := filepath.Dir(basePath)
	if _, err := os.Stat(profileBase); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile base directory not found at %s", profileBase)
	}

	// Scan for profile directories (e.g., Default, Profile 1)
	entries, err := os.ReadDir(profileBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile directory: %v", err)
	}

	var allExtensions []Extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		profileName := entry.Name()
		if profileName != "Default" && !strings.HasPrefix(profileName, "Profile") {
			continue // Skip non-profile directories
		}

		extensionsPath := filepath.Join(profileBase, profileName, "Extensions")
		if _, err := os.Stat(extensionsPath); os.IsNotExist(err) {
			if debug {
				fmt.Printf("Note: Extensions directory not found at %s, skipping profile %s\n", extensionsPath, profileName)
			}
			continue
		}

		if debug {
			fmt.Printf("Resolved extensions path for profile %s: %s\n", profileName, extensionsPath)
		}

		dirs, err := os.ReadDir(extensionsPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read extensions directory %s: %v", extensionsPath, err)
		}

		for _, dir := range dirs {
			if !dir.IsDir() {
				continue
			}
			extensionID := dir.Name()
			versions, err := os.ReadDir(filepath.Join(extensionsPath, extensionID))
			if err != nil {
				if debug {
					fmt.Printf("Warning: Failed to read version directory for %s: %v\n", extensionID, err)
				}
				continue
			}

			for _, ver := range versions {
				if !ver.IsDir() {
					continue
				}
				manifestPath := filepath.Join(extensionsPath, extensionID, ver.Name(), config.ManifestFile)
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					if debug {
						fmt.Printf("Warning: Failed to read manifest %s: %v\n", manifestPath, err)
					}
					continue
				}

				var manifest struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}
				if err := json.Unmarshal(data, &manifest); err != nil {
					if debug {
						fmt.Printf("Warning: Failed to parse manifest %s: %v\n", manifestPath, err)
					}
					continue
				}

				resolvedName := manifest.Name
				if strings.HasPrefix(resolvedName, "__MSG_") {
					resolvedName = resolveMessage(resolvedName, filepath.Join(extensionsPath, extensionID, ver.Name()), debug)
				}

				allExtensions = append(allExtensions, Extension{
					Name:    resolvedName,
					Version: manifest.Version,
					ID:      extensionID,
					Enabled: true, // Chromium assumes enabled if present
					Browser: config.Name,
					Profile: profileName,
				})
			}
		}
	}

	if len(allExtensions) == 0 {
		if debug {
			fmt.Printf("Note: No extensions found across profiles in %s\n", profileBase)
		}
	}

	return allExtensions, nil
}

// resolveMessage handles __MSG_ placeholders in Chromium manifest names
func resolveMessage(msg, basePath string, debug bool) string {
	msg = strings.TrimPrefix(msg, "__MSG_")
	msg = strings.TrimSuffix(msg, "__")
	localesPath := filepath.Join(basePath, "locales")
	localeDirs, err := os.ReadDir(localesPath)
	if err != nil {
		if debug {
			fmt.Printf("Warning: Failed to read locales directory %s: %v\n", localesPath, err)
		}
		return msg
	}

	for _, dir := range localeDirs {
		if !dir.IsDir() {
			continue
		}
		messagesPath := filepath.Join(localesPath, dir.Name(), "messages.json")
		data, err := os.ReadFile(messagesPath)
		if err != nil {
			continue
		}

		var messages map[string]struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(data, &messages); err != nil {
			continue
		}

		if val, ok := messages[msg]; ok {
			return val.Message
		}
	}

	return msg
}
