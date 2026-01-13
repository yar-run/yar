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
- [ ] Write test for `ConfigError.Error()` formatting
- [ ] Write test for `ConfigError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `ConfigError` struct and methods
- [ ] Verify tests pass (green)

### A2. ValidationError

**Test First:**
- [ ] Write test for `ValidationError.Error()` with single error
- [ ] Write test for `ValidationError.Error()` with multiple errors
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `ValidationError` struct and methods
- [ ] Verify tests pass (green)

### A3. NotFoundError

**Test First:**
- [ ] Write test for `NotFoundError.Error()` formatting
- [ ] Verify test fails (red)

**Implement:**
- [ ] Implement `NotFoundError` struct and methods
- [ ] Verify test passes (green)

### A4. SecretError

**Test First:**
- [ ] Write test for `SecretError.Error()` formatting
- [ ] Write test for `SecretError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `SecretError` struct and methods
- [ ] Verify tests pass (green)

### A5. PackError

**Test First:**
- [ ] Write test for `PackError.Error()` formatting
- [ ] Write test for `PackError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `PackError` struct and methods
- [ ] Verify tests pass (green)

### A6. DockerError

**Test First:**
- [ ] Write test for `DockerError.Error()` formatting
- [ ] Write test for `DockerError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `DockerError` struct and methods
- [ ] Verify tests pass (green)

### A7. KubernetesError

**Test First:**
- [ ] Write test for `KubernetesError.Error()` formatting
- [ ] Write test for `KubernetesError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `KubernetesError` struct and methods
- [ ] Verify tests pass (green)

### A8. NetworkError

**Test First:**
- [ ] Write test for `NetworkError.Error()` formatting
- [ ] Write test for `NetworkError.Unwrap()`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `NetworkError` struct and methods
- [ ] Verify tests pass (green)

### A9. Interface Compliance

**Test First:**
- [ ] Write test verifying all types satisfy `error` interface
- [ ] Write test verifying Unwrap types satisfy `interface{ Unwrap() error }`

**Implement:**
- [ ] Ensure all types compile and satisfy interfaces
- [ ] Verify tests pass (green)

---

## Phase B: Platform Detection

### B1. Platform Detection

**Test First:**
- [ ] Write test for `Platform()` returns valid OS constant
- [ ] Verify test fails (red)

**Implement:**
- [ ] Implement `Platform()` using `runtime.GOOS`
- [ ] Verify test passes (green)

### B2. HomeDir

**Test First:**
- [ ] Write test for `HomeDir()` returns non-empty path
- [ ] Verify test fails (red)

**Implement:**
- [ ] Implement `HomeDir()` using `os.UserHomeDir()`
- [ ] Verify test passes (green)

### B3. ConfigDir

**Test First:**
- [ ] Write test for `ConfigDir()` returns non-empty path
- [ ] Write test for `ConfigDir()` respects `$XDG_CONFIG_HOME`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `ConfigDir()` with XDG fallback logic
- [ ] Verify tests pass (green)

### B4. CacheDir

**Test First:**
- [ ] Write test for `CacheDir()` returns non-empty path
- [ ] Write test for `CacheDir()` respects `$XDG_CACHE_HOME`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `CacheDir()` with XDG fallback logic
- [ ] Verify tests pass (green)

### B5. DataDir

**Test First:**
- [ ] Write test for `DataDir()` returns non-empty path
- [ ] Write test for `DataDir()` respects `$XDG_DATA_HOME`
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `DataDir()` with XDG fallback logic
- [ ] Verify tests pass (green)

### B6. ExpandPath

**Test First:**
- [ ] Write test for `ExpandPath()` expands `~` to home
- [ ] Write test for `ExpandPath()` expands `$VAR` and `${VAR}`
- [ ] Write test for `ExpandPath()` handles no expansion needed
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement `ExpandPath()` with tilde and env var expansion
- [ ] Verify tests pass (green)

---

## Completion Checklist

- [ ] All tests written and passing
- [ ] All error types match SPEC.md interfaces
- [ ] All platform functions match SPEC.md interfaces
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] TASKS.md fully checked off
- [ ] Exit criteria from SPEC.md verified
