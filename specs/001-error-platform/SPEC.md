# Iteration 001: Foundation - Error Types, Platform Detection & CLI Skeleton

## Overview

This iteration implements the foundational layer for yar: typed error definitions, platform detection utilities, and a complete CLI skeleton with all commands as stubs. This provides the infrastructure for all subsequent iterations and enables functional testing from day one.

## Scope

### Included
- **Error Types**: `ConfigError`, `ValidationError`, `NotFoundError`, `SecretError`, `PackError`, `DockerError`, `KubernetesError`, `NetworkError`
- **Platform Detection**: OS detection, XDG-compliant directory paths
- **CLI Skeleton**: Complete command structure with all 15+ commands as stubs

### NOT Included (deferred)
- Error wrapping strategies (will evolve with usage)
- Platform-specific privilege escalation (iteration 028)
- Platform-specific keyring access (iteration 014)
- Actual command implementations (subsequent iterations)

---

## Interfaces

### Error Types

All error types implement the `error` interface and provide structured fields for programmatic handling.

```go
// ConfigError represents configuration-related errors
type ConfigError struct {
    Path    string // file path that caused the error
    Field   string // specific field, if applicable
    Message string // human-readable description
    Err     error  // underlying error, if any
}

func (e *ConfigError) Error() string
func (e *ConfigError) Unwrap() error
```

```go
// ValidationError represents schema or value validation failures
type ValidationError struct {
    Field   string   // field that failed validation
    Value   any      // the invalid value
    Message string   // description of why validation failed
    Errors  []string // multiple validation errors, if applicable
}

func (e *ValidationError) Error() string
```

```go
// NotFoundError represents missing resources
type NotFoundError struct {
    Resource string // type of resource (file, secret, pack, service)
    Name     string // name/identifier of the resource
    Message  string // additional context
}

func (e *NotFoundError) Error() string
```

```go
// SecretError represents secret operation failures
type SecretError struct {
    Provider string // provider name (pass, keychain, azure, etc.)
    Key      string // secret key
    Op       string // operation: get, set, delete, list, sync
    Err      error  // underlying error
}

func (e *SecretError) Error() string
func (e *SecretError) Unwrap() error
```

```go
// PackError represents pack-related errors
type PackError struct {
    Pack    string // pack name
    Message string // description
    Err     error  // underlying error
}

func (e *PackError) Error() string
func (e *PackError) Unwrap() error
```

```go
// DockerError represents Docker operation failures
type DockerError struct {
    Op      string // operation: create, start, stop, remove, etc.
    Target  string // container/network/volume name
    Message string // description
    Err     error  // underlying error
}

func (e *DockerError) Error() string
func (e *DockerError) Unwrap() error
```

```go
// KubernetesError represents Kubernetes operation failures
type KubernetesError struct {
    Op        string // operation: apply, delete, get, etc.
    Resource  string // resource type (deployment, service, etc.)
    Name      string // resource name
    Namespace string // namespace
    Err       error  // underlying error
}

func (e *KubernetesError) Error() string
func (e *KubernetesError) Unwrap() error
```

```go
// NetworkError represents network-related failures
type NetworkError struct {
    Op      string // operation: vpn, dns, hosts
    Target  string // target (hostname, IP, etc.)
    Message string // description
    Err     error  // underlying error
}

func (e *NetworkError) Error() string
func (e *NetworkError) Unwrap() error
```

### Platform Detection

```go
// OS represents the operating system
type OS string

const (
    Darwin  OS = "darwin"
    Linux   OS = "linux"
    Windows OS = "windows"
)

// Platform returns the current operating system
func Platform() OS

// ConfigDir returns the XDG-compliant config directory for yar
// - macOS: ~/Library/Application Support/yar or $XDG_CONFIG_HOME/yar
// - Linux: $XDG_CONFIG_HOME/yar or ~/.config/yar
// - Windows: %APPDATA%\yar
func ConfigDir() (string, error)

// CacheDir returns the XDG-compliant cache directory for yar
// - macOS: ~/Library/Caches/yar or $XDG_CACHE_HOME/yar
// - Linux: $XDG_CACHE_HOME/yar or ~/.cache/yar
// - Windows: %LOCALAPPDATA%\yar\cache
func CacheDir() (string, error)

// DataDir returns the XDG-compliant data directory for yar
// - macOS: ~/Library/Application Support/yar or $XDG_DATA_HOME/yar
// - Linux: $XDG_DATA_HOME/yar or ~/.local/share/yar
// - Windows: %LOCALAPPDATA%\yar\data
func DataDir() (string, error)

// HomeDir returns the user's home directory
func HomeDir() (string, error)

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) (string, error)
```

### CLI Commands (Stubs)

All commands implemented as stubs that print placeholder messages:

| Command | Subcommands | Description |
|---------|-------------|-------------|
| `yar fleet` | up, down, destroy, restart, status, update | Fleet lifecycle management |
| `yar config` | get, edit | Configuration management |
| `yar project` | init, get, edit | Project management |
| `yar pack` | list, install, remove | Pack management |
| `yar template` | build, render, publish | Template operations |
| `yar secret` | set, get, delete, list, sync | Secret management |
| `yar hosts` | set, get, delete, list | Hosts file management |
| `yar doctor` | run | System health checks |

**Aliases** (nautical theme):
- `hoist` → `fleet up`
- `dock` → `fleet down`
- `scuttle` → `fleet destroy`
- `swab` → `doctor run --fix`
- `up` → `fleet up`
- `down` → `fleet down`

**Global Flags**:
- `--verbose, -v` - Verbose output
- `--output, -o` - Output format (text, json, yaml)

---

## Data Structures

Error types are defined in the Interfaces section above.

Platform types:

```go
type OS string

const (
    Darwin  OS = "darwin"
    Linux   OS = "linux"
    Windows OS = "windows"
)
```

---

## Dependencies

### External Packages
- `github.com/spf13/cobra` - CLI framework

### Internal Packages
- None (this is a foundational iteration)

---

## Invariants

This iteration establishes foundations for these invariants:

- **INV-SEC-005**: Failed secret resolution MUST halt operations with a clear error; secrets MUST NOT fall back to empty strings. → `SecretError` provides structured error for this.
- **INV-CFG-001**: All configuration files MUST validate against their JSON Schema before use. → `ValidationError` provides structured error for schema failures.

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/errors/errors.go` | All error type definitions |
| `internal/errors/errors_test.go` | Unit tests for error formatting |
| `internal/platform/platform.go` | Platform detection and path utilities |
| `internal/platform/platform_test.go` | Unit tests for platform detection |
| `cmd/root.go` | Root command with global flags |
| `cmd/fleet.go` | Fleet commands (up/down/destroy/restart/status/update) |
| `cmd/config.go` | Config commands (get/edit) |
| `cmd/project.go` | Project commands (init/get/edit) |
| `cmd/pack.go` | Pack commands (list/install/remove) |
| `cmd/template.go` | Template commands (build/render/publish) |
| `cmd/secret.go` | Secret commands (set/get/delete/list/sync) |
| `cmd/hosts.go` | Hosts commands (set/get/delete/list) |
| `cmd/doctor.go` | Doctor command (run) |
| `cmd/aliases.go` | Nautical aliases (hoist/dock/scuttle/swab/up/down) |

---

## Test Requirements

### Unit Tests

#### errors_test.go
- [x] `ConfigError.Error()` returns formatted message with path and field
- [x] `ConfigError.Unwrap()` returns underlying error
- [x] `ValidationError.Error()` returns formatted message with field and value
- [x] `ValidationError.Error()` handles multiple errors
- [x] `NotFoundError.Error()` returns formatted message with resource and name
- [x] `SecretError.Error()` returns formatted message with provider, key, and op
- [x] `SecretError.Unwrap()` returns underlying error
- [x] `PackError.Error()` returns formatted message
- [x] `PackError.Unwrap()` returns underlying error
- [x] `DockerError.Error()` returns formatted message
- [x] `DockerError.Unwrap()` returns underlying error
- [x] `KubernetesError.Error()` returns formatted message with namespace
- [x] `KubernetesError.Unwrap()` returns underlying error
- [x] `NetworkError.Error()` returns formatted message
- [x] `NetworkError.Unwrap()` returns underlying error
- [x] All error types satisfy `error` interface
- [x] Error types with `Err` field satisfy `errors.Unwrap` interface

#### platform_test.go
- [x] `Platform()` returns one of darwin, linux, windows
- [x] `ConfigDir()` returns non-empty path
- [x] `ConfigDir()` respects `$XDG_CONFIG_HOME` when set
- [x] `CacheDir()` returns non-empty path
- [x] `CacheDir()` respects `$XDG_CACHE_HOME` when set
- [x] `DataDir()` returns non-empty path
- [x] `DataDir()` respects `$XDG_DATA_HOME` when set
- [x] `HomeDir()` returns non-empty path
- [x] `ExpandPath()` expands `~` to home directory
- [x] `ExpandPath()` expands environment variables
- [x] `ExpandPath()` handles paths without expansion

### Functional Tests

| Command | Expected Result |
|---------|-----------------|
| `yar --help` | Shows all commands with descriptions |
| `yar fleet up` | Prints "starting services for environment 'local'" |
| `yar fleet down` | Prints "stopping services for environment 'local'" |
| `yar config get` | Shows config directory path from platform.ConfigDir() |
| `yar doctor run` | Shows health check table |
| `yar hoist` | Same as `yar fleet up` |
| `yar --version` | Shows version string |

### Integration Tests
- None for this iteration

---

## Exit Criteria

- [x] All error types defined with `Error()` method
- [x] Error types with underlying errors implement `Unwrap()`
- [x] `Platform()` returns correct OS
- [x] `ConfigDir()` returns valid XDG-compliant path
- [x] `CacheDir()` returns valid XDG-compliant path
- [x] All CLI commands callable (stubs)
- [x] All aliases work correctly
- [x] All unit tests pass
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] `./yar --help` shows all commands

---

## Clarifications

*Scope expanded during implementation to include CLI skeleton, enabling functional testing from iteration 001. This ensures every iteration delivers user-testable functionality.*
