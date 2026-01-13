# Iteration 001: Error & Platform Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Error Types

### A1. ConfigError

**Test First:**
- [x] Write test for `ConfigError.Error()` formatting
- [x] Write test for `ConfigError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `ConfigError` struct and methods
- [x] Verify tests pass (green)

### A2. ValidationError

**Test First:**
- [x] Write test for `ValidationError.Error()` with single error
- [x] Write test for `ValidationError.Error()` with multiple errors
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `ValidationError` struct and methods
- [x] Verify tests pass (green)

### A3. NotFoundError

**Test First:**
- [x] Write test for `NotFoundError.Error()` formatting
- [x] Verify test fails (red)

**Implement:**
- [x] Implement `NotFoundError` struct and methods
- [x] Verify test passes (green)

### A4. SecretError

**Test First:**
- [x] Write test for `SecretError.Error()` formatting
- [x] Write test for `SecretError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `SecretError` struct and methods
- [x] Verify tests pass (green)

### A5. PackError

**Test First:**
- [x] Write test for `PackError.Error()` formatting
- [x] Write test for `PackError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `PackError` struct and methods
- [x] Verify tests pass (green)

### A6. DockerError

**Test First:**
- [x] Write test for `DockerError.Error()` formatting
- [x] Write test for `DockerError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `DockerError` struct and methods
- [x] Verify tests pass (green)

### A7. KubernetesError

**Test First:**
- [x] Write test for `KubernetesError.Error()` formatting
- [x] Write test for `KubernetesError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `KubernetesError` struct and methods
- [x] Verify tests pass (green)

### A8. NetworkError

**Test First:**
- [x] Write test for `NetworkError.Error()` formatting
- [x] Write test for `NetworkError.Unwrap()`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `NetworkError` struct and methods
- [x] Verify tests pass (green)

### A9. Interface Compliance

**Test First:**
- [x] Write test verifying all types satisfy `error` interface
- [x] Write test verifying Unwrap types satisfy `interface{ Unwrap() error }`

**Implement:**
- [x] Ensure all types compile and satisfy interfaces
- [x] Verify tests pass (green)

---

## Phase B: Platform Detection

### B1. Platform Detection

**Test First:**
- [x] Write test for `Platform()` returns valid OS constant
- [x] Verify test fails (red)

**Implement:**
- [x] Implement `Platform()` using `runtime.GOOS`
- [x] Verify test passes (green)

### B2. HomeDir

**Test First:**
- [x] Write test for `HomeDir()` returns non-empty path
- [x] Verify test fails (red)

**Implement:**
- [x] Implement `HomeDir()` using `os.UserHomeDir()`
- [x] Verify test passes (green)

### B3. ConfigDir

**Test First:**
- [x] Write test for `ConfigDir()` returns non-empty path
- [x] Write test for `ConfigDir()` respects `$XDG_CONFIG_HOME`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `ConfigDir()` with XDG fallback logic
- [x] Verify tests pass (green)

### B4. CacheDir

**Test First:**
- [x] Write test for `CacheDir()` returns non-empty path
- [x] Write test for `CacheDir()` respects `$XDG_CACHE_HOME`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `CacheDir()` with XDG fallback logic
- [x] Verify tests pass (green)

### B5. DataDir

**Test First:**
- [x] Write test for `DataDir()` returns non-empty path
- [x] Write test for `DataDir()` respects `$XDG_DATA_HOME`
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `DataDir()` with XDG fallback logic
- [x] Verify tests pass (green)

### B6. ExpandPath

**Test First:**
- [x] Write test for `ExpandPath()` expands `~` to home
- [x] Write test for `ExpandPath()` expands `$VAR` and `${VAR}`
- [x] Write test for `ExpandPath()` handles no expansion needed
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement `ExpandPath()` with tilde and env var expansion
- [x] Verify tests pass (green)

---

## Completion Checklist

- [x] All tests written and passing
- [x] All error types match SPEC.md interfaces
- [x] All platform functions match SPEC.md interfaces
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] TASKS.md fully checked off
- [x] Exit criteria from SPEC.md verified
