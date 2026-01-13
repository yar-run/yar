# Iteration 006: Config Commands Plan

## Overview

This iteration completes the config and project get/edit commands. The `get` commands are partially implemented; the `edit` commands need real editor integration.

---

## Phases

### Phase A: Editor Module

**Duration**: 30 min

**Objective**: Create editor detection and execution utilities.

**Deliverables**:
- `internal/editor/editor.go` - DetectEditor(), OpenInEditor()
- `internal/editor/editor_test.go` - Unit tests

**Dependencies**: None

### Phase B: Config Edit Command

**Duration**: 15 min

**Objective**: Wire editor into `yar config edit`.

**Deliverables**:
- Updated `cmd/config.go` - Real editor invocation
- Create default config file if missing

**Dependencies**: Phase A

### Phase C: Project Edit Command

**Duration**: 15 min

**Objective**: Wire editor into `yar project edit`.

**Deliverables**:
- Updated `cmd/project.go` - Real editor invocation

**Dependencies**: Phase A

---

## Verification

After completion:
- [ ] `yar config get -o json` outputs JSON
- [ ] `yar config edit` opens $EDITOR
- [ ] `yar project get` works with yar.yaml present
- [ ] `yar project edit` opens $EDITOR
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
