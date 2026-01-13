# Iteration 004: Config Loading Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Path Resolution

### A1. GlobalConfigPath

**Test First:**
- [x] Write test for GlobalConfigPath returns correct path

**Implement:**
- [x] Implement GlobalConfigPath() using platform.ConfigDir()

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/config/...` passes

### A2. FindProjectConfig

**Test First:**
- [x] Write test for FindProjectConfig in current dir
- [x] Write test for FindProjectConfig in parent dir
- [x] Write test for FindProjectConfig not found

**Implement:**
- [x] Implement FindProjectConfig() with parent traversal

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/config/...` passes

---

## Phase B: Schema Validation

### B1. JSON Schemas

**Implement:**
- [x] Create schemas/config.schema.json
- [x] Create schemas/project.schema.json

**Verify:**
- [x] Schemas are valid JSON

### B2. Validation Functions

**Test First:**
- [x] Write test for ValidateConfig with valid config
- [x] Write test for ValidateProject with missing required field

**Implement:**
- [x] Implement ValidateConfig()
- [x] Implement ValidateProject()

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/config/...` passes

---

## Phase C: Loader Implementation

### C1. Loader Struct

**Implement:**
- [x] Define Loader struct
- [x] Implement NewLoader() with functional options
- [x] Implement GlobalPath() and ProjectPath()

**Verify:**
- [x] `go build ./...` succeeds

### C2. LoadGlobal

**Test First:**
- [x] Write test for LoadGlobal with missing file (returns defaults)
- [x] Write test for LoadGlobal with valid file
- [x] Write test for LoadGlobal with bad YAML
- [x] Write test for LoadGlobal with schema violation

**Implement:**
- [x] Implement LoadGlobal()

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/config/...` passes

### C3. LoadProject

**Test First:**
- [x] Write test for LoadProject not found
- [x] Write test for LoadProject valid file
- [x] Write test for LoadProject missing required field
- [x] Write test for LoadProject searches parent dirs

**Implement:**
- [x] Implement LoadProject()

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/config/...` passes

### C4. Test Fixtures

**Implement:**
- [x] Create testdata/valid/config.yaml
- [x] Create testdata/valid/project.yaml
- [x] Create testdata/invalid/config-bad-yaml.yaml
- [x] Create testdata/invalid/project-missing-field.yaml

---

## Functional Tests

This iteration has no user-facing CLI changes.

**Build and verify:**
```bash
cd ~/code/yar
go build ./...
go test ./internal/config/...
```

---

## Completion Checklist

- [x] Path resolution functions implemented
- [x] Schema validation functions implemented
- [x] Loader struct with LoadGlobal/LoadProject implemented
- [x] JSON Schemas created
- [x] Test fixtures created
- [x] All unit tests written and passing
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] Exit criteria from SPEC.md verified

---

## Status

**COMPLETE** - Implemented as part of iteration 002.
