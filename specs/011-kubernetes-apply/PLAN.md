# Iteration 011: Kubernetes Apply Plan

## Overview

Convert rendered manifests into safe Kubernetes API operations. The implementation uses server-side apply for desired-state convergence and ownership-checked delete for safe cleanup. Proof comes from real dynamic-client HTTP requests against a TLS test server rather than a live cluster.

---

## Phases

### Phase A: Client Mutation Foundation

**Duration**: 20 minutes

**Objective**: Extend the private Kubernetes client state and public interface without exposing credentials or making requests during construction.

**Deliverables**:
- Retained private REST config and shared HTTP client.
- Owner and field manager client options.
- Apply, delete, get, and list option types.
- Lazy dynamic client and REST mapper initialization.

### Phase B: Manifest Decode and Server-Side Apply

**Duration**: 25 minutes

**Objective**: Decode manifest streams, validate object identity, map GVKs, stamp ownership labels, and issue SSA requests.

**Deliverables**:
- Multi-document YAML/JSON decoder.
- Dynamic resource resolution with namespaced and cluster-scoped paths.
- `Apply` with field manager, force, dry-run, and context behavior.

### Phase C: Safe Delete, Get, and List

**Duration**: 25 minutes

**Objective**: Add resource inspection and cleanup that never adopts or removes external resources.

**Deliverables**:
- `Get` and `List` dynamic-client operations.
- Ownership-checked `Delete` with idempotent absent-resource behavior.
- API error mapping and deletion propagation handling.

### Phase D: Functional SDK Proof

**Duration**: 20 minutes

**Objective**: Verify actual HTTP method, path, query, headers, bodies, ownership refusal, and context behavior against TLS test servers.

**Deliverables**:
- `apply_test.go` with isolated test API server and REST mapper.
- Full module, build, test, vet, race, and coverage verification.

---

## Verification

After completion:
- [x] All mutations carry Yar ownership labels.
- [x] Applies refuse external resources before issuing POST or PATCH; deletes refuse them before DELETE.
- [x] Applying a desired state is idempotent at the client boundary.
- [x] Kubeconfig stays byte-for-byte unchanged.
- [x] All behavior is proven without a live Kubernetes cluster.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
