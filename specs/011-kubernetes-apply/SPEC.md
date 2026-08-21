# Iteration 011: Kubernetes Apply Specification

## Overview

Extend Yar's Kubernetes client with safe manifest operations: server-side apply, ownership-checked delete, get, and list. These operations turn rendered Kubernetes artifacts into a testable platform boundary while retaining Yar's central rule: it may mutate only resources it can identify as belonging to the declared Yar project and environment.

## Scope

### Included
- Decode multi-document YAML or JSON manifest streams into unstructured Kubernetes objects.
- Resolve GroupVersionKind to GroupVersionResource through a REST mapper.
- Handle namespaced and cluster-scoped resources correctly.
- Apply manifests through Kubernetes server-side apply (SSA).
- Stamp each applied resource with Yar ownership labels.
- Delete only live resources proven to be Yar-owned by matching ownership labels.
- Get one resource and list resources through the dynamic client.
- Preserve Kubernetes API status errors and context cancellation.
- Add end-to-end SDK proof through a TLS `httptest` API server, without a live cluster.

### NOT Included (deferred)
- Watch and wait operations.
- Helm release operations.
- Fleet orchestration or command wiring.
- Namespace creation or implicit namespace selection beyond the configured client namespace.
- Adoption or deletion of resources created outside Yar.
- Client-side apply, create/update fallback, or Kubernetes versions without SSA support.
- CRD discovery refresh guarantees beyond client-go's normal deferred REST mapper retry behavior.

---

## Safety Model

### Ownership Labels

Every object passed to `Apply` is stamped before the API request with:

```yaml
app.kubernetes.io/managed-by: yar
yar.io/project: <project>
yar.io/environment: <environment>
```

`WithOwner(project, environment)` supplies the required identity. Both values are required for mutation operations and must be valid Kubernetes label values.

### Mutation Rules

1. `Apply` creates a confirmed-absent resource with Yar ownership labels. For an existing exact Yar ownership match, it uses SSA with the default field manager `yar`, making repeated updates idempotent.
2. `Apply` first GETs an existing resource. It may create a missing resource or apply an exact Yar ownership match; external, unlabeled, partially labeled, or mismatched resources are refused before POST or PATCH. If a concurrent creator wins the create race, Yar re-reads and proves ownership before applying.
3. `Apply` adds or updates its ownership labels but does not copy unmanaged live labels into its desired payload; their ownership remains with their existing managers.
4. `Delete` first performs a GET for every declared object.
5. A resource is deletable only if all three ownership labels exactly match this client's owner identity.
6. `Delete` sends live UID and resource-version preconditions when available, preventing a GET-to-DELETE race from deleting a replacement or modified object.
7. A missing resource is a successful no-op.
8. Apply and delete are fail-fast but not transactional. A later failure does not roll back earlier successful objects; the error must identify the failed resource.
9. Cluster-scoped resources never receive a namespace path. For namespaced resources, manifest namespace wins over the client's selected namespace.

These rules enforce **INV-OWN-001**: ambiguous or external resources remain untouched.

---

## Interfaces

```go
// WithOwner identifies resources this client may manage.
func WithOwner(project, environment string) Option

// WithFieldManager overrides the server-side apply field manager.
func WithFieldManager(name string) Option

type ApplyOptions struct {
    FieldManager string
    Force        *bool // nil defaults to true; false preserves SSA conflicts
    DryRun       bool
}

type DeleteOptions struct {
    PropagationPolicy string // foreground, background, or orphan; empty uses the API default
}

type ListOptions struct {
    LabelSelector string
    Limit         int64
    Continue      string
}

// Apply applies a multi-document manifest stream through server-side apply.
func (c *client) Apply(ctx context.Context, manifests []byte, opts ApplyOptions) error

// Delete removes only Yar-owned resources declared by a multi-document manifest stream.
func (c *client) Delete(ctx context.Context, manifests []byte, opts DeleteOptions) error

// Get retrieves one resource by GVK, namespace, and name.
func (c *client) Get(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string) (*unstructured.Unstructured, error)

// List retrieves resources matching the supplied GVK, namespace, and options.
func (c *client) List(ctx context.Context, gvk schema.GroupVersionKind, namespace string, opts ListOptions) (*unstructured.UnstructuredList, error)
```

`Client` is extended with these operations. Watch and wait remain deferred.

### Namespace Rules

- Apply: `metadata.namespace` wins; otherwise use `Client.Namespace()` for namespaced kinds.
- Get: supplied namespace wins; empty means `Client.Namespace()` for namespaced kinds.
- List: supplied namespace lists that namespace; empty lists all namespaces for namespaced kinds.
- Delete: `metadata.namespace` wins; otherwise use `Client.Namespace()` for namespaced kinds.
- Cluster-scoped kinds ignore all namespace values.

---

## Implementation Design

### Client State

Iteration 010's private client retains the selected `*rest.Config` and shared `*http.Client` so iteration 011 can lazily construct:

- `dynamic.Interface` using `dynamic.NewForConfigAndClient`
- `meta.RESTMapper` using a deferred discovery mapper

The REST configuration, client transport, credentials, and mapper remain private. Production construction uses the discovery mapper; tests may inject an unexported REST mapper to isolate resource-operation behavior from discovery response fixtures.

### Manifest Decoding

- Use `yaml.NewYAMLOrJSONDecoder` over the provided bytes.
- Decode each document into `unstructured.Unstructured`.
- Ignore empty/comment-only documents.
- Require `apiVersion`, `kind`, and `metadata.name` for every non-empty document.
- Reject manifest documents containing `metadata.managedFields`; this prevents applying server-owned state.
- Never log or persist manifest bytes. This iteration receives generated artifacts and must not write secrets embedded by caller error to disk or logs.

### REST Mapping

- Resolve a GVK through `meta.RESTMapper.RESTMapping`.
- Use the mapping's resource and scope to construct a namespaced or cluster-scoped dynamic resource interface.
- Map failures to `*errors.KubernetesError` with `Op: "mapping"` and the requested GVK identity.

### Apply Behavior

- Serialize the sanitized unstructured object to JSON.
- Create confirmed-absent resources with a field-managed Kubernetes `POST`; update existing Yar-owned resources with `PATCH` using `types.ApplyPatchType` and `metav1.PatchOptions{FieldManager: ...}`.
- Default field manager: `yar`.
- `ApplyOptions.Force == nil` defaults to true for Yar-owned fields; an explicit false preserves conflicts for caller review.
- `ApplyOptions.DryRun` sends `DryRunAll` and does not alter server state.

### Delete Behavior

- Decode the same manifest format as `Apply`.
- GET each live resource, inspect ownership labels, then DELETE only exact ownership matches.
- Translate `DeleteOptions.PropagationPolicy` to Kubernetes deletion propagation values.
- Do not offer a force flag that bypasses ownership verification.

---

## Error Behavior

| Condition | Error |
|-----------|-------|
| Empty or invalid manifest document | `*errors.ValidationError` |
| Missing GVK/name in manifest | `*errors.ValidationError` |
| Invalid owner identity, field manager, selector, limit, or propagation policy | `*errors.ValidationError` |
| REST mapping failure | `*errors.KubernetesError{Op: "mapping"}` |
| Apply/get/list/delete API or decode failure | `*errors.KubernetesError` with the operation and resource identity |
| Apply or delete target is not owned by this Yar project/environment | `*errors.KubernetesError`; no POST, PATCH, or DELETE request |
| Delete target does not exist | Success (idempotent no-op) |
| `context.Canceled` or `context.DeadlineExceeded` | Returned unchanged |

Kubernetes status errors remain unwrapped under `KubernetesError` so callers can use `apierrors.IsNotFound`, `IsForbidden`, `IsConflict`, `IsInvalid`, and related predicates.

---

## Invariants

Reference applicable root invariants:

- **INV-OWN-001**: Delete only exact Yar-owned resource matches.
- **INV-OWN-002**: Client construction remains read-only for machine configuration.
- **INV-K8S-001**: Kubeconfig state remains in-memory and unmodified.
- **INV-K8S-002**: Every API operation respects cancellation and deadlines.
- **INV-FLT-004**: Server-side apply makes repeated desired-state application idempotent.
- **INV-SEC-001** and **INV-SEC-002**: Manifests and errors must not cause secret values to be written or logged.

---

## File Manifest

| File | Purpose |
|------|---------|
| `SPEC.md` | Extend root Kubernetes client contract with options |
| `internal/kubernetes/client.go` | Retain private REST state; add owner and field-manager options |
| `internal/kubernetes/apply.go` | Manifest decode, mapping, apply, delete, get, and list |
| `internal/kubernetes/apply_test.go` | TLS API-server proof tests for all resource operations |
| `internal/kubernetes/client_test.go` | Owner/field-manager option validation tests |

---

## Test Requirements

### Unit and SDK Boundary Tests
- [x] Valid multi-document YAML creates a Deployment and Service through real dynamic-client HTTP requests.
- [x] Apply uses SSA `PATCH` for exact Yar-owned resources with apply content type, field manager, and force settings.
- [x] Apply stamps exact Yar ownership labels without removing unrelated labels.
- [x] Reapplying an exact Yar-owned manifest converges through the client boundary without client-side state.
- [x] Dry-run sends the Kubernetes dry-run query and does not rely on resource state.
- [x] Namespaced apply uses manifest namespace over client namespace; absent namespace uses client namespace.
- [x] Cluster-scoped apply does not add a namespace URL segment.
- [x] Empty documents are ignored; malformed YAML and incomplete documents return validation errors.
- [x] Apply rejects input containing managed fields.
- [x] Get and List construct correct scoped API requests and map returned unstructured objects.
- [x] Empty namespace List requests all namespaces for namespaced resources.
- [x] Delete first GETs the live object and DELETEs only an exact Yar ownership match.
- [x] Delete missing objects is idempotent success.
- [x] Delete external or mismatched resources refuses without a DELETE request.
- [x] API status, malformed response, transport failure, cancellation, and timeout preserve typed causes.
- [x] Kubeconfig remains unchanged and no cluster mutation occurs outside expected test-server requests.

### Test Infrastructure
- Reuse temporary kubeconfigs and `httptest.NewTLSServer` from iteration 010.
- Use a package-private REST mapper seam with real dynamic HTTP requests.
- Tests must not require Docker, kubectl, or an accessible Kubernetes cluster.

---

## Exit Criteria

- [x] Manifests are decoded and validated without being written to disk.
- [x] Apply creates missing resources or server-side applies exact Yar-owned resources with ownership labels.
- [x] Delete refuses external or ambiguous resources.
- [x] Get and List correctly handle namespaced and cluster-scoped resources.
- [x] Kubernetes API errors and context cancellation remain inspectable.
- [x] TLS-backed functional SDK tests cover mutation and refusal paths.
- [x] `go mod tidy -diff -compat=1.24` succeeds.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
