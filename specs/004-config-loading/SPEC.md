# Iteration 004: Config Loading Specification

## Overview

This iteration implements config file loading, path resolution, and JSON Schema validation for global configuration (`~/.config/yar/config.yaml`) and project configuration (`./yar.yaml`).

## Scope

### Included
- `internal/config/loader.go` - Loader struct with LoadGlobal() and LoadProject()
- `internal/config/paths.go` - GlobalConfigPath() and FindProjectConfig()
- `internal/config/schema.go` - ValidateConfig() and ValidateProject()
- `schemas/config.schema.json` - JSON Schema for global config
- `schemas/project.schema.json` - JSON Schema for project config
- Test fixtures for valid/invalid configs

### NOT Included (deferred)
- Config editing commands (iteration 006)
- Config migration between versions
- Remote config sync

---

## Interfaces

### Loader

```go
// Loader handles loading configuration files.
type Loader struct {
    globalPath  string
    projectPath string
}

// NewLoader creates a new Loader with default paths.
func NewLoader(opts ...LoaderOption) *Loader

// LoadGlobal loads the global configuration.
// Returns defaults if file doesn't exist.
func (l *Loader) LoadGlobal() (*Config, error)

// LoadProject loads the project configuration.
// Returns NotFoundError if no yar.yaml found.
func (l *Loader) LoadProject() (*Project, error)

// GlobalPath returns the path to global config.
func (l *Loader) GlobalPath() (string, error)

// ProjectPath returns the path to project config.
func (l *Loader) ProjectPath() (string, error)
```

### Path Resolution

```go
// GlobalConfigPath returns the path to ~/.config/yar/config.yaml
func GlobalConfigPath() (string, error)

// FindProjectConfig searches current and parent directories for yar.yaml
func FindProjectConfig(startDir string) (string, error)
```

### Schema Validation

```go
// ValidateConfig validates a Config against the JSON Schema.
func ValidateConfig(cfg *Config) error

// ValidateProject validates a Project against the JSON Schema.
func ValidateProject(proj *Project) error
```

---

## Dependencies

### External Packages
- `gopkg.in/yaml.v3` - YAML parsing
- `os` - File operations (stdlib)
- `path/filepath` - Path manipulation (stdlib)

### Internal Packages
- `internal/config` (types.go, defaults.go) - From iteration 003
- `internal/platform` - For ConfigDir()
- `internal/errors` - For NotFoundError, ConfigError

---

## Invariants

- **INV-CFG-001**: All configuration files MUST validate against their JSON Schema before use
- **INV-CFG-002**: Project config MUST be present for project-scoped commands
- **INV-CFG-003**: Global config is optional; sensible defaults apply when absent

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/config/loader.go` | Loader struct and Load methods |
| `internal/config/paths.go` | Path resolution functions |
| `internal/config/schema.go` | Schema validation functions |
| `internal/config/loader_test.go` | Unit tests for loader |
| `internal/config/paths_test.go` | Unit tests for paths |
| `schemas/config.schema.json` | JSON Schema for global config |
| `schemas/project.schema.json` | JSON Schema for project config |
| `internal/config/testdata/` | Test fixtures |

---

## Test Requirements

### Unit Tests
- [ ] Test LoadGlobal returns defaults when file missing
- [ ] Test LoadGlobal loads valid config file
- [ ] Test LoadGlobal returns error on invalid YAML
- [ ] Test LoadGlobal returns error on schema violation
- [ ] Test LoadProject returns NotFoundError when missing
- [ ] Test LoadProject loads valid project file
- [ ] Test LoadProject returns error on missing required field
- [ ] Test LoadProject searches parent directories
- [ ] Test GlobalConfigPath returns correct path
- [ ] Test FindProjectConfig finds yar.yaml in current dir
- [ ] Test FindProjectConfig finds yar.yaml in parent dir
- [ ] Test FindProjectConfig returns error when not found

### Test Fixtures
- `internal/config/testdata/valid/config.yaml`
- `internal/config/testdata/valid/project.yaml`
- `internal/config/testdata/invalid/config-bad-yaml.yaml`
- `internal/config/testdata/invalid/project-missing-field.yaml`

---

## Exit Criteria

- [ ] Can load valid config files
- [ ] Invalid configs produce validation errors
- [ ] Missing global config returns defaults
- [ ] Missing project config returns clear error
- [ ] Unit tests pass
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
