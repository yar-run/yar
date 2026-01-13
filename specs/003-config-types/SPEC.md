# Iteration 003: Config Types Specification

## Overview

This iteration defines all configuration struct types for global configuration (`~/.config/yar/config.yaml`) and project configuration (`./yar.yaml`). These types provide the data structures that the rest of Yar uses for configuration management.

**Note**: This iteration was implemented as part of iteration 002. This spec documents what was delivered.

## Scope

### Included
- Global Config struct with all nested types
- Project struct with environments and services
- Default configuration values
- YAML and JSON struct tags for serialization
- Unit tests for marshaling/unmarshaling

### NOT Included (deferred)
- Config file loading (iteration 004)
- JSON Schema validation (iteration 004)
- Path resolution (iteration 004)

---

## Data Structures

### Global Configuration Types

```go
// Config represents global yar configuration (~/.config/yar/config.yaml)
type Config struct {
    Container string                    `yaml:"container" json:"container"`
    VPN       *VPNConfig                `yaml:"vpn,omitempty" json:"vpn,omitempty"`
    Hosts     *HostsConfig              `yaml:"hosts,omitempty" json:"hosts,omitempty"`
    Network   *NetworkConfig            `yaml:"network,omitempty" json:"network,omitempty"`
    Secrets   *SecretsConfig            `yaml:"secrets,omitempty" json:"secrets,omitempty"`
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
    Local     *LocalSecretConfig               `yaml:"local" json:"local"`
    Providers map[string]*SecretProviderConfig `yaml:"providers,omitempty" json:"providers,omitempty"`
}

type LocalSecretConfig struct {
    Provider string `yaml:"provider" json:"provider"`
    Store    string `yaml:"store,omitempty" json:"store,omitempty"`
    Fallback bool   `yaml:"fallback" json:"fallback"`
}

type SecretProviderConfig struct {
    Type   string         `yaml:"type" json:"type"`
    Config map[string]any `yaml:",inline" json:"-"`
}

type ClusterConfig struct {
    Provider  string `yaml:"provider" json:"provider"`
    Context   string `yaml:"context,omitempty" json:"context,omitempty"`
    Namespace string `yaml:"namespace,omitempty" json:"namespace,omitempty"`
}
```

### Project Configuration Types

```go
// Project represents project configuration (yar.yaml)
type Project struct {
    Project      string                  `yaml:"project" json:"project"`
    Environments map[string]*Environment `yaml:"environments" json:"environments"`
    Services     []*Service              `yaml:"services" json:"services"`
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

### Default Configuration

```go
// DefaultConfig returns sensible defaults for global configuration.
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

## Dependencies

### External Packages
- `gopkg.in/yaml.v3` - YAML marshaling
- `encoding/json` - JSON marshaling (stdlib)

### Internal Packages
- None (this is a foundational package)

---

## Invariants

- **INV-CFG-003**: Global config always loadable (returns defaults if missing)
- All struct fields have both `yaml` and `json` tags
- Optional fields use `omitempty` tag
- Pointer types for optional nested structs

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/config/types.go` | All configuration struct definitions |
| `internal/config/defaults.go` | DefaultConfig() function |
| `internal/config/types_test.go` | Unit tests for marshaling |

---

## Test Requirements

### Unit Tests
- [x] Test Config YAML marshal/unmarshal round-trip
- [x] Test Config JSON marshal/unmarshal round-trip
- [x] Test Config omitempty behavior
- [x] Test Project YAML marshal/unmarshal round-trip
- [x] Test Project JSON marshal/unmarshal round-trip
- [x] Test Service omitempty behavior
- [x] Test DefaultConfig returns valid config with expected values

---

## Exit Criteria

- [x] All types compile
- [x] Default config is valid
- [x] YAML/JSON tags present on all fields
- [x] All unit tests pass
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean

---

## Status

**COMPLETE** - Implemented as part of iteration 002.
