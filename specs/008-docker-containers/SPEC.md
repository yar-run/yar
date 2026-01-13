# Iteration 008: Docker Containers Specification

## Overview

Extend the Docker client wrapper to support container lifecycle operations. This iteration adds the ability to create, start, stop, remove, and inspect containers, as well as stream container logs. These primitives enable the Compose driver (iteration 024) to manage containerized services.

## Scope

### Included
- Container types (`Container`, `ContainerConfig`, `ContainerState`)
- Container operations: Create, Start, Stop, Remove, Inspect, List
- Log streaming with follow and tail options
- Port mapping and volume mount configuration
- Environment variable injection
- Error types for container operations
- MockClient extensions for container testing
- Comprehensive unit tests following Go idioms

### NOT Included (deferred)
- Image operations (pull, build) - iteration TBD
- Container exec (running commands in containers)
- Container attach (interactive TTY)
- Docker Compose integration - iteration 009
- Health check monitoring
- Container networking beyond port exposure

---

## Interfaces

### Extended Client Interface

```go
type Client interface {
    // Existing from 007
    NetworkCreate(ctx context.Context, name string, opts NetworkCreateOptions) (string, error)
    NetworkRemove(ctx context.Context, name string) error
    NetworkList(ctx context.Context, opts NetworkListOptions) ([]Network, error)
    NetworkInspect(ctx context.Context, name string) (*Network, error)
    Ping(ctx context.Context) error
    Close() error

    // New container operations
    ContainerCreate(ctx context.Context, config ContainerConfig) (string, error)
    ContainerStart(ctx context.Context, id string) error
    ContainerStop(ctx context.Context, id string, timeout *time.Duration) error
    ContainerRemove(ctx context.Context, id string, opts ContainerRemoveOptions) error
    ContainerInspect(ctx context.Context, id string) (*Container, error)
    ContainerList(ctx context.Context, opts ContainerListOptions) ([]Container, error)
    ContainerLogs(ctx context.Context, id string, opts ContainerLogOptions) (io.ReadCloser, error)
    ContainerWait(ctx context.Context, id string, condition WaitCondition) (<-chan ContainerWaitResult, <-chan error)
}
```

### ContainerCreate

Creates a container from the provided configuration. Returns the container ID.
- Does NOT start the container
- Container name must be unique or create fails
- Image must exist locally (pull not automatic in this iteration)

### ContainerStart

Starts a stopped or created container by ID or name.
- Idempotent: starting an already-running container returns nil
- Returns error if container doesn't exist

### ContainerStop

Stops a running container with optional timeout.
- Sends SIGTERM, then SIGKILL after timeout
- Default timeout: 10 seconds
- Idempotent: stopping an already-stopped container returns nil

### ContainerRemove

Removes a container by ID or name.
- Fails if container is running unless Force=true
- Can optionally remove associated volumes

### ContainerInspect

Returns detailed information about a container.
- Returns error if container doesn't exist

### ContainerList

Lists containers with optional filters.
- Can filter by name, label, status, network
- Returns running containers by default; use All=true for all

### ContainerLogs

Streams container logs.
- Returns io.ReadCloser that must be closed by caller
- Supports Follow (tail -f behavior), Tail (last N lines)
- Supports stdout/stderr filtering

### ContainerWait

Waits for a container to reach a condition.
- Conditions: NotRunning, NextExit, Removed
- Returns channels for result and error (async)

---

## Data Structures

### Container

```go
// Container represents a Docker container.
type Container struct {
    ID         string            `json:"id"`
    Name       string            `json:"name"`
    Image      string            `json:"image"`
    Command    string            `json:"command"`
    Created    time.Time         `json:"created"`
    State      ContainerState    `json:"state"`
    Ports      []PortBinding     `json:"ports,omitempty"`
    Labels     map[string]string `json:"labels,omitempty"`
    NetworkIDs []string          `json:"network_ids,omitempty"`
}

// ContainerState represents the current state of a container.
type ContainerState struct {
    Status     string    `json:"status"`     // created, running, paused, restarting, removing, exited, dead
    Running    bool      `json:"running"`
    Paused     bool      `json:"paused"`
    ExitCode   int       `json:"exit_code"`
    StartedAt  time.Time `json:"started_at"`
    FinishedAt time.Time `json:"finished_at"`
    Error      string    `json:"error,omitempty"`
}
```

### ContainerConfig

```go
// ContainerConfig specifies how to create a container.
type ContainerConfig struct {
    Name        string            // Container name (required)
    Image       string            // Image reference (required)
    Command     []string          // Override CMD
    Entrypoint  []string          // Override ENTRYPOINT
    Env         []string          // Environment variables ("KEY=value")
    Labels      map[string]string // Container labels
    WorkingDir  string            // Working directory inside container
    
    // Host configuration
    Ports       []PortMapping     // Port mappings
    Volumes     []VolumeMount     // Volume mounts
    Networks    []string          // Networks to connect to
    RestartPolicy RestartPolicy   // Restart policy
    
    // Resource constraints
    Memory      int64             // Memory limit in bytes (0 = unlimited)
    CPUs        float64           // CPU limit (0 = unlimited)
}

// PortMapping maps a host port to a container port.
type PortMapping struct {
    HostIP        string // Host IP to bind (default: "0.0.0.0")
    HostPort      string // Host port (can be range "8080-8090")
    ContainerPort string // Container port (required)
    Protocol      string // tcp or udp (default: tcp)
}

// VolumeMount mounts a volume or bind mount into a container.
type VolumeMount struct {
    Source   string // Host path or volume name
    Target   string // Container path (required)
    ReadOnly bool   // Mount as read-only
    Type     string // bind, volume, or tmpfs (default: bind)
}

// RestartPolicy defines when to restart a container.
type RestartPolicy struct {
    Name              string // no, always, on-failure, unless-stopped
    MaximumRetryCount int    // For on-failure policy
}
```

### Option Types

```go
// ContainerRemoveOptions configures container removal.
type ContainerRemoveOptions struct {
    Force         bool // Remove even if running
    RemoveVolumes bool // Remove associated volumes
}

// ContainerListOptions configures container listing.
type ContainerListOptions struct {
    All     bool                // Include stopped containers
    Limit   int                 // Max number to return (0 = no limit)
    Filters map[string][]string // Filter by label, name, status, etc.
}

// ContainerLogOptions configures log streaming.
type ContainerLogOptions struct {
    ShowStdout bool          // Include stdout
    ShowStderr bool          // Include stderr
    Follow     bool          // Follow log output (tail -f)
    Tail       string        // Number of lines from end ("all" or number)
    Since      time.Time     // Show logs since timestamp
    Until      time.Time     // Show logs until timestamp
    Timestamps bool          // Include timestamps
}

// WaitCondition specifies what container state to wait for.
type WaitCondition string

const (
    WaitConditionNotRunning WaitCondition = "not-running"
    WaitConditionNextExit   WaitCondition = "next-exit"
    WaitConditionRemoved    WaitCondition = "removed"
)

// ContainerWaitResult is returned when container wait completes.
type ContainerWaitResult struct {
    StatusCode int64  // Exit code
    Error      string // Error message if any
}
```

---

## Error Types

```go
// ErrContainerCreate creates a container creation error.
func ErrContainerCreate(name string, err error) *DockerError

// ErrContainerStart creates a container start error.
func ErrContainerStart(id string, err error) *DockerError

// ErrContainerStop creates a container stop error.
func ErrContainerStop(id string, err error) *DockerError

// ErrContainerRemove creates a container removal error.
func ErrContainerRemove(id string, err error) *DockerError

// ErrContainerInspect creates a container inspect error.
func ErrContainerInspect(id string, err error) *DockerError

// ErrContainerList creates a container list error.
func ErrContainerList(err error) *DockerError

// ErrContainerLogs creates a container logs error.
func ErrContainerLogs(id string, err error) *DockerError

// ErrContainerNotFound creates a container not found error.
func ErrContainerNotFound(id string) *DockerError

// ErrContainerAlreadyExists creates a container already exists error.
func ErrContainerAlreadyExists(name string) *DockerError

// ErrContainerRunning creates an error for operations on running containers.
func ErrContainerRunning(id string) *DockerError
```

---

## Dependencies

### External Packages
- `github.com/docker/docker/client` - Docker SDK client
- `github.com/docker/docker/api/types/container` - Container types
- `github.com/docker/go-connections/nat` - Port mapping utilities
- `github.com/google/go-cmp/cmp` - Test comparisons (already added)

### Internal Packages
- `internal/docker` - Existing client, network, types, errors, mock

---

## Invariants

Reference applicable invariants from root SPEC.md:

- **INV-FLT-004**: Fleet operations MUST be idempotent; ContainerStart/Stop must be idempotent.

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/docker/container.go` | Container operations on dockerClient |
| `internal/docker/container_types.go` | Container, ContainerConfig, and option types |
| `internal/docker/container_test.go` | Unit tests for container types |
| `internal/docker/mock.go` | Extend MockClient with container operations |
| `internal/docker/mock_container_test.go` | Tests for MockClient container behavior |
| `internal/docker/errors.go` | Add container error constructors |

---

## Test Requirements

### Unit Tests

#### container_types_test.go
- [ ] Container struct field access
- [ ] ContainerState status values
- [ ] ContainerConfig with all fields populated
- [ ] PortMapping defaults (protocol, host IP)
- [ ] VolumeMount types (bind, volume, tmpfs)
- [ ] RestartPolicy values
- [ ] ContainerRemoveOptions defaults
- [ ] ContainerListOptions filters
- [ ] ContainerLogOptions combinations
- [ ] WaitCondition constants

#### mock_container_test.go
- [ ] MockClient.ContainerCreate returns configured ID
- [ ] MockClient.ContainerCreate records call details
- [ ] MockClient.ContainerCreate returns configured error
- [ ] MockClient.ContainerCreate callback behavior
- [ ] MockClient.ContainerStart success
- [ ] MockClient.ContainerStart records ID
- [ ] MockClient.ContainerStart returns error
- [ ] MockClient.ContainerStop success
- [ ] MockClient.ContainerStop records ID and timeout
- [ ] MockClient.ContainerStop idempotent for stopped container
- [ ] MockClient.ContainerRemove success
- [ ] MockClient.ContainerRemove with force
- [ ] MockClient.ContainerRemove records options
- [ ] MockClient.ContainerInspect returns result
- [ ] MockClient.ContainerInspect not found error
- [ ] MockClient.ContainerList returns configured results
- [ ] MockClient.ContainerList with filters
- [ ] MockClient.ContainerLogs returns reader
- [ ] MockClient.ContainerWait returns channel
- [ ] MockClient.Reset clears container call records

#### errors_test.go (additions)
- [ ] ErrContainerCreate formats correctly
- [ ] ErrContainerNotFound formats correctly
- [ ] ErrContainerAlreadyExists formats correctly
- [ ] ErrContainerRunning formats correctly

### Integration Tests
- Deferred to iteration 032 (requires Docker daemon)

---

## Exit Criteria

- [ ] Client interface extended with all container methods
- [ ] All container types defined with proper JSON/YAML tags
- [ ] All error constructors implemented
- [ ] MockClient supports all container operations
- [ ] All unit tests pass following Go idioms:
  - Table-driven tests with map[string]struct{}
  - t.Run() with descriptive names
  - t.Parallel() for isolation
  - cmp.Diff() for struct comparisons
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
