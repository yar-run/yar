# Iteration 007: Docker Client Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Client Foundation

### A1. Data Structures

**Implement:**
- [x] Create `internal/docker/types.go` with Network, IPAM, IPAMConfig structs
- [x] Create NetworkCreateOptions, NetworkListOptions structs
- [x] Add JSON tags for serialization

**Verify:**
- [x] `go build ./...` succeeds

### A2. Error Types

**Implement:**
- [x] Create `internal/docker/errors.go` with DockerError type
- [x] Implement Error() and Unwrap() methods
- [x] Add helper constructors for common errors

**Verify:**
- [x] `go build ./...` succeeds

### A3. Client Interface and Options

**Test First:**
- [x] Write test for NewClient with default options
- [x] Write test for NewClient with WithHost option
- [x] Write test for NewClient with WithTimeout option

**Implement:**
- [x] Create `internal/docker/client.go` with Client interface
- [x] Implement Option type and functional options (WithHost, WithTimeout, WithAPIVersion)
- [x] Implement NewClient constructor
- [x] Implement Ping method
- [x] Implement Close method

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/docker/...` passes

---

## Phase B: Mock Client

### B1. Mock Implementation

**Implement:**
- [x] Create `internal/docker/mock.go` with MockClient struct
- [x] Implement all Client interface methods on MockClient
- [x] Add fields for configuring mock responses
- [x] Add fields for recording method calls

**Verify:**
- [x] `go build ./...` succeeds
- [x] MockClient compiles and satisfies Client interface

---

## Phase C: Network Operations

### C1. NetworkCreate

**Test First:**
- [x] Write test for NetworkCreate with default driver
- [x] Write test for NetworkCreate with custom subnet
- [x] Write test for NetworkCreate with labels
- [x] Write test for NetworkCreate idempotency (already exists)
- [x] Write test for NetworkCreate with invalid subnet

**Implement:**
- [x] Implement NetworkCreate in `internal/docker/network.go`
- [x] Handle IPAM configuration for custom subnets
- [x] Handle "already exists" case idempotently
- [x] Wrap errors with DockerError

**Verify:**
- [x] `go test ./internal/docker/...` passes

### C2. NetworkRemove

**Test First:**
- [x] Write test for NetworkRemove success
- [x] Write test for NetworkRemove non-existent (idempotent)
- [x] Write test for NetworkRemove with attached containers

**Implement:**
- [x] Implement NetworkRemove
- [x] Handle "not found" case idempotently
- [x] Return clear error when containers attached

**Verify:**
- [x] `go test ./internal/docker/...` passes

### C3. NetworkList

**Test First:**
- [x] Write test for NetworkList returns all networks
- [x] Write test for NetworkList with name filter
- [x] Write test for NetworkList with label filter

**Implement:**
- [x] Implement NetworkList
- [x] Convert Docker API filters to our filter format
- [x] Map Docker network types to our Network struct

**Verify:**
- [x] `go test ./internal/docker/...` passes

### C4. NetworkInspect

**Test First:**
- [x] Write test for NetworkInspect returns details
- [x] Write test for NetworkInspect not found error

**Implement:**
- [x] Implement NetworkInspect
- [x] Map Docker network inspect response to our Network struct
- [x] Include attached container IDs

**Verify:**
- [x] `go test ./internal/docker/...` passes

---

## Functional Tests

After this iteration, verify with a running Docker daemon:

| Test | Command/Action | Expected Result | Status |
|------|----------------|-----------------|--------|
| Client connects | `client.Ping(ctx)` | Returns nil | ✓ (requires Docker) |
| Create network | `client.NetworkCreate(ctx, "yar-test", opts)` | Returns network ID | ✓ (requires Docker) |
| List networks | `client.NetworkList(ctx, opts)` | Includes "yar-test" | ✓ (requires Docker) |
| Inspect network | `client.NetworkInspect(ctx, "yar-test")` | Returns network details | ✓ (requires Docker) |
| Remove network | `client.NetworkRemove(ctx, "yar-test")` | Returns nil | ✓ (requires Docker) |
| Remove again | `client.NetworkRemove(ctx, "yar-test")` | Returns nil (idempotent) | ✓ (requires Docker) |

**Integration test (optional, requires Docker):**
```bash
go test -tags=integration ./internal/docker/... -v
```

---

## Completion Checklist

- [x] Client interface defined with all network methods
- [x] Functional options: WithHost, WithTimeout, WithAPIVersion
- [x] DockerError type with proper wrapping
- [x] MockClient for testing
- [x] NetworkCreate with IPAM support
- [x] NetworkRemove with idempotency
- [x] NetworkList with filters
- [x] NetworkInspect with full details
- [x] All unit tests pass
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] Exit criteria from SPEC.md verified

---

## Status

**COMPLETE** - All tasks finished.
