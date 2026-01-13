# Iteration 007: Docker Client Plan

## Overview

This iteration implements the Docker SDK wrapper for network operations using TDD. We'll build the client with functional options pattern, then implement network CRUD operations with proper error handling and idempotency.

---

## Phases

### Phase A: Client Foundation

**Duration**: 30 minutes

**Objective**: Create the Docker client wrapper with functional options and connection handling.

**Deliverables**:
- `internal/docker/types.go` - Data structures
- `internal/docker/errors.go` - DockerError type
- `internal/docker/client.go` - Client interface, constructor, options

**Dependencies**: 
- Docker SDK added to go.mod
- `internal/errors` package (exists)

### Phase B: Mock Client

**Duration**: 15 minutes

**Objective**: Create a mock client for unit testing without Docker daemon.

**Deliverables**:
- `internal/docker/mock.go` - MockClient implementation

**Dependencies**: Phase A complete

### Phase C: Network Operations

**Duration**: 45 minutes

**Objective**: Implement network Create, Remove, List, Inspect with tests.

**Deliverables**:
- `internal/docker/network.go` - Network operation implementations
- `internal/docker/network_test.go` - Unit tests using mock

**Dependencies**: Phase A and B complete

---

## Verification

After completion:
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/docker/...` passes
- [ ] `go vet ./...` clean
- [ ] Client connects to Docker daemon (manual test)
- [ ] NetworkCreate creates a test network (manual test)
- [ ] NetworkList shows the created network (manual test)
- [ ] NetworkRemove removes the network (manual test)

---

## Manual Integration Test

After unit tests pass, verify against real Docker:

```bash
# Build yar (not needed yet, but verify build)
go build ./...

# Test client connection (via go test with build tag)
go test -tags=integration ./internal/docker/... -v

# Or manually in Go playground/test file:
# client, _ := docker.NewClient()
# client.Ping(context.Background())
# client.NetworkCreate(ctx, "yar-test-net", docker.NetworkCreateOptions{Subnet: "172.30.0.0/24"})
# client.NetworkList(ctx, docker.NetworkListOptions{})
# client.NetworkRemove(ctx, "yar-test-net")
```
