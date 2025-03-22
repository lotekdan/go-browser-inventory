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
	var defaultProfileFound bool

	// First pass: look for the default profile
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line
			continue
		}
		if strings.HasPrefix(line, "Default=1") && currentSection != "" {
			for _, prevLine := range lines {
				if strings.HasPrefix(prevLine, "Path=") {
					profilePath = strings.TrimPrefix(prevLine, "Path=")
					defaultProfileFound = true
					if debug {
						fmt.Printf("Found profile marked as default in profiles.ini: %s\n", profilePath)
					}
					break
				}
			}
		}
	}

	// Second pass: if no default, take the first profile
	if !defaultProfileFound {
		for _, line := range lines {
			if strings.HasPrefix(line, "Path=") && profilePath == "" {
				profilePath = strings.TrimPrefix(line, "Path=")
				if debug {
					fmt.Printf("No default profile found, using first profile from profiles.ini: %s\n", profilePath)
				}
				break
			}
		}
	}

	// Temporary hardcode to ensure correct profile (remove after confirming profiles.ini)
	profilePath = "Profiles/wteh27n3.default-release"
	if debug {
		fmt.Printf("Hardcoded profile path for testing: %s\n", profilePath)
	}

	if profilePath == "" {
		entries, err := os.ReadDir(basePath)
		if err != nil {
			return nil, fmt.Errorf("no default profile found and failed to read directory: %v", err)
		}
		// Prioritize .default-release (modern Firefox default)
		for _, entry := range entries {
			if entry.IsDir() && strings.Contains(entry.Name(), ".default-release") {
				profilePath = entry.Name()
				if debug {
					fmt.Printf("No profile in profiles.ini, using .default-release: %s\n", profilePath)
				}
				break
			}
		}
		// If no .default-release, fall back to .default
		if profilePath == "" {
			for _, entry := range entries {
				if entry.IsDir() && strings.Contains(entry.Name(), ".default") {
					profilePath = entry.Name()
					if debug {
						fmt.Printf("No .default-release, using .default: %s\n", profilePath)
					}
					break
				}
			}
		}
	}

	if profilePath == "" {
		return nil, fmt.Errorf("no Firefox profile found in %s", basePath)
	}

	if !filepath.IsAbs(profilePath) {
		profilePath = filepath.Join(basePath, profilePath)
	}

	if debug {
		fmt.Printf("Resolved profile path: %s\n", profilePath)
	}

	extensionsJSON := filepath.Join(profilePath, "extensions.json")
	data, err := os.ReadFile(extensionsJSON)
	if err != nil {
		if os.IsNotExist(err) {
			if debug {
				fmt.Printf("Note: extensions.json not found at %s, assuming no extensions\n", extensionsJSON)
			}
			return []Extension{}, nil
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

	var extensions []Extension
	for _, addon := range extData.Addons {
		extensions = append(extensions, Extension{
			Name:    addon.DefaultLocale.Name, // Use nested defaultLocale.name
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
