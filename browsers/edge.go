package browsers

import (
	"go-browser-inventory/config"
	"go-browser-inventory/models"
	"os"
)

type Edge struct{}

func NewEdge() Browser {
	return &Edge{}
}

func (e *Edge) Name() string {
	return "edge"
}

func (e *Edge) Detect() bool {
	_, err := os.Stat(config.EdgeExtensionsPath())
	return err == nil
}

func (e *Edge) Extensions() ([]models.Extension, error) {
	return getChromiumExtensions(config.EdgeExtensionsPath())
}
