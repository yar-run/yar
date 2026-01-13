# Iteration 008: Docker Containers Plan

## Overview

This iteration extends the Docker client wrapper with container lifecycle operations. We follow the same patterns established in iteration 007: types file, implementation file, error constructors, MockClient extensions, and idiomatic Go tests.

---

## Phases

### Phase A: Container Types

**Duration**: 20 minutes

**Objective**: Define all container-related types and option structs.

**Deliverables**:
- `internal/docker/container_types.go` - Container, ContainerConfig, option types
- `internal/docker/container_types_test.go` - Type tests

**Dependencies**: Existing docker package from iteration 007.

### Phase B: Container Errors

**Duration**: 10 minutes

**Objective**: Add error constructors for container operations.

**Deliverables**:
- Extend `internal/docker/errors.go` with container error constructors
- Extend `internal/docker/errors_test.go` with container error tests

**Dependencies**: Phase A types.

### Phase C: MockClient Container Support

**Duration**: 25 minutes

**Objective**: Extend MockClient with container operation stubs.

**Deliverables**:
- Extend `internal/docker/mock.go` with container methods
- `internal/docker/mock_container_test.go` - Container mock tests

**Dependencies**: Phase A types, Phase B errors.

### Phase D: Container Implementation

**Duration**: 35 minutes

**Objective**: Implement container operations on dockerClient.

**Deliverables**:
- `internal/docker/container.go` - Container operations using Docker SDK
- Extend Client interface in `internal/docker/client.go`

**Dependencies**: Phase A, B, C complete.

---

## Verification

After completion:
- [ ] All new types are exported and documented
- [ ] Client interface includes all container methods
- [ ] MockClient implements extended Client interface
- [ ] All tests follow idiomatic Go patterns
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (50+ tests in docker package)
- [ ] `go vet ./...` clean
