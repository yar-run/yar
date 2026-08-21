# Iteration 010: Kubernetes Client Specification

## Overview

Add the read-only, SDK-native Kubernetes client foundation that lets Yar load a trusted kubeconfig, select a context and namespace in memory, and verify API-server connectivity. This is the Kubernetes counterpart to the Docker client wrapper: it establishes safe, typed platform access without yet applying, deleting, or otherwise mutating cluster resources.

## Scope

### Included
- Add compatible `client-go` and `apimachinery` dependencies.
- Load an explicit kubeconfig or the user's standard kubeconfig precedence.
- Select a kubeconfig context without changing the source kubeconfig's current context.
- Resolve a namespace from an explicit option, selected kubeconfig context, or the `default` fallback.
- Validate explicit namespace overrides as Kubernetes DNS-1123 labels.
- Create a private REST and discovery client from the selected context.
- Probe `GET /version` with the caller's context and return `version.Info`.
- Return typed Yar errors with preserved underlying Kubernetes status and transport errors.
- Ensure read-only kubeconfig loading and read-only cluster probing through tests.

### NOT Included (deferred)
- Applying, deleting, getting, listing, watching, or waiting on Kubernetes resources (iteration 011).
- Helm operations (iteration 012).
- Fleet Kubernetes orchestration (iteration 025).
- In-cluster configuration, kubeconfig generation, or kubeconfig editing.
- Namespace creation, context persistence, credential persistence, or any cluster mutation.
- An insecure TLS bypass option.

---

## Interfaces

### Client

```go
// Client provides read-only Kubernetes connection operations.
type Client interface {
    // Context returns the kubeconfig context selected for this client.
    Context() string

    // Namespace returns the namespace selected for this client.
    Namespace() string

    // Probe verifies API-server connectivity and returns server version information.
    Probe(ctx context.Context) (*version.Info, error)
}
```

Iteration 011 extends this interface with resource operations from the root `SPEC.md`. This iteration deliberately exposes only the connection boundary needed to construct and verify a safe client.

### Construction

```go
// NewClient loads a trusted kubeconfig and creates a read-only Kubernetes client.
func NewClient(opts ...Option) (Client, error)

// WithKubeconfig selects an explicit kubeconfig path.
func WithKubeconfig(path string) Option

// WithContext selects a kubeconfig context in memory.
func WithContext(name string) Option

// WithNamespace overrides the namespace selected from kubeconfig.
func WithNamespace(namespace string) Option

// WithTimeout sets the default timeout for connection operations.
func WithTimeout(timeout time.Duration) Option
```

### Selection Rules

1. `WithKubeconfig` selects one explicit kubeconfig. Without it, client-go's standard `KUBECONFIG` / `~/.kube/config` read precedence applies.
2. `WithContext` selects a context in memory. Without it, the kubeconfig current context is used.
3. `WithNamespace` wins over the selected context namespace. If neither supplies a namespace, Yar uses `default`.
4. The selected context and namespace apply only to this client instance. Yar never calls `kubectl`, `ModifyConfig`, `WriteToFile`, or any credential persister.
5. Kubeconfig files are trusted user inputs. Yar never auto-loads a project-provided or downloaded kubeconfig. An `exec` authentication plugin contained in a kubeconfig may execute when an authenticated request is made, so users must trust the kubeconfig they select.

### Probe

- `Probe` issues only `GET /version` against the selected API server.
- It uses the supplied `context.Context`; cancellation and deadlines propagate unchanged.
- A configured timeout bounds the operation without overriding an earlier caller deadline.
- It returns a decoded `*version.Info` on success.
- It never creates, updates, deletes, or lists cluster resources.

---

## Error Behavior

| Condition | Error |
|-----------|-------|
| Explicit kubeconfig path does not exist | `*errors.NotFoundError` with resource `kubeconfig` |
| No usable default kubeconfig exists | `*errors.NotFoundError` with resource `kubeconfig` |
| Selected context does not exist | `*errors.NotFoundError` with resource `kubeconfig context` |
| Kubeconfig read, decode, TLS, or client configuration failure | `*errors.ConfigError` |
| Invalid explicit namespace | `*errors.ValidationError` |
| API-server transport, status, or version-decode failure | `*errors.KubernetesError` with `Op: "probe"` |
| `context.Canceled` or `context.DeadlineExceeded` | Returned unchanged |

`KubernetesError.Error()` must include its underlying cause when one exists while preserving `Unwrap`, so users receive actionable connection and API status details without losing programmatic error inspection.

---

## Data Structures

```go
type clientOptions struct {
    kubeconfig string
    context    string
    namespace  string
    timeout    time.Duration
}

type client struct {
    context   string
    namespace string
    timeout   time.Duration
    discovery discovery.DiscoveryInterface
}
```

The REST configuration, bearer tokens, certificates, key material, and auth-provider settings remain private. They must never be logged or returned from the public interface.

---

## Dependencies

### External Packages
- `k8s.io/client-go v0.34.10` - Kubeconfig loading, REST configuration, and discovery client.
- `k8s.io/apimachinery v0.34.10` - Server version types and Kubernetes namespace validation.

All direct `k8s.io/*` modules must use the same `v0.34.10` release line. This is compatible with Yar's `go 1.24.5` baseline; this iteration must not raise the project's Go version.

### Internal Packages
- `internal/errors` - Typed configuration, validation, not-found, and Kubernetes operation errors.
- `internal/config` - Future callers pass its cluster context and namespace configuration into this client through options.

---

## Invariants

Reference applicable invariants from root `SPEC.md`:

- **INV-OWN-001**: This iteration must not mutate external cluster resources.
- **INV-OWN-002**: Kubeconfig loading must be read-only.
- **INV-K8S-001**: Context and namespace selection must be in memory only.
- **INV-K8S-002**: Probes must be read-only and context-aware.
- **INV-SEC-002**: Kubeconfig credentials, tokens, certificates, headers, and request bodies must never appear in logs.
- **INV-DEV-003**: Implementation follows TDD.

---

## File Manifest

| File | Purpose |
|------|---------|
| `go.mod` | Direct client-go and apimachinery dependencies |
| `go.sum` | Dependency checksums |
| `internal/kubernetes/client.go` | Read-only kubeconfig client and probe implementation |
| `internal/kubernetes/client_test.go` | Kubeconfig, context, namespace, probe, and immutability tests |
| `internal/errors/errors.go` | Include underlying Kubernetes errors in user-facing messages |
| `internal/errors/errors_test.go` | Error-message regression test |

---

## Test Requirements

### Unit Tests
- [x] `NewClient` implements the `Client` interface.
- [x] An explicit kubeconfig selects its current context and context namespace.
- [x] `WithContext` selects a different context in memory and reaches only that API server.
- [x] `WithNamespace` overrides the context namespace.
- [x] Missing namespace defaults to `default`.
- [x] Invalid namespace returns `*errors.ValidationError`.
- [x] Missing explicit kubeconfig returns `*errors.NotFoundError`.
- [x] Missing selected context returns `*errors.NotFoundError`.
- [x] Invalid kubeconfig returns `*errors.ConfigError` and preserves the cause.
- [x] Constructor and probe leave kubeconfig bytes unchanged and create no lock file.
- [x] Probe sends the selected context's credentials and decodes `/version`.
- [x] Unauthorized, forbidden, service-unavailable, malformed-version, and transport failures return `*errors.KubernetesError` with preserved causes.
- [x] Cancellation and timeout propagate through `Probe`.
- [x] `KubernetesError.Error()` includes its underlying cause.

### Test Infrastructure
- Use `httptest.NewTLSServer` with CA data embedded in temporary kubeconfigs.
- Use temporary kubeconfig files and HTTP handlers rather than a live cluster or a fake clientset.
- Tests must not require Docker, kubectl, or an accessible Kubernetes cluster.

### Integration Tests
- Deferred to iteration 032.

---

## Exit Criteria

- [x] Client-go dependencies are pinned to compatible, matching Kubernetes module versions.
- [x] A kubeconfig can be loaded and context switched without changing source files.
- [x] Client namespace selection follows documented precedence.
- [x] A read-only, cancellable API-server probe works against a TLS test server.
- [x] All error paths are typed and actionable without exposing credentials.
- [x] No kubeconfig or cluster resource mutation occurs.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
