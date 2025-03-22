package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getFirefoxExtensions handles Firefox extensions
func (bi *BrowserInventory) getFirefoxExtensions(basePath string, config BrowserConfig, debug bool) ([]Extension, error) {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profiles directory not found at %s", basePath)
	}

	profilesIni := filepath.Join(basePath, "profiles.ini")
	iniData, err := os.ReadFile(profilesIni)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles.ini at %s: %v", profilesIni, err)
	}

	lines := strings.Split(string(iniData), "\n")
	var profiles []string
	var currentSection string
	var defaultProfilePath string

	// Collect all profiles and find default
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}
		if strings.HasPrefix(line, "Path=") && currentSection != "" {
			profile := strings.TrimPrefix(line, "Path=")
			profiles = append(profiles, profile)
			if debug {
				fmt.Printf("Found profile in profiles.ini: %s\n", profile)
			}
		}
		if strings.HasPrefix(line, "Default=1") && currentSection != "" {
			for _, prevLine := range lines {
				if strings.HasPrefix(prevLine, "Path=") {
					defaultProfilePath = strings.TrimPrefix(prevLine, "Path=")
					if debug {
						fmt.Printf("Found default profile in profiles.ini: %s\n", defaultProfilePath)
					}
					break
				}
			}
		}
	}

	var allExtensions []Extension
	for _, profilePath := range profiles {
		if !filepath.IsAbs(profilePath) {
			profilePath = filepath.Join(basePath, profilePath)
		}
		if debug {
			fmt.Printf("Checking profile: %s\n", profilePath)
		}

		extensionsJSON := filepath.Join(profilePath, "extensions.json")
		data, err := os.ReadFile(extensionsJSON)
		if err != nil {
			if os.IsNotExist(err) {
				if debug {
					fmt.Printf("Note: extensions.json not found at %s, skipping profile\n", extensionsJSON)
				}
				continue
			}
			return nil, fmt.Errorf("failed to read extensions.json at %s: %v", extensionsJSON, err)
		}

		var extData struct {
			Addons []struct {
				ID            string `json:"id"`
				Version       string `json:"version"`
				Active        bool   `json:"active"`
				DefaultLocale struct {
					Name string `json:"name"`
				} `json:"defaultLocale"`
			} `json:"addons"`
		}
		if err := json.Unmarshal(data, &extData); err != nil {
			return nil, fmt.Errorf("failed to parse extensions.json at %s: %v", extensionsJSON, err)
		}

		for _, addon := range extData.Addons {
			profileName := filepath.Base(profilePath) // Extract profile name
			allExtensions = append(allExtensions, Extension{
				Name:    addon.DefaultLocale.Name,
				Version: addon.Version,
				ID:      addon.ID,
				Enabled: addon.Active,
				Browser: config.Name,
				Profile: profileName,
			})
		}
	}

	if len(allExtensions) == 0 && debug {
		fmt.Printf("Note: No extensions found across all profiles in %s\n", basePath)
	}

	return allExtensions, nil
}
