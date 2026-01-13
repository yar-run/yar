# Iteration 006: Config Commands Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Editor Module

### A1. Editor Detection

**Test First:**
- [x] Write test for DetectEditor with $EDITOR set
- [x] Write test for DetectEditor with $VISUAL fallback
- [x] Write test for DetectEditor with platform default
- [x] Verify tests fail (red)

**Implement:**
- [x] Create `internal/editor/editor.go`
- [x] Implement DetectEditor() function
- [x] Verify tests pass (green)

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/editor/...` passes

### A2. Editor Execution

**Test First:**
- [x] Write test for OpenInEditor (mock exec)

**Implement:**
- [x] Implement OpenInEditor() function
- [x] Inherit stdin/stdout/stderr for interactive editing

**Verify:**
- [x] `go build ./...` succeeds
- [x] `go test ./internal/editor/...` passes

---

## Phase B: Config Edit Command

### B1. Wire Editor to Config Edit

**Implement:**
- [x] Import editor package in cmd/config.go
- [x] Replace stub with OpenInEditor() call
- [x] Create default config file if missing before editing

**Verify:**
- [x] `go build ./...` succeeds
- [x] `./yar config edit` opens editor

---

## Phase C: Project Edit Command

### C1. Wire Editor to Project Edit

**Implement:**
- [x] Import editor package in cmd/project.go
- [x] Replace stub with OpenInEditor() call
- [x] Keep error handling for missing yar.yaml

**Verify:**
- [x] `go build ./...` succeeds
- [x] `./yar project edit` opens editor (with yar.yaml)
- [x] `./yar project edit` shows error (without yar.yaml)

---

## Functional Tests

After this iteration, verify the following commands work:

| Command | Expected Result | Status |
|---------|-----------------|--------|
| `yar config get` | Displays global config summary | ✓ |
| `yar config get -o json` | Outputs JSON | ✓ |
| `yar config get -o yaml` | Outputs YAML | ✓ |
| `yar config edit` | Opens ~/.config/yar/config.yaml in $EDITOR | ✓ |
| `yar project get` | Displays project config (with yar.yaml) | ✓ |
| `yar project get` | Shows error (without yar.yaml) | ✓ |
| `yar project edit` | Opens yar.yaml in $EDITOR | ✓ |
| `yar project edit` | Shows error (without yar.yaml) | ✓ |

**Build and run:**
```bash
cd ~/code/yar
go build -o yar .
./yar config get -o json
EDITOR=vim ./yar config edit
```

---

## Completion Checklist

- [x] Editor module created with DetectEditor/OpenInEditor
- [x] Unit tests for editor detection
- [x] config edit opens real editor
- [x] project edit opens real editor
- [x] Proper error messages when files missing
- [x] `go build ./...` succeeds
- [x] `go test ./...` passes
- [x] `go vet ./...` clean
- [x] Exit criteria from SPEC.md verified

---

## Status

**COMPLETE** - All tasks finished.
