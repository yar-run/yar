# Yar

**Yar** is a fleet bootstrapper CLI that bridges local development with production Kubernetes clusters, ensuring developers work in environments identical to production.

## Vision

Development environments should be indistinguishable from production. A developer running `yar up` on their laptop should have the same service topology, networking, secret management, and configuration as a production cluster. When code works locally, it works in production.

## Problem Statement

Modern microservice development suffers from environment drift:

1. **Local vs Production Gap**: Developers use docker-compose locally but deploy to Kubernetes, leading to "works on my machine" failures
2. **Secret Sprawl**: Secrets end up in `.env` files, committed accidentally, or managed inconsistently across environments
3. **Configuration Fragmentation**: Each service has its own setup scripts, compose files, and deployment manifests with no unified management
4. **Onboarding Friction**: New developers spend days configuring local environments instead of shipping code
5. **Network Isolation**: Containers aren't accessible by name/IP on macOS/Windows without complex tunneling
6. **Secret Distribution**: Teams share `.env` files via Slack/email because there's no standard way to distribute secrets; when one developer adds a new secret requirement, others discover it by runtime failure or broken builds

## Solution

Yar provides:

- **Unified CLI**: Single tool for local dev (`yar up`) and production deployment (`yar template build`)
- **Pack System**: Portable service definitions that generate both docker-compose and Helm/K8s manifests
- **Secret References**: Secrets are never stored in files; they're referenced by key and resolved at runtime from secure providers
- **Network Transparency**: VPN/DNS/hosts management so containers are accessible by name across all platforms
- **Environment Parity**: Same `yar.yaml` project config drives local and production, with environment-specific overrides
- **Secret Inventory**: `yar.yaml` declares all required secrets by reference; `yar fleet up` validates all secrets exist locally before starting, giving developers immediate feedback on what's missing

## Objectives

### Primary Objectives

1. **Local-Production Parity**: Developers run the same services locally as in production, with identical networking and configuration
2. **Zero-Secret Exposure**: Secrets never appear in `.env` files, git history, or environment variables; they're resolved from secure stores at runtime
3. **Cross-Platform Support**: First-class support for macOS (Intel/Apple Silicon), Linux, and Windows
4. **SDK-Native Operations**: Use Docker, Kubernetes, and Helm Go SDKs directly—no shelling out to CLIs for core operations
5. **Single Source of Truth**: One pack definition generates all deployment artifacts (compose, Helm, raw K8s manifests)

### Secondary Objectives

1. **Developer Experience**: Simple commands (`yar up`, `yar down`) with sensible defaults
2. **Team Standardization**: Shared packs and configurations ensure all developers have identical setups
3. **CI/CD Integration**: `yar template build` produces artifacts suitable for GitOps pipelines
4. **Extensibility**: Custom packs for organization-specific services

## Core Concepts

### Fleet
A fleet is the collection of services defined in a project's `yar.yaml`. Running `yar fleet up` starts all services in dependency order.

### Pack
A pack is a portable service definition with a schema, defaults, and templates. Packs generate docker-compose services, Helm charts, or raw Kubernetes manifests.

### Environment
An environment (local, dev, staging, prod) defines which cluster to target and which secret provider to use. The same services run in each environment with appropriate configuration.

### Secret Reference
A secret reference (`passwordRef: redis_pass`) points to a secret by name. Yar resolves it at runtime from the configured provider (pass, Keychain, Azure Key Vault, etc.) without ever writing the value to disk.

## Non-Goals

- **Replacing Kubernetes**: Yar generates K8s manifests; it doesn't replace kubectl or cluster management
- **Replacing Helm**: Yar uses Helm SDK internally and can output Helm charts; it's complementary, not a replacement
- **Production Orchestration**: Yar focuses on local dev and artifact generation; production deployment is handled by GitOps tools (ArgoCD, Flux)
- **Container Registry**: Yar doesn't host images; it references them from existing registries

## Success Criteria

1. A new developer can clone a repo with `yar.yaml` and run `yar up` to have a fully functional local environment in under 5 minutes
2. The same `yar.yaml` produces identical service configurations for local docker-compose and production Kubernetes
3. Zero secrets exist in any file in the repository or local filesystem (outside encrypted stores)
4. All standard services (Redis, Kafka, Postgres, etc.) are available as built-in packs
5. Cross-platform: works on macOS (Intel + Apple Silicon), Linux, and Windows

## Target Users

1. **Backend Developers**: Primary users running local services while developing
2. **DevOps Engineers**: Creating and maintaining packs, configuring environments
3. **Platform Teams**: Standardizing development environments across organizations
4. **New Team Members**: Onboarding with minimal friction

## Terminology

| Term | Definition |
|------|------------|
| **Fleet** | Collection of services in a project |
| **Pack** | Portable service definition (schema + templates) |
| **Environment** | Target context (local, dev, prod) with cluster and secret config |
| **Secret Reference** | Pointer to a secret by name, resolved at runtime |
| **Hoist** | Alias for `fleet up` (start services) |
| **Dock** | Alias for `fleet down` (stop services) |
| **Scuttle** | Alias for `fleet destroy` (remove everything) |
| **Swab** | Alias for `doctor run --fix-cache` (cleanup) |

---

## CLI Reference

This is the complete command-line interface specification. This is the **true north** for all development—every feature, iteration, and design decision serves this UX.

### Usage

```
yar <object> <verb> [args] [flags]
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help |
| `--verbose` | `-v` | Verbose output |
| `--output <fmt>` | `-o` | Output format: `yaml`, `json`, `table` (default: `table`) |

---

### fleet — Service Lifecycle

Manage the fleet of services defined in `yar.yaml`.

| Command | Description |
|---------|-------------|
| `yar fleet up [env]` | Start all services for environment (default: `local`). Bootstraps Colima/VPN/DNS, validates secrets, starts containers in dependency order. |
| `yar fleet down [env]` | Stop all services. Containers are stopped but not removed. |
| `yar fleet destroy [env]` | Stop and remove all services, networks, and volumes. |
| `yar fleet restart [env]` | Restart all services, applying any config changes. |
| `yar fleet status [env]` | Show status of all services (running, stopped, health). |
| `yar fleet update` | Update yar binary and pack catalog. |

**Flags for `fleet up`:**
| Flag | Description |
|------|-------------|
| `--detach` | Run in background (default: true) |
| `--build` | Build images before starting |
| `--force-recreate` | Recreate containers even if unchanged |

**Flags for `fleet destroy`:**
| Flag | Description |
|------|-------------|
| `--keep-volumes` | Don't remove volumes |
| `--force` | Skip confirmation prompt |

---

### config — Global Configuration

Manage machine-wide configuration at `~/.config/yar/config.yaml`.

| Command | Description |
|---------|-------------|
| `yar config get` | Display global configuration |
| `yar config edit` | Open global config in `$EDITOR` |

---

### project — Project Configuration

Manage project configuration at `./yar.yaml`.

| Command | Description |
|---------|-------------|
| `yar project init` | Interactive guided setup, creates `./yar.yaml` |
| `yar project get` | Display project configuration |
| `yar project edit` | Open project config in `$EDITOR` |

---

### pack — Service Packs

Manage portable service definitions.

| Command | Description |
|---------|-------------|
| `yar pack list` | List available packs (built-in and installed) |
| `yar pack install <name>` | Install a pack from catalog |
| `yar pack remove <name>` | Remove an installed pack |

---

### template — Deployment Artifacts

Generate deployment assets from packs.

| Command | Description |
|---------|-------------|
| `yar template build [--env <e>]` | Generate Helm charts, Compose files, or K8s manifests |
| `yar template render [--env <e>]` | Render templates to stdout (dry-run) |
| `yar template publish` | Publish charts to artifact repository |

**Flags for `template build`:**
| Flag | Description |
|------|-------------|
| `--env <env>` | Target environment (default: `local`) |
| `--format <fmt>` | Output format: `helm`, `compose`, `manifest` |
| `--output-dir <dir>` | Output directory (default: `./dist`) |
| `--package` | Package Helm chart as `.tgz` |
| `--push <url>` | Push to OCI registry (e.g., `oci://ghcr.io/org/charts`) |
| `--values-only` | Only update values, no template changes |
| `--lock` | Lock dependency versions for reproducibility |

---

### secret — Secret Management

Manage secrets by reference. Values are stored in encrypted stores, never in files.

| Command | Description |
|---------|-------------|
| `yar secret list` | List all secrets required by `yar.yaml` with status (present/missing) |
| `yar secret set <key> <value>` | Set a secret in local store |
| `yar secret get <key>` | Show secret metadata (redacted value) |
| `yar secret delete <key>` | Delete a secret from local store |
| `yar secret sync` | Sync secrets from remote provider to local store |

**Flags for `secret set`:**
| Flag | Description |
|------|-------------|
| `--env <env>` | Scope secret to environment |
| `--store <store>` | Override target store |

**Flags for `secret sync`:**
| Flag | Description |
|------|-------------|
| `--from <provider>` | Source provider (e.g., `github`, `azure`) |
| `--to <provider>` | Destination provider (default: `pass`) |
| `--prefix <prefix>` | Key prefix in destination (default: `yar/`) |

**Example workflow:**
```bash
# Developer A adds new secret reference to yar.yaml, commits
# Developer B pulls, runs:
$ yar fleet up
Error: Missing required secrets:
  - new_api_key (service 'app', env.API_KEY)

To resolve:
  yar secret sync --from github
  # or: pass insert yar/new_api_key

# Developer B syncs from team's remote store:
$ yar secret sync --from github
Synced 3 secrets from github to pass

$ yar fleet up
✓ All secrets resolved
✓ Starting services...
```

---

### hosts — Host Resolution

Manage `/etc/hosts` entries for container name resolution.

| Command | Description |
|---------|-------------|
| `yar hosts list` | List yar-managed host entries |
| `yar hosts set <name> <ip>` | Add or update a host entry |
| `yar hosts get <name>` | Show a host entry |
| `yar hosts delete <name>` | Remove a host entry |

All entries are marked with `# yar:managed` for safe cleanup.

---

### doctor — Health Checks

Diagnose and repair environment issues.

| Command | Description |
|---------|-------------|
| `yar doctor run` | Run all health checks (VPN, DNS, hosts, clusters, secrets) |
| `yar doctor run --fix` | Attempt to auto-repair issues |
| `yar doctor run --fix-cache` | Clear caches and regenerate state |

---

### Aliases

Ergonomic shortcuts for common operations.

| Alias | Expands To |
|-------|------------|
| `yar hoist [env]` | `yar fleet up [env]` |
| `yar dock [env]` | `yar fleet down [env]` |
| `yar scuttle [env]` | `yar fleet destroy [env]` |
| `yar swab` | `yar doctor run --fix-cache` |
| `yar up [env]` | `yar fleet up [env]` |
| `yar down [env]` | `yar fleet down [env]` |

---

### Other Commands

| Command | Description |
|---------|-------------|
| `yar version` | Print yar version |
| `yar help [command]` | Show help for a command |

---

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Configuration error |
| 3 | Secret resolution error (missing secrets) |
| 4 | Docker/container error |
| 5 | Kubernetes error |
| 6 | Network error |
| 7 | Pack error |
| 64 | Usage error (invalid arguments) |
| 130 | Interrupted (SIGINT) |

---

### Example Session

```bash
# New developer clones repo
$ git clone git@github.com:acme/backend.git && cd backend

# See what's needed
$ yar secret list
SECRET          STATUS    SERVICE     REFERENCE
redis_pass      missing   redis       params.passwordRef
kafka_pass      missing   kafka       params.passwordRef
api_key         missing   app         env.API_KEY

# Sync from team's secret store
$ yar secret sync --from github
Synced 3 secrets from github to pass

# Start everything
$ yar fleet up
✓ All secrets resolved
✓ Network yar-net created
✓ redis.ai-agents started
✓ kafka.ai-agents started
✓ app.ai-agents started (3 replicas)

Fleet is up. Services:
  redis.ai-agents-redis    172.16.34.2:6379
  kafka.ai-agents-kafka    172.16.34.3:9092
  app.ai-agents            172.16.34.4:8080

# Check status
$ yar fleet status
SERVICE                   STATUS    REPLICAS    ENDPOINTS
redis.ai-agents-redis     running   1/1         172.16.34.2:6379
kafka.ai-agents-kafka     running   1/1         172.16.34.3:9092
app.ai-agents             running   3/3         172.16.34.4:8080

# Stop when done
$ yar fleet down
✓ Stopped 3 services

# Generate production artifacts
$ yar template build --env prod --package --push oci://ghcr.io/acme/charts
✓ Generated Helm chart for redis
✓ Generated Helm chart for kafka
✓ Generated Helm chart for app
✓ Packaged charts
✓ Pushed to oci://ghcr.io/acme/charts
```
