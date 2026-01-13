# Iteration 002: Configuration Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Configuration Types

### A1. Global Config Types

**Test First:**
- [x] Write test for Config struct YAML marshal/unmarshal
- [x] Write test for Config struct JSON marshal/unmarshal
- [x] Write test for nested structs (VPNConfig, HostsConfig, etc.)
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/config/types.go`
- [x] Define Config struct with all fields
- [x] Define VPNConfig, HostsConfig, NetworkConfig
- [x] Define SecretsConfig, LocalSecretConfig, SecretProviderConfig
- [x] Define ClusterConfig
- [x] Add YAML and JSON struct tags
- [x] Verify tests pass (green)

### A2. Project Config Types

**Test First:**
- [x] Write test for Project struct YAML marshal/unmarshal
- [x] Write test for nested Service struct
- [x] Write test for IngressConfig
- [x] Verify tests fail (red)

**Implement:**
- [x] Define Project struct
- [x] Define Environment struct
- [x] Define Service struct
- [x] Define IngressConfig struct
- [x] Add YAML and JSON struct tags
- [x] Verify tests pass (green)

### A3. Default Configuration

**Test First:**
- [x] Write test for DefaultConfig() returns valid config
- [x] Write test for default values match spec
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/config/defaults.go`
- [x] Implement DefaultConfig() function
- [x] Verify tests pass (green)

**Verify:**
```bash
go build ./...  # PASS
go test ./internal/config/...  # PASS (7 tests)
go vet ./...  # PASS
```

---

## Phase B: Path Resolution

### B1. Global Config Path

**Test First:**
- [x] Write test for GlobalConfigPath() returns XDG path
- [x] Write test for path uses platform.ConfigDir()
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/config/paths.go`
- [x] Implement GlobalConfigPath() function
- [x] Verify tests pass (green)

### B2. Project Config Path

**Test First:**
- [x] Write test for FindProjectConfig() in current directory
- [x] Write test for FindProjectConfig() in parent directory
- [x] Write test for FindProjectConfig() returns error when not found
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement FindProjectConfig() with directory traversal
- [x] Verify tests pass (green)

**Verify:**
```bash
go build ./...  # PASS
go test ./internal/config/...  # PASS (12 tests)
go vet ./...  # PASS
```

---

## Phase C: Loading & Validation

### C1. JSON Schemas

**Implement:**
- [x] Create `schemas/` directory
- [x] Create `schemas/config.schema.json` for global config
- [x] Create `schemas/project.schema.json` for project config
- [x] Verify schemas are valid JSON

### C2. Test Fixtures

**Implement:**
- [x] Create `internal/config/testdata/valid/` directory
- [x] Create `internal/config/testdata/invalid/` directory
- [x] Create valid config fixtures
- [x] Create invalid config fixtures

### C3. Loader Implementation

**Test First:**
- [x] Write test for NewLoader() with default options
- [x] Write test for NewLoader() with WithGlobalPath option
- [x] Write test for NewLoader() with WithProjectPath option
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/config/loader.go`
- [x] Implement Loader struct
- [x] Implement Option type and WithGlobalPath/WithProjectPath
- [x] Implement NewLoader()
- [x] Verify tests pass (green)

### C4. LoadGlobal Implementation

**Test First:**
- [x] Write test for LoadGlobal() returns defaults when file missing
- [x] Write test for LoadGlobal() loads valid file
- [x] Write test for LoadGlobal() returns ConfigError for invalid YAML
- [x] Write test for LoadGlobal() returns ValidationError for schema failure
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement LoadGlobal() method
- [x] Handle missing file → return defaults
- [x] Parse YAML with yaml.v3
- [x] Validate against schema
- [x] Return appropriate errors
- [x] Verify tests pass (green)

### C5. LoadProject Implementation

**Test First:**
- [x] Write test for LoadProject() returns NotFoundError when missing
- [x] Write test for LoadProject() loads valid file
- [x] Write test for LoadProject() returns ConfigError for invalid YAML
- [x] Write test for LoadProject() returns ValidationError for schema failure
- [x] Verify tests fail (red)

**Implement:**
- [x] Implement LoadProject() method
- [x] Use FindProjectConfig() to locate file
- [x] Parse YAML
- [x] Validate against schema
- [x] Return appropriate errors
- [x] Verify tests pass (green)

### C6. Schema Validation

**Test First:**
- [x] Write test for Validate() accepts valid config
- [x] Write test for Validate() rejects missing required fields
- [x] Write test for Validate() rejects invalid types
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/config/schema.go`
- [x] Implement Validate() function
- [x] Programmatic validation (enum checks, required fields)
- [x] Convert validation errors to ValidationError
- [x] Verify tests pass (green)

**Verify:**
```bash
go build ./...  # PASS
go test ./internal/config/...  # PASS (25 tests)
go vet ./...  # PASS
```

---

## Phase D: CLI Integration

### D1. Update config get Command

**Implement:**
- [x] Update `cmd/config.go` to use config.Loader
- [x] Load global config with LoadGlobal()
- [x] Support --output flag (yaml, json, table)
- [x] Display config in requested format
- [x] Handle errors gracefully

### D2. Update project get Command

**Implement:**
- [x] Update `cmd/project.go` to use config.Loader
- [x] Load project config with LoadProject()
- [x] Support --output flag (yaml, json, table)
- [x] Display project in requested format
- [x] Show clear error when yar.yaml not found

**Verify:**
```bash
go build ./...  # PASS
go test ./...  # PASS
go vet ./...  # PASS
./yar config get  # Shows default config (table)
./yar config get -o json  # Shows JSON
./yar project get  # Shows "no yar.yaml found" error
```

---

## Functional Tests

| Command | Expected Result | Status |
|---------|-----------------|--------|
| `yar config get` | Shows default config (table) | [x] Pass |
| `yar config get -o json` | Shows default config (JSON) | [x] Pass |
| `yar config get -o yaml` | Shows default config (YAML) | [x] Pass |
| `yar project get` | Shows error: no yar.yaml found | [x] Pass |
| `yar project get` (with yar.yaml) | Shows project config | [x] Pass |

---

## Completion Checklist

- [x] All configuration types defined
- [x] YAML/JSON tags on all structs
- [x] DefaultConfig() returns valid defaults
- [x] JSON Schemas created and valid
- [x] Test fixtures created
- [x] Loader implementation complete
- [x] Schema validation works
- [x] CLI commands use real config
- [x] All tests pass (25 in config, 48 total)
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] Functional tests verified
- [x] TASKS.md fully checked off
- [x] Exit criteria from SPEC.md verified
