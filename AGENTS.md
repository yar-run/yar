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
| `specify ###` | Create specs for iteration, notify, STOP |
| `implement ###` | Execute iteration from specs, notify, STOP |
| `continue` | Resume current iteration |

### Command Flow

**Each command is a discrete unit. Complete it, notify, then STOP and wait for the next command.**

```
User: specify 003
Agent: [creates specs/003-*/SPEC.md, PLAN.md, TASKS.md]
Agent: [notifies user]
Agent: [STOPS - waits for next command]

User: implement 003
Agent: [implements from specs]
Agent: [notifies user]
Agent: [STOPS - waits for next command]
```

**NEVER chain commands.** Do not proceed from `specify` to `implement` without user issuing `implement`.

### Pre-Specification Check

Before creating specs for iteration ###:

1. Read PLAN.md to identify what iteration ### should deliver
2. Check if those deliverables already exist in codebase
3. If deliverables exist from an earlier iteration:
   - STOP - do not create retroactive specs
   - Report: "Iteration ### scope was already implemented in iteration ###"
   - Wait for user direction

### Command Recognition

**These commands are IMPERATIVES, not questions.**

When the user issues a command, execute it immediately:
- `specify ###` → Read docs, check for drift, CREATE specs, notify, STOP.
- `implement ###` → Read specs, START coding, notify when complete, STOP.
- `continue` → Resume work immediately. Do not recap unless stuck.

**DO NOT:**
- Chain commands without user instruction
- Proceed from specify to implement automatically
- Create retroactive specs for already-implemented work
- Report status instead of acting
- Treat commands as information requests

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
