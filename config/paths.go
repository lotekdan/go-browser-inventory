package config

import (
	"os"
	"path/filepath"
	"runtime"
)

func ChromeExtensionsPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default", "Extensions")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Google", "Chrome", "Default", "Extensions")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "google-chrome", "Default", "Extensions")
	default:
		return ""
	}
}

func FirefoxExtensionsPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles", "*", "extensions.json")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Firefox", "Profiles", "*", "extensions.json")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".mozilla", "firefox", "*", "extensions.json")
	default:
		return ""
	}
}

func EdgeExtensionsPath() string {
	switch runtime.GOOS {
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "microsoft-edge", "Default", "Extensions")
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Edge", "User Data", "Default", "Extensions")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "Microsoft Edge", "Default", "Extensions")
	default:
		return ""
	}
}
func BraveExtensionsPath() string {
	switch runtime.GOOS {
	case "windows":
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "BraveSoftware", "Brave-Browser", "User Data", "Default", "Extensions")
	case "darwin":
		return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "BraveSoftware", "Brave-Browser", "Default", "Extensions")
	case "linux":
		return filepath.Join(os.Getenv("HOME"), ".config", "BraveSoftware", "Brave-Browser", "Default", "Extensions")
	default:
		return ""
	}
}
