package browsers

import (
	"os"
	"path/filepath"

	"go-browser-inventory/debug"
	"go-browser-inventory/models"
)

type Edge struct{}

func (e *Edge) Name() string { return "edge" }

func (e *Edge) Detect() bool {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		debug.Printf("Error getting home directory: %v", err)
		return false
	}
	edgePath := filepath.Join(homeDir, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions")
	_, err = os.Stat(edgePath)
	return !os.IsNotExist(err)
}

func (e *Edge) Extensions() ([]models.Extension, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	edgePath := filepath.Join(homeDir, "AppData", "Local", "Microsoft", "Edge", "User Data", "Default", "Extensions")
	return getChromiumExtensions(edgePath, "edge")
}
