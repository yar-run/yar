# Iteration 005: CLI Skeleton Plan

## Overview

This iteration creates the complete CLI structure with all commands as stubs. Users can explore the full command tree via `--help`.

---

## Phases

### Phase A: Root Command

**Duration**: 15 min

**Objective**: Set up root command with global flags.

**Deliverables**:
- `cmd/root.go` - Root command with -v, -o flags

**Dependencies**: None

### Phase B: Core Commands

**Duration**: 45 min

**Objective**: Implement main command groups.

**Deliverables**:
- `cmd/fleet.go` - fleet up/down/destroy/restart/status/update
- `cmd/config.go` - config get/edit
- `cmd/project.go` - project init/get/edit

**Dependencies**: Phase A

### Phase C: Supporting Commands

**Duration**: 30 min

**Objective**: Implement remaining command groups.

**Deliverables**:
- `cmd/pack.go` - pack list/install/remove
- `cmd/template.go` - template build/render/publish
- `cmd/secret.go` - secret set/get/delete/list/sync
- `cmd/hosts.go` - hosts set/get/delete/list
- `cmd/doctor.go` - doctor run

**Dependencies**: Phase A

### Phase D: Aliases

**Duration**: 15 min

**Objective**: Implement nautical aliases.

**Deliverables**:
- `cmd/aliases.go` - hoist/dock/scuttle/swab

**Dependencies**: Phase B (fleet commands)

---

## Verification

After completion:
- [ ] `yar --help` lists all commands
- [ ] Each command group has working `--help`
- [ ] Aliases execute correctly
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean
