package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Extension represents a browser extension
type Extension struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Browser string `json:"browser"`
}

// BrowserConfig defines browser-specific configuration
type BrowserConfig struct {
	Name         string
	WindowsPath  []string
	MacOSPath    []string
	LinuxPath    []string
	IsFirefox    bool
	ManifestFile string
}

// BrowserInventory holds the utility's main functionality
type BrowserInventory struct {
	configs []BrowserConfig
}

// InventoryOutput struct for JSON output
type InventoryOutput struct {
	Extensions []Extension `json:"extensions"`
	Total      int         `json:"total"`
}

// NewBrowserInventory creates a new inventory instance
func NewBrowserInventory() *BrowserInventory {
	return &BrowserInventory{
		configs: []BrowserConfig{
			{
				Name: "Chrome",
				WindowsPath: []string{
					"AppData", "Local", "Google", "Chrome", "User Data", "Default",
				},
				MacOSPath: []string{
					"Library", "Application Support", "Google", "Chrome", "Default",
				},
				LinuxPath: []string{
					".config", "google-chrome", "Default",
				},
				IsFirefox:    false,
				ManifestFile: "manifest.json",
			},
			{
				Name: "Edge",
				WindowsPath: []string{
					"AppData", "Local", "Microsoft", "Edge", "User Data", "Default",
				},
				MacOSPath: []string{
					"Library", "Application Support", "Microsoft Edge", "Default",
				},
				LinuxPath: []string{
					".config", "microsoft-edge", "Default",
				},
				IsFirefox:    false,
				ManifestFile: "manifest.json",
			},
			{
				Name: "Firefox",
				WindowsPath: []string{
					"AppData", "Roaming", "Mozilla", "Firefox", "Profiles",
				},
				MacOSPath: []string{
					"Library", "Application Support", "Firefox", "Profiles",
				},
				LinuxPath: []string{
					".mozilla", "firefox",
				},
				IsFirefox:    true,
				ManifestFile: "manifest.json",
			},
		},
	}
}

// GetExtensions retrieves extensions based on browser selection
func (bi *BrowserInventory) GetExtensions(selectedBrowser string) ([]Extension, error) {
	var allExtensions []Extension

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	for _, config := range bi.configs {
		// Skip if a specific browser is selected and it doesn't match
		if selectedBrowser != "" && strings.ToLower(config.Name) != strings.ToLower(selectedBrowser) {
			continue
		}

		var basePath string
		switch runtime.GOOS {
		case "windows":
			basePath = filepath.Join(homeDir, filepath.Join(config.WindowsPath...))
		case "darwin": // macOS
			basePath = filepath.Join(homeDir, filepath.Join(config.MacOSPath...))
		case "linux":
			basePath = filepath.Join(homeDir, filepath.Join(config.LinuxPath...))
		default:
			fmt.Printf("Warning: Unsupported OS %s for %s\n", runtime.GOOS, config.Name)
			continue
		}

		exts, err := bi.getBrowserExtensions(basePath, config)
		if err != nil {
			fmt.Printf("Warning: Failed to get %s extensions: %v\n", config.Name, err)
			continue
		}
		allExtensions = append(allExtensions, exts...)
	}

	return allExtensions, nil
}

// getBrowserExtensions handles extension retrieval for a specific browser
func (bi *BrowserInventory) getBrowserExtensions(basePath string, config BrowserConfig) ([]Extension, error) {
	if config.IsFirefox {
		return bi.getFirefoxExtensions(basePath, config)
	}
	return bi.getChromiumExtensions(basePath, config)
}

// getChromiumExtensions handles Chrome and Edge extensions
func (bi *BrowserInventory) getChromiumExtensions(basePath string, config BrowserConfig) ([]Extension, error) {
	extensionsPath := filepath.Join(basePath, "Extensions")
	if _, err := os.Stat(extensionsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("extensions directory not found")
	}

	var extensions []Extension
	err := filepath.Walk(extensionsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.Name() == config.ManifestFile {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil // Skip this extension
			}

			var manifest struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			}
			if err := json.Unmarshal(data, &manifest); err != nil {
				return nil // Skip this extension
			}

			extensionID := filepath.Base(filepath.Dir(filepath.Dir(path)))
			extensions = append(extensions, Extension{
				Name:    manifest.Name,
				Version: manifest.Version,
				ID:      extensionID,
				Enabled: true, // Chromium doesn't store enabled state here
				Browser: config.Name,
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning extensions: %v", err)
	}
	return extensions, nil
}

// getFirefoxExtensions handles Firefox extensions
func (bi *BrowserInventory) getFirefoxExtensions(basePath string, config BrowserConfig) ([]Extension, error) {
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("profiles directory not found")
	}

	profilesIni := filepath.Join(basePath, "profiles.ini")
	iniData, err := ioutil.ReadFile(profilesIni)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles.ini: %v", err)
	}

	var profilePath string
	lines := strings.Split(string(iniData), "\n")
	for i, line := range lines {
		if strings.Contains(line, "Default=") && !strings.Contains(line, "Default=0") {
			for j := i - 1; j >= 0; j-- {
				if strings.HasPrefix(lines[j], "Path=") {
					profilePath = strings.TrimPrefix(lines[j], "Path=")
					break
				}
			}
			break
		}
	}

	if profilePath == "" {
		return nil, fmt.Errorf("no default profile found")
	}

	extensionsJSON := filepath.Join(basePath, profilePath, "extensions.json")
	data, err := ioutil.ReadFile(extensionsJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to read extensions.json: %v", err)
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
		return nil, fmt.Errorf("failed to parse extensions.json: %v", err)
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

	return extensions, nil
}

func main() {
	// Define command-line flag
	browserPtr := flag.String("browser", "", "Specify a browser to scan (chrome, edge, firefox). Leave empty for all.")
	flag.Parse()

	// Validate browser selection
	selectedBrowser := *browserPtr
	if selectedBrowser != "" {
		validBrowsers := map[string]bool{
			"chrome":  true,
			"edge":    true,
			"firefox": true,
		}
		if !validBrowsers[strings.ToLower(selectedBrowser)] {
			fmt.Printf("Error: Invalid browser '%s'. Use 'chrome', 'edge', or 'firefox'.\n", selectedBrowser)
			fmt.Println("Usage:")
			flag.PrintDefaults()
			os.Exit(1)
		}
	}

	inventory := NewBrowserInventory()
	extensions, err := inventory.GetExtensions(selectedBrowser)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	output := InventoryOutput{
		Extensions: extensions,
		Total:      len(extensions),
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
