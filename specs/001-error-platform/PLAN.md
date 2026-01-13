# Iteration 001: Error & Platform Plan

## Overview

This iteration implements error types and platform detection in two phases. Phase A implements all error types with TDD. Phase B implements platform detection utilities with TDD. Total estimated time: 1 hour.

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

---

## Verification

After completion:
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/errors/...` passes
- [ ] `go test ./internal/platform/...` passes
- [ ] `go vet ./...` clean
- [ ] All error types format correctly
- [ ] Platform detection returns correct OS
- [ ] Path utilities respect XDG environment variables
