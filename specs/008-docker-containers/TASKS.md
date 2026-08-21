# Iteration 008: Docker Containers Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Container Types

### A1. Container and State Types

**Test First:**
- [x] Write tests for Container struct field access
- [x] Write tests for ContainerState status values
- [x] Verify tests compile

**Implement:**
- [x] Create `internal/docker/container_types.go`
- [x] Define Container struct with ID, Name, Image, Command, Created, State, Ports, Labels, NetworkIDs
- [x] Define ContainerState with Status, Running, Paused, ExitCode, StartedAt, FinishedAt, Error

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### A2. ContainerConfig Types

**Test First:**
- [x] Write tests for ContainerConfig with all fields
- [x] Write tests for PortMapping defaults
- [x] Write tests for VolumeMount types
- [x] Write tests for RestartPolicy values

**Implement:**
- [x] Define ContainerConfig struct
- [x] Define PortMapping struct
- [x] Define VolumeMount struct
- [x] Define RestartPolicy struct

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### A3. Option Types

**Test First:**
- [x] Write tests for ContainerRemoveOptions
- [x] Write tests for ContainerListOptions
- [x] Write tests for ContainerLogOptions
- [x] Write tests for WaitCondition constants
- [x] Write tests for ContainerWaitResult

**Implement:**
- [x] Define ContainerRemoveOptions
- [x] Define ContainerListOptions
- [x] Define ContainerLogOptions
- [x] Define WaitCondition type and constants
- [x] Define ContainerWaitResult

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

---

## Phase B: Container Errors

### B1. Error Constructors

**Test First:**
- [x] Write tests for ErrContainerCreate
- [x] Write tests for ErrContainerStart
- [x] Write tests for ErrContainerStop
- [x] Write tests for ErrContainerRemove
- [x] Write tests for ErrContainerInspect
- [x] Write tests for ErrContainerList
- [x] Write tests for ErrContainerLogs
- [x] Write tests for ErrContainerNotFound
- [x] Write tests for ErrContainerAlreadyExists
- [x] Write tests for ErrContainerRunning

**Implement:**
- [x] Add container error constructors to errors.go

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

---

## Phase C: MockClient Container Support

### C1. MockClient Container Fields

**Implement:**
- [x] Add container mock response fields to MockClient
- [x] Add container call recording fields
- [x] Add container callback fields

### C2. MockClient ContainerCreate

**Test First:**
- [x] Test ContainerCreate returns configured ID
- [x] Test ContainerCreate records call details
- [x] Test ContainerCreate returns configured error
- [x] Test ContainerCreate callback behavior

**Implement:**
- [x] Implement MockClient.ContainerCreate

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### C3. MockClient ContainerStart/Stop

**Test First:**
- [x] Test ContainerStart success
- [x] Test ContainerStart records ID
- [x] Test ContainerStart returns error
- [x] Test ContainerStop success
- [x] Test ContainerStop records ID and timeout
- [x] Test ContainerStop idempotent behavior

**Implement:**
- [x] Implement MockClient.ContainerStart
- [x] Implement MockClient.ContainerStop

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### C4. MockClient ContainerRemove/Inspect

**Test First:**
- [x] Test ContainerRemove success
- [x] Test ContainerRemove with force
- [x] Test ContainerRemove records options
- [x] Test ContainerInspect returns result
- [x] Test ContainerInspect not found error

**Implement:**
- [x] Implement MockClient.ContainerRemove
- [x] Implement MockClient.ContainerInspect

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### C5. MockClient ContainerList/Logs/Wait

**Test First:**
- [x] Test ContainerList returns configured results
- [x] Test ContainerList with filters
- [x] Test ContainerLogs returns reader
- [x] Test ContainerWait returns channels

**Implement:**
- [x] Implement MockClient.ContainerList
- [x] Implement MockClient.ContainerLogs
- [x] Implement MockClient.ContainerWait

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

### C6. MockClient Reset Extension

**Test First:**
- [x] Test Reset clears container call records

**Implement:**
- [x] Extend MockClient.Reset to clear container fields

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes

---

## Phase D: Container Implementation

### D1. Extend Client Interface

**Implement:**
- [x] Add container methods to Client interface in client.go
- [x] Verify MockClient still implements Client (compile check)

**Verify:**
- [x] `go build ./...` succeeds

### D2. ContainerCreate Implementation

**Implement:**
- [x] Implement dockerClient.ContainerCreate
- [x] Map ContainerConfig to Docker SDK types
- [x] Handle port mappings
- [x] Handle volume mounts
- [x] Handle network connections

**Verify:**
- [x] `go build ./...` succeeds

### D3. ContainerStart/Stop Implementation

**Implement:**
- [x] Implement dockerClient.ContainerStart
- [x] Implement dockerClient.ContainerStop with timeout

**Verify:**
- [x] `go build ./...` succeeds

### D4. ContainerRemove/Inspect Implementation

**Implement:**
- [x] Implement dockerClient.ContainerRemove with options
- [x] Implement dockerClient.ContainerInspect
- [x] Convert Docker SDK response to Container type

**Verify:**
- [x] `go build ./...` succeeds

### D5. ContainerList Implementation

**Implement:**
- [x] Implement dockerClient.ContainerList
- [x] Handle filter options
- [x] Convert Docker SDK responses

**Verify:**
- [x] `go build ./...` succeeds

### D6. ContainerLogs/Wait Implementation

**Implement:**
- [x] Implement dockerClient.ContainerLogs
- [x] Implement dockerClient.ContainerWait

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean

---

## Completion Checklist

- [x] All container types defined with JSON/YAML tags
- [x] All error constructors implemented and tested
- [x] Client interface extended with container methods
- [x] MockClient implements all container operations
- [x] dockerClient implements all container operations
- [x] All unit tests follow Go idioms (table-driven, t.Parallel, cmp.Diff)
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes (expect 70+ tests in docker package)
- [x] `go vet ./...` clean
- [x] TASKS.md fully checked off

---

## Status

**COMPLETE** - All tasks finished.
