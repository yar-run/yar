# Iteration 001: Foundation Plan

## Overview

This iteration implements foundational infrastructure in three phases: error types, platform detection, and CLI skeleton. All phases follow TDD. Total estimated time: 2 hours.

---

## Phases

### Phase A: Error Types

**Duration**: 30 minutes

**Objective**: Implement all typed error definitions with proper formatting and unwrapping.

**Deliverables**:
- `internal/errors/errors.go` - All error type definitions
- `internal/errors/errors_test.go` - Unit tests

**Dependencies**: None (foundational package)

**Approach**:
1. Write tests for `ConfigError` first (red)
2. Implement `ConfigError` (green)
3. Repeat for each error type in order:
   - `ValidationError`
   - `NotFoundError`
   - `SecretError`
   - `PackError`
   - `DockerError`
   - `KubernetesError`
   - `NetworkError`

### Phase B: Platform Detection

**Duration**: 30 minutes

**Objective**: Implement platform detection and XDG-compliant path utilities.

**Deliverables**:
- `internal/platform/platform.go` - Platform detection and paths
- `internal/platform/platform_test.go` - Unit tests

**Dependencies**: Phase A (for error types, though not strictly required)

**Approach**:
1. Write test for `Platform()` (red)
2. Implement `Platform()` (green)
3. Write test for `HomeDir()` (red)
4. Implement `HomeDir()` (green)
5. Write test for `ConfigDir()` with XDG (red)
6. Implement `ConfigDir()` (green)
7. Write test for `CacheDir()` with XDG (red)
8. Implement `CacheDir()` (green)
9. Write test for `DataDir()` with XDG (red)
10. Implement `DataDir()` (green)
11. Write test for `ExpandPath()` (red)
12. Implement `ExpandPath()` (green)

### Phase C: CLI Skeleton

**Duration**: 1 hour

**Objective**: Implement complete CLI command structure with all commands as stubs.

**Deliverables**:
- `cmd/root.go` - Root command with global flags
- `cmd/fleet.go` - Fleet commands
- `cmd/config.go` - Config commands
- `cmd/project.go` - Project commands
- `cmd/pack.go` - Pack commands
- `cmd/template.go` - Template commands
- `cmd/secret.go` - Secret commands
- `cmd/hosts.go` - Hosts commands
- `cmd/doctor.go` - Doctor command
- `cmd/aliases.go` - Nautical aliases

**Dependencies**: Phase B (for platform.ConfigDir() in config command)

**Approach**:
1. Update root.go with global flags
2. Implement each command group as stubs
3. Implement aliases
4. Verify all commands work

---

## Verification

After completion:
- [x] `go build ./...` succeeds
- [x] `go test ./internal/errors/...` passes
- [x] `go test ./internal/platform/...` passes
- [x] `go vet ./...` clean
- [x] All error types format correctly
- [x] Platform detection returns correct OS
- [x] Path utilities respect XDG environment variables
- [x] `./yar --help` shows all commands
- [x] All stub commands execute without error
- [x] All aliases work correctly
