# Yar Technical Specification

This document defines the technical specifications, invariants, data structures, and DSL schemas that govern Yar's behavior.

## Table of Contents

1. [Invariants](#invariants)
2. [CLI Specification](#cli-specification)
3. [Configuration Schemas](#configuration-schemas)
4. [Pack DSL Specification](#pack-dsl-specification)
5. [Secret Reference Specification](#secret-reference-specification)
6. [Generator Output Specifications](#generator-output-specifications)
7. [API Contracts](#api-contracts)

---

## Invariants

These invariants MUST hold true at all times. Violations are bugs.

### Secret Invariants

1. **INV-SEC-001**: Secret values MUST NEVER be written to any file outside of encrypted stores (pass, Keychain, Credential Manager).

2. **INV-SEC-002**: Secret values MUST NEVER appear in logs at any log level.

3. **INV-SEC-003**: Secret values MUST NEVER be stored in environment variables in `.env` files or process environment exports visible in config.

4. **INV-SEC-004**: Secret references (`passwordRef`, `secretRef`) MUST be resolved at runtime, not at configuration load time.

5. **INV-SEC-005**: Failed secret resolution MUST halt operations with a clear error; secrets MUST NOT fall back to empty strings.

6. **INV-SEC-006**: Before starting any service, Yar MUST extract all secret references from `yar.yaml`, check each against the local store, and fail immediately if any are missing. The error MUST list all missing secrets with instructions to obtain them.

### Configuration Invariants

6. **INV-CFG-001**: All configuration files MUST validate against their JSON Schema before use.

7. **INV-CFG-002**: Project config (`yar.yaml`) MUST be present in the current directory or a parent directory for project-scoped commands.

8. **INV-CFG-003**: Global config (`~/.config/yar/config.yaml`) is optional; sensible defaults apply when absent.

9. **INV-CFG-004**: Environment names MUST be unique within a project.

10. **INV-CFG-005**: Service names MUST be unique within a project.

### Pack Invariants

11. **INV-PCK-001**: A pack MUST have `meta.yaml`, `schema.json`, and at least one template file.

12. **INV-PCK-002**: Pack parameters MUST validate against the pack's `schema.json` before generation.

13. **INV-PCK-003**: Pack templates MUST be valid Go templates that compile without error.

14. **INV-PCK-004**: Generated output MUST be valid for the target format (valid Compose YAML, valid Helm chart, valid K8s manifest).

### Fleet Invariants

15. **INV-FLT-001**: `fleet up` MUST start services in dependency order (services with `requires` wait for dependencies).

16. **INV-FLT-002**: `fleet down` MUST stop services in reverse dependency order.

17. **INV-FLT-003**: `fleet destroy` MUST remove all resources created by `fleet up`, including networks and volumes (unless `--keep-volumes` is specified).

18. **INV-FLT-004**: Fleet operations MUST be idempotent; running `fleet up` twice has the same result as running it once.

### Network Invariants

19. **INV-NET-001**: Container hostnames MUST be resolvable from the host machine when `fleet up` completes successfully.

20. **INV-NET-002**: `/etc/hosts` modifications MUST be reversible by `fleet down` or `hosts delete`.

21. **INV-NET-003**: Yar-managed host entries MUST be clearly marked with comments (e.g., `# yar:managed`).

### Development Process Invariants

22. **INV-DEV-001**: Each iteration MUST have specs created before implementation begins (`specs/{###}-{name}/SPEC.md`, `PLAN.md`, `TASKS.md`).

23. **INV-DEV-002**: During implementation, TASKS.md MUST be updated in real-time:
    - Mark task `[~]` when starting
    - Mark task `[x]` immediately upon completion
    - Mark task `[!]` if blocked, with note explaining why

24. **INV-DEV-003**: All iterations with testable code MUST follow TDD:
    - Write test first (red)
    - Implement until test passes (green)
    - Refactor if needed

25. **INV-DEV-004**: `go build ./...`, `go test ./...`, and `go vet ./...` MUST pass before marking an iteration complete.

26. **INV-DEV-005**: PROJECT.md CLI Reference is the true north. All implementation MUST align with the specified CLI behavior.

---

## CLI Specification

### Command Structure

```
yar <object> <verb> [args] [flags]
```

### Objects and Verbs

| Object | Verb | Arguments | Description |
|--------|------|-----------|-------------|
| `fleet` | `up` | `[env]` | Start services for environment (default: local) |
| `fleet` | `down` | `[env]` | Stop services |
| `fleet` | `destroy` | `[env]` | Remove all resources |
| `fleet` | `restart` | `[env]` | Restart services |
| `fleet` | `status` | `[env]` | Show service status |
| `fleet` | `update` | | Update yar and pack catalog |
| `config` | `get` | | Show global config |
| `config` | `edit` | | Open global config in editor |
| `project` | `init` | | Interactive project setup |
| `project` | `get` | | Show project config |
| `project` | `edit` | | Open project config in editor |
| `pack` | `list` | | List available packs |
| `pack` | `install` | `<name>` | Install a pack |
| `pack` | `remove` | `<name>` | Remove a pack |
| `template` | `build` | | Generate deployment artifacts |
| `template` | `render` | | Render templates to stdout |
| `template` | `publish` | | Push artifacts to registry |
| `secret` | `set` | `<key> <value>` | Set a secret |
| `secret` | `get` | `<key>` | Get a secret (redacted) |
| `secret` | `delete` | `<key>` | Delete a secret |
| `secret` | `list` | | List secret keys |
| `secret` | `sync` | | Sync remote secrets to local |
| `hosts` | `set` | `<name> <ip>` | Add host entry |
| `hosts` | `get` | `<name>` | Show host entry |
| `hosts` | `delete` | `<name>` | Remove host entry |
| `hosts` | `list` | | List yar-managed hosts |
| `doctor` | `run` | | Run health checks |

### Aliases

| Alias | Expands To |
|-------|------------|
| `yar hoist [env]` | `yar fleet up [env]` |
| `yar dock [env]` | `yar fleet down [env]` |
| `yar scuttle [env]` | `yar fleet destroy [env]` |
| `yar swab` | `yar doctor run --fix-cache` |

### Global Flags

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--help` | `-h` | bool | false | Show help |
| `--verbose` | `-v` | bool | false | Verbose output |
| `--output` | `-o` | string | "table" | Output format (yaml\|json\|table) |
| `--config` | `-c` | string | "" | Override config file path |
| `--project` | `-p` | string | "" | Override project file path |

### Command-Specific Flags

#### `fleet up`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--detach` | bool | true | Run in background |
| `--build` | bool | false | Build images before starting |
| `--force-recreate` | bool | false | Recreate containers |

#### `fleet destroy`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--keep-volumes` | bool | false | Don't remove volumes |
| `--force` | bool | false | Skip confirmation |

#### `template build`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--env` | string | "local" | Target environment |
| `--format` | string | "helm" | Output format (helm\|compose\|manifest) |
| `--output-dir` | string | "./dist" | Output directory |
| `--package` | bool | false | Package Helm chart |
| `--push` | string | "" | Push to registry URL |

#### `secret set`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--env` | string | "" | Environment scope |
| `--provider` | string | "" | Override secret provider |

#### `secret sync`
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--from` | string | "" | Source provider |
| `--to` | string | "pass" | Destination provider |
| `--prefix` | string | "yar/" | Key prefix in destination |

---

## Configuration Schemas

### Global Configuration Schema

**File**: `~/.config/yar/config.yaml`

```yaml
# yaml-language-server: $schema=https://yar.io/schemas/config.schema.json

# Container runtime for local development
# REQUIRED for local fleet operations
container: colima  # enum: colima, docker, nerdctl, podman

# VPN configuration for network access
vpn:
  provider: openvpn  # enum: openvpn, wireguard, tailscale
  configPath: ~/.config/yar/vpn/client.ovpn  # path to VPN config

# Host resolution configuration
hosts:
  mode: etc  # enum: etc, kubedns
  suffix: ""  # optional suffix (e.g., ".local")

# Default Docker network for yar services
network:
  name: yar-net  # network name
  cidr: 172.16.34.0/23  # network CIDR

# Secret provider configurations
secrets:
  # Local secret store (required)
  local:
    provider: pass  # enum: pass, keychain, credential-manager, auto
    store: ~/.password-store  # for pass provider
    fallback: true  # use OS-native if primary unavailable

  # Remote providers (optional, for sync and direct reference)
  providers:
    <name>:  # arbitrary provider name
      type: string  # enum: github, azure, hashicorp, onepassword, aws, gcp
      # Provider-specific configuration (see provider schemas below)

# Cluster configurations
clusters:
  <name>:  # arbitrary cluster name (e.g., "local", "dev", "prod")
    provider: string  # enum: compose, k8s
    context: string   # kubeconfig context name (for k8s)
    namespace: string # default namespace (for k8s)
```

#### Secret Provider Schemas

**GitHub**:
```yaml
providers:
  github:
    type: github
    organization: string  # GitHub org name
    # For ESO integration:
    clusterSecretStore: string  # ESO store name
    bootstrapSecretName: string  # K8s secret with GitHub App creds
```

**Azure Key Vault**:
```yaml
providers:
  azure:
    type: azure
    vaultName: string  # Key Vault name
    tenantId: string   # optional, inferred from env if not set
    clientId: string   # optional, for service principal auth
```

**HashiCorp Vault**:
```yaml
providers:
  vault:
    type: hashicorp
    address: string    # Vault server URL
    namespace: string  # Vault namespace
    authMethod: string # enum: token, kubernetes, approle
```

**1Password**:
```yaml
providers:
  onepassword:
    type: onepassword
    vault: string      # 1Password vault name
    account: string    # optional, account identifier
```

### Project Configuration Schema

**File**: `./yar.yaml`

```yaml
# yaml-language-server: $schema=https://yar.io/schemas/project.schema.json

# Project name (used for container prefixes, namespaces)
# REQUIRED
project: string  # pattern: ^[a-z][a-z0-9-]*$

# Environment definitions
# REQUIRED: at least "local" environment
environments:
  <name>:  # environment name (e.g., "local", "dev", "prod")
    cluster: string   # reference to cluster in global config
    secrets: string   # reference to secret provider in global config

# Service definitions
# REQUIRED: at least one service
services:
  - name: string      # REQUIRED: unique service name
    namespace: string # optional: override namespace
    pack: string      # REQUIRED: pack name
    requires: [string] # optional: dependency service names
    replicas: integer  # optional: replica count (default: 1)
    
    # Pack parameters (validated against pack schema)
    params:
      <key>: <value>
    
    # Ingress configuration (optional)
    ingress:
      host: string     # hostname
      path: string     # path prefix (default: "/")
      tls: boolean     # enable TLS (default: false)
    
    # Environment variables (non-secret)
    env:
      <KEY>: <value>
    
    # Secret references (resolved at runtime)
    secretRefs:
      <ENV_VAR_NAME>: string  # secret key reference
```

### Pack Schema

**File**: `packs/<pack-name>/schema.json`

Standard JSON Schema (draft 2020-12) defining pack parameters.

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://yar.io/packs/<pack-name>/schema.json",
  "type": "object",
  "properties": {
    "passwordRef": {
      "type": "string",
      "description": "Secret reference for password"
    },
    "port": {
      "type": "integer",
      "default": 6379,
      "minimum": 1,
      "maximum": 65535
    }
  },
  "required": ["passwordRef"]
}
```

---

## Pack DSL Specification

### Pack Structure

```
packs/<pack-name>/
├── meta.yaml        # Pack metadata (REQUIRED)
├── schema.json      # Parameter schema (REQUIRED)
├── defaults.yaml    # Default parameter values (optional)
└── templates/
    ├── resources.yaml  # Main resource template (REQUIRED)
    └── files/          # Additional files (optional)
        └── *.conf
```

### meta.yaml Schema

```yaml
# Pack name (must match directory name)
name: string  # pattern: ^[a-z][a-z0-9-]*$

# Semantic version
version: string  # pattern: ^\d+\.\d+\.\d+$

# Human-readable description
description: string

# Maintainer info
maintainer: string

# Searchable tags
tags: [string]

# Minimum yar version required
minYarVersion: string  # optional
```

### resources.yaml DSL

The pack DSL uses a Kubernetes-inspired structure with yar-specific extensions.

```yaml
apiVersion: yar.io/v1
kind: Pack
metadata:
  name: string  # pack name

spec:
  # Container definitions
  containers:
    - name: string           # container name
      image: string          # image reference (supports templating)
      command: [string]      # optional: override entrypoint
      args: [string]         # optional: arguments
      
      # Port mappings
      ports:
        - containerPort: integer
          hostPort: integer  # optional
          protocol: string   # tcp|udp (default: tcp)
      
      # Environment variables
      env:
        - name: string
          value: string      # literal value
        - name: string
          secretRef: string  # secret reference
      
      # Volume mounts
      volumes:
        - name: string
          mountPath: string
          content: string    # inline content (for configmaps)
        - name: string
          mountPath: string
          persistent: boolean # create PVC/volume
          size: string       # e.g., "1Gi"
      
      # Resource limits
      resources:
        requests:
          memory: string
          cpu: string
        limits:
          memory: string
          cpu: string
      
      # Health checks
      livenessProbe:
        httpGet:
          path: string
          port: integer
        initialDelaySeconds: integer
        periodSeconds: integer
      
      readinessProbe:
        tcpSocket:
          port: integer
        initialDelaySeconds: integer
        periodSeconds: integer

  # Service definitions
  services:
    - name: string
      port: integer
      targetPort: integer
      type: string  # ClusterIP|NodePort|LoadBalancer

  # ConfigMap definitions
  configMaps:
    - name: string
      data:
        <filename>: string  # content or template

  # Ingress definitions (optional)
  ingress:
    - name: string
      host: string
      path: string
      serviceName: string
      servicePort: integer
      tls: boolean
```

### Template Functions

Templates have access to these contexts:

| Variable | Type | Description |
|----------|------|-------------|
| `.Params` | map[string]any | Validated pack parameters |
| `.Project` | Project | Project configuration |
| `.Service` | Service | Current service configuration |
| `.Environment` | Environment | Current environment |
| `.Secrets` | SecretResolver | Secret resolution (for K8s refs) |

Built-in functions (Sprig + custom):

| Function | Description |
|----------|-------------|
| `default` | Default value if empty |
| `required` | Fail if value is empty |
| `include` | Include another template file |
| `toYaml` | Convert to YAML string |
| `toJson` | Convert to JSON string |
| `indent` | Indent string |
| `nindent` | Newline + indent |
| `secretRef` | Generate secret reference for target |
| `configMapRef` | Generate configmap reference |
| `quote` | Quote string |
| `squote` | Single-quote string |

---

## Secret Reference Specification

### Multi-Developer Workflow

The secret reference system solves the "works on my machine" problem for secrets:

1. **Declaration**: Developer A adds `secretRef: new_api_key` to `yar.yaml`, commits
2. **Discovery**: Developer B pulls, runs `yar fleet up`
3. **Inventory**: Yar extracts all secret references from `yar.yaml`
4. **Validation**: Yar checks each reference against local store (pass/keychain)
5. **Fail Fast**: Missing secrets → immediate failure with actionable error:
   ```
   Error: Missing required secrets:
     - new_api_key (service 'app', env.API_KEY)
     - stripe_secret (service 'payments', params.secretRef)
   
   To resolve:
     yar secret sync --from github   # pull from team's remote store
     pass insert yar/new_api_key     # or add manually
   ```
6. **Resolution**: Developer B syncs or adds secrets, runs `yar fleet up` again
7. **Success**: All secrets resolved → services start

This ensures:
- No `.env` files shared between developers
- New secret requirements are immediately visible on pull
- Clear path to resolution without asking "what am I missing?"

### Reference Syntax

**Simple reference** (uses default provider):
```yaml
passwordRef: redis_pass
```

**Provider-qualified reference**:
```yaml
passwordRef:
  provider: azure
  key: redis-password
```

**Environment-scoped reference**:
```yaml
passwordRef:
  key: redis_pass
  env: prod  # only used in prod environment
```

### Resolution Order

1. If `provider` is specified, use that provider directly
2. If `env` is specified, only resolve in that environment
3. Otherwise, resolve from the environment's configured secret provider
4. For local environments, resolve from local pass store
5. If local fails and fallback enabled, try OS-native keychain

### Resolution Output by Target

**Docker Compose**:
```yaml
services:
  redis:
    secrets:
      - redis_password
    environment:
      - REDIS_PASSWORD_FILE=/run/secrets/redis_password

secrets:
  redis_password:
    external: true  # yar creates Docker secret from pass
```

**Kubernetes**:
```yaml
env:
  - name: REDIS_PASSWORD
    valueFrom:
      secretKeyRef:
        name: redis-secrets
        key: password
```

**ESO ExternalSecret**:
```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: redis-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: github  # or azure, vault
    kind: ClusterSecretStore
  target:
    name: redis-secrets
  data:
    - secretKey: password
      remoteRef:
        key: redis_pass
```

---

## Generator Output Specifications

### Compose Generator

**Input**: Pack resources + Service config + Environment  
**Output**: `docker-compose.yaml`

Mapping rules:

| Pack DSL | Compose |
|----------|---------|
| `spec.containers[*]` | `services.<name>` |
| `spec.containers[*].ports` | `services.<name>.ports` |
| `spec.containers[*].env` | `services.<name>.environment` |
| `spec.containers[*].volumes` (content) | `configs` + `services.<name>.configs` |
| `spec.containers[*].volumes` (persistent) | `volumes` + `services.<name>.volumes` |
| `spec.services` | N/A (internal networking) |
| `secretRef` | `secrets` + env with `_FILE` suffix |

### Helm Generator

**Input**: Pack resources + Service config + Environment  
**Output**: Complete Helm chart

```
charts/<service>/
├── Chart.yaml
├── values.yaml
├── templates/
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   ├── pvc.yaml (if persistent volumes)
│   ├── ingress.yaml (if ingress defined)
│   └── externalsecret.yaml (if ESO enabled)
└── charts/  (subcharts if nested packs)
```

### Manifest Generator

**Input**: Pack resources + Service config + Environment  
**Output**: Raw Kubernetes manifests

```
manifests/
├── deployment.yaml
├── service.yaml
├── configmap.yaml
├── persistentvolumeclaim.yaml
├── ingress.yaml
└── externalsecret.yaml
```

---

## API Contracts

### Internal Package APIs

#### config.Load

```go
// Load loads and validates configuration
func Load(opts ...Option) (*Config, error)

// Options:
// - WithGlobalPath(path string) - override global config path
// - WithProjectPath(path string) - override project config path  
// - WithEnvironment(env string) - set active environment
```

#### docker.Client

```go
type Client interface {
    // Network operations
    NetworkCreate(ctx context.Context, name string, opts NetworkOptions) error
    NetworkRemove(ctx context.Context, name string) error
    NetworkList(ctx context.Context) ([]Network, error)
    
    // Container operations
    ContainerCreate(ctx context.Context, config ContainerConfig) (string, error)
    ContainerStart(ctx context.Context, id string) error
    ContainerStop(ctx context.Context, id string, timeout time.Duration) error
    ContainerRemove(ctx context.Context, id string, opts RemoveOptions) error
    ContainerList(ctx context.Context, opts ListOptions) ([]Container, error)
    ContainerLogs(ctx context.Context, id string, opts LogOptions) (io.ReadCloser, error)
    
    // Image operations
    ImagePull(ctx context.Context, ref string) error
    ImageBuild(ctx context.Context, context io.Reader, opts BuildOptions) error
    
    // Secret operations (Docker secrets for Swarm mode, or mounted files)
    SecretCreate(ctx context.Context, name string, data []byte) error
    SecretRemove(ctx context.Context, name string) error
}
```

#### kubernetes.Client

```go
type Client interface {
    // Apply manifests (like kubectl apply)
    Apply(ctx context.Context, manifests []byte, opts ApplyOptions) error
    
    // Delete resources
    Delete(ctx context.Context, manifests []byte) error
    
    // Get resources
    Get(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string) (*unstructured.Unstructured, error)
    
    // List resources
    List(ctx context.Context, gvk schema.GroupVersionKind, namespace string, opts ListOptions) (*unstructured.UnstructuredList, error)
    
    // Watch resources
    Watch(ctx context.Context, gvk schema.GroupVersionKind, namespace string, opts WatchOptions) (watch.Interface, error)
    
    // Wait for condition
    WaitFor(ctx context.Context, gvk schema.GroupVersionKind, namespace, name string, condition WaitCondition, timeout time.Duration) error
}
```

#### secrets.Provider

```go
type Provider interface {
    // Name returns the provider name
    Name() string
    
    // Get retrieves a secret value
    Get(ctx context.Context, key string) (string, error)
    
    // Set stores a secret value
    Set(ctx context.Context, key, value string) error
    
    // Delete removes a secret
    Delete(ctx context.Context, key string) error
    
    // List returns all secret keys (not values)
    List(ctx context.Context) ([]string, error)
    
    // Exists checks if a secret exists
    Exists(ctx context.Context, key string) (bool, error)
}

// Syncer syncs secrets between providers
type Syncer interface {
    // Sync copies secrets from source to destination
    Sync(ctx context.Context, source, dest Provider, opts SyncOptions) (*SyncResult, error)
}
```

#### packs.Generator

```go
type Generator interface {
    // Generate produces output for the target format
    Generate(ctx context.Context, pack *Pack, svc *Service, env *Environment) (*Output, error)
}

type Output struct {
    Format   OutputFormat  // compose, helm, manifest
    Files    map[string][]byte  // filename -> content
    Secrets  []SecretRef   // secrets that need to be created
}

type OutputFormat string

const (
    OutputCompose   OutputFormat = "compose"
    OutputHelm      OutputFormat = "helm"
    OutputManifest  OutputFormat = "manifest"
)
```

#### fleet.Driver

```go
type Driver interface {
    // Up starts services
    Up(ctx context.Context, project *Project, env string, opts UpOptions) error
    
    // Down stops services
    Down(ctx context.Context, project *Project, env string, opts DownOptions) error
    
    // Destroy removes all resources
    Destroy(ctx context.Context, project *Project, env string, opts DestroyOptions) error
    
    // Status returns current state
    Status(ctx context.Context, project *Project, env string) (*FleetStatus, error)
    
    // Restart restarts services
    Restart(ctx context.Context, project *Project, env string, opts RestartOptions) error
}

type FleetStatus struct {
    Environment string
    Services    []ServiceStatus
    Networks    []NetworkStatus
    Healthy     bool
}

type ServiceStatus struct {
    Name      string
    Status    string  // running, stopped, error
    Replicas  int
    Ready     int
    Endpoints []string
}
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Secret resolution error |
| 4 | Docker/container error |
| 5 | Kubernetes error |
| 6 | Network error |
| 7 | Pack error |
| 64 | Usage error (invalid arguments) |
| 130 | Interrupted (SIGINT) |

---

## File Locations

| File | Path | Description |
|------|------|-------------|
| Global config | `~/.config/yar/config.yaml` | Machine-wide settings |
| Global config (XDG) | `$XDG_CONFIG_HOME/yar/config.yaml` | XDG-compliant path |
| Project config | `./yar.yaml` | Project settings |
| VPN config | `~/.config/yar/vpn/` | VPN configuration files |
| Installed packs | `~/.config/yar/packs/` | User-installed packs |
| Cache | `~/.cache/yar/` | Cached data |
| Pass store | `~/.password-store/` | GNU pass secrets |
| Pass prefix | `yar/` | Prefix for yar-managed secrets in pass |
