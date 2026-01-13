# Spec-Driven Development (SDD) Process

This document describes how to execute the Yar project using Spec-Driven Development. Each phase from PLAN.md becomes an SDD iteration with its own specification, plan, and task list.

## SDD Philosophy

**Specification First**: Write the specification before writing code. The spec defines what the code must do, the interfaces it exposes, and the invariants it maintains. Code is written to satisfy the spec.

**Executable Documentation**: Specs are not just documents—they drive tests. Every invariant in a spec becomes a test case. Every interface becomes a contract that tests verify.

**Iterative Refinement**: Each iteration builds on the previous. Specs may be refined as implementation reveals constraints, but changes are explicit and documented.

**Functional Testing Per Iteration**: Every iteration MUST deliver user-testable functionality. Internal refactors or foundational work must be bundled with user-facing features. After each iteration, the user should be able to run the `yar` binary and observe new behavior.

**Build Verification**: Every task that modifies code MUST include explicit build and test verification steps. Never assume code compiles—verify it.

## Iteration Scope Requirements

**PLAN.md is the 1:1 source of truth for iteration scope.**

Each iteration in PLAN.md defines exactly what that iteration delivers. When creating specs for iteration ###:

1. Read PLAN.md § Phase ### to identify scope
2. Create specs for EXACTLY that scope - no more, no less
3. Do NOT bundle multiple iterations together
4. Do NOT pull work forward from future iterations

### Scope Discipline

- Iteration 003 specs contain ONLY what PLAN.md § Phase 003 defines
- If iteration 003 depends on 002, that's fine - 002 must be complete first
- Foundational work (types, utilities) ships in its designated iteration even if not yet user-testable
- User-testable functionality comes when PLAN.md schedules it

## Directory Structure

```
specs/
├── 001-foundation/
│   ├── SPEC.md      # Technical specification for this iteration
│   ├── PLAN.md      # Phased approach within this iteration
│   └── TASKS.md     # Detailed task checklist
├── 002-docker/
│   ├── SPEC.md
│   ├── PLAN.md
│   └── TASKS.md
├── 003-kubernetes/
│   └── ...
└── ...
```

## Iteration Lifecycle

### 1. Specification Phase

**Goal**: Define exactly what will be built in this iteration.

**Inputs**:
- Root SPEC.md (overall technical specification)
- Root PLAN.md (phase objectives)
- Previous iteration deliverables

**Outputs**:
- `specs/{iteration}/SPEC.md`

**SPEC.md Template**:

```markdown
# {Iteration Name} Specification

## Overview
One paragraph describing what this iteration delivers.

## Scope
What IS included in this iteration:
- Item 1
- Item 2

What is NOT included (deferred to later):
- Item 1

## Interfaces

### {InterfaceName}

```go
type InterfaceName interface {
    Method(ctx context.Context, args) (returns, error)
}
```

**Method**: Description of what it does.

### Dependencies
- External packages required
- Internal packages from previous iterations

## Data Structures

```go
type StructName struct {
    Field type `json:"field" yaml:"field"`
}
```

## Invariants
Reference to root SPEC.md invariants that apply, plus any new ones.

## File Manifest
List of files to be created/modified with descriptions.

## Test Requirements
- Unit tests required
- Integration tests required
- Test fixtures needed

## Exit Criteria
Checkable criteria that define "done".
```

### 2. Planning Phase

**Goal**: Break the spec into an executable plan.

**Inputs**:
- `specs/{iteration}/SPEC.md`

**Outputs**:
- `specs/{iteration}/PLAN.md`

**PLAN.md Template**:

```markdown
# {Iteration Name} Plan

## Phases

### Phase A: {Name}
**Duration**: {time estimate}

**Objective**: What this phase accomplishes.

**Deliverables**:
- File or feature 1
- File or feature 2

**Dependencies**: What must exist before this phase.

### Phase B: {Name}
...

## Sequence Diagram
Visual representation of component interactions.

## Risk Assessment
| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| ... | ... | ... | ... |

## Rollback Plan
How to revert if the iteration fails.
```

### 3. Task Breakdown Phase

**Goal**: Create actionable checklist for implementation.

**Inputs**:
- `specs/{iteration}/PLAN.md`

**Outputs**:
- `specs/{iteration}/TASKS.md`

**TASKS.md Template**:

```markdown
# {Iteration Name} Tasks

## Status
- [ ] Not started
- [~] In progress
- [x] Complete

## Phase A: {Name}

### Setup
- [ ] Task 1: Description
- [ ] Task 2: Description

### Implementation
- [ ] Task 3: Description
  - [ ] Subtask 3.1
  - [ ] Subtask 3.2
- [ ] Task 4: Description

### Verification
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `go vet ./...` clean

## Phase B: {Name}
...

## Functional Tests

After this iteration, verify:

| Command | Expected Result |
|---------|-----------------|
| `yar <command>` | Description of expected output |

## Completion Checklist
- [ ] All tests pass
- [ ] All functional tests verified manually
- [ ] Code reviewed
- [ ] Documentation updated
- [ ] Exit criteria from SPEC.md met
```

### 4. Implementation Phase

**Goal**: Write code to satisfy the spec using Test-Driven Development.

**TDD Procedure** (for all iterations with testable code):

1. **Red**: Write a failing test that defines expected behavior
2. **Green**: Write minimal code to make the test pass
3. **Refactor**: Clean up code while keeping tests green

**Task Execution**:
1. Read TASKS.md, mark current task as `[~]` (in-progress)
2. Write test first → verify it fails (red)
3. Implement the code → verify test passes (green)
4. Refactor if needed → verify tests still pass
5. Mark task as `[x]` (complete) in TASKS.md immediately
6. Commit with descriptive message
7. Repeat for next task

**TASKS.md Format for TDD**:
```markdown
### A1. {Feature Category}

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
```

**CRITICAL**: The "Verify" section is mandatory for every task group. Never skip build verification.

**Note**: Scaffolding iterations (like 000) with no testable logic are exempt from TDD.

**Commit Message Format**:
```
[{iteration}] {type}: {description}

{body}

Refs: specs/{iteration}/TASKS.md#{task-number}
```

Types: `feat`, `fix`, `test`, `docs`, `refactor`

### 5. Verification Phase

**Goal**: Confirm the iteration meets its spec.

**Process**:
1. Run all tests (`go test ./...`)
2. Check each exit criterion from SPEC.md
3. Review invariants—do they hold?
4. Run linter (`golangci-lint run`)
5. Build for all platforms
6. Manual smoke test of new commands

### 6. Functional Testing Phase

**Goal**: Verify user-facing functionality works as expected.

**Requirement**: Every iteration MUST have a "Functional Tests" section in TASKS.md that describes:
- Commands the user can run
- Expected output or behavior
- How to verify the feature works

**Example**:
```markdown
## Functional Tests

After this iteration, verify:

| Command | Expected Result |
|---------|-----------------|
| `yar fleet up` | Shows "starting services for environment 'local'" |
| `yar config get` | Shows path to config file |
| `yar --help` | Lists all commands including new ones |
```

This ensures:
- No iteration ships "invisible" internal-only changes
- User can validate work immediately
- Features are demonstrable, not theoretical

### 7. Documentation Phase

**Goal**: Update docs to reflect new functionality.

**Process**:
1. Update root README.md if user-facing changes
2. Add/update command docs in `docs/commands/`
3. Update ARCHITECTURE.md if structural changes
4. Add examples for new features

### 8. Completion Phase

**Goal**: Mark iteration complete and prepare for next.

**Process**:
1. Final commit with all changes
2. Tag release if appropriate
3. Update TASKS.md—all items should be checked
4. Create specs/{next-iteration}/ directory
5. Begin next iteration's spec

---

## Writing Effective Specs

### Be Specific
Bad: "The config loader loads configuration"
Good: "Load() reads ~/.config/yar/config.yaml, validates against config.schema.json, and returns a *Config or error"

### Define Interfaces First
Define Go interfaces in the spec before implementation. This:
- Clarifies what the component does
- Enables parallel development
- Makes testing straightforward

### Include Examples
Show example input and output:

```yaml
# Input: yar.yaml
project: my-app
services:
  - name: redis
    pack: redis
```

```yaml
# Output: docker-compose.yaml
version: "3"
services:
  redis.my-app:
    image: redis:latest
```

### Reference Invariants
Instead of restating invariants, reference them:
"This implementation MUST satisfy INV-SEC-001 through INV-SEC-005."

### State What's NOT Included
Explicitly list what's out of scope to prevent scope creep:
"NOT included: Remote pack installation, pack versioning, pack signing"

---

## Task Granularity

### Too Coarse
- [ ] Implement Docker integration

### Too Fine
- [ ] Create docker/client.go file
- [ ] Add import for docker SDK
- [ ] Define Client struct
- [ ] Add dockerClient field

### Just Right
- [ ] Implement docker.Client struct with NewClient() constructor
- [ ] Implement NetworkCreate and NetworkRemove methods
- [ ] Implement ContainerCreate, ContainerStart, ContainerStop methods
- [ ] Add unit tests for Client with mock Docker API

---

## Updating TASKS.md During Implementation

As you work, update TASKS.md in real-time:

1. **Before starting a task**: Change `- [ ]` to `- [~]` (in progress)
2. **After completing**: Change `- [~]` to `- [x]`
3. **If blocked**: Add note below task explaining blocker
4. **If discovering new tasks**: Add them under appropriate phase

Example:
```markdown
- [x] Implement docker.Client struct
- [~] Implement network operations
  - [x] NetworkCreate
  - [ ] NetworkRemove (blocked: need to handle existing containers)
- [ ] Implement container operations
```

---

## Handling Spec Changes

Sometimes implementation reveals spec issues. Handle them explicitly:

1. **Minor clarifications**: Update spec inline, commit with message "spec: clarify {topic}"

2. **Interface changes**: 
   - Document the change and rationale in spec
   - Update all affected files
   - Commit with message "spec: revise {interface} - {reason}"

3. **Scope changes**:
   - If removing scope: Move to "NOT included" section
   - If adding scope: Add to scope with note "(added during implementation)"
   - Commit with message "spec: adjust scope - {reason}"

---

## Integration with Git

### Branch Strategy
```
main
  └── feat/001-foundation
        ├── commit: [001] feat: add config loading
        ├── commit: [001] feat: add schema validation
        └── commit: [001] test: add config tests
```

### Commit Frequency
- Commit after each completed task or logical unit
- Don't batch multiple tasks into one commit
- Each commit should leave the code in a working state

### Pull Request
At iteration end:
- Create PR from `feat/{iteration}` to `main`
- PR description references `specs/{iteration}/SPEC.md`
- Ensure all TASKS.md items are checked
- Squash merge to keep history clean

---

## Example: Starting Iteration 001

### Step 1: Create spec directory
```bash
mkdir -p specs/001-foundation
```

### Step 2: Write SPEC.md
Define interfaces, data structures, invariants, files, tests, exit criteria.

### Step 3: Write PLAN.md
Break into phases: setup, config loading, CLI commands, testing.

### Step 4: Write TASKS.md
List every task with checkboxes.

### Step 5: Create branch
```bash
git checkout -b feat/001-foundation
```

### Step 6: Implement
Work through TASKS.md, committing after each task.

### Step 7: Verify
Run tests, check exit criteria, review invariants.

### Step 8: Complete
Merge to main, create next iteration specs.

---

## Checklist for Each Iteration

Before starting:
- [ ] SPEC.md written and reviewed
- [ ] PLAN.md created from spec
- [ ] TASKS.md created from plan
- [ ] Branch created

During implementation:
- [ ] TASKS.md updated in real-time
- [ ] Tests written before/with implementation
- [ ] Commits reference task numbers
- [ ] Blockers documented

After completion:
- [ ] All tests pass
- [ ] All TASKS.md items checked
- [ ] Exit criteria verified
- [ ] Documentation updated
- [ ] PR created and merged
