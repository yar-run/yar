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
- [ ] Create `internal/docker/types.go` with Network, IPAM, IPAMConfig structs
- [ ] Create NetworkCreateOptions, NetworkListOptions structs
- [ ] Add JSON tags for serialization

**Verify:**
- [ ] `go build ./...` succeeds

### A2. Error Types

**Implement:**
- [ ] Create `internal/docker/errors.go` with DockerError type
- [ ] Implement Error() and Unwrap() methods
- [ ] Add helper constructors for common errors

**Verify:**
- [ ] `go build ./...` succeeds

### A3. Client Interface and Options

**Test First:**
- [ ] Write test for NewClient with default options
- [ ] Write test for NewClient with WithHost option
- [ ] Write test for NewClient with WithTimeout option

**Implement:**
- [ ] Create `internal/docker/client.go` with Client interface
- [ ] Implement Option type and functional options (WithHost, WithTimeout, WithAPIVersion)
- [ ] Implement NewClient constructor
- [ ] Implement Ping method
- [ ] Implement Close method

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/docker/...` passes

---

## Phase B: Mock Client

### B1. Mock Implementation

**Implement:**
- [ ] Create `internal/docker/mock.go` with MockClient struct
- [ ] Implement all Client interface methods on MockClient
- [ ] Add fields for configuring mock responses
- [ ] Add fields for recording method calls

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] MockClient compiles and satisfies Client interface

---

## Phase C: Network Operations

### C1. NetworkCreate

**Test First:**
- [ ] Write test for NetworkCreate with default driver
- [ ] Write test for NetworkCreate with custom subnet
- [ ] Write test for NetworkCreate with labels
- [ ] Write test for NetworkCreate idempotency (already exists)
- [ ] Write test for NetworkCreate with invalid subnet

**Implement:**
- [ ] Implement NetworkCreate in `internal/docker/network.go`
- [ ] Handle IPAM configuration for custom subnets
- [ ] Handle "already exists" case idempotently
- [ ] Wrap errors with DockerError

**Verify:**
- [ ] `go test ./internal/docker/...` passes

### C2. NetworkRemove

**Test First:**
- [ ] Write test for NetworkRemove success
- [ ] Write test for NetworkRemove non-existent (idempotent)
- [ ] Write test for NetworkRemove with attached containers

**Implement:**
- [ ] Implement NetworkRemove
- [ ] Handle "not found" case idempotently
- [ ] Return clear error when containers attached

**Verify:**
- [ ] `go test ./internal/docker/...` passes

### C3. NetworkList

**Test First:**
- [ ] Write test for NetworkList returns all networks
- [ ] Write test for NetworkList with name filter
- [ ] Write test for NetworkList with label filter

**Implement:**
- [ ] Implement NetworkList
- [ ] Convert Docker API filters to our filter format
- [ ] Map Docker network types to our Network struct

**Verify:**
- [ ] `go test ./internal/docker/...` passes

### C4. NetworkInspect

**Test First:**
- [ ] Write test for NetworkInspect returns details
- [ ] Write test for NetworkInspect not found error

**Implement:**
- [ ] Implement NetworkInspect
- [ ] Map Docker network inspect response to our Network struct
- [ ] Include attached container IDs

**Verify:**
- [ ] `go test ./internal/docker/...` passes

---

## Functional Tests

After this iteration, verify with a running Docker daemon:

| Test | Command/Action | Expected Result |
|------|----------------|-----------------|
| Client connects | `client.Ping(ctx)` | Returns nil |
| Create network | `client.NetworkCreate(ctx, "yar-test", opts)` | Returns network ID |
| List networks | `client.NetworkList(ctx, opts)` | Includes "yar-test" |
| Inspect network | `client.NetworkInspect(ctx, "yar-test")` | Returns network details |
| Remove network | `client.NetworkRemove(ctx, "yar-test")` | Returns nil |
| Remove again | `client.NetworkRemove(ctx, "yar-test")` | Returns nil (idempotent) |

**Integration test (optional, requires Docker):**
```bash
go test -tags=integration ./internal/docker/... -v
```

---

## Completion Checklist

- [ ] Client interface defined with all network methods
- [ ] Functional options: WithHost, WithTimeout, WithAPIVersion
- [ ] DockerError type with proper wrapping
- [ ] MockClient for testing
- [ ] NetworkCreate with IPAM support
- [ ] NetworkRemove with idempotency
- [ ] NetworkList with filters
- [ ] NetworkInspect with full details
- [ ] All unit tests pass
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] Exit criteria from SPEC.md verified

---

## Status

**PENDING** - Ready for implementation.
