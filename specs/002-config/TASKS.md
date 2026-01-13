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
- [ ] Write test for Config struct YAML marshal/unmarshal
- [ ] Write test for Config struct JSON marshal/unmarshal
- [ ] Write test for nested structs (VPNConfig, HostsConfig, etc.)
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Create `internal/config/types.go`
- [ ] Define Config struct with all fields
- [ ] Define VPNConfig, HostsConfig, NetworkConfig
- [ ] Define SecretsConfig, LocalSecretConfig, SecretProviderConfig
- [ ] Define ClusterConfig
- [ ] Add YAML and JSON struct tags
- [ ] Verify tests pass (green)

### A2. Project Config Types

**Test First:**
- [ ] Write test for Project struct YAML marshal/unmarshal
- [ ] Write test for nested Service struct
- [ ] Write test for IngressConfig
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Define Project struct
- [ ] Define Environment struct
- [ ] Define Service struct
- [ ] Define IngressConfig struct
- [ ] Add YAML and JSON struct tags
- [ ] Verify tests pass (green)

### A3. Default Configuration

**Test First:**
- [ ] Write test for DefaultConfig() returns valid config
- [ ] Write test for default values match spec
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Create `internal/config/defaults.go`
- [ ] Implement DefaultConfig() function
- [ ] Verify tests pass (green)

**Verify:**
```bash
go build ./...
go test ./internal/config/...
go vet ./...
```

---

## Phase B: Path Resolution

### B1. Global Config Path

**Test First:**
- [ ] Write test for GlobalConfigPath() returns XDG path
- [ ] Write test for path uses platform.ConfigDir()
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Create `internal/config/paths.go`
- [ ] Implement GlobalConfigPath() function
- [ ] Verify tests pass (green)

### B2. Project Config Path

**Test First:**
- [ ] Write test for FindProjectConfig() in current directory
- [ ] Write test for FindProjectConfig() in parent directory
- [ ] Write test for FindProjectConfig() returns error when not found
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement FindProjectConfig() with directory traversal
- [ ] Verify tests pass (green)

**Verify:**
```bash
go build ./...
go test ./internal/config/...
go vet ./...
```

---

## Phase C: Loading & Validation

### C1. JSON Schemas

**Implement:**
- [ ] Create `schemas/` directory
- [ ] Create `schemas/config.schema.json` for global config
- [ ] Create `schemas/project.schema.json` for project config
- [ ] Verify schemas are valid JSON

### C2. Test Fixtures

**Implement:**
- [ ] Create `internal/config/testdata/valid/` directory
- [ ] Create `internal/config/testdata/invalid/` directory
- [ ] Create valid config fixtures
- [ ] Create invalid config fixtures

### C3. Loader Implementation

**Test First:**
- [ ] Write test for NewLoader() with default options
- [ ] Write test for NewLoader() with WithGlobalPath option
- [ ] Write test for NewLoader() with WithProjectPath option
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Create `internal/config/loader.go`
- [ ] Implement Loader struct
- [ ] Implement Option type and WithGlobalPath/WithProjectPath
- [ ] Implement NewLoader()
- [ ] Verify tests pass (green)

### C4. LoadGlobal Implementation

**Test First:**
- [ ] Write test for LoadGlobal() returns defaults when file missing
- [ ] Write test for LoadGlobal() loads valid file
- [ ] Write test for LoadGlobal() returns ConfigError for invalid YAML
- [ ] Write test for LoadGlobal() returns ValidationError for schema failure
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement LoadGlobal() method
- [ ] Handle missing file → return defaults
- [ ] Parse YAML with yaml.v3
- [ ] Validate against schema
- [ ] Return appropriate errors
- [ ] Verify tests pass (green)

### C5. LoadProject Implementation

**Test First:**
- [ ] Write test for LoadProject() returns NotFoundError when missing
- [ ] Write test for LoadProject() loads valid file
- [ ] Write test for LoadProject() returns ConfigError for invalid YAML
- [ ] Write test for LoadProject() returns ValidationError for schema failure
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Implement LoadProject() method
- [ ] Use FindProjectConfig() to locate file
- [ ] Parse YAML
- [ ] Validate against schema
- [ ] Return appropriate errors
- [ ] Verify tests pass (green)

### C6. Schema Validation

**Test First:**
- [ ] Write test for Validate() accepts valid config
- [ ] Write test for Validate() rejects missing required fields
- [ ] Write test for Validate() rejects invalid types
- [ ] Verify tests fail (red)

**Implement:**
- [ ] Create `internal/config/schema.go`
- [ ] Implement Validate() function
- [ ] Use jsonschema library
- [ ] Convert validation errors to ValidationError
- [ ] Verify tests pass (green)

**Verify:**
```bash
go build ./...
go test ./internal/config/...
go vet ./...
```

---

## Phase D: CLI Integration

### D1. Update config get Command

**Implement:**
- [ ] Update `cmd/config.go` to use config.Loader
- [ ] Load global config with LoadGlobal()
- [ ] Support --output flag (yaml, json, table)
- [ ] Display config in requested format
- [ ] Handle errors gracefully

### D2. Update project get Command

**Implement:**
- [ ] Update `cmd/project.go` to use config.Loader
- [ ] Load project config with LoadProject()
- [ ] Support --output flag (yaml, json, table)
- [ ] Display project in requested format
- [ ] Show clear error when yar.yaml not found

**Verify:**
```bash
go build ./...
go test ./...
go vet ./...
./yar config get
./yar config get -o json
./yar project get
```

---

## Functional Tests

| Command | Expected Result | Status |
|---------|-----------------|--------|
| `yar config get` | Shows default config (YAML) | [ ] Pass |
| `yar config get -o json` | Shows default config (JSON) | [ ] Pass |
| `yar config get -o yaml` | Shows default config (YAML) | [ ] Pass |
| `yar project get` | Shows error: no yar.yaml found | [ ] Pass |
| `yar project get` (with yar.yaml) | Shows project config | [ ] Pass |

---

## Completion Checklist

- [ ] All configuration types defined
- [ ] YAML/JSON tags on all structs
- [ ] DefaultConfig() returns valid defaults
- [ ] JSON Schemas created and valid
- [ ] Test fixtures created
- [ ] Loader implementation complete
- [ ] Schema validation works
- [ ] CLI commands use real config
- [ ] All tests pass
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
- [ ] Functional tests verified
- [ ] TASKS.md fully checked off
- [ ] Exit criteria from SPEC.md verified
