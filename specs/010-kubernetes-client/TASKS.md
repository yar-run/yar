# Iteration 010: Kubernetes Client Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Dependency and Read-Only Client Contract

### A1. Kubernetes SDK Baseline

**Test First:**
- [x] Add a compile-level assertion that the concrete client implements `Client`.
- [x] Verify the test fails before implementation.

**Implement:**
- [x] Add matching `k8s.io/client-go` and `k8s.io/apimachinery` dependencies at `v0.34.10`.
- [x] Preserve the Go 1.24.5 module baseline.
- [x] Define `Client`, `Option`, `NewClient`, and functional options.

**Verify:**
- [x] `go build ./...` succeeds.
- [x] `go test ./internal/kubernetes` passes.

### A2. Kubeconfig and Selection Rules

**Test First:**
- [x] Add temporary kubeconfig tests for current-context selection, explicit context selection, and namespace precedence.
- [x] Add tests for missing kubeconfig, missing context, and invalid namespace.

**Implement:**
- [x] Load explicit kubeconfig paths or standard user kubeconfig precedence read-only.
- [x] Disable migration and credential persistence.
- [x] Select context and namespace only in memory.
- [x] Map configuration, not-found, and validation errors to Yar types.

**Verify:**
- [x] Kubeconfig bytes are unchanged after construction.
- [x] No kubeconfig lock file exists.

---

## Phase B: Connectivity and Typed Errors

### B1. Read-Only API Probe

**Test First:**
- [x] Add TLS server tests for successful version probe and selected bearer token.
- [x] Add API status, malformed response, transport, cancellation, and timeout tests.

**Implement:**
- [x] Create a private discovery REST client from the selected REST configuration.
- [x] Implement `Probe(ctx)` with `GET /version`.
- [x] Apply configured timeout while preserving earlier caller deadlines.
- [x] Return `*errors.KubernetesError` for non-context probe failures.

### B2. Actionable Kubernetes Errors

**Test First:**
- [x] Add a `KubernetesError` formatting test with an underlying cause.

**Implement:**
- [x] Include the underlying cause in `KubernetesError.Error()`.
- [x] Preserve `Unwrap` behavior.

**Verify:**
- [x] `go test ./internal/errors ./internal/kubernetes` passes.

---

## Phase C: Final Verification

- [x] Run `go mod tidy -diff -compat=1.24`.
- [x] Run `go build ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go vet ./...`.
- [x] Run `go test -race -count=1 ./internal/kubernetes`.
- [x] Update `SPEC.md`, `PLAN.md`, and `TASKS.md` from verified results.

---

## Functional Tests

This iteration exposes an internal read-only client API only. No new CLI behavior is expected.

---

## Completion Checklist

- [x] All unit tests written and passing.
- [x] All interfaces match `SPEC.md`.
- [x] Kubeconfig and cluster resources remain unmodified.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
- [x] TASKS.md fully checked off.
- [x] Exit criteria from `SPEC.md` verified.

---

## Status

**COMPLETE** - All tasks finished.
