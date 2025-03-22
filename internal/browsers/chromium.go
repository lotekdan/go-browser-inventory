package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (bi *BrowserInventory) getChromiumExtensions(basePath string, config BrowserConfig, debug bool) ([]Extension, error) {
	profileBase := filepath.Dir(basePath)
	if _, err := os.Stat(profileBase); os.IsNotExist(err) {
		return nil, fmt.Errorf("profile base directory not found at %s", profileBase)
	}

	profileNames := make(map[string]string)
	localStatePath := filepath.Join(profileBase, "Local State")
	if data, err := os.ReadFile(localStatePath); err == nil {
		var localState struct {
			Profile struct {
				InfoCache map[string]struct {
					Name string `json:"name"`
				} `json:"info_cache"`
			} `json:"profile"`
		}
		if err := json.Unmarshal(data, &localState); err == nil {
			for dir, info := range localState.Profile.InfoCache {
				profileNames[dir] = info.Name
			}
			if debug {
				fmt.Printf("Loaded profile names from Local State: %v\n", profileNames)
			}
		} else if debug {
			fmt.Printf("Warning: Failed to parse Local State at %s: %v\n", localStatePath, err)
		}
	} else if debug {
		fmt.Printf("Note: Local State not found at %s, using directory names\n", localStatePath)
	}

	entries, err := os.ReadDir(profileBase)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile directory: %v", err)
	}

	var allExtensions []Extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		profileDir := entry.Name()
		if profileDir != "Default" && !strings.HasPrefix(profileDir, "Profile") {
			continue
		}

		profileName := profileNames[profileDir]
		if profileName == "" {
			profileName = profileDir
		}

		extensionsPath := filepath.Join(profileBase, profileDir, "Extensions")
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
					Name          string `json:"name"`
					Version       string `json:"version"`
					DefaultLocale string `json:"default_locale"`
				}
				if err := json.Unmarshal(data, &manifest); err != nil {
					if debug {
						fmt.Printf("Warning: Failed to parse manifest %s: %v\n", manifestPath, err)
					}
					continue
				}

				resolvedName := manifest.Name
				if strings.HasPrefix(resolvedName, "__MSG_") {
					resolvedName = resolveMessage(resolvedName, filepath.Join(extensionsPath, extensionID, ver.Name()), manifest.DefaultLocale, debug)
				}

				allExtensions = append(allExtensions, Extension{
					Name:    resolvedName,
					Version: manifest.Version,
					ID:      extensionID,
					Enabled: true,
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
