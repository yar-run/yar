# Iteration 001: Error & Platform Specification

## Overview

This iteration implements foundational error types and platform detection utilities. Error types provide structured, typed errors for all yar operations. Platform utilities detect the operating system and provide XDG-compliant configuration and cache directory paths.

## Scope

### Included
- Typed error definitions: `ConfigError`, `ValidationError`, `NotFoundError`, `SecretError`, `PackError`, `DockerError`, `KubernetesError`, `NetworkError`
- Error formatting with `Error()` method returning actionable messages
- Platform detection returning `darwin`, `linux`, or `windows`
- `ConfigDir()` returning XDG-compliant config directory
- `CacheDir()` returning XDG-compliant cache directory
- Path helper utilities

### NOT Included (deferred)
- Error wrapping strategies (will evolve with usage)
- Platform-specific privilege escalation (iteration 028)
- Platform-specific keyring access (iteration 014)

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
- None (uses only stdlib)

### Internal Packages
- None (this is a foundational package)

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

---

## Test Requirements

### Unit Tests

#### errors_test.go
- [ ] `ConfigError.Error()` returns formatted message with path and field
- [ ] `ConfigError.Unwrap()` returns underlying error
- [ ] `ValidationError.Error()` returns formatted message with field and value
- [ ] `ValidationError.Error()` handles multiple errors
- [ ] `NotFoundError.Error()` returns formatted message with resource and name
- [ ] `SecretError.Error()` returns formatted message with provider, key, and op
- [ ] `SecretError.Unwrap()` returns underlying error
- [ ] `PackError.Error()` returns formatted message
- [ ] `PackError.Unwrap()` returns underlying error
- [ ] `DockerError.Error()` returns formatted message
- [ ] `DockerError.Unwrap()` returns underlying error
- [ ] `KubernetesError.Error()` returns formatted message with namespace
- [ ] `KubernetesError.Unwrap()` returns underlying error
- [ ] `NetworkError.Error()` returns formatted message
- [ ] `NetworkError.Unwrap()` returns underlying error
- [ ] All error types satisfy `error` interface
- [ ] Error types with `Err` field satisfy `errors.Unwrap` interface

#### platform_test.go
- [ ] `Platform()` returns one of darwin, linux, windows
- [ ] `ConfigDir()` returns non-empty path
- [ ] `ConfigDir()` respects `$XDG_CONFIG_HOME` when set
- [ ] `CacheDir()` returns non-empty path
- [ ] `CacheDir()` respects `$XDG_CACHE_HOME` when set
- [ ] `DataDir()` returns non-empty path
- [ ] `DataDir()` respects `$XDG_DATA_HOME` when set
- [ ] `HomeDir()` returns non-empty path
- [ ] `ExpandPath()` expands `~` to home directory
- [ ] `ExpandPath()` expands environment variables
- [ ] `ExpandPath()` handles paths without expansion

### Integration Tests
- None for this iteration

### Test Fixtures
- None for this iteration

---

## Exit Criteria

- [x] All error types defined with `Error()` method
- [x] Error types with underlying errors implement `Unwrap()`
- [x] `Platform()` returns correct OS
- [x] `ConfigDir()` returns valid XDG-compliant path
- [x] `CacheDir()` returns valid XDG-compliant path
- [x] All unit tests pass
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean

---

## Clarifications

*Document any spec ambiguities resolved during implementation here.*
