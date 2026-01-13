# Iteration 005: CLI Skeleton Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: Root Command

### A1. Root Command Setup

**Implement:**
- [x] Create cmd/root.go with rootCmd
- [x] Add --verbose (-v) flag
- [x] Add --output (-o) flag with yaml/json/table options
- [x] Set up Execute() function

**Verify:**
- [x] `go build ./...` succeeds
- [x] `./yar --help` works

---

## Phase B: Core Commands

### B1. Fleet Commands

**Implement:**
- [x] Create cmd/fleet.go with fleetCmd parent
- [x] Add fleet up subcommand
- [x] Add fleet down subcommand
- [x] Add fleet destroy subcommand
- [x] Add fleet restart subcommand
- [x] Add fleet status subcommand
- [x] Add fleet update subcommand

**Verify:**
- [x] `./yar fleet --help` shows all subcommands

### B2. Config Commands

**Implement:**
- [x] Create cmd/config.go with configCmd parent
- [x] Add config get subcommand
- [x] Add config edit subcommand

**Verify:**
- [x] `./yar config --help` shows all subcommands

### B3. Project Commands

**Implement:**
- [x] Create cmd/project.go with projectCmd parent
- [x] Add project init subcommand
- [x] Add project get subcommand
- [x] Add project edit subcommand

**Verify:**
- [x] `./yar project --help` shows all subcommands

---

## Phase C: Supporting Commands

### C1. Pack Commands

**Implement:**
- [x] Create cmd/pack.go with packCmd parent
- [x] Add pack list subcommand
- [x] Add pack install subcommand
- [x] Add pack remove subcommand

**Verify:**
- [x] `./yar pack --help` shows all subcommands

### C2. Template Commands

**Implement:**
- [x] Create cmd/template.go with templateCmd parent
- [x] Add template build subcommand
- [x] Add template render subcommand
- [x] Add template publish subcommand

**Verify:**
- [x] `./yar template --help` shows all subcommands

### C3. Secret Commands

**Implement:**
- [x] Create cmd/secret.go with secretCmd parent
- [x] Add secret set subcommand
- [x] Add secret get subcommand
- [x] Add secret delete subcommand
- [x] Add secret list subcommand
- [x] Add secret sync subcommand

**Verify:**
- [x] `./yar secret --help` shows all subcommands

### C4. Hosts Commands

**Implement:**
- [x] Create cmd/hosts.go with hostsCmd parent
- [x] Add hosts set subcommand
- [x] Add hosts get subcommand
- [x] Add hosts delete subcommand
- [x] Add hosts list subcommand

**Verify:**
- [x] `./yar hosts --help` shows all subcommands

### C5. Doctor Commands

**Implement:**
- [x] Create cmd/doctor.go with doctorCmd parent
- [x] Add doctor run subcommand

**Verify:**
- [x] `./yar doctor --help` shows all subcommands

---

## Phase D: Aliases

### D1. Nautical Aliases

**Implement:**
- [x] Create cmd/aliases.go
- [x] Add hoist as alias for fleet up
- [x] Add dock as alias for fleet down
- [x] Add scuttle as alias for fleet destroy
- [x] Add swab as alias for doctor run --fix-cache

**Verify:**
- [x] `./yar hoist` works
- [x] `./yar dock` works
- [x] `./yar scuttle` works
- [x] `./yar swab` works

---

## Functional Tests

After this iteration, verify the following commands work:

| Command | Expected Result |
|---------|-----------------|
| `yar --help` | Lists all top-level commands |
| `yar fleet --help` | Lists fleet subcommands |
| `yar config --help` | Lists config subcommands |
| `yar project --help` | Lists project subcommands |
| `yar pack --help` | Lists pack subcommands |
| `yar template --help` | Lists template subcommands |
| `yar secret --help` | Lists secret subcommands |
| `yar hosts --help` | Lists hosts subcommands |
| `yar doctor --help` | Lists doctor subcommands |
| `yar hoist` | Runs fleet up alias |
| `yar dock` | Runs fleet down alias |
| `yar scuttle` | Runs fleet destroy alias |
| `yar swab` | Runs doctor fix alias |

**Build and run:**
```bash
cd ~/code/yar
go build -o yar .
./yar --help
./yar fleet --help
./yar hoist
```

---

## Completion Checklist

- [x] Root command with global flags
- [x] All command groups implemented
- [x] All subcommands print stub messages
- [x] Aliases work correctly
- [x] `go build ./...` succeeds
- [x] `go vet ./...` clean
- [x] Exit criteria from SPEC.md verified

---

## Status

**COMPLETE** - Implemented as part of iteration 001.
