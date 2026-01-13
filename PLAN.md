# Yar Development Plan

This document outlines the phased development approach for Yar. Each phase is a 1-2 hour iteration that corresponds to an SDD spec in `specs/`.

## Phase Overview

| Phase | Name | Duration | Focus |
|-------|------|----------|-------|
| 001 | Project Structure | 1 hr | Directories, go.mod, placeholder files |
| 002 | Error & Platform | 1 hr | Error types, platform detection |
| 003 | Config Types | 1 hr | Configuration struct definitions |
| 004 | Config Loading | 1.5 hr | YAML loading, schema validation |
| 005 | CLI Skeleton | 1.5 hr | All Cobra commands as stubs |
| 006 | Config Commands | 1 hr | Implement config get/edit, project get/edit |
| 007 | Docker Client | 1.5 hr | Docker SDK wrapper, network ops |
| 008 | Docker Containers | 1.5 hr | Container lifecycle via SDK |
| 009 | Compose Parsing | 1 hr | Parse compose specs with compose-go |
| 010 | K8s Client | 1.5 hr | client-go wrapper, kubeconfig |
| 011 | K8s Apply | 1 hr | Apply/delete manifests |
| 012 | Helm Client | 1.5 hr | Helm SDK wrapper, chart ops |
| 013 | Secret Interface | 1 hr | Provider interface, pass provider |
| 014 | Secret Keyring | 1 hr | Keychain, Credential Manager, go-keyring |
| 015 | Secret Remote | 1.5 hr | Azure, Vault, 1Password providers |
| 016 | Secret Sync | 1 hr | Sync remote to local |
| 017 | Pack Types | 1 hr | Pack DSL types, schema |
| 018 | Pack Loader | 1 hr | Load and validate packs |
| 019 | Compose Generator | 1.5 hr | Pack DSL → docker-compose.yaml |
| 020 | Helm Generator | 1.5 hr | Pack DSL → Helm chart |
| 021 | Manifest Generator | 1 hr | Pack DSL → raw K8s YAML |
| 022 | Built-in Packs | 1.5 hr | Redis, Postgres, Kafka packs |
| 023 | Fleet Driver | 1 hr | Driver interface, orchestrator |
| 024 | Fleet Compose | 1.5 hr | Compose driver implementation |
| 025 | Fleet K8s | 1.5 hr | Kubernetes driver implementation |
| 026 | Dependency Order | 1 hr | Topological sort for requires |
| 027 | Colima Management | 1 hr | Colima start/stop/status |
| 028 | Hosts Management | 1 hr | /etc/hosts read/write |
| 029 | VPN Management | 1 hr | OpenVPN start/stop |
| 030 | Doctor Checks | 1 hr | Health check implementations |
| 031 | Doctor Repair | 1 hr | Auto-fix functionality |
| 032 | Integration Tests | 2 hr | End-to-end test suite |

---

## Phase 001: Project Structure

**Duration**: 1 hour

**Objective**: Create complete directory structure and configure dependencies.

**Deliverables**:
- All directories from DESIGN.md
- Updated go.mod with dependencies
- Placeholder doc.go files

**Exit Criteria**:
- `go build ./...` succeeds
- All internal/ directories exist

---

## Phase 002: Error & Platform

**Duration**: 1 hour

**Objective**: Implement error types and platform detection utilities.

**Deliverables**:
- `internal/errors/` - ConfigError, ValidationError, NotFoundError
- `internal/platform/` - Platform detection, path helpers

**Exit Criteria**:
- Error types format correctly
- Platform detection returns darwin/linux/windows
- ConfigDir/CacheDir return valid paths
- Unit tests pass

---

## Phase 003: Config Types

**Duration**: 1 hour

**Objective**: Define all configuration struct types.

**Deliverables**:
- `internal/config/types.go` - All config structs
- `internal/config/defaults.go` - Default values

**Exit Criteria**:
- All types compile
- Default config is valid
- YAML/JSON tags present

---

## Phase 004: Config Loading

**Duration**: 1.5 hours

**Objective**: Implement config file loading and JSON Schema validation.

**Deliverables**:
- `internal/config/loader.go` - Load global/project config
- `internal/config/paths.go` - Path resolution
- `internal/config/schema.go` - JSON Schema validation
- `schemas/*.json` - Config schemas
- Test fixtures

**Exit Criteria**:
- Can load valid config files
- Invalid configs produce validation errors
- Missing global config returns defaults
- Missing project config returns clear error
- Unit tests pass

---

## Phase 005: CLI Skeleton

**Duration**: 1.5 hours

**Objective**: Define all Cobra commands as stubs.

**Deliverables**:
- `cmd/root.go` - Updated with global flags
- `cmd/fleet.go` - fleet up/down/destroy/restart/status/update
- `cmd/config.go` - config get/edit
- `cmd/project.go` - project init/get/edit
- `cmd/pack.go` - pack list/install/remove
- `cmd/template.go` - template build/render/publish
- `cmd/secret.go` - secret set/get/delete/list/sync
- `cmd/hosts.go` - hosts set/get/delete/list
- `cmd/doctor.go` - doctor run
- `cmd/aliases.go` - hoist/dock/scuttle/swab

**Exit Criteria**:
- `yar --help` shows all commands
- `yar <cmd> --help` shows subcommands
- All stubs print placeholder messages
- Aliases work

---

## Phase 006: Config Commands

**Duration**: 1 hour

**Objective**: Implement config get/edit and project get/edit commands.

**Deliverables**:
- `yar config get` displays global config
- `yar config edit` opens in $EDITOR
- `yar project get` displays project config
- `yar project edit` opens in $EDITOR

**Exit Criteria**:
- Commands output YAML/JSON based on --output flag
- Edit opens correct file in editor
- Errors are clear when files don't exist

---

## Phase 007: Docker Client

**Duration**: 1.5 hours

**Objective**: Create Docker SDK wrapper for network operations.

**Deliverables**:
- `internal/docker/client.go` - Client wrapper
- `internal/docker/network.go` - NetworkCreate, NetworkRemove, NetworkList

**Exit Criteria**:
- Can create Docker network
- Can list Docker networks
- Can remove Docker network
- Unit tests with mocks pass

---

## Phase 008: Docker Containers

**Duration**: 1.5 hours

**Objective**: Implement container lifecycle operations.

**Deliverables**:
- `internal/docker/container.go` - Container ops
- Create, Start, Stop, Remove, List, Logs

**Exit Criteria**:
- Can create container from config
- Can start/stop containers
- Can remove containers
- Can stream logs
- Unit tests pass

---

## Phase 009: Compose Parsing

**Duration**: 1 hour

**Objective**: Parse Compose specs using compose-go.

**Deliverables**:
- `internal/docker/compose.go` - Compose integration
- Parse docker-compose.yaml to typed structs

**Exit Criteria**:
- Can parse valid compose files
- Invalid compose files produce errors
- Can extract service definitions
- Unit tests pass

---

## Phase 010: K8s Client

**Duration**: 1.5 hours

**Objective**: Create Kubernetes client-go wrapper.

**Deliverables**:
- `internal/kubernetes/client.go` - Client wrapper
- Kubeconfig loading
- Context switching

**Exit Criteria**:
- Can connect to cluster from kubeconfig
- Can switch contexts
- Connection errors are clear
- Unit tests pass

---

## Phase 011: K8s Apply

**Duration**: 1 hour

**Objective**: Implement manifest apply/delete operations.

**Deliverables**:
- `internal/kubernetes/apply.go` - Apply, Delete, Get, List

**Exit Criteria**:
- Can apply YAML manifests
- Can delete resources
- Can get/list resources
- Unit tests pass

---

## Phase 012: Helm Client

**Duration**: 1.5 hours

**Objective**: Create Helm SDK wrapper.

**Deliverables**:
- `internal/helm/client.go` - Helm wrapper
- `internal/helm/chart.go` - Chart loading
- Template, Install, Upgrade, Uninstall

**Exit Criteria**:
- Can render Helm templates
- Can install releases
- Can upgrade releases
- Can uninstall releases
- Unit tests pass

---

## Phase 013: Secret Interface

**Duration**: 1 hour

**Objective**: Define secret provider interface and implement pass provider.

**Deliverables**:
- `internal/secrets/provider.go` - Provider interface
- `internal/secrets/pass.go` - GNU pass provider

**Exit Criteria**:
- Interface defines Get, Set, Delete, List
- Pass provider wraps CLI commands
- Fallback to direct file access
- Unit tests pass

---

## Phase 014: Secret Keyring

**Duration**: 1 hour

**Objective**: Implement OS-native keyring providers.

**Deliverables**:
- `internal/secrets/keyring.go` - go-keyring abstraction
- `internal/secrets/keychain.go` - macOS Keychain
- `internal/secrets/credman.go` - Windows Credential Manager

**Exit Criteria**:
- Keychain works on macOS
- Credential Manager works on Windows
- Abstraction selects correct provider
- Unit tests pass

---

## Phase 015: Secret Remote

**Duration**: 1.5 hours

**Objective**: Implement remote secret providers.

**Deliverables**:
- `internal/secrets/azure.go` - Azure Key Vault
- `internal/secrets/vault.go` - HashiCorp Vault
- `internal/secrets/onepassword.go` - 1Password
- `internal/secrets/github.go` - GitHub Secrets

**Exit Criteria**:
- Each provider implements interface
- SDK authentication works
- Unit tests with mocks pass

---

## Phase 016: Secret Sync

**Duration**: 1 hour

**Objective**: Implement secret sync between providers.

**Deliverables**:
- `internal/secrets/sync.go` - Sync logic
- `yar secret sync` command implementation

**Exit Criteria**:
- Can sync from remote to local
- Handles conflicts/updates
- Reports sync results
- Unit tests pass

---

## Phase 017: Pack Types

**Duration**: 1 hour

**Objective**: Define Pack DSL types and schema.

**Deliverables**:
- `internal/packs/types.go` - Pack, Container, Service, etc.
- `schemas/pack.schema.json` - Pack JSON Schema

**Exit Criteria**:
- All DSL types defined
- Schema validates pack structure
- Types compile

---

## Phase 018: Pack Loader

**Duration**: 1 hour

**Objective**: Implement pack loading and validation.

**Deliverables**:
- `internal/packs/loader.go` - Load pack from directory
- `internal/packs/registry.go` - Pack discovery
- `internal/packs/schema.go` - Param validation

**Exit Criteria**:
- Can load pack from packs/ directory
- Schema validates pack params
- Registry lists available packs
- Unit tests pass

---

## Phase 019: Compose Generator

**Duration**: 1.5 hours

**Objective**: Generate docker-compose.yaml from Pack DSL.

**Deliverables**:
- `internal/packs/compose.go` - Compose generator
- Template rendering with Go templates

**Exit Criteria**:
- Generates valid docker-compose.yaml
- Handles all Pack DSL features
- Secret refs become Docker secrets
- Unit tests pass

---

## Phase 020: Helm Generator

**Duration**: 1.5 hours

**Objective**: Generate Helm chart from Pack DSL.

**Deliverables**:
- `internal/packs/helm.go` - Helm generator
- Generates Chart.yaml, values.yaml, templates/

**Exit Criteria**:
- Generates valid Helm chart structure
- Templates compile with Helm SDK
- Values are correctly parameterized
- Unit tests pass

---

## Phase 021: Manifest Generator

**Duration**: 1 hour

**Objective**: Generate raw K8s manifests from Pack DSL.

**Deliverables**:
- `internal/packs/manifest.go` - Manifest generator
- `internal/packs/eso.go` - ExternalSecret generator

**Exit Criteria**:
- Generates Deployment, Service, ConfigMap
- Generates ExternalSecret for secret refs
- Manifests are valid K8s YAML
- Unit tests pass

---

## Phase 022: Built-in Packs

**Duration**: 1.5 hours

**Objective**: Create built-in service packs.

**Deliverables**:
- `packs/redis/` - Redis pack
- `packs/postgres/` - PostgreSQL pack
- `packs/kafka/` - Kafka pack

**Exit Criteria**:
- Each pack has meta.yaml, schema.json, resources.yaml
- Each pack generates valid Compose and Helm
- Unit tests validate all packs

---

## Phase 023: Fleet Driver

**Duration**: 1 hour

**Objective**: Define fleet driver interface and orchestrator.

**Deliverables**:
- `internal/fleet/driver.go` - Driver interface
- `internal/fleet/orchestrator.go` - Driver selection

**Exit Criteria**:
- Interface defines Up, Down, Destroy, Status
- Orchestrator selects driver by environment
- Unit tests pass

---

## Phase 024: Fleet Compose

**Duration**: 1.5 hours

**Objective**: Implement Compose driver for local development.

**Deliverables**:
- `internal/fleet/compose.go` - Compose driver

**Exit Criteria**:
- `fleet up` starts containers via Docker SDK
- `fleet down` stops containers
- `fleet destroy` removes all resources
- `fleet status` reports container state
- Integration tests pass

---

## Phase 025: Fleet K8s

**Duration**: 1.5 hours

**Objective**: Implement Kubernetes driver for cluster deployment.

**Deliverables**:
- `internal/fleet/kubernetes.go` - K8s driver

**Exit Criteria**:
- `fleet up` applies manifests or installs Helm
- `fleet down` scales to zero or stops
- `fleet destroy` deletes resources
- `fleet status` reports pod state
- Integration tests pass

---

## Phase 026: Dependency Order

**Duration**: 1 hour

**Objective**: Implement dependency ordering for services.

**Deliverables**:
- `internal/fleet/deps.go` - Dependency graph, topological sort

**Exit Criteria**:
- Services start in dependency order
- Circular dependencies detected and error
- Unit tests pass

---

## Phase 027: Colima Management

**Duration**: 1 hour

**Objective**: Implement Colima lifecycle management.

**Deliverables**:
- `internal/fleet/colima.go` - Colima start/stop/status

**Exit Criteria**:
- Can check if Colima running
- Can start Colima if not running
- Can stop Colima
- Works only on macOS
- Unit tests pass

---

## Phase 028: Hosts Management

**Duration**: 1 hour

**Objective**: Implement /etc/hosts read/write.

**Deliverables**:
- `internal/network/hosts.go` - Hosts file management
- `yar hosts` command implementation

**Exit Criteria**:
- Can read yar-managed entries
- Can add entries with `# yar:managed` marker
- Can remove entries
- Handles privilege escalation
- Unit tests pass

---

## Phase 029: VPN Management

**Duration**: 1 hour

**Objective**: Implement OpenVPN client management.

**Deliverables**:
- `internal/network/vpn.go` - VPN management

**Exit Criteria**:
- Can start OpenVPN with config
- Can stop OpenVPN
- Can check VPN status
- Unit tests pass

---

## Phase 030: Doctor Checks

**Duration**: 1 hour

**Objective**: Implement health check functionality.

**Deliverables**:
- `internal/doctor/checks.go` - Health checks

**Exit Criteria**:
- Checks: Docker running, K8s reachable, VPN connected, hosts configured, secrets accessible
- Reports issues clearly
- Unit tests pass

---

## Phase 031: Doctor Repair

**Duration**: 1 hour

**Objective**: Implement auto-fix functionality.

**Deliverables**:
- `internal/doctor/repair.go` - Auto-repair

**Exit Criteria**:
- Can start Colima if not running
- Can clear cache
- Can regenerate hosts entries
- Reports what was fixed
- Unit tests pass

---

## Phase 032: Integration Tests

**Duration**: 2 hours

**Objective**: Create end-to-end test suite.

**Deliverables**:
- `test/integration/` - Integration tests
- Full workflow tests

**Exit Criteria**:
- Test: project init → fleet up → fleet status → fleet down
- Test: secret sync → template build
- All tests pass in CI
- Coverage > 70%

---

## Dependency Graph

```
001 Project Structure
 │
 ▼
002 Error & Platform
 │
 ▼
003 Config Types
 │
 ▼
004 Config Loading
 │
 ├─────────────────────────────────────────────────────┐
 ▼                                                     │
005 CLI Skeleton                                       │
 │                                                     │
 ▼                                                     │
006 Config Commands                                    │
 │                                                     │
 ├──────────────┬──────────────┬──────────────────────┤
 ▼              ▼              ▼                      │
007 Docker    010 K8s       013 Secret               │
 │            Client        Interface                 │
 ▼              │              │                      │
008 Docker      ▼              ▼                      │
Containers   011 K8s       014 Secret                │
 │           Apply         Keyring                    │
 ▼              │              │                      │
009 Compose     ▼              ▼                      │
Parsing      012 Helm      015 Secret                │
 │           Client        Remote                     │
 │              │              │                      │
 │              │              ▼                      │
 │              │          016 Secret                 │
 │              │          Sync                       │
 │              │              │                      │
 └──────────────┴──────────────┴──────────────────────┘
                               │
                               ▼
                    017 Pack Types
                               │
                               ▼
                    018 Pack Loader
                               │
              ┌────────────────┼────────────────┐
              ▼                ▼                ▼
        019 Compose      020 Helm        021 Manifest
        Generator        Generator       Generator
              │                │                │
              └────────────────┼────────────────┘
                               │
                               ▼
                    022 Built-in Packs
                               │
                               ▼
                    023 Fleet Driver
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
        024 Fleet                         025 Fleet
        Compose                           K8s
              │                                 │
              └────────────────┬────────────────┘
                               │
                               ▼
                    026 Dependency Order
                               │
                               ▼
                    027 Colima Management
                               │
              ┌────────────────┴────────────────┐
              ▼                                 ▼
        028 Hosts                         029 VPN
        Management                        Management
              │                                 │
              └────────────────┬────────────────┘
                               │
                               ▼
                    030 Doctor Checks
                               │
                               ▼
                    031 Doctor Repair
                               │
                               ▼
                    032 Integration Tests
```

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Iteration completion | 1-2 hours each |
| Unit test coverage | > 80% |
| Build time | < 30 seconds |
| Binary size | < 50 MB |
| Platform support | macOS, Linux, Windows |
