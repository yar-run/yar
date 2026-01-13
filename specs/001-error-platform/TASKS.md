# Iteration 001: Foundation Tasks

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

**Verify:**
```bash
go build ./...
go test ./internal/errors/...
go vet ./...
```

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

**Verify:**
```bash
go build ./...
go test ./internal/platform/...
go vet ./...
```

---

## Phase C: CLI Skeleton

### C1. Root Command

**Implement:**
- [x] Update `cmd/root.go` with global flags (--verbose, --output)
- [x] Verify `./yar --help` works

### C2. Fleet Commands

**Implement:**
- [x] Create `cmd/fleet.go` with up/down/destroy/restart/status/update subcommands
- [x] Each prints stub message
- [x] Verify `./yar fleet --help` works

### C3. Config Commands

**Implement:**
- [x] Create `cmd/config.go` with get/edit subcommands
- [x] `get` uses `platform.ConfigDir()` for real path
- [x] Verify `./yar config get` shows path

### C4. Project Commands

**Implement:**
- [x] Create `cmd/project.go` with init/get/edit subcommands
- [x] Verify `./yar project --help` works

### C5. Pack Commands

**Implement:**
- [x] Create `cmd/pack.go` with list/install/remove subcommands
- [x] Verify `./yar pack --help` works

### C6. Template Commands

**Implement:**
- [x] Create `cmd/template.go` with build/render/publish subcommands
- [x] Verify `./yar template --help` works

### C7. Secret Commands

**Implement:**
- [x] Create `cmd/secret.go` with set/get/delete/list/sync subcommands
- [x] Verify `./yar secret --help` works

### C8. Hosts Commands

**Implement:**
- [x] Create `cmd/hosts.go` with set/get/delete/list subcommands
- [x] Verify `./yar hosts --help` works

### C9. Doctor Command

**Implement:**
- [x] Create `cmd/doctor.go` with run subcommand
- [x] Accepts --fix and --fix-cache flags
- [x] Prints health check table
- [x] Verify `./yar doctor run` works

### C10. Aliases

**Implement:**
- [x] Create `cmd/aliases.go` with hoist/dock/scuttle/swab/up/down
- [x] Each alias calls appropriate fleet/doctor command
- [x] Verify `./yar hoist` works

**Verify:**
```bash
go build ./...
go test ./...
go vet ./...
./yar --help
./yar fleet up
./yar config get
./yar doctor run
./yar hoist
```

---

## Functional Tests

| Command | Expected Result | Status |
|---------|-----------------|--------|
| `yar --help` | Shows all commands | [x] Pass |
| `yar fleet up` | Prints stub message | [x] Pass |
| `yar fleet down` | Prints stub message | [x] Pass |
| `yar config get` | Shows real config path | [x] Pass |
| `yar doctor run` | Shows health table | [x] Pass |
| `yar hoist` | Same as fleet up | [x] Pass |
| `yar --version` | Shows version | [x] Pass |

---

## Completion Checklist

- [x] All tests written and passing
- [x] All error types match SPEC.md interfaces
- [x] All platform functions match SPEC.md interfaces
- [x] All CLI commands callable (stubs)
- [x] All aliases work correctly
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] Functional tests verified
- [x] TASKS.md fully checked off
- [x] Exit criteria from SPEC.md verified
