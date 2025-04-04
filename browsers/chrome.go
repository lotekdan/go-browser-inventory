package browsers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"go-browser-inventory/config"
	"go-browser-inventory/debug"
	"go-browser-inventory/models"
)

type Chrome struct{}

func (c *Chrome) Name() string { return "chrome" }

func (c *Chrome) Detect() bool {
	_, err := os.Stat(config.ChromeExtensionsPath())
	return !os.IsNotExist(err)
}

func (c *Chrome) Extensions() ([]models.Extension, error) {
	return getChromiumExtensions(config.ChromeExtensionsPath(), "chrome")
}

func getChromiumExtensions(extDir, browserName string) ([]models.Extension, error) {
	if extDir == "" {
		debug.Printf("Extension directory is empty for %s", browserName)
		return nil, nil
	}
	debug.Printf("Scanning %s extensions in: %s", browserName, extDir)

	var extensions []models.Extension
	err := filepath.Walk(extDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			debug.Printf("Error walking path %s: %v", path, err)
			return nil
		}
		if info.IsDir() || filepath.Base(path) != "manifest.json" {
			return nil
		}

		debug.Printf("Processing manifest: %s", path)
		data, err := os.ReadFile(path)
		if err != nil {
			debug.Printf("Error reading %s: %v", path, err)
			return nil
		}

		var manifest struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := json.Unmarshal(data, &manifest); err != nil {
			debug.Printf("Error parsing JSON in %s: %v", path, err)
			return nil
		}

		parts := strings.Split(filepath.Dir(path), string(filepath.Separator))
		if len(parts) < 2 {
			debug.Printf("Invalid path structure for %s", path)
			return nil
		}
		extID := parts[len(parts)-2]
		if len(extID) != 32 {
			return nil
		}

		name := manifest.Name
		if strings.HasPrefix(name, "__MSG_") {
			msgKey := strings.TrimSuffix(strings.TrimPrefix(name, "__MSG_"), "__")
			localePath := filepath.Join(filepath.Dir(path), "_locales", "en", "messages.json")
			if resolvedName, ok := getMessageName(localePath, msgKey); ok {
				name = resolvedName
			} else {
				localeDir := filepath.Join(filepath.Dir(path), "_locales")
				dirs, err := os.ReadDir(localeDir)
				if err == nil {
					for _, dir := range dirs {
						if dir.IsDir() && dir.Name() != "en" {
							localePath = filepath.Join(localeDir, dir.Name(), "messages.json")
							if resolvedName, ok := getMessageName(localePath, msgKey); ok {
								name = resolvedName
								break
							}
						}
					}
				}
				if strings.HasPrefix(name, "__MSG_") {
					name = extID // Final fallback
					debug.Printf("Using fallback name for %s: %s (original: %s)", path, name, manifest.Name)
				}
			}
		} else if name == "" {
			name = extID
			debug.Printf("Using fallback name for %s: %s (no name in manifest)", path, name)
		}

		debug.Printf("Found valid extension: ID=%s, Name=%s, Version=%s", extID, name, manifest.Version)
		extensions = append(extensions, models.Extension{
			ID:      extID,
			Name:    name,
			Version: manifest.Version,
			Enabled: true, // Chromium assumes enabled unless disabled via policy
			Browser: browserName,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return extensions, nil
}

func getMessageName(localePath, key string) (string, bool) {
	data, err := os.ReadFile(localePath)
	if err != nil {
		debug.Printf("Error reading locale file %s: %v", localePath, err)
		return "", false
	}

	var messages map[string]struct {
		Message string `json:"message"`
	}
	if err := json.Unmarshal(data, &messages); err != nil {
		debug.Printf("Error parsing locale JSON %s: %v", localePath, err)
		return "", false
	}

	if msg, ok := messages[key]; ok && msg.Message != "" {
		return msg.Message, true
	}
	return "", false
}
