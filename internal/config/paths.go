package config

import (
	"os"
	"path/filepath"

	"github.com/yar-run/yar/internal/errors"
	"github.com/yar-run/yar/internal/platform"
)

const (
	// ConfigFileName is the name of the global config file
	ConfigFileName = "config.yaml"
	// ProjectFileName is the name of the project config file
	ProjectFileName = "yar.yaml"
)

// GlobalConfigPath returns the path to the global configuration file.
// Uses XDG-compliant directory from platform.ConfigDir().
func GlobalConfigPath() (string, error) {
	configDir, err := platform.ConfigDir()
	if err != nil {
		return "", &errors.ConfigError{
			Path:    "",
			Message: "failed to determine config directory",
			Err:     err,
		}
	}
	return filepath.Join(configDir, ConfigFileName), nil
}

// FindProjectConfig searches for yar.yaml starting from the given directory
// and traversing up to parent directories until found or root is reached.
// If startDir is empty, the current working directory is used.
func FindProjectConfig(startDir string) (string, error) {
	if startDir == "" {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			return "", &errors.ConfigError{
				Path:    "",
				Message: "failed to get working directory",
				Err:     err,
			}
		}
	}

	// Make path absolute
	absDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", &errors.ConfigError{
			Path:    startDir,
			Message: "failed to resolve absolute path",
			Err:     err,
		}
	}

	// Traverse up directory tree
	dir := absDir
	for {
		candidate := filepath.Join(dir, ProjectFileName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return "", &errors.NotFoundError{
		Resource: "file",
		Name:     ProjectFileName,
		Message:  "no yar.yaml found in current directory or any parent",
	}
}
