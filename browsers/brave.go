package browsers

import (
	"os"

	"go-browser-inventory/config"
	"go-browser-inventory/models"
)

type Brave struct{}

func NewBrave() Browser {
	return &Brave{}
}

func (b *Brave) Name() string {
	return "brave"
}

func (b *Brave) Detect() bool {
	_, err := os.Stat(config.BraveExtensionsPath())
	return err == nil
}

func (b *Brave) Extensions() ([]models.Extension, error) {
	return getChromiumExtensions(config.BraveExtensionsPath())
}
