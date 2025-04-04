package browsers

import "go-browser-inventory/models"

type Browser interface {
	Name() string
	Extensions() ([]models.Extension, error)
	Detect() bool
}
