package browsers

import (
	"os"
	"path/filepath"

	"go-browser-inventory/debug"
	"go-browser-inventory/models"
)

type Brave struct{}

func (b *Brave) Name() string { return "brave" }

func (b *Brave) Detect() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		debug.Printf("Error getting home directory: %v", err)
		return false
	}
	bravePath := filepath.Join(homeDir, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	_, err = os.Stat(bravePath)
	return !os.IsNotExist(err)
}

func (b *Brave) Extensions() ([]models.Extension, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	bravePath := filepath.Join(homeDir, "AppData", "Local", "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	return getChromiumExtensions(bravePath, "brave")
}
