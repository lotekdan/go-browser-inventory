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

	var profilePath string
	lines := strings.Split(string(iniData), "\n")
	var currentSection string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}
		if strings.HasPrefix(line, "Default=1") && currentSection != "" {
			for _, prevLine := range lines {
				if strings.HasPrefix(prevLine, "Path=") && strings.Contains(prevLine, filepath.Base(basePath)) {
					profilePath = strings.TrimPrefix(prevLine, "Path=")
					break
				}
			}
		}
		if strings.HasPrefix(line, "Path=") && profilePath == "" {
			profilePath = strings.TrimPrefix(line, "Path=")
		}
	}

	if profilePath == "" {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			return nil, fmt.Errorf("no default profile found and failed to read directory: %v", err)
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(entry.Name(), ".default") {
				profilePath = entry.Name()
				break
			}
		}
		if profilePath == "" {
			return nil, fmt.Errorf("no Firefox profile found in %s", basePath)
		}
	}

	if !filepath.IsAbs(profilePath) {
		profilePath = filepath.Join(basePath, profilePath)
	}

	extensionsJSON := filepath.Join(profilePath, "extensions.json")
	data, err := os.ReadFile(extensionsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read extensions.json at %s: %v", extensionsJSON, err)
	}

	var extData struct {
		Addons []struct {
			ID      string `json:"id"`
			Name    string `json:"defaultLocaleName"`
			Version string `json:"version"`
			Active  bool   `json:"active"`
		} `json:"addons"`
	}
	if err := json.Unmarshal(data, &extData); err != nil {
		return nil, fmt.Errorf("failed to parse extensions.json at %s: %v", extensionsJSON, err)
	}

	var extensions []Extension
	for _, addon := range extData.Addons {
		extensions = append(extensions, Extension{
			Name:    addon.Name,
			Version: addon.Version,
			ID:      addon.ID,
			Enabled: addon.Active,
			Browser: config.Name,
		})
	}

	if len(extensions) == 0 && debug {
		fmt.Printf("Note: No extensions found in Firefox profile at %s\n", profilePath)
	}

	return extensions, nil
}
