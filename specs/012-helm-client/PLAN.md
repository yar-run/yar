# Iteration 012: Helm Client Plan

## Overview

Build Helm as a safe rendering and release boundary for Yar. Begin with pure local chart rendering, then add memory-proven lifecycle logic. The production Kubernetes adapter must reuse Yar's selected in-memory context instead of Helm's disk-backed defaults.

---

## Phases

### Phase A: Compatible Helm SDK and Chart Boundary

**Duration**: 25 minutes

**Objective**: Add Helm 3.19.5 without moving Yar's Go or Kubernetes compatibility line, then expose typed chart loading.

**Deliverables**:
- Helm and matching Kubernetes staging module dependencies.
- Chart file, release, render, and lifecycle types.
- Chart load methods with typed error mapping.

### Phase B: Deterministic Local Rendering

**Duration**: 25 minutes

**Objective**: Render charts safely without cluster access, disk output, or secret exposure.

**Deliverables**:
- Client-only Helm render action.
- Pinned capabilities and disabled DNS.
- Values coalescing and secret suppression.
- Fixture chart tests.

### Phase C: In-Memory Kubernetes Adapter and Release Safety

**Duration**: 25 minutes

**Objective**: Bridge Helm lifecycle actions to Yar's selected Kubernetes context without file-backed configuration or release adoption.

**Deliverables**:
- Package-private Kubernetes REST getter adapter.
- Helm install/upgrade/uninstall wrappers.
- Deterministic release identity and ownership label enforcement.
- Memory-driver and fake-kube test harness.

### Phase D: Functional Proof and Verification

**Duration**: 15 minutes

**Objective**: Verify Helm SDK behavior without external platform dependencies.

**Deliverables**:
- Offline render and lifecycle tests.
- Module consistency, full build/test/vet/race verification.

---

## Verification

After completion:
- [x] Chart rendering is offline, deterministic, and secret-safe.
- [x] Lifecycle never adopts or removes an external release.
- [x] Helm uses Yar-selected in-memory Kubernetes configuration only.
- [x] No kubeconfig/discovery cache file is written.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
