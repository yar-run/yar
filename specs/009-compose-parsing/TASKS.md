# Iteration 009: Compose Parsing Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Safe Compose Loader

### A1. Dependency and Contract

**Test First:**
- [x] Add a compile-level test for `LoadCompose` returning a typed project.
- [x] Verify the test fails before implementation.

**Implement:**
- [x] Add compose-go as a direct dependency.
- [x] Add `LoadCompose(ctx, paths...)` in `internal/docker/compose.go`.
- [x] Load only explicitly supplied file paths.
- [x] Disable interpolation and ambient environment, `.env`, `env_file`, and `label_file` resolution.
- [x] Preserve context cancellation errors.
- [x] Map other failures to `*errors.ConfigError`.

**Verify:**
- [x] `go build ./...` succeeds.
- [x] `go test ./internal/docker` passes.

---

## Phase B: Fixtures and Validation Tests

### B1. Valid Compose Projects

**Test First:**
- [x] Add a valid Compose fixture with services, dependencies, networks, volumes, ports, and a health check.
- [x] Add an override fixture.

**Implement:**
- [x] Verify typed service extraction.
- [x] Verify override merge behavior.
- [x] Verify relative path resolution from the Compose directory.

### B2. Invalid and Secret-Safe Input

**Test First:**
- [x] Add invalid YAML, schema, and consistency fixtures.
- [x] Add a literal interpolation fixture.

**Implement:**
- [x] Verify typed configuration errors for input failures.
- [x] Verify interpolation ignores ambient environment values.
- [x] Verify canceled contexts propagate unchanged.

**Verify:**
- [x] `go test ./internal/docker` passes.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.

---

## Functional Tests

This iteration exposes an internal parsing API only. No new CLI behavior is expected.

---

## Completion Checklist

- [x] All unit tests written and passing.
- [x] All interfaces match `SPEC.md`.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
- [x] TASKS.md fully checked off.
- [x] Exit criteria from `SPEC.md` verified.

---

## Status

**COMPLETE** - All tasks finished.
