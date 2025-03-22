package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getChromiumExtensions handles Chrome and Edge extensions
func (bi *BrowserInventory) getChromiumExtensions(basePath string, config BrowserConfig, debug bool) ([]Extension, error) {
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
			data, err := os.ReadFile(path)
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

			extPath := filepath.Dir(path)
			resolvedName := bi.resolveMessage(manifest.Name, extPath, debug)

			extensionID := filepath.Base(filepath.Dir(filepath.Dir(path)))
			extensions = append(extensions, Extension{
				Name:    resolvedName,
				Version: manifest.Version,
				ID:      extensionID,
				Enabled: true,
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

// resolveMessage resolves __MSG_ placeholders to human-readable names
func (bi *BrowserInventory) resolveMessage(name, extPath string, debug bool) string {
	if !strings.HasPrefix(name, "__MSG_") || !strings.HasSuffix(name, "__") {
		return name
	}

	msgKey := strings.TrimPrefix(strings.TrimSuffix(name, "__"), "__MSG_")
	localesPath := filepath.Join(extPath, "_locales")

	keyVariations := []string{
		msgKey,
		strings.ToLower(msgKey),
		"appName",
		"extensionName",
		strings.ToLower(msgKey[:1]) + msgKey[1:],
	}

	if debug {
		fmt.Printf("Resolving message key '%s' for extension at %s\n", msgKey, extPath)
	}

	// Try default locale (en) first
	enMessagesPath := filepath.Join(localesPath, "en", "messages.json")
	if data, err := os.ReadFile(enMessagesPath); err == nil {
		var messages map[string]struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(data, &messages); err == nil {
			for _, key := range keyVariations {
				if msg, ok := messages[key]; ok {
					if debug {
						fmt.Printf("Found message '%s' for key '%s' in en/messages.json\n", msg.Message, key)
					}
					return msg.Message
				}
			}
			if debug {
				fmt.Printf("No matching key found in en/messages.json for variations: %v\n", keyVariations)
			}
		} else if debug {
			fmt.Printf("Failed to parse en/messages.json: %v\n", err)
		}
	} else if debug {
		fmt.Printf("No en/messages.json found at %s: %v\n", enMessagesPath, err)
	}

	// Try any available locale
	localeDirs, err := os.ReadDir(localesPath)
	if err != nil {
		if debug {
			fmt.Printf("Failed to read _locales directory: %v\n", err)
		}
		return name
	}

	for _, dir := range localeDirs {
		if dir.IsDir() {
			messagesPath := filepath.Join(localesPath, dir.Name(), "messages.json")
			data, err := os.ReadFile(messagesPath)
			if err != nil {
				if debug {
					fmt.Printf("Skipping %s: %v\n", messagesPath, err)
				}
				continue
			}
			var messages map[string]struct {
				Message string `json:"message"`
			}
			if err := json.Unmarshal(data, &messages); err == nil {
				for _, key := range keyVariations {
					if msg, ok := messages[key]; ok {
						if debug {
							fmt.Printf("Found message '%s' for key '%s' in %s/messages.json\n", msg.Message, key, dir.Name())
						}
						return msg.Message
					}
				}
				if debug {
					fmt.Printf("No matching key found in %s/messages.json for variations: %v\n", dir.Name(), keyVariations)
				}
			} else if debug {
				fmt.Printf("Failed to parse %s: %v\n", messagesPath, err)
			}
		}
	}

	if debug {
		fmt.Printf("Could not resolve '%s', returning original name\n", name)
	}
	return name
}
