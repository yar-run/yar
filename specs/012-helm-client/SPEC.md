# Iteration 012: Helm Client Specification

## Overview

Add Yar's Helm SDK boundary for chart loading, deterministic local rendering, and release lifecycle operations. Helm gives Yar a second production representation of the same declarative system model, but it must operate as a contained SDK integration: no Helm CLI, no kubeconfig rewrites, no discovery cache files, no cluster calls during local rendering, and no adoption of external releases.

## Scope

### Included
- Add `helm.sh/helm/v3 v3.19.5` with the Kubernetes module set pinned to Yar's existing `v0.34.10` release line.
- Load and validate Helm charts from an explicit directory, archive, or in-memory files.
- Merge user values with chart defaults using Helm's coalescing rules.
- Deterministically render a chart locally using pinned capabilities, no cluster connectivity, and no filesystem output.
- Suppress core Kubernetes `Secret` objects from rendered output.
- Provide a private, in-memory Helm REST client getter backed by Yar's already-selected Kubernetes client configuration.
- Provide release install, upgrade, and uninstall wrappers with deterministic release identity and ownership enforcement.
- Prove behavior using Helm's in-memory release driver and fake Kubernetes client, without a Kubernetes cluster or Helm CLI.

### NOT Included (deferred)
- OCI registry login, chart pull/push, repository management, dependency download, packaging, or publishing.
- Helm CLI invocation.
- Kubernetes live-cluster integration tests.
- Server-side Helm dry-run, hooks, waits, atomic rollback, rollback history, and release status/list commands.
- Bundled or remote chart dependency processing. This iteration accepts only self-contained charts without dependencies to keep rendering offline and prevent Helm's dependency warning paths from logging values.
- Helm generator output (iteration 020) and fleet Kubernetes driver integration (iteration 025).
- Adoption of existing Helm releases or Kubernetes resources.

---

## Compatibility

Use:

```text
helm.sh/helm/v3 v3.19.5
k8s.io/api v0.34.10
k8s.io/apimachinery v0.34.10
k8s.io/apiserver v0.34.10
k8s.io/cli-runtime v0.34.10
k8s.io/client-go v0.34.10
```

Helm 3.19.5 requires Go 1.24 and Kubernetes 0.34.x, so it remains compatible with Yar's `go 1.24.5` baseline and Kubernetes 0.34.10 implementation. This iteration must not upgrade Yar to Go 1.25 or a different Kubernetes minor release.

---

## Safety Model

### Render

`Render` is a local transformation, not an operation on a cluster.

- It uses `action.Install` with `ClientOnly: true` and `DryRun: true`.
- It pins Kubernetes version and API versions from `RenderOptions`; no live discovery is permitted.
- DNS template functions are disabled.
- It never writes a rendered manifest to disk and never logs it.
- It sets Helm's `HideSecret` option, suppressing rendered `v1/Secret` objects. Callers must still treat all rendered output as sensitive because a chart can place values in non-Secret objects.
- Values must contain secret references or target-native secret mechanisms, never resolved secret values.

### Release Ownership

Helm stores release state separately from Yar's direct server-side-apply resources. Helm ownership is therefore established by both the release identity and release storage labels:

```yaml
yar.io/project: <project>
yar.io/environment: <environment>
```

`Release` must include non-empty, Kubernetes-valid `Name`, `Namespace`, `Project`, and `Environment` fields.

Mutation rules:

1. `Install` validates that a release name is absent before creating it. Existing releases are never adopted.
2. `Install` and `Upgrade` attach exact Yar project/environment labels to Helm release storage.
3. `Upgrade` and `Uninstall` load the existing release first and require exact Yar label matches.
4. External, unlabeled, partially labeled, or mismatched releases are refused before mutation.
5. Helm's `TakeOwnership` option remains false for all operations.
6. `Uninstall` ignores only a missing release, making cleanup idempotent; it never ignores ownership mismatches.
7. Release mutations use the in-memory Kubernetes selection derived by Yar, never a second kubeconfig lookup or disk cache.

These rules satisfy **INV-OWN-001** and **INV-HLM-003**.

---

## Interfaces

```go
// Client provides chart and release operations.
type Client interface {
    LoadChart(path string) (*chart.Chart, error)
    LoadChartFiles(files []ChartFile) (*chart.Chart, error)
    Render(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts RenderOptions) (*RenderResult, error)
    Install(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts InstallOptions) (*release.Release, error)
    Upgrade(ctx context.Context, chart *chart.Chart, release Release, values map[string]any, opts UpgradeOptions) (*release.Release, error)
    Uninstall(ctx context.Context, release Release, opts UninstallOptions) error
}

type ChartFile struct {
    Name string
    Data []byte
}

type Release struct {
    Name        string
    Namespace   string
    Project     string
    Environment string
}

type RenderOptions struct {
    KubeVersion string
    APIVersions []string
    IncludeCRDs bool
}

type RenderResult struct {
    Manifest string
    Notes    string
}

type InstallOptions struct {
    DryRun bool
}

type UpgradeOptions struct {
    DryRun bool
}

type UninstallOptions struct {
    DryRun bool
}
```

### Construction

```go
// NewClient creates an isolated Helm client.
func NewClient(opts ...Option) (Client, error)

// WithKubernetesClient supplies Yar's already-selected in-memory Kubernetes client.
func WithKubernetesClient(client kubernetes.Client) Option

// WithTimeout sets a default lifecycle-operation timeout.
func WithTimeout(timeout time.Duration) Option
```

Pure chart loading and rendering do not require a Kubernetes client. Install, upgrade, and uninstall do.

The internal Kubernetes-to-Helm adapter is package-private. It exposes only a copied REST configuration, in-memory discovery cache, and REST mapper to Helm's `RESTClientGetter` interface; it never exposes bearer tokens or client credentials in public fields or errors.

---

## Error Behavior

| Condition | Error |
|-----------|-------|
| Missing chart path | `*errors.NotFoundError` with resource `chart` |
| Invalid chart/archive/files or values | `*errors.ValidationError` or `*errors.PackError` with preserved cause |
| Invalid release/project/environment/name/namespace | `*errors.ValidationError` |
| Render failure | `*errors.PackError` with `Pack` set to the chart name |
| Helm install/upgrade/uninstall failure | `*errors.KubernetesError` with operation `helm.install`, `helm.upgrade`, or `helm.uninstall` and a redacted operational cause |
| Existing external or mismatched release | `*errors.KubernetesError`; no release mutation |
| Missing uninstall target | success |
| `context.Canceled` or `context.DeadlineExceeded` | returned unchanged |

Helm chart content, values, rendered manifests, release values, kubeconfig credentials, and secret material must not appear in errors or logs.

---

## Implementation Design

### Chart Loading and Values

- Use `chart/loader.Load` for explicit filesystem paths.
- Use `chart/loader.LoadFiles` for `ChartFile` input.
- Validate chart metadata through Helm's loader and chart validation.
- Copy caller values before coalescing. Yar must not mutate caller-owned maps.
- Use Helm coalescing semantics for defaults and user overrides.

### Deterministic Render

- Create a fresh `action.Configuration` for each render.
- Use `action.NewInstall` with `ClientOnly`, `DryRun`, `HideSecret`, `EnableDNS: false`, and `OutputDir: ""`.
- Pin `KubeVersion` and `APIVersions`; defaults are Helm's static default capabilities if options are empty.
- Check cancellation before rendering, then render synchronously and return only manifest and notes in `RenderResult`; this prevents Helm from continuing background work after Yar has returned.
- Do not initialize Helm action configuration with a disk-backed REST getter for render.

### Lifecycle Operations

- Build action configuration from the internal in-memory Kubernetes adapter.
- Production construction uses Helm's secret release driver in the selected namespace; test construction may use its in-memory storage driver and fake kube client.
- `Install` runs only after confirming no release with the same name exists; it sets release storage labels.
- `Upgrade` and `Uninstall` load the current release from Helm storage and verify exact Yar ownership labels before acting.
- Set `TakeOwnership: false`, disable hooks, and do not configure waits, atomic rollback, or server-side dry run in this iteration.
- Helm 3 lifecycle actions are not safely cancellable: `RunWithContext` can return while installation continues in the background. Yar checks cancellation before starting an action, then runs lifecycle operations synchronously to ensure it never returns while a release mutation continues. `WithTimeout` configures Helm's action timeout for supported internal waits; waits and hooks are disabled in this iteration.

---

## Invariants

Reference applicable root invariants:

- **INV-SEC-001**, **INV-SEC-002**, **INV-SEC-003**, **INV-SEC-004**: Never persist, log, or resolve secret values during chart handling.
- **INV-OWN-001**: Never mutate an external Helm release or adopt resources.
- **INV-OWN-002**: Never mutate kubeconfig or platform configuration.
- **INV-K8S-001**: Context selection and kubeconfig use remain in-memory.
- **INV-HLM-001**: Render is deterministic and offline.
- **INV-HLM-002**: Lifecycle uses Yar's selected in-memory Kubernetes context.
- **INV-HLM-003**: Lifecycle uses exact Yar release ownership checks.
- **INV-DEV-003**: Implementation follows TDD.

---

## File Manifest

| File | Purpose |
|------|---------|
| `go.mod` | Helm 3.19.5 and matching Kubernetes staging module requirements |
| `go.sum` | Dependency checksums |
| `internal/helm/client.go` | Client options, lifecycle configuration, ownership checks |
| `internal/helm/chart.go` | Chart loading and deterministic rendering |
| `internal/helm/client_test.go` | Lifecycle and ownership tests using Helm memory storage |
| `internal/helm/chart_test.go` | Chart/load/render fixture tests |
| `internal/helm/testdata/` | Minimal valid, invalid, secret, and capability fixture charts |
| `internal/kubernetes/helm.go` | Package-private in-memory Helm REST getter adapter |
| `internal/kubernetes/helm_test.go` | Adapter immutability tests |

---

## Test Requirements

### Chart and Render Tests
- [x] Load valid charts from a directory and in-memory files.
- [x] Reject missing and invalid charts with typed errors.
- [x] Render is deterministic for identical chart, values, capabilities, and release identity.
- [x] Render uses specified Kubernetes version and API versions without contacting a server.
- [x] Render does not resolve process environment values, DNS, kubeconfig, or a cluster.
- [x] Render suppresses core `v1/Secret` output and never writes output to disk.
- [x] Values merge with chart defaults without mutating caller values.
- [x] Render errors do not include literal supplied secret fixture values.

### Lifecycle Tests
- [x] Install stamps exact Yar release labels and refuses an existing release.
- [x] Upgrade mutates only an exact Yar-owned release.
- [x] Upgrade refuses external, partial-label, and mismatched-project releases.
- [x] Uninstall is idempotent for a missing release.
- [x] Uninstall refuses external or mismatched releases before mutation.
- [x] Dry-run lifecycle operations do not commit release-storage changes.
- [x] Context cancellation and operation timeout remain inspectable.
- [x] Lifecycle errors retain their cause through `KubernetesError`.

### Kubernetes Adapter Tests
- [x] Adapter uses the selected Kubernetes client context/namespace in memory.
- [x] Adapter construction leaves kubeconfig bytes unchanged and creates no discovery/cache files.
- [x] Adapter never invokes a second kubeconfig load.

### Test Infrastructure
- Use fixture charts plus Helm memory release storage and fake kube clients.
- Do not require Docker, Helm CLI, kubectl, network access, or a Kubernetes cluster.

---

## Exit Criteria

- [x] Helm 3.19.5 and Kubernetes dependencies retain Yar's Go 1.24.5 / K8s 0.34.10 baseline.
- [x] Chart loading and local rendering work without a cluster or filesystem output.
- [x] Render suppresses core Kubernetes Secret manifests.
- [x] Lifecycle operations honor exact Yar release ownership.
- [x] No Helm operation reads/writes kubeconfig or discovery caches outside Yar's in-memory selection.
- [x] Tests prove safe operation without Helm CLI or live cluster.
- [x] `go mod tidy -diff -compat=1.24` succeeds.
- [x] `go build ./...` succeeds.
- [x] `go test ./...` passes.
- [x] `go vet ./...` is clean.
