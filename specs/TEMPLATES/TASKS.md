# Iteration {###}: {Name} Tasks

## Status Legend
- [ ] Not started
- [~] In progress
- [x] Complete
- [!] Blocked

---

## Phase A: {Name}

### A1. {Category}

**Test First:**
- [ ] Write test for {functionality}
- [ ] Verify test fails (red)

**Implement:**
- [ ] Implement {functionality}
- [ ] Verify test passes (green)
- [ ] Refactor if needed

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

### A2. {Category}

**Test First:**
- [ ] Write test for {functionality}

**Implement:**
- [ ] Implement {functionality}

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Phase B: {Name} (if applicable)

### B1. {Category}

**Test First:**
- [ ] Write test for {functionality}

**Implement:**
- [ ] Implement {functionality}

**Verify:**
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes

---

## Functional Tests

After this iteration, verify the following commands work:

| Command | Expected Result |
|---------|-----------------|
| `yar <command>` | Description of expected output |
| `yar <command> --flag` | Description with flag behavior |

**Build and run:**
```bash
cd ~/code/yar
go build -o yar .
./yar <command>
```

---

## Completion Checklist

- [ ] All unit tests written and passing
- [ ] All functional tests verified manually
- [ ] All interfaces match SPEC.md
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes  
- [ ] `go vet ./...` clean
- [ ] TASKS.md fully checked off
- [ ] Exit criteria from SPEC.md verified
