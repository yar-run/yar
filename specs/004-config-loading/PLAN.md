# Iteration 004: Config Loading Plan

## Overview

This iteration implements config file loading with path resolution and schema validation. The Loader provides the primary interface for accessing configuration.

---

## Phases

### Phase A: Path Resolution

**Duration**: 20 min

**Objective**: Implement functions to locate config files.

**Deliverables**:
- `internal/config/paths.go` - GlobalConfigPath(), FindProjectConfig()
- `internal/config/paths_test.go` - Unit tests

**Dependencies**: internal/platform (from iteration 001)

### Phase B: Schema Validation

**Duration**: 30 min

**Objective**: Implement programmatic schema validation.

**Deliverables**:
- `internal/config/schema.go` - ValidateConfig(), ValidateProject()
- `schemas/config.schema.json` - Global config schema
- `schemas/project.schema.json` - Project config schema

**Dependencies**: Phase A (for test integration)

### Phase C: Loader Implementation

**Duration**: 30 min

**Objective**: Implement the Loader struct with Load methods.

**Deliverables**:
- `internal/config/loader.go` - Loader struct, LoadGlobal(), LoadProject()
- `internal/config/loader_test.go` - Unit tests
- `internal/config/testdata/` - Test fixtures

**Dependencies**: Phases A and B

---

## Verification

After completion:
- [ ] `go build ./...` succeeds
- [ ] `go test ./internal/config/...` passes
- [ ] `go vet ./...` clean
- [ ] LoadGlobal returns defaults when no file
- [ ] LoadProject returns NotFoundError when no yar.yaml
