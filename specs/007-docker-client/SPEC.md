# Iteration 007: Docker Client Specification

## Overview

This iteration creates a Docker SDK wrapper for network operations. The wrapper provides a clean interface for creating, listing, and removing Docker networks, abstracting the Docker Engine API behind a testable interface with functional options for configuration.

## Scope

### Included
- Docker client wrapper with functional options pattern
- Network operations: Create, Remove, List, Inspect
- Mock interface for unit testing
- Connection handling with configurable host/timeout
- Error wrapping with actionable messages

### NOT Included (deferred)
- Container operations (iteration 008)
- Compose parsing (iteration 009)
- Image operations (future iteration)
- Secret operations (future iteration)

---

## Interfaces

### Client

```go
// Client provides Docker operations.
type Client interface {
    // Network operations
    NetworkCreate(ctx context.Context, name string, opts NetworkCreateOptions) (string, error)
    NetworkRemove(ctx context.Context, name string) error
    NetworkList(ctx context.Context, opts NetworkListOptions) ([]Network, error)
    NetworkInspect(ctx context.Context, name string) (*Network, error)
    
    // Ping checks Docker daemon connectivity
    Ping(ctx context.Context) error
    
    // Close releases resources
    Close() error
}
```

**NetworkCreate**: Creates a Docker network with the given name and options. Returns the network ID on success.

**NetworkRemove**: Removes a Docker network by name. Fails if containers are attached (unless force option used).

**NetworkList**: Lists Docker networks, optionally filtered by options.

**NetworkInspect**: Returns detailed information about a specific network.

**Ping**: Verifies the Docker daemon is reachable.

**Close**: Closes the underlying Docker client connection.

---

## Data Structures

```go
// Network represents a Docker network.
type Network struct {
    ID         string            `json:"id"`
    Name       string            `json:"name"`
    Driver     string            `json:"driver"`
    Scope      string            `json:"scope"`
    IPAM       *IPAM             `json:"ipam,omitempty"`
    Labels     map[string]string `json:"labels,omitempty"`
    Containers []string          `json:"containers,omitempty"` // container IDs
    Created    time.Time         `json:"created"`
}

// IPAM represents IP Address Management configuration.
type IPAM struct {
    Driver  string       `json:"driver"`
    Config  []IPAMConfig `json:"config,omitempty"`
}

// IPAMConfig represents IPAM pool configuration.
type IPAMConfig struct {
    Subnet  string `json:"subnet,omitempty"`
    Gateway string `json:"gateway,omitempty"`
}

// NetworkCreateOptions configures network creation.
type NetworkCreateOptions struct {
    Driver     string            // Network driver (default: "bridge")
    Subnet     string            // CIDR notation (e.g., "172.16.34.0/23")
    Gateway    string            // Gateway IP (optional, derived from subnet)
    Labels     map[string]string // Network labels
    Internal   bool              // Restrict external access
    Attachable bool              // Allow manual container attachment
}

// NetworkListOptions configures network listing.
type NetworkListOptions struct {
    Filters map[string][]string // Filter by name, id, driver, label, etc.
}
```

### Functional Options

```go
// Option configures the Docker client.
type Option func(*clientOptions)

// WithHost sets the Docker host (e.g., "unix:///var/run/docker.sock").
func WithHost(host string) Option

// WithTimeout sets the operation timeout.
func WithTimeout(timeout time.Duration) Option

// WithAPIVersion sets the Docker API version.
func WithAPIVersion(version string) Option

// WithTLSConfig sets TLS configuration for remote Docker hosts.
func WithTLSConfig(config *tls.Config) Option
```

---

## Dependencies

### External Packages
- `github.com/docker/docker/client` - Docker Engine SDK client
- `github.com/docker/docker/api/types` - Docker API types
- `github.com/docker/docker/api/types/network` - Network-specific types

### Internal Packages
- `internal/errors` - Error types (ConfigError for connection failures)

---

## Invariants

Reference applicable invariants from root SPEC.md:

- **INV-CFG-001**: Configuration validation (Docker host configuration)
- **INV-FLT-004**: Fleet operations MUST be idempotent - network creation should handle "already exists" gracefully

---

## Error Handling

### Error Types

```go
// DockerError represents a Docker operation failure.
type DockerError struct {
    Op      string // Operation: "network.create", "network.remove", etc.
    Name    string // Resource name
    Message string // Human-readable message
    Err     error  // Underlying error
}

func (e *DockerError) Error() string
func (e *DockerError) Unwrap() error
```

### Error Scenarios

| Scenario | Error Handling |
|----------|----------------|
| Docker daemon not running | Return actionable error with fix suggestion |
| Network already exists | Return nil (idempotent) or wrapped error with context |
| Network not found (remove) | Return nil (idempotent) |
| Network has attached containers | Return error listing attached containers |
| Permission denied | Return error with sudo/permission fix suggestion |

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/docker/client.go` | Client interface and constructor with options |
| `internal/docker/network.go` | Network operation implementations |
| `internal/docker/types.go` | Network, IPAM, options structs |
| `internal/docker/errors.go` | DockerError type |
| `internal/docker/mock.go` | Mock client for testing |
| `internal/docker/client_test.go` | Unit tests for client construction |
| `internal/docker/network_test.go` | Unit tests for network operations |

---

## Test Requirements

### Unit Tests

#### Client Construction
- [ ] NewClient with default options connects to default socket
- [ ] NewClient with WithHost option uses specified host
- [ ] NewClient with WithTimeout option respects timeout
- [ ] NewClient returns error when Docker daemon unavailable

#### Network Create
- [ ] NetworkCreate creates network with default driver
- [ ] NetworkCreate with subnet configures IPAM
- [ ] NetworkCreate with labels applies labels
- [ ] NetworkCreate returns existing network ID if name exists (idempotent)
- [ ] NetworkCreate returns error for invalid subnet CIDR

#### Network Remove
- [ ] NetworkRemove removes existing network
- [ ] NetworkRemove returns nil for non-existent network (idempotent)
- [ ] NetworkRemove returns error if containers attached

#### Network List
- [ ] NetworkList returns all networks
- [ ] NetworkList with name filter returns matching networks
- [ ] NetworkList with label filter returns labeled networks

#### Network Inspect
- [ ] NetworkInspect returns network details
- [ ] NetworkInspect returns error for non-existent network

### Mock Client
- [ ] MockClient implements Client interface
- [ ] MockClient allows configuring responses
- [ ] MockClient tracks method calls for assertions

---

## Exit Criteria

- [ ] Client interface defined with all network methods
- [ ] Functional options pattern implemented (WithHost, WithTimeout)
- [ ] NetworkCreate creates networks with configurable CIDR
- [ ] NetworkRemove removes networks idempotently
- [ ] NetworkList lists networks with optional filters
- [ ] NetworkInspect returns network details
- [ ] Mock client available for testing
- [ ] All unit tests pass with mocks
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

---

## Usage Examples

### Creating the Client

```go
// Default client (uses DOCKER_HOST or default socket)
client, err := docker.NewClient()
if err != nil {
    return fmt.Errorf("failed to create Docker client: %w", err)
}
defer client.Close()

// Custom host with timeout
client, err := docker.NewClient(
    docker.WithHost("unix:///var/run/docker.sock"),
    docker.WithTimeout(30 * time.Second),
)
```

### Network Operations

```go
// Create network with yar defaults
networkID, err := client.NetworkCreate(ctx, "yar-net", docker.NetworkCreateOptions{
    Subnet: "172.16.34.0/23",
    Labels: map[string]string{
        "yar.managed": "true",
    },
})

// List yar-managed networks
networks, err := client.NetworkList(ctx, docker.NetworkListOptions{
    Filters: map[string][]string{
        "label": {"yar.managed=true"},
    },
})

// Remove network
err := client.NetworkRemove(ctx, "yar-net")
```
