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
- [ ] Write tests for Container struct field access
- [ ] Write tests for ContainerState status values
- [ ] Verify tests compile

**Implement:**
- [ ] Create `internal/docker/container_types.go`
- [ ] Define Container struct with ID, Name, Image, Command, Created, State, Ports, Labels, NetworkIDs
- [ ] Define ContainerState with Status, Running, Paused, ExitCode, StartedAt, FinishedAt, Error

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### A2. ContainerConfig Types

**Test First:**
- [ ] Write tests for ContainerConfig with all fields
- [ ] Write tests for PortMapping defaults
- [ ] Write tests for VolumeMount types
- [ ] Write tests for RestartPolicy values

**Implement:**
- [ ] Define ContainerConfig struct
- [ ] Define PortMapping struct
- [ ] Define VolumeMount struct
- [ ] Define RestartPolicy struct

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### A3. Option Types

**Test First:**
- [ ] Write tests for ContainerRemoveOptions
- [ ] Write tests for ContainerListOptions
- [ ] Write tests for ContainerLogOptions
- [ ] Write tests for WaitCondition constants
- [ ] Write tests for ContainerWaitResult

**Implement:**
- [ ] Define ContainerRemoveOptions
- [ ] Define ContainerListOptions
- [ ] Define ContainerLogOptions
- [ ] Define WaitCondition type and constants
- [ ] Define ContainerWaitResult

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Phase B: Container Errors

### B1. Error Constructors

**Test First:**
- [ ] Write tests for ErrContainerCreate
- [ ] Write tests for ErrContainerStart
- [ ] Write tests for ErrContainerStop
- [ ] Write tests for ErrContainerRemove
- [ ] Write tests for ErrContainerInspect
- [ ] Write tests for ErrContainerList
- [ ] Write tests for ErrContainerLogs
- [ ] Write tests for ErrContainerNotFound
- [ ] Write tests for ErrContainerAlreadyExists
- [ ] Write tests for ErrContainerRunning

**Implement:**
- [ ] Add container error constructors to errors.go

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Phase C: MockClient Container Support

### C1. MockClient Container Fields

**Implement:**
- [ ] Add container mock response fields to MockClient
- [ ] Add container call recording fields
- [ ] Add container callback fields

### C2. MockClient ContainerCreate

**Test First:**
- [ ] Test ContainerCreate returns configured ID
- [ ] Test ContainerCreate records call details
- [ ] Test ContainerCreate returns configured error
- [ ] Test ContainerCreate callback behavior

**Implement:**
- [ ] Implement MockClient.ContainerCreate

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### C3. MockClient ContainerStart/Stop

**Test First:**
- [ ] Test ContainerStart success
- [ ] Test ContainerStart records ID
- [ ] Test ContainerStart returns error
- [ ] Test ContainerStop success
- [ ] Test ContainerStop records ID and timeout
- [ ] Test ContainerStop idempotent behavior

**Implement:**
- [ ] Implement MockClient.ContainerStart
- [ ] Implement MockClient.ContainerStop

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### C4. MockClient ContainerRemove/Inspect

**Test First:**
- [ ] Test ContainerRemove success
- [ ] Test ContainerRemove with force
- [ ] Test ContainerRemove records options
- [ ] Test ContainerInspect returns result
- [ ] Test ContainerInspect not found error

**Implement:**
- [ ] Implement MockClient.ContainerRemove
- [ ] Implement MockClient.ContainerInspect

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### C5. MockClient ContainerList/Logs/Wait

**Test First:**
- [ ] Test ContainerList returns configured results
- [ ] Test ContainerList with filters
- [ ] Test ContainerLogs returns reader
- [ ] Test ContainerWait returns channels

**Implement:**
- [ ] Implement MockClient.ContainerList
- [ ] Implement MockClient.ContainerLogs
- [ ] Implement MockClient.ContainerWait

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

### C6. MockClient Reset Extension

**Test First:**
- [ ] Test Reset clears container call records

**Implement:**
- [ ] Extend MockClient.Reset to clear container fields

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Phase D: Container Implementation

### D1. Extend Client Interface

**Implement:**
- [ ] Add container methods to Client interface in client.go
- [ ] Verify MockClient still implements Client (compile check)

**Verify:**
- [ ] `go build ./...` succeeds

### D2. ContainerCreate Implementation

**Implement:**
- [ ] Implement dockerClient.ContainerCreate
- [ ] Map ContainerConfig to Docker SDK types
- [ ] Handle port mappings
- [ ] Handle volume mounts
- [ ] Handle network connections

**Verify:**
- [ ] `go build ./...` succeeds

### D3. ContainerStart/Stop Implementation

**Implement:**
- [ ] Implement dockerClient.ContainerStart
- [ ] Implement dockerClient.ContainerStop with timeout

**Verify:**
- [ ] `go build ./...` succeeds

### D4. ContainerRemove/Inspect Implementation

**Implement:**
- [ ] Implement dockerClient.ContainerRemove with options
- [ ] Implement dockerClient.ContainerInspect
- [ ] Convert Docker SDK response to Container type

**Verify:**
- [ ] `go build ./...` succeeds

### D5. ContainerList Implementation

**Implement:**
- [ ] Implement dockerClient.ContainerList
- [ ] Handle filter options
- [ ] Convert Docker SDK responses

**Verify:**
- [ ] `go build ./...` succeeds

### D6. ContainerLogs/Wait Implementation

**Implement:**
- [ ] Implement dockerClient.ContainerLogs
- [ ] Implement dockerClient.ContainerWait

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

---

## Completion Checklist

- [ ] All container types defined with JSON/YAML tags
- [ ] All error constructors implemented and tested
- [ ] Client interface extended with container methods
- [ ] MockClient implements all container operations
- [ ] dockerClient implements all container operations
- [ ] All unit tests follow Go idioms (table-driven, t.Parallel, cmp.Diff)
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (expect 70+ tests in docker package)
- [ ] `go vet ./...` clean
- [ ] TASKS.md fully checked off
