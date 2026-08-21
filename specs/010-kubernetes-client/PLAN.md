# Iteration 010: Kubernetes Client Plan

## Overview

Build the smallest safe Kubernetes boundary first: read trusted kubeconfig state, select an existing context and namespace in memory, and make one context-aware read-only API call. The result gives later apply, pack, and fleet work a tested client foundation without granting this iteration mutation capability.

---

## Phases

### Phase A: Dependency and Read-Only Client Contract

**Duration**: 30 minutes

**Objective**: Add the compatible Kubernetes SDK dependencies and define the client/options contract.

**Deliverables**:
- `k8s.io/client-go` and `k8s.io/apimachinery` pinned at `v0.34.10`.
- `internal/kubernetes/client.go` with `Client`, `Option`, and `NewClient`.
- Kubeconfig loading with migration and credential persistence disabled.
- In-memory context and namespace selection.

**Dependencies**: Iterations 002, 003, and 004.

### Phase B: Connectivity and Typed Errors

**Duration**: 30 minutes

**Objective**: Implement an authenticated, cancellable read-only API-server probe with useful error mapping.

**Deliverables**:
- `Client.Probe(ctx)` using `GET /version`.
- Context timeout handling.
- Typed error mapping and improved `KubernetesError` messages.

**Dependencies**: Phase A.

### Phase C: Isolated Client Tests

**Duration**: 30 minutes

**Objective**: Prove behavior using temporary kubeconfigs and TLS API servers, without a cluster.

**Deliverables**:
- Context and namespace precedence tests.
- Kubeconfig immutability tests.
- Probe success, status failure, malformed response, transport, cancellation, and timeout tests.

**Dependencies**: Phases A and B.

---

## Verification

After completion:
- [x] Kubeconfig source files remain byte-for-byte unchanged.
- [x] Context switching is local to a Yar client instance.
- [x] Probe performs only a read-only request and honors cancellation.
- [x] Credentials are never logged or exposed through Yar's public API.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
