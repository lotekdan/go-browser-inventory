package browsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
					"AppData", "Roaming", "Mozilla", "Firefox",
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

// resolveMessage handles __MSG_ placeholders for extension names
func resolveMessage(msg, basePath, defaultLocale string, debug bool) string {
	msgKey := strings.TrimPrefix(msg, "__MSG_")
	msgKey = strings.TrimSuffix(msgKey, "__")
	lookupKey := strings.ToLower(msgKey) // Lowercase for consistency
	lookupKeyOriginal := msgKey          // Original case for exact match
	localesPath := filepath.Join(basePath, "_locales")
	if debug {
		fmt.Printf("Debug: Resolving %s, lookupKey: %s, lookupKeyOriginal: %s, basePath: %s\n", msgKey, lookupKey, lookupKeyOriginal, basePath)
	}

	if _, err := os.Stat(localesPath); os.IsNotExist(err) {
		if debug {
			fmt.Printf("Note: No _locales directory found at %s for %s\n", localesPath, msgKey)
		}
		return msgKey
	}

	localeDirs, err := os.ReadDir(localesPath)
	if err != nil {
		if debug {
			fmt.Printf("Warning: Failed to read _locales directory %s: %v\n", localesPath, err)
		}
		return msgKey
	}

	// Try default_locale
	if defaultLocale != "" {
		messagesPath := filepath.Join(localesPath, defaultLocale, "messages.json")
		if data, err := os.ReadFile(messagesPath); err == nil {
			var messages map[string]struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(data, &messages); err == nil {
				if debug {
					fmt.Printf("Debug: Parsed %s, keys: %v\n", messagesPath, reflect.ValueOf(messages).MapKeys())
					fmt.Printf("Debug: Checking for key %s (original) and %s (lowercase) in map: %v\n", lookupKeyOriginal, lookupKey, messages)
				}
				// Try original case first
				if val, ok := messages[lookupKeyOriginal]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (default locale, original case)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
				// Then try lowercase
				if val, ok := messages[lookupKey]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (default locale, lowercase)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
				if debug {
					fmt.Printf("Note: Key %s (lookup: %s or %s) not found in %s (default locale)\n", msgKey, lookupKeyOriginal, lookupKey, messagesPath)
				}
			} else if debug {
				fmt.Printf("Warning: Failed to parse %s: %v\n", messagesPath, err)
			}
		} else if debug {
			fmt.Printf("Note: Failed to read %s: %v\n", messagesPath, err)
		}
	}

	// Try English locales
	for _, enLocale := range []string{"en", "en_US"} {
		if enLocale == defaultLocale {
			continue
		}
		messagesPath := filepath.Join(localesPath, enLocale, "messages.json")
		if data, err := os.ReadFile(messagesPath); err == nil {
			var messages map[string]struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(data, &messages); err == nil {
				if debug {
					fmt.Printf("Debug: Parsed %s, keys: %v\n", messagesPath, reflect.ValueOf(messages).MapKeys())
					fmt.Printf("Debug: Checking for key %s (original) and %s (lowercase) in map: %v\n", lookupKeyOriginal, lookupKey, messages)
				}
				// Try original case first
				if val, ok := messages[lookupKeyOriginal]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (English fallback, original case)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
				// Then try lowercase
				if val, ok := messages[lookupKey]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (English fallback, lowercase)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
				if debug {
					fmt.Printf("Note: Key %s (lookup: %s or %s) not found in %s (English fallback)\n", msgKey, lookupKeyOriginal, lookupKey, messagesPath)
				}
			} else if debug {
				fmt.Printf("Warning: Failed to parse %s: %v\n", messagesPath, err)
			}
		} else if debug {
			fmt.Printf("Note: English fallback file %s not found for %s\n", messagesPath, msgKey)
		}
	}

	// Fallback to other locales
	for _, dir := range localeDirs {
		if !dir.IsDir() || dir.Name() == defaultLocale || dir.Name() == "en" || dir.Name() == "en_US" {
			continue
		}
		messagesPath := filepath.Join(localesPath, dir.Name(), "messages.json")
		if data, err := os.ReadFile(messagesPath); err == nil {
			var messages map[string]struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(data, &messages); err == nil {
				if val, ok := messages[lookupKeyOriginal]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (other locale, original case)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
				if val, ok := messages[lookupKey]; ok {
					if debug {
						fmt.Printf("Resolved %s to %s from %s (other locale, lowercase)\n", msgKey, val.Message, messagesPath)
					}
					return val.Message
				}
			}
		}
	}

	if debug {
		fmt.Printf("Note: No matching message found for %s (lookup: %s or %s) in %s\n", msgKey, lookupKeyOriginal, lookupKey, localesPath)
	}
	return msgKey
}
