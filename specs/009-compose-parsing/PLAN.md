# Iteration 009: Compose Parsing Plan

## Overview

Implement a minimal, safe boundary around compose-go. The parser consumes only explicitly named files and returns compose-go's typed project representation so future pack and fleet work can operate against the Compose specification without duplicating it.

---

## Phases

### Phase A: Safe Compose Loader

**Duration**: 30 minutes

**Objective**: Establish the dependency and a typed loading API with deliberate environment behavior.

**Deliverables**:
- Direct compose-go dependency in `go.mod`.
- `internal/docker/compose.go` with `LoadCompose`.
- Typed configuration-error mapping.

**Dependencies**: Iterations 002, 007, and 008.

### Phase B: Fixtures and Validation Tests

**Duration**: 30 minutes

**Objective**: Prove valid loading, Compose-spec validation, merge behavior, path resolution, and secret-safe interpolation behavior.

**Deliverables**:
- Compose fixtures under `internal/docker/testdata/compose/`.
- `internal/docker/compose_test.go`.

**Dependencies**: Phase A.

---

## Verification

After completion:
- [x] Valid Compose projects load without a Docker daemon.
- [x] Invalid input returns typed configuration errors.
- [x] Ambient environment and `.env` values do not affect parsed projects.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
