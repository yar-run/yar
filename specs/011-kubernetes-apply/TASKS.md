# Iteration 011: Kubernetes Apply Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Client Mutation Foundation

### A1. Interface and Private Client State

**Test First:**
- [x] Add compile-level assertions for the extended client interface.
- [x] Add option validation tests for owner and field manager identity.

**Implement:**
- [x] Extend `Client` with Apply, Delete, Get, and List.
- [x] Define ApplyOptions, DeleteOptions, and ListOptions.
- [x] Retain private REST config and HTTP client from iteration 010.
- [x] Add WithOwner and WithFieldManager options.
- [x] Add lazy dynamic client and REST mapper initialization.

**Verify:**
- [x] Client construction remains kubeconfig-read-only.
- [x] `go test ./internal/kubernetes` passes.

---

## Phase B: Manifest Decode and Apply

### B1. Decode and Validate Manifest Streams

**Test First:**
- [x] Add tests for multi-document YAML, empty documents, malformed YAML, missing GVK/name, and managed fields.

**Implement:**
- [x] Decode YAML/JSON streams into unstructured objects.
- [x] Validate manifest identity without logging or persisting content.
- [x] Resolve GVKs and dynamic resource interfaces by scope.

### B2. Server-Side Apply

**Test First:**
- [x] Add TLS API-server tests asserting SSA method, path, headers, query, and labels.
- [x] Add namespaced, cluster-scoped, dry-run, force, cancellation, and timeout tests.

**Implement:**
- [x] Stamp Yar ownership labels.
- [x] Apply owned resources with `types.ApplyPatchType`; create confirmed-absent resources with field-managed POST.
- [x] Preserve API status and context errors.

**Verify:**
- [x] Reapplying manifests is client-side idempotent.
- [x] `go test ./internal/kubernetes` passes.

---

## Phase C: Get, List, and Safe Delete

### C1. Get and List

**Test First:**
- [x] Add scoped Get and List API-server tests.

**Implement:**
- [x] Implement Get and List with documented namespace behavior.
- [x] Map status and transport failures to typed Kubernetes errors.

### C2. Ownership-Checked Delete

**Test First:**
- [x] Add tests for owned, missing, external, partial-label, and mismatched-project delete targets.

**Implement:**
- [x] GET before DELETE.
- [x] Verify exact ownership labels.
- [x] Treat NotFound as an idempotent success.
- [x] Apply validated deletion propagation settings and live UID/resource-version preconditions.

**Verify:**
- [x] External resources never receive DELETE requests.

---

## Phase D: Final Verification

- [x] Run `go mod tidy -diff -compat=1.24`.
- [x] Run `go build ./...`.
- [x] Run `go test -count=1 ./...`.
- [x] Run `go vet ./...`.
- [x] Run `go test -race -count=1 ./internal/kubernetes`.
- [x] Update iteration specs and tasks from verified results.

---

## Functional Tests

This iteration exposes an internal Kubernetes resource API only. Its TLS test server proves actual SDK requests; it does not contact a live cluster or add CLI behavior.

---

## Completion Checklist

- [x] All unit and SDK boundary tests pass.
- [x] All interfaces match root and iteration `SPEC.md`.
- [x] Apply creates missing resources or server-side applies exact Yar-owned matches with ownership labels.
- [x] Delete only removes exact Yar-owned matches.
- [x] Kubeconfig and unrelated cluster resources remain unmodified.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
- [x] TASKS.md fully checked off.
- [x] Exit criteria from `SPEC.md` verified.

---

## Status

**COMPLETE** - All tasks finished.
