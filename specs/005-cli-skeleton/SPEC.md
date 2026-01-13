# Iteration 005: CLI Skeleton Specification

## Overview

This iteration defines all Cobra commands as stubs. Every command from PROJECT.md CLI Reference is implemented with placeholder output, allowing the full CLI structure to be explored via `--help`.

## Scope

### Included
- `cmd/root.go` - Root command with global flags (-v, -o, -h)
- `cmd/fleet.go` - fleet up/down/destroy/restart/status/update
- `cmd/config.go` - config get/edit
- `cmd/project.go` - project init/get/edit
- `cmd/pack.go` - pack list/install/remove
- `cmd/template.go` - template build/render/publish
- `cmd/secret.go` - secret set/get/delete/list/sync
- `cmd/hosts.go` - hosts set/get/delete/list
- `cmd/doctor.go` - doctor run
- `cmd/aliases.go` - hoist/dock/scuttle/swab

### NOT Included (deferred)
- Actual command implementations (later iterations)
- Command-specific flags beyond stubs
- Integration with internal packages

---

## Interfaces

All commands follow the Cobra pattern:

```go
var cmdName = &cobra.Command{
    Use:   "name",
    Short: "Short description",
    Long:  "Long description",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("command: stub message")
    },
}

func init() {
    parentCmd.AddCommand(cmdName)
}
```

---

## Data Structures

### Global Flags

```go
var (
    verbose      bool   // -v, --verbose
    outputFormat string // -o, --output (yaml|json|table)
)
```

---

## Dependencies

### External Packages
- `github.com/spf13/cobra` - CLI framework

### Internal Packages
- None (stubs only)

---

## Invariants

- **INV-DEV-005**: PROJECT.md CLI Reference is the true north. All commands MUST match the CLI Reference.

---

## File Manifest

| File | Purpose |
|------|---------|
| `cmd/root.go` | Root command, global flags |
| `cmd/fleet.go` | fleet subcommands |
| `cmd/config.go` | config subcommands |
| `cmd/project.go` | project subcommands |
| `cmd/pack.go` | pack subcommands |
| `cmd/template.go` | template subcommands |
| `cmd/secret.go` | secret subcommands |
| `cmd/hosts.go` | hosts subcommands |
| `cmd/doctor.go` | doctor subcommands |
| `cmd/aliases.go` | hoist/dock/scuttle/swab aliases |

---

## Test Requirements

### Functional Tests (Manual)
- [ ] `yar --help` shows all top-level commands
- [ ] `yar fleet --help` shows fleet subcommands
- [ ] `yar config --help` shows config subcommands
- [ ] `yar project --help` shows project subcommands
- [ ] `yar pack --help` shows pack subcommands
- [ ] `yar template --help` shows template subcommands
- [ ] `yar secret --help` shows secret subcommands
- [ ] `yar hosts --help` shows hosts subcommands
- [ ] `yar doctor --help` shows doctor subcommands
- [ ] `yar hoist` runs as alias for `yar fleet up`
- [ ] `yar dock` runs as alias for `yar fleet down`
- [ ] `yar scuttle` runs as alias for `yar fleet destroy`
- [ ] `yar swab` runs as alias for `yar doctor run --fix-cache`

---

## Exit Criteria

- [ ] `yar --help` shows all commands
- [ ] `yar <cmd> --help` shows subcommands
- [ ] All stubs print placeholder messages
- [ ] Aliases work
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` clean
