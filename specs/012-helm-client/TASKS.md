# Iteration 012: Helm Client Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Compatible Helm SDK and Chart Boundary

### A1. Dependency and Client Contract

**Test First:**
- [x] Add compile-level assertions for the Helm client interface.
- [x] Verify the test fails before implementation.

**Implement:**
- [x] Add Helm 3.19.5 and matching K8s staging module requirements.
- [x] Preserve Go 1.24.5 and K8s 0.34.10.
- [x] Define Client, Release, ChartFile, render, and lifecycle option types.
- [x] Define typed errors without exposing values or credentials.

### A2. Chart Loading

**Test First:**
- [x] Add valid directory, in-memory-file, invalid, and missing chart fixtures.

**Implement:**
- [x] Load explicit chart paths with Helm loader.
- [x] Load in-memory chart files.
- [x] Map chart failures to typed Yar errors.

---

## Phase B: Deterministic Local Rendering

### B1. Render Contract

**Test First:**
- [x] Add a deterministic render fixture with default values, overrides, pinned capability conditional, and a Secret.
- [x] Add a secret-literal regression fixture.

**Implement:**
- [x] Coalesce values without mutating caller input.
- [x] Render with client-only, dry-run Helm actions.
- [x] Pin capability inputs, disable DNS, avoid output directories, and hide core Secret output.
- [x] Return manifest and notes without logging/persisting output.

**Verify:**
- [x] Rendering needs no cluster, kubeconfig, Helm CLI, or network.
- [x] Supplied secret literal is absent from error/log output.

---

## Phase C: In-Memory Kubernetes Adapter and Release Safety

### C1. Adapter

**Test First:**
- [x] Add tests for an adapter built from Yar's selected Kubernetes client.
- [x] Add kubeconfig/cache immutability tests.

**Implement:**
- [x] Add package-private in-memory REST getter adapter.
- [x] Reuse selected REST config, discovery cache, mapper, and namespace.
- [x] Avoid Helm's default disk-backed config flags.

### C2. Release Lifecycle

**Test First:**
- [x] Add Helm memory-driver tests for install, upgrade, uninstall, ownership mismatch, missing release, dry run, cancellation, and timeout.

**Implement:**
- [x] Install only absent releases with exact Yar labels.
- [x] Upgrade and uninstall only exact Yar-owned releases.
- [x] Keep Helm takeover, hooks, waits, rollback, and server-side dry-run disabled.
- [x] Preserve typed causes and context errors.

**Verify:**
- [x] External releases never receive lifecycle mutation calls.

---

## Phase D: Final Verification

- [x] Run `go mod tidy -diff -compat=1.24`.
- [x] Run `go build ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go vet ./...`.
- [x] Run `go test -race -count=1 ./internal/helm`.
- [x] Update `SPEC.md`, `PLAN.md`, and `TASKS.md` from verified results.

---

## Functional Tests

This iteration exposes internal Helm APIs only. Functional proof uses offline chart rendering and Helm's memory release driver/fake kube client, never a Helm CLI or a live cluster.

---

## Completion Checklist

- [x] All unit and SDK boundary tests pass.
- [x] All interfaces match root and iteration `SPEC.md`.
- [x] Local Helm render is deterministic and never contacts a cluster.
- [x] Release mutations enforce exact Yar ownership.
- [x] Kubeconfig and external platform state remain unmodified.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
- [x] TASKS.md fully checked off.
- [x] Exit criteria from `SPEC.md` verified.

---

## Status

**COMPLETE** - All tasks finished.
