# Iteration {###}: {Name} Specification

## Overview

One paragraph describing what this iteration delivers.

## Scope

### Included
- Item 1
- Item 2

### NOT Included (deferred)
- Item 1

---

## Interfaces

### {InterfaceName}

```go
type InterfaceName interface {
    Method(ctx context.Context, arg Type) (Result, error)
}
```

**Method**: Description of what it does.

---

## Data Structures

```go
type StructName struct {
    Field Type `json:"field" yaml:"field"`
}
```

---

## Dependencies

### External Packages
- `package/name` - purpose

### Internal Packages
- `internal/package` - purpose

---

## Invariants

Reference applicable invariants from root SPEC.md:
- **INV-XXX-NNN**: Description

---

## File Manifest

| File | Purpose |
|------|---------|
| `internal/package/file.go` | Description |
| `internal/package/file_test.go` | Unit tests |

---

## Test Requirements

### Unit Tests
- [ ] Test case 1
- [ ] Test case 2

### Integration Tests
- [ ] Test case 1 (if applicable)

### Test Fixtures
- `test/fixtures/path/file` - purpose

---

## Exit Criteria

- [ ] All interfaces implemented
- [ ] All unit tests pass
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean
