# Iteration 000: Project Structure Specification

## Overview

This iteration creates the complete directory structure for Yar and configures Go module dependencies. No code logic is implemented—only the scaffolding.

## Scope

### Included
- All directories from DESIGN.md
- Updated go.mod with all dependencies
- Placeholder doc.go files for each package
- go mod tidy to download dependencies

### NOT Included
- Any implementation code
- Error types (iteration 002)
- Config types (iteration 003)
- CLI commands (iteration 005)

---

## Directory Structure

```
yar/
├── main.go                     (exists)
├── go.mod                      (update)
├── go.sum                      (generated)
├── cmd/
│   └── root.go                 (exists)
├── internal/
│   ├── config/
│   │   └── doc.go
│   ├── errors/
│   │   └── doc.go
│   ├── platform/
│   │   └── doc.go
│   ├── docker/
│   │   └── doc.go
│   ├── kubernetes/
│   │   └── doc.go
│   ├── helm/
│   │   └── doc.go
│   ├── secrets/
│   │   └── doc.go
│   ├── packs/
│   │   └── doc.go
│   ├── fleet/
│   │   └── doc.go
│   ├── network/
│   │   └── doc.go
│   └── doctor/
│       └── doc.go
├── schemas/
│   └── .gitkeep
├── packs/
│   └── .gitkeep
├── docs/
│   └── commands/
│       └── .gitkeep
└── test/
    └── fixtures/
        ├── config/
        │   └── .gitkeep
        └── project/
            └── .gitkeep
```

---

## Dependencies

### go.mod additions

```go
require (
    // CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    
    // YAML/JSON
    gopkg.in/yaml.v3 v3.0.1
    github.com/santhosh-tekuri/jsonschema/v5 v5.3.1
    
    // Docker
    github.com/docker/docker v24.0.7+incompatible
    github.com/docker/go-connections v0.4.0
    github.com/compose-spec/compose-go/v2 v2.1.0
    
    // Kubernetes
    k8s.io/client-go v0.29.0
    k8s.io/apimachinery v0.29.0
    
    // Helm
    helm.sh/helm/v3 v3.14.0
    
    // Secrets
    github.com/zalando/go-keyring v0.2.3
    github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets v1.0.0
    github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.4.0
    github.com/hashicorp/vault/api v1.10.0
    github.com/1password/onepassword-sdk-go v0.1.0
    github.com/google/go-github/v57 v57.0.0
    
    // Testing
    github.com/stretchr/testify v1.8.4
)
```

---

## File Contents

### doc.go template

Each package gets a doc.go with package documentation:

```go
// Package {name} provides {description}.
package {name}
```

Descriptions:
- `config`: configuration loading, parsing, and validation
- `errors`: typed error definitions for yar
- `platform`: platform detection and OS-specific utilities
- `docker`: Docker SDK wrapper for container operations
- `kubernetes`: Kubernetes client-go wrapper for cluster operations
- `helm`: Helm SDK wrapper for chart operations
- `secrets`: secret provider interface and implementations
- `packs`: pack loading, validation, and generation
- `fleet`: fleet orchestration and driver implementations
- `network`: VPN, DNS, and hosts management
- `doctor`: health checks and auto-repair

---

## Exit Criteria

- [x] All directories exist as specified
- [x] All doc.go files created with package documentation
- [x] go.mod updated with dependencies
- [x] `go mod tidy` succeeds
- [x] `go build ./...` succeeds
- [x] No compilation errors

---

## Invariants

None for this iteration—no logic implemented.

---

## Test Requirements

None for this iteration—no testable code.
