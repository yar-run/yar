# Yar Design Document

This document captures the architectural decisions, technology choices, design patterns, and principles that guide Yar's implementation.

## Design Principles

### 1. SDK Over CLI
Use native Go SDKs for Docker, Kubernetes, and Helm rather than shelling out to CLI tools. This provides:
- Better error handling and typed responses
- No dependency on CLI tool versions
- Faster execution (no process spawning)
- Testability via mocks

**Exception**: Where no Go SDK exists (Colima, OpenVPN, pass), use CLI wrappers with structured output parsing.

### 2. Configuration as Code
All configuration is declarative YAML validated against JSON Schemas:
- Global config: `~/.config/yar/config.yaml`
- Project config: `./yar.yaml`
- Pack definitions: `meta.yaml`, `schema.json`, `resources.yaml`

### 3. Single Source of Truth
One pack definition generates all output formats:
- Docker Compose for local development
- Helm charts for Kubernetes
- Raw K8s manifests for simple deployments
- ESO ExternalSecret manifests for secret management

### 4. Secrets Never Touch Disk
Secret values are:
- Stored in encrypted stores (pass, Keychain, Credential Manager)
- Referenced by key in configuration (`passwordRef: redis_pass`)
- Resolved at runtime, never written to `.env` or config files
- Injected via Docker secrets or Kubernetes secret references

**Multi-developer guarantee**: When a developer adds a new secret reference to `yar.yaml` and commits, other developers will get an immediate, actionable error on `yar fleet up` listing exactly what secrets they need and how to obtain them. No more "why isn't this working?" debugging sessions.

### 5. Cross-Platform First
Design for macOS, Linux, and Windows from the start:
- Abstract platform-specific operations behind interfaces
- Use `go-keyring` for cross-platform secret store access
- Handle path separators, permissions, and privilege escalation per-platform

### 6. Fail Fast with Clear Errors
- Validate configuration before any operations
- Provide actionable error messages with fix suggestions
- `yar doctor` diagnoses and repairs common issues

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                                   │
│  cmd/                                                                    │
│  ├── root.go          Cobra root command, global flags                  │
│  ├── fleet.go         fleet up/down/destroy/restart/status              │
│  ├── config.go        config get/edit                                   │
│  ├── project.go       project init/get/edit                             │
│  ├── pack.go          pack list/install/remove                          │
│  ├── template.go      template build/render/publish                     │
│  ├── secret.go        secret set/get/delete/list/sync                   │
│  ├── hosts.go         hosts set/get/delete/list                         │
│  ├── doctor.go        doctor run                                        │
│  └── aliases.go       hoist/dock/scuttle/swab                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           Internal Packages                              │
├─────────────────────────────────────────────────────────────────────────┤
│  internal/config/      Configuration loading, validation, schemas       │
│  internal/fleet/       Fleet orchestration (compose + k8s drivers)      │
│  internal/docker/      Docker SDK wrapper                               │
│  internal/kubernetes/  client-go wrapper                                │
│  internal/helm/        Helm SDK wrapper                                 │
│  internal/secrets/     Secret provider interface + implementations      │
│  internal/packs/       Pack loading, validation, generation             │
│  internal/network/     VPN, DNS, hosts management                       │
│  internal/doctor/      Health checks and repairs                        │
│  internal/platform/    Platform-specific abstractions                   │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                           External SDKs                                  │
├─────────────────────────────────────────────────────────────────────────┤
│  github.com/docker/docker/client      Docker Engine API                 │
│  github.com/compose-spec/compose-go   Compose spec parsing              │
│  k8s.io/client-go                     Kubernetes API                    │
│  helm.sh/helm/v3                      Helm operations                   │
│  github.com/Azure/azure-sdk-for-go    Azure Key Vault                   │
│  github.com/hashicorp/vault/api       HashiCorp Vault                   │
│  github.com/zalando/go-keyring        Cross-platform keyring            │
└─────────────────────────────────────────────────────────────────────────┘
```

## Technology Choices

### Language: Go 1.22+
- Native compilation to single binary
- Excellent SDK support for Docker, K8s, Helm
- Strong concurrency primitives for parallel operations
- Cross-compilation for all target platforms

### CLI Framework: Cobra + Viper
- Cobra: Industry standard for Go CLIs (kubectl, helm, gh use it)
- Viper: Configuration loading from files, env vars, flags
- Automatic help generation and shell completion

### Configuration: YAML + JSON Schema
- YAML for human readability and editing
- JSON Schema for validation and IDE support
- `santhosh-tekuri/jsonschema` for Go-native validation

### Templating: Go text/template
- Same engine used by Helm
- Familiar to K8s ecosystem users
- Sprig functions for additional utilities

### Docker: docker/docker SDK
- Direct Docker Engine API access
- No dependency on docker CLI version
- Full control over container lifecycle

### Compose: compose-spec/compose-go
- Parse Compose files to typed structs
- Generate Compose files from structs
- Validate against Compose spec

### Kubernetes: client-go
- Official Kubernetes Go client
- Dynamic client for arbitrary resources
- Kubeconfig loading built-in

### Helm: helm.sh/helm/v3
- Chart loading and rendering
- Values merging
- No Tiller dependency (Helm 3)

### Secret Stores

| Platform | Primary | Fallback |
|----------|---------|----------|
| macOS | pass (GPG) | Keychain |
| Linux | pass (GPG) | Secret Service (GNOME Keyring, KWallet) |
| Windows | pass (via WSL or GoPass) | Credential Manager |

Cross-platform abstraction via `zalando/go-keyring` for OS-native stores.

### Remote Secret Providers

| Provider | SDK |
|----------|-----|
| Azure Key Vault | `Azure/azure-sdk-for-go/sdk/keyvault/azsecrets` |
| HashiCorp Vault | `hashicorp/vault/api` |
| 1Password | `1password/onepassword-sdk-go` |
| GitHub Secrets | `google/go-github` (REST API) |
| AWS Secrets Manager | `aws/aws-sdk-go-v2/service/secretsmanager` |
| GCP Secret Manager | `cloud.google.com/go/secretmanager` |

## Design Patterns

### 1. Provider Interface Pattern
All pluggable components implement interfaces:

```go
// Secret providers
type SecretProvider interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key, value string) error
    Delete(ctx context.Context, key string) error
    List(ctx context.Context) ([]string, error)
}

// Fleet drivers
type FleetDriver interface {
    Up(ctx context.Context, project *Project, env string) error
    Down(ctx context.Context, project *Project, env string) error
    Destroy(ctx context.Context, project *Project, env string) error
    Status(ctx context.Context, project *Project, env string) (*FleetStatus, error)
}

// Pack generators
type PackGenerator interface {
    Generate(pack *Pack, params map[string]any, env *Environment) ([]byte, error)
}
```

### 2. Registry Pattern
Packs and providers are registered in typed registries:

```go
type PackRegistry struct {
    builtin  map[string]*Pack  // Built-in packs (redis, kafka, etc.)
    installed map[string]*Pack // User-installed packs
    remote   []PackSource      // Remote pack sources
}

func (r *PackRegistry) Get(name string) (*Pack, error)
func (r *PackRegistry) List() []*PackInfo
func (r *PackRegistry) Install(name string, source string) error
```

### 3. Builder Pattern
Complex objects use builders for clarity:

```go
project := NewProjectBuilder().
    WithName("my-service").
    WithEnvironment("local", EnvConfig{Cluster: "local", Secrets: "pass"}).
    AddService(ServiceBuilder().
        WithName("redis").
        WithPack("redis").
        WithParam("passwordRef", "redis_pass").
        Build()).
    Build()
```

### 4. Functional Options
Configuration uses functional options for extensibility:

```go
client, err := docker.NewClient(
    docker.WithHost("unix:///var/run/docker.sock"),
    docker.WithTimeout(30 * time.Second),
    docker.WithTLSConfig(tlsConfig),
)
```

### 5. Context Propagation
All operations accept `context.Context` for cancellation and timeouts:

```go
func (f *Fleet) Up(ctx context.Context, env string) error {
    // Respect context cancellation throughout
}
```

## Package Structure

```
yar/
├── main.go                 Entry point
├── cmd/                    Cobra commands (thin wrappers)
│   └── *.go
├── internal/               Private packages
│   ├── config/
│   │   ├── global.go       ~/.config/yar/config.yaml
│   │   ├── project.go      ./yar.yaml
│   │   ├── schema.go       JSON Schema validation
│   │   └── types.go        Config structs
│   ├── docker/
│   │   ├── client.go       Docker SDK wrapper
│   │   ├── network.go      Network operations
│   │   ├── container.go    Container operations
│   │   └── compose.go      Compose integration
│   ├── kubernetes/
│   │   ├── client.go       client-go wrapper
│   │   ├── apply.go        Apply manifests
│   │   └── resources.go    Resource helpers
│   ├── helm/
│   │   ├── client.go       Helm SDK wrapper
│   │   ├── chart.go        Chart operations
│   │   └── values.go       Values merging
│   ├── secrets/
│   │   ├── provider.go     Provider interface
│   │   ├── pass.go         GNU pass
│   │   ├── keychain.go     macOS Keychain
│   │   ├── credman.go      Windows Credential Manager
│   │   ├── keyring.go      go-keyring abstraction
│   │   ├── azure.go        Azure Key Vault
│   │   ├── vault.go        HashiCorp Vault
│   │   ├── onepassword.go  1Password
│   │   ├── github.go       GitHub Secrets
│   │   └── sync.go         Remote → local sync
│   ├── packs/
│   │   ├── registry.go     Pack discovery
│   │   ├── loader.go       Load pack definitions
│   │   ├── schema.go       Validate pack params
│   │   ├── generator.go    Generator interface
│   │   ├── compose.go      → docker-compose.yaml
│   │   ├── helm.go         → Helm chart
│   │   ├── manifest.go     → raw K8s YAML
│   │   └── eso.go          → ExternalSecret
│   ├── fleet/
│   │   ├── driver.go       Driver interface
│   │   ├── compose.go      Compose driver
│   │   ├── kubernetes.go   K8s driver
│   │   └── orchestrator.go Driver selection
│   ├── network/
│   │   ├── vpn.go          OpenVPN management
│   │   ├── hosts.go        /etc/hosts management
│   │   └── dns.go          DNS configuration
│   ├── doctor/
│   │   ├── checks.go       Health checks
│   │   └── repair.go       Auto-fix
│   └── platform/
│       ├── platform.go     Platform detection
│       ├── darwin.go       macOS-specific
│       ├── linux.go        Linux-specific
│       └── windows.go      Windows-specific
├── packs/                  Built-in packs
│   ├── redis/
│   ├── kafka/
│   ├── postgres/
│   └── ...
├── schemas/                JSON Schemas
│   ├── config.schema.json
│   ├── project.schema.json
│   └── pack.schema.json
├── docs/                   Documentation
└── test/                   Test fixtures and integration tests
```

## Error Handling

### Error Types
Define typed errors for programmatic handling:

```go
type ConfigError struct {
    Path    string
    Field   string
    Message string
}

type SecretError struct {
    Provider string
    Key      string
    Op       string // "get", "set", "delete"
    Err      error
}

type PackError struct {
    Pack    string
    Message string
    Err     error
}
```

### Error Wrapping
Use `fmt.Errorf` with `%w` for error chains:

```go
if err != nil {
    return fmt.Errorf("failed to start container %s: %w", name, err)
}
```

### User-Facing Errors
Errors shown to users include:
1. What failed
2. Why it failed (if known)
3. How to fix it (if possible)

```
Error: Failed to connect to Docker daemon
Cause: Docker socket not found at /var/run/docker.sock
Fix:   Start Colima with 'colima start' or ensure Docker is running
```

## Testing Strategy

### Unit Tests
- All internal packages have unit tests
- Mock interfaces for external dependencies
- Table-driven tests for validation logic

### Integration Tests
- Use `testcontainers-go` for Docker integration
- Kind or k3d for Kubernetes integration
- Temporary directories for config file tests

### Test Fixtures
- Sample configs in `test/fixtures/`
- Sample packs in `test/fixtures/packs/`
- Golden files for generated output comparison

## Security Considerations

### Secret Handling
- Never log secret values
- Clear secret strings from memory after use
- Use secure comparison for secret validation

### Privilege Escalation
- `/etc/hosts` modification requires sudo
- Prompt user explicitly before privilege escalation
- Cache sudo credentials where OS allows

### Network Security
- VPN connections use TLS
- Validate server certificates
- Support mTLS for secure clusters

## Performance Considerations

### Parallel Operations
- Start independent containers in parallel
- Fetch secrets in parallel
- Use worker pools for bulk operations

### Caching
- Cache parsed pack definitions
- Cache validated schemas
- Invalidate on file modification time change

### Lazy Loading
- Don't load all packs at startup
- Load Kubernetes client only when needed
- Defer Docker connection until first use
