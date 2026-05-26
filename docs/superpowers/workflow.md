# Enhanced Workflow — Process Reference

## Workflow Phases (Mandatory Order)

```
Phase 1: decomposing-requirements  (only for large requirements)
    ↓
Phase 2: brainstorming              → specs/
    ↓
Phase 3: writing-plans              → plans/
    ↓
Phase 4: execution (per task loop)
    ├── documenting-execution       → execution-log/
    ├── documenting-verification    → execution-log/ (appended)
    ├── documenting-debugging       → debugging-log/ (if bugs found)
    └── documenting-review          → review-log/
    ↓
Phase 5: documenting-completion     → completion/
    ↓
Phase 6: finishing-a-development-branch → merge / PR
```

**Each phase MUST complete before the next begins.** Skipping phases is not allowed.

## Per-Task Execution Loop

During Phase 4, every task follows this sub-loop:

```
Implement → commit → documenting-execution
    ↓
Verify (tests, build, lint) → documenting-verification
    ↓ (if fails)
Debug → documenting-debugging → re-verify → re-record
    ↓
Code review → documenting-review
    ↓ (if issues)
Fix → verification → prefer same reviewer re-check
    ↓
Fallback to fresh reviewer if needed → update review log
    ↓
Next task
```

## Phase Details

### Phase 1: Requirement Decomposition (conditional)

- **When:** Requirement involves 3+ features, is vague, or user asks "how to approach this"
- **Skill:** `decomposing-requirements`
- **Output:** `decomposition/YYYY-MM-DD-<topic>.md`
- **Next:** User picks first sub-project → Phase 2
- **Skip if:** Single, well-scoped feature

### Phase 2: Brainstorming

- **When:** Always, before any code
- **Skill:** `brainstorming` (Superpowers)
- **Output:** `specs/YYYY-MM-DD-<feature-name>-design.md`
- **Next:** Phase 3

### Phase 3: Planning

- **When:** Always, after spec is approved
- **Skill:** `writing-plans` (Superpowers)
- **Output:** `plans/YYYY-MM-DD-<feature-name>.md`
- **Next:** Phase 4

### Phase 4: Execution

- **When:** Plan is ready
- **Skill:** `subagent-driven-development` or `executing-plans` (Superpowers)
- **Output:** Working code + git commits
- **Next:** Phase 5

### Phase 5: Completion

- **When:** All tasks done, all reviews passed
- **Skill:** `documenting-completion`
- **Output:** `completion/YYYY-MM-DD-<feature-name>-summary.md`
- **Next:** Phase 6

### Phase 6: Branch Finish

- **When:** Completion summary is written
- **Skill:** `finishing-a-development-branch` (Superpowers)
- **Output:** Merged branch or PR

## Resuming Mid-Project

When joining a project that already has docs:

1. Check for **completion summary** in `completion/` — if it exists, the feature is done
2. Check **execution-log** — shows which tasks are done, blocked, or remaining
3. Check **review-log** — shows deferred items that might need attention
4. Check **debugging-log** — shows known issues and workarounds
5. Resume from where the last entry left off

Do NOT restart the workflow from scratch.
