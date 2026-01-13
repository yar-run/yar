package config

import (
	"os"

	"github.com/yar-run/yar/internal/errors"
	"gopkg.in/yaml.v3"
)

// Loader loads and validates configuration files
type Loader struct {
	globalPath  string
	projectPath string
}

// Option configures a Loader
type Option func(*Loader)

// WithGlobalPath sets a custom global config path
func WithGlobalPath(path string) Option {
	return func(l *Loader) {
		l.globalPath = path
	}
}

// WithProjectPath sets a custom project config path
func WithProjectPath(path string) Option {
	return func(l *Loader) {
		l.projectPath = path
	}
}

// NewLoader creates a new configuration loader
func NewLoader(opts ...Option) *Loader {
	l := &Loader{}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// GlobalPath returns the path to the global configuration file
func (l *Loader) GlobalPath() (string, error) {
	if l.globalPath != "" {
		return l.globalPath, nil
	}
	return GlobalConfigPath()
}

// ProjectPath returns the path to the project configuration file
func (l *Loader) ProjectPath() (string, error) {
	if l.projectPath != "" {
		return l.projectPath, nil
	}
	return FindProjectConfig("")
}

// LoadGlobal loads global configuration from file.
// Returns default configuration if file doesn't exist.
func (l *Loader) LoadGlobal() (*Config, error) {
	path, err := l.GlobalPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Return defaults when file doesn't exist
			return DefaultConfig(), nil
		}
		return nil, &errors.ConfigError{
			Path:    path,
			Message: "failed to read config file",
			Err:     err,
		}
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, &errors.ConfigError{
			Path:    path,
			Message: "failed to parse YAML",
			Err:     err,
		}
	}

	// Validate against schema
	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadProject loads project configuration from file.
// Returns NotFoundError if file doesn't exist.
func (l *Loader) LoadProject() (*Project, error) {
	path, err := l.ProjectPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &errors.NotFoundError{
				Resource: "file",
				Name:     path,
				Message:  "project configuration file not found",
			}
		}
		return nil, &errors.ConfigError{
			Path:    path,
			Message: "failed to read project file",
			Err:     err,
		}
	}

	var proj Project
	if err := yaml.Unmarshal(data, &proj); err != nil {
		return nil, &errors.ConfigError{
			Path:    path,
			Message: "failed to parse YAML",
			Err:     err,
		}
	}

	// Validate against schema
	if err := ValidateProject(&proj); err != nil {
		return nil, err
	}

	return &proj, nil
}
