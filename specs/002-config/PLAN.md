# Iteration 002: Configuration Plan

## Overview

This iteration implements the configuration system in four phases: type definitions, path utilities, loading/validation, and CLI integration. All phases follow TDD. Total estimated time: 2.5 hours.

---

## Phases

### Phase A: Configuration Types

**Duration**: 30 minutes

**Objective**: Define all configuration structs with proper YAML/JSON tags.

**Deliverables**:
- `internal/config/types.go` - All type definitions
- `internal/config/defaults.go` - Default values
- `internal/config/types_test.go` - Marshal/unmarshal tests

**Dependencies**: None

**Approach**:
1. Define Config struct with all nested types
2. Define Project struct with all nested types
3. Add YAML and JSON struct tags
4. Implement DefaultConfig()
5. Write tests for marshal/unmarshal

### Phase B: Path Resolution

**Duration**: 20 minutes

**Objective**: Implement config file path resolution.

**Deliverables**:
- `internal/config/paths.go` - Path resolution
- Tests in loader_test.go

**Dependencies**: Phase A, internal/platform

**Approach**:
1. Implement GlobalConfigPath() using platform.ConfigDir()
2. Implement FindProjectConfig() with parent directory search
3. Write tests for path resolution

### Phase C: Loading & Validation

**Duration**: 1 hour

**Objective**: Implement YAML loading and JSON Schema validation.

**Deliverables**:
- `internal/config/loader.go` - Loader implementation
- `internal/config/schema.go` - Validation
- `internal/config/loader_test.go` - Tests
- `internal/config/schema_test.go` - Tests
- `schemas/config.schema.json` - Global schema
- `schemas/project.schema.json` - Project schema
- `internal/config/testdata/` - Test fixtures

**Dependencies**: Phase A, Phase B

**Approach**:
1. Write JSON Schemas for config and project
2. Implement Loader struct with options
3. Implement LoadGlobal() with default fallback
4. Implement LoadProject() with parent search
5. Implement Validate() with jsonschema
6. Write comprehensive tests with fixtures

### Phase D: CLI Integration

**Duration**: 40 minutes

**Objective**: Update CLI commands to display real configuration.

**Deliverables**:
- `cmd/config.go` - Updated with real config loading
- `cmd/project.go` - Updated with real project loading

**Dependencies**: Phase C

**Approach**:
1. Update `yar config get` to load and display global config
2. Support --output flag (yaml, json, table)
3. Update `yar project get` to load and display project config
4. Show clear error when yar.yaml not found
5. Test all commands manually

---

## Verification

After completion:
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/config/...` passes
- [ ] `go vet ./...` clean
- [ ] `yar config get` displays real config
- [ ] `yar config get -o json` outputs JSON
- [ ] `yar project get` shows error (no yar.yaml in repo)
- [ ] Create test yar.yaml, verify `yar project get` works
