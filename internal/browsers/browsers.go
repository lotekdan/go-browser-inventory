package browsers

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

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
func (bi *BrowserInventory) GetExtensions(selectedBrowser string, debug bool) ([]Extension, error) {
	var allExtensions []Extension

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	for _, config := range bi.configs {
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
			if debug {
				fmt.Printf("Warning: Unsupported OS %s for %s\n", runtime.GOOS, config.Name)
			}
			continue
		}

		var exts []Extension
		if config.IsFirefox {
			exts, err = bi.getFirefoxExtensions(basePath, config, debug)
		} else {
			exts, err = bi.getChromiumExtensions(basePath, config, debug)
		}
		if err != nil {
			fmt.Printf("Warning: Failed to get %s extensions: %v\n", config.Name, err)
			continue
		}
		allExtensions = append(allExtensions, exts...)
	}

	return allExtensions, nil
}
