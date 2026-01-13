# Iteration 002: Configuration Types & Loading

## Overview

This iteration implements the configuration system for yar: type definitions for global and project configuration, YAML loading, JSON Schema validation, and integration with CLI commands. This enables `yar config get` and `yar project get` to display real configuration data.

## Scope

### Included
- **Config Types**: All structs for global config (`~/.config/yar/config.yaml`)
- **Project Types**: All structs for project config (`./yar.yaml`)
- **YAML Loading**: Load and parse configuration files with YAML tags
- **JSON Schema Validation**: Validate configs against JSON Schema
- **Path Resolution**: Find config files in correct locations
- **CLI Integration**: `yar config get` and `yar project get` display real config

### NOT Included (deferred)
- Config editing (iteration 003)
- Secret provider implementations (iteration 013+)
- Pack loading (iteration 018)
- Environment resolution (iteration 023)

---

## Interfaces

### Configuration Types

```go
// Config represents global yar configuration
type Config struct {
    Container string         `yaml:"container" json:"container"`
    VPN       *VPNConfig     `yaml:"vpn,omitempty" json:"vpn,omitempty"`
    Hosts     *HostsConfig   `yaml:"hosts,omitempty" json:"hosts,omitempty"`
    Network   *NetworkConfig `yaml:"network,omitempty" json:"network,omitempty"`
    Secrets   *SecretsConfig `yaml:"secrets,omitempty" json:"secrets,omitempty"`
    Clusters  map[string]*ClusterConfig `yaml:"clusters,omitempty" json:"clusters,omitempty"`
}

type VPNConfig struct {
    Provider   string `yaml:"provider" json:"provider"`
    ConfigPath string `yaml:"configPath" json:"configPath"`
}

type HostsConfig struct {
    Mode   string `yaml:"mode" json:"mode"`
    Suffix string `yaml:"suffix,omitempty" json:"suffix,omitempty"`
}

type NetworkConfig struct {
    Name string `yaml:"name" json:"name"`
    CIDR string `yaml:"cidr" json:"cidr"`
}

type SecretsConfig struct {
    Local     *LocalSecretConfig            `yaml:"local" json:"local"`
    Providers map[string]*SecretProviderConfig `yaml:"providers,omitempty" json:"providers,omitempty"`
}

type LocalSecretConfig struct {
    Provider string `yaml:"provider" json:"provider"`
    Store    string `yaml:"store,omitempty" json:"store,omitempty"`
    Fallback bool   `yaml:"fallback" json:"fallback"`
}

type SecretProviderConfig struct {
    Type string `yaml:"type" json:"type"`
    // Provider-specific fields stored as map for flexibility
    Config map[string]any `yaml:",inline" json:",inline"`
}

type ClusterConfig struct {
    Provider  string `yaml:"provider" json:"provider"`
    Context   string `yaml:"context,omitempty" json:"context,omitempty"`
    Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}
```

### Project Types

```go
// Project represents project configuration (yar.yaml)
type Project struct {
    Project      string                    `yaml:"project" json:"project"`
    Environments map[string]*Environment   `yaml:"environments" json:"environments"`
    Services     []*Service                `yaml:"services" json:"services"`
}

type Environment struct {
    Cluster string `yaml:"cluster" json:"cluster"`
    Secrets string `yaml:"secrets" json:"secrets"`
}

type Service struct {
    Name       string            `yaml:"name" json:"name"`
    Namespace  string            `yaml:"namespace,omitempty" json:"namespace,omitempty"`
    Pack       string            `yaml:"pack" json:"pack"`
    Requires   []string          `yaml:"requires,omitempty" json:"requires,omitempty"`
    Replicas   int               `yaml:"replicas,omitempty" json:"replicas,omitempty"`
    Params     map[string]any    `yaml:"params,omitempty" json:"params,omitempty"`
    Ingress    *IngressConfig    `yaml:"ingress,omitempty" json:"ingress,omitempty"`
    Env        map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
    SecretRefs map[string]string `yaml:"secretRefs,omitempty" json:"secretRefs,omitempty"`
}

type IngressConfig struct {
    Host string `yaml:"host" json:"host"`
    Path string `yaml:"path,omitempty" json:"path,omitempty"`
    TLS  bool   `yaml:"tls,omitempty" json:"tls,omitempty"`
}
```

### Loader Interface

```go
// Loader loads and validates configuration
type Loader struct {
    globalPath  string
    projectPath string
}

// NewLoader creates a new config loader with optional path overrides
func NewLoader(opts ...Option) *Loader

type Option func(*Loader)

func WithGlobalPath(path string) Option
func WithProjectPath(path string) Option

// LoadGlobal loads global configuration
// Returns default config if file doesn't exist
func (l *Loader) LoadGlobal() (*Config, error)

// LoadProject loads project configuration
// Returns NotFoundError if file doesn't exist
func (l *Loader) LoadProject() (*Project, error)

// GlobalPath returns the resolved global config path
func (l *Loader) GlobalPath() (string, error)

// ProjectPath returns the resolved project config path
// Searches current directory and parents for yar.yaml
func (l *Loader) ProjectPath() (string, error)

// Validate validates config against JSON Schema
func Validate(cfg any, schemaPath string) error
```

### Default Configuration

```go
// DefaultConfig returns sensible defaults for global config
func DefaultConfig() *Config {
    return &Config{
        Container: "colima",
        Hosts: &HostsConfig{
            Mode: "etc",
        },
        Network: &NetworkConfig{
            Name: "yar-net",
            CIDR: "172.16.34.0/23",
        },
        Secrets: &SecretsConfig{
            Local: &LocalSecretConfig{
                Provider: "pass",
                Fallback: true,
            },
        },
    }
}
```

---

## Data Structures

Configuration types are defined in the Interfaces section above.

---

## Dependencies

### External Packages
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/santhosh-tekuri/jsonschema/v6` - JSON Schema validation

### Internal Packages
- `internal/errors` - Typed errors (ConfigError, ValidationError, NotFoundError)
- `internal/platform` - Path utilities (ConfigDir, ExpandPath)

---

## Invariants

This iteration implements these invariants:

- **INV-CFG-001**: All configuration files MUST validate against their JSON Schema before use.
- **INV-CFG-002**: Project config (`yar.yaml`) MUST be present in the current directory or a parent directory for project-scoped commands.
- **INV-CFG-003**: Global config (`~/.config/yar/config.yaml`) is optional; sensible defaults apply when absent.
- **INV-CFG-004**: Environment names MUST be unique within a project.
- **INV-CFG-005**: Service names MUST be unique within a project.

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/config/types.go` | Config and Project type definitions |
| `internal/config/defaults.go` | Default configuration values |
| `internal/config/loader.go` | Load and parse config files |
| `internal/config/paths.go` | Path resolution utilities |
| `internal/config/schema.go` | JSON Schema validation |
| `internal/config/types_test.go` | Type tests |
| `internal/config/loader_test.go` | Loader tests |
| `internal/config/schema_test.go` | Validation tests |
| `schemas/config.schema.json` | Global config JSON Schema |
| `schemas/project.schema.json` | Project config JSON Schema |
| `cmd/config.go` | Updated config command with real output |
| `cmd/project.go` | Updated project command with real output |

---

## Test Requirements

### Unit Tests

#### types_test.go
- [ ] Config struct has correct YAML tags
- [ ] Config struct has correct JSON tags
- [ ] Project struct has correct YAML tags
- [ ] Project struct has correct JSON tags
- [ ] Nested structs marshal/unmarshal correctly
- [ ] Optional fields omit when empty

#### loader_test.go
- [ ] LoadGlobal returns default when file missing
- [ ] LoadGlobal loads valid YAML file
- [ ] LoadGlobal returns ConfigError for invalid YAML
- [ ] LoadGlobal returns ValidationError for schema failure
- [ ] LoadProject returns NotFoundError when file missing
- [ ] LoadProject loads valid yar.yaml
- [ ] LoadProject returns ConfigError for invalid YAML
- [ ] LoadProject returns ValidationError for schema failure
- [ ] LoadProject finds yar.yaml in parent directories
- [ ] GlobalPath returns XDG-compliant path
- [ ] ProjectPath returns found path
- [ ] WithGlobalPath overrides default path
- [ ] WithProjectPath overrides default path

#### schema_test.go
- [ ] Validate accepts valid global config
- [ ] Validate rejects invalid global config
- [ ] Validate accepts valid project config
- [ ] Validate rejects project without required fields
- [ ] Validate rejects invalid environment names
- [ ] Validate rejects duplicate service names
- [ ] Validate rejects invalid cluster references

### Functional Tests

| Command | Expected Result |
|---------|-----------------|
| `yar config get` | Shows default config (YAML format) |
| `yar config get -o json` | Shows default config (JSON format) |
| `yar project get` | Shows "no yar.yaml found" error |
| `yar project get` (with yar.yaml) | Shows project config |

---

## Test Fixtures

Create `testdata/` directory with:

```
internal/config/testdata/
├── valid/
│   ├── config.yaml           # Valid global config
│   ├── config-minimal.yaml   # Minimal valid global config
│   ├── project.yaml          # Valid project config
│   └── project-minimal.yaml  # Minimal valid project config
├── invalid/
│   ├── config-bad-yaml.yaml  # Malformed YAML
│   ├── config-bad-schema.yaml # Valid YAML, invalid schema
│   ├── project-missing-name.yaml # Missing required field
│   └── project-duplicate-svc.yaml # Duplicate service names
└── schemas/
    ├── config.schema.json
    └── project.schema.json
```

---

## Exit Criteria

- [ ] All configuration types defined with YAML/JSON tags
- [ ] DefaultConfig() returns valid defaults
- [ ] LoadGlobal() works with missing file (returns defaults)
- [ ] LoadGlobal() works with valid file
- [ ] LoadProject() fails gracefully when file missing
- [ ] LoadProject() works with valid file
- [ ] LoadProject() searches parent directories
- [ ] JSON Schema validation works for both configs
- [ ] `yar config get` displays real configuration
- [ ] `yar project get` displays project config or error
- [ ] All unit tests pass
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

---

## Clarifications

*Document any spec ambiguities resolved during implementation here.*
