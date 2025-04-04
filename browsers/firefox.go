package browsers

import (
	"encoding/json"
	"os"
	"path/filepath"

	"go-browser-inventory/config"
	"go-browser-inventory/debug"
	"go-browser-inventory/models"
)

type Firefox struct{}

func (f *Firefox) Name() string { return "firefox" }

func (f *Firefox) Detect() bool {
	profiles, _ := filepath.Glob(config.FirefoxExtensionsPath())
	return len(profiles) > 0
}

func (f *Firefox) Extensions() ([]models.Extension, error) {
	profileGlob := config.FirefoxExtensionsPath()
	debug.Printf("Scanning Firefox profiles with glob: %s", profileGlob)
	profiles, err := filepath.Glob(profileGlob)
	if err != nil {
		debug.Printf("Glob error: %v", err)
		return nil, err
	}
	if len(profiles) == 0 {
		debug.Printf("No Firefox profiles matched glob: %s", profileGlob)
		return nil, nil
	}

	var allExtensions []models.Extension
	for _, profile := range profiles {
		debug.Printf("Checking profile: %s", profile)
		data, err := os.ReadFile(profile)
		if err != nil {
			debug.Printf("Error reading %s: %v", profile, err)
			continue
		}

		var addons struct {
			Addons []struct {
				ID            string `json:"id"`
				Name          string `json:"name,omitempty"`
				Version       string `json:"version"`
				Enabled       bool   `json:"active"`
				DefaultLocale struct {
					Name string `json:"name"`
				} `json:"defaultLocale,omitempty"`
			} `json:"addons"`
		}
		if err := json.Unmarshal(data, &addons); err != nil {
			debug.Printf("Error parsing JSON in %s: %v", profile, err)
			continue
		}

		for _, addon := range addons.Addons {
			name := addon.Name
			if name == "" {
				name = addon.DefaultLocale.Name // Fallback to defaultLocale.name
			}
			if addon.ID == "" || name == "" {
				debug.Printf("Skipping invalid addon in %s: ID=%s, Name=%s", profile, addon.ID, name)
				continue
			}
			debug.Printf("Found Firefox extension: ID=%s, Name=%s, Version=%s, Enabled=%v", addon.ID, name, addon.Version, addon.Enabled)
			allExtensions = append(allExtensions, models.Extension{
				ID:      addon.ID,
				Name:    name,
				Version: addon.Version,
				Enabled: addon.Enabled,
			})
		}
	}
	return allExtensions, nil
}
