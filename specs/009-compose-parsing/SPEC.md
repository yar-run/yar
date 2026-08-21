# Iteration 009: Compose Parsing Specification

## Overview

Add a safe, typed Docker Compose parser to `internal/docker`. This parser gives Yar an authoritative way to read existing Compose files while it transitions teams away from bespoke local-environment workflows. It uses compose-go's Compose-spec validation and typed project model without reading ambient process environment variables or `.env` files.

## Scope

### Included
- Add `github.com/compose-spec/compose-go/v2` as a direct dependency.
- Load one or more explicitly supplied Compose files into `*types.Project`.
- Validate Compose syntax, schema, service references, and project consistency through compose-go.
- Resolve paths relative to the primary Compose file.
- Disable interpolation and ambient environment, `.env`, `env_file`, and `label_file` resolution.
- Return typed Yar configuration errors for Compose input failures.
- Add fixtures and unit tests for valid and invalid Compose files.

### NOT Included (deferred)
- Generating Compose files from Yar packs (iteration 019).
- Starting or managing Compose services (iteration 024).
- Parsing `COMPOSE_FILE`, auto-discovering files, or reading process environment variables.
- Resolving secret values or Compose interpolation variables.
- Docker Compose CLI invocation.
- Health checks or runtime inspection.

---

## Interfaces

### LoadCompose

```go
// LoadCompose loads and validates explicitly supplied Compose files.
// It never reads ambient process environment variables or .env files.
func LoadCompose(ctx context.Context, paths ...string) (*composetypes.Project, error)
```

`LoadCompose` returns compose-go's typed, immutable project model. Callers can use `ServiceNames`, `GetService`, `NetworkNames`, `VolumeNames`, and related Compose-spec types without Yar introducing a duplicate model.

### Behavior

- At least one path is required.
- Paths are loaded in the provided order using Compose override semantics.
- The working directory is the directory containing the first file.
- Compose files are schema-validated and consistency-checked.
- Relative paths are normalized against the Compose working directory.
- Interpolation remains literal in this iteration.
- `context.Canceled` and `context.DeadlineExceeded` propagate unchanged.
- Other load failures return `*errors.ConfigError`, retaining the original error through `Unwrap`.

---

## Dependencies

### External Packages
- `github.com/compose-spec/compose-go/v2` - Compose specification parsing, validation, and typed project model.
- `github.com/compose-spec/compose-go/v2/cli` - Explicit project loading options.
- `github.com/compose-spec/compose-go/v2/types` - Returned `Project` and service types.

### Internal Packages
- `internal/errors` - Typed configuration errors for invalid Compose input.

---

## Invariants

Reference applicable invariants from root `SPEC.md`:

- **INV-CFG-001**: Compose configuration must validate before use.
- **INV-PCK-004**: Generated and consumed Compose configuration must be valid Compose YAML.
- **INV-SEC-001**: Parsing must not write secrets to disk.
- **INV-SEC-003**: Parsing must not load secret values from `.env` files or ambient process environment variables.
- **INV-DEV-003**: Implementation follows TDD.

---

## File Manifest

| File | Purpose |
|------|---------|
| `go.mod` | Direct compose-go dependency |
| `go.sum` | Dependency checksums |
| `internal/docker/compose.go` | Safe Compose loader |
| `internal/docker/compose_test.go` | Parser and error-mapping tests |
| `internal/docker/testdata/compose/valid.yaml` | Valid Compose project fixture |
| `internal/docker/testdata/compose/override.yaml` | Override merge fixture |
| `internal/docker/testdata/compose/invalid-yaml.yaml` | Invalid YAML fixture |
| `internal/docker/testdata/compose/invalid-schema.yaml` | Invalid Compose-schema fixture |
| `internal/docker/testdata/compose/invalid-consistency.yaml` | Invalid dependency reference fixture |
| `internal/docker/testdata/compose/literal-interpolation.yaml` | No-ambient-interpolation fixture |

---

## Test Requirements

### Unit Tests
- [x] Valid Compose file loads into a typed project.
- [x] Service, dependency, network, volume, port, and health-check details are available through the typed project.
- [x] Multiple files merge using Compose override semantics.
- [x] Relative build and bind-mount paths resolve from the Compose file directory.
- [x] Missing path returns `*errors.ConfigError` with the requested path.
- [x] Invalid YAML returns `*errors.ConfigError` and preserves the underlying error.
- [x] Invalid schema returns `*errors.ConfigError`.
- [x] Undefined service dependency returns `*errors.ConfigError`.
- [x] Interpolation remains literal and ignores ambient environment values.
- [x] Empty path input returns `*errors.ConfigError`.
- [x] Canceled contexts propagate without wrapping.

### Integration Tests
- Deferred to iteration 032. Parsing tests must not require a Docker daemon.

---

## Exit Criteria

- [x] Compose-go is a direct module dependency.
- [x] `LoadCompose` returns validated typed projects.
- [x] Parser never consumes ambient environment variables or `.env` files.
- [x] Compose input failures return typed configuration errors.
- [x] Unit tests are table-driven where appropriate and use `t.Parallel` safely.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
