package platform

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// OS represents the operating system
type OS string

const (
	Darwin  OS = "darwin"
	Linux   OS = "linux"
	Windows OS = "windows"
)

// Platform returns the current operating system
func Platform() OS {
	return OS(runtime.GOOS)
}

// HomeDir returns the user's home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}

// ConfigDir returns the XDG-compliant config directory for yar
// - macOS: $XDG_CONFIG_HOME/yar or ~/.config/yar
// - Linux: $XDG_CONFIG_HOME/yar or ~/.config/yar
// - Windows: %APPDATA%\yar
func ConfigDir() (string, error) {
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "yar"), nil
	}

	// Unix-like systems (macOS, Linux)
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "yar"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "yar"), nil
}

// CacheDir returns the XDG-compliant cache directory for yar
// - macOS: $XDG_CACHE_HOME/yar or ~/Library/Caches/yar
// - Linux: $XDG_CACHE_HOME/yar or ~/.cache/yar
// - Windows: %LOCALAPPDATA%\yar\cache
func CacheDir() (string, error) {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "yar", "cache"), nil
	}

	// Unix-like systems (macOS, Linux)
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "yar"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// macOS uses ~/Library/Caches by default
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Caches", "yar"), nil
	}

	// Linux uses ~/.cache
	return filepath.Join(home, ".cache", "yar"), nil
}

// DataDir returns the XDG-compliant data directory for yar
// - macOS: $XDG_DATA_HOME/yar or ~/Library/Application Support/yar
// - Linux: $XDG_DATA_HOME/yar or ~/.local/share/yar
// - Windows: %LOCALAPPDATA%\yar\data
func DataDir() (string, error) {
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "yar", "data"), nil
	}

	// Unix-like systems (macOS, Linux)
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "yar"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// macOS uses ~/Library/Application Support by default
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "yar"), nil
	}

	// Linux uses ~/.local/share
	return filepath.Join(home, ".local", "share", "yar"), nil
}

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	// Expand tilde
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path, nil
}
