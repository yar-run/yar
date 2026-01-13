# Yar Agent Instructions

## Document Map

| File | Purpose | When to Read |
|------|---------|--------------|
| PROJECT.md | Vision, objectives, **CLI Reference** (true north) | Always first |
| DESIGN.md | Architecture, SDKs, patterns | Before implementing |
| SPEC.md | Invariants, schemas, API contracts | Before implementing |
| PLAN.md | All 32 iterations with objectives | To find iteration scope |
| SDD.md | Full spec-driven development process | For detailed process |

## Iteration Structure

```
specs/{###}-{name}/
  SPEC.md   - Technical specification for this iteration
  PLAN.md   - Phased approach within iteration
  TASKS.md  - Step-by-step task checklist
```

---

## Commands

| Command | Action |
|---------|--------|
| `specify ###` | Create specs for iteration (SPEC.md, PLAN.md, TASKS.md) |
| `implement ###` | Execute iteration from specs |
| `continue` | Resume current iteration |

---

## Quick Reference

### `specify ###`

1. Read: PROJECT.md (true north), PLAN.md (find iteration), SPEC.md, DESIGN.md
2. Create: `mkdir -p specs/{###}-{name}`
3. Write: SPEC.md, PLAN.md, TASKS.md using templates in `specs/TEMPLATES/`
4. Branch: `git checkout -b feat/{###}-{name}`

See **SDD.md § "Iteration Lifecycle"** for details.

### `implement ###`

1. **Read** (in order):
   - `PROJECT.md` - CLI Reference (true north)
   - `specs/{iteration}/SPEC.md` - what to build
   - `specs/{iteration}/PLAN.md` - phases
   - `specs/{iteration}/TASKS.md` - checklist

2. **Branch**: `git checkout -b feat/{iteration-name}`

3. **Execute** (TDD): For each task in TASKS.md:
   - Mark `[~]` in progress
   - Write test first (when applicable)
   - Implement until test passes
   - Mark `[x]` complete
   - Commit: `[{iteration}] {type}: {description}`

4. **Verify**:
   ```bash
   go build ./...
   go test ./...
   go vet ./...
   ```

5. **Complete**: `git commit -m "[{iteration}] complete: {summary}"`

See **SDD.md § "Implementation Phase"** for full process.

---

## Rules

### DO:
- Read PROJECT.md CLI Reference as true north
- Read iteration SPEC.md fully before coding
- Update TASKS.md in real-time
- Commit after each logical unit
- Follow interfaces exactly as specified

### DON'T:
- Skip reading documentation
- Deviate from spec without documenting
- Batch unrelated changes in one commit
- Mark tasks complete before verifying
- Proceed to next iteration before current passes exit criteria

---

## Handling Issues

| Situation | Action |
|-----------|--------|
| Spec ambiguous | Decide, document in SPEC.md "Clarifications" section |
| Spec wrong | Document change + rationale, commit as `spec: {change}` |
| Task blocked | Mark `[!]`, add note, continue with unblocked tasks |

See **SDD.md § "Handling Spec Changes"** for details.

---

## Commit Format

```
type(scope): description

Types: feat, fix, test, docs, refactor
Scope: optional, e.g., config, docker, secrets
```

Examples:
```bash
git commit -m "feat(config): add YAML loader"
git commit -m "test(docker): add client unit tests"
git commit -m "docs: update CLI reference"
```
