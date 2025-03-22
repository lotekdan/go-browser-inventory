package main

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
