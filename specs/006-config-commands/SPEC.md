# Iteration 006: Config Commands Specification

## Overview

This iteration implements the `config get/edit` and `project get/edit` commands. The `get` commands display configuration with format options, and the `edit` commands open files in the user's preferred editor.

## Scope

### Included
- `yar config get` - Display global config with -o yaml/json/table
- `yar config edit` - Open global config in $EDITOR
- `yar project get` - Display project config with -o yaml/json/table
- `yar project edit` - Open project config in $EDITOR
- Editor detection ($EDITOR, $VISUAL, platform default)
- File creation with defaults when missing (for config edit)

### NOT Included (deferred)
- `yar project init` - Interactive project creation (later iteration)
- Schema validation after editing
- Config diff/merge functionality

---

## Interfaces

### Editor

```go
// OpenInEditor opens the file at path in the user's preferred editor.
// Blocks until the editor exits.
func OpenInEditor(path string) error

// DetectEditor returns the editor command to use.
// Priority: $EDITOR -> $VISUAL -> platform default
func DetectEditor() string
```

---

## Behavior Specifications

### `yar config get`

1. Load global config via Loader.LoadGlobal()
2. Format output based on --output flag:
   - `table` (default): Human-readable summary
   - `yaml`: Full YAML output
   - `json`: Full JSON output
3. Print to stdout

### `yar config edit`

1. Get global config path via Loader.GlobalPath()
2. If file doesn't exist, create with DefaultConfig() as YAML
3. Detect editor command
4. Execute editor with file path, inheriting stdin/stdout/stderr
5. Wait for editor to exit
6. Exit 0 on success, non-zero on editor failure

### `yar project get`

1. Load project config via Loader.LoadProject()
2. If not found, print error with suggestion to run `project init`, exit 2
3. Format output based on --output flag
4. Print to stdout

### `yar project edit`

1. Find project config via Loader.ProjectPath()
2. If not found, print error with suggestion to run `project init`, exit 2
3. Detect and execute editor
4. Wait for editor to exit

### Editor Detection

Priority order:
1. `$EDITOR` environment variable
2. `$VISUAL` environment variable
3. Platform default:
   - macOS/Linux: `vim`
   - Windows: `notepad`

---

## Dependencies

### External Packages
- `os/exec` - Executing editor command

### Internal Packages
- `internal/config` - Loader, types, defaults
- `internal/errors` - NotFoundError

---

## Invariants

- **INV-CFG-003**: Global config always loadable (returns defaults if missing)
- **INV-CFG-002**: Project config MUST be present for project-scoped commands

---

## File Manifest

| File | Purpose |
|------|---------|
| `cmd/config.go` | Updated config get/edit commands |
| `cmd/project.go` | Updated project get/edit commands |
| `internal/editor/editor.go` | Editor detection and execution |
| `internal/editor/editor_test.go` | Unit tests for editor |

---

## Test Requirements

### Unit Tests
- [ ] Test DetectEditor with $EDITOR set
- [ ] Test DetectEditor with $VISUAL fallback
- [ ] Test DetectEditor with platform default

### Functional Tests (Manual)
- [ ] `yar config get` displays config
- [ ] `yar config get -o json` outputs JSON
- [ ] `yar config get -o yaml` outputs YAML
- [ ] `yar config edit` opens editor
- [ ] `yar project get` displays project (with yar.yaml present)
- [ ] `yar project get` shows error (without yar.yaml)
- [ ] `yar project edit` opens editor (with yar.yaml present)
- [ ] `yar project edit` shows error (without yar.yaml)

---

## Exit Criteria

- [ ] Commands output YAML/JSON based on --output flag
- [ ] Edit opens correct file in editor
- [ ] Errors are clear when files don't exist
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
