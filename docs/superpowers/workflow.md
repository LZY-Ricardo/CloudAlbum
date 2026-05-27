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
    ├── documenting-verification    → execution-log/ (preferred) or verification-log/ (fallback)
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

- **When:** Requirement involves 3+ features, is vague, or the user asks how to approach it
- **Skill:** `decomposing-requirements`
- **Output:** `decomposition/YYYY-MM-DD-<topic>.md`
- **Next:** User picks the first sub-project → Phase 2
- **Skip if:** Single, well-scoped feature

### Phase 2: Brainstorming

- **When:** Always, before any code
- **Skill:** `brainstorming` (Superpowers)
- **Output:** `specs/YYYY-MM-DD-<feature-name>-design.md`
- **Next:** Phase 3

### Phase 3: Planning

- **When:** Always, after the spec is approved
- **Skill:** `writing-plans` (Superpowers)
- **Output:** `plans/YYYY-MM-DD-<feature-name>.md`
- **Next:** Phase 4

### Phase 4: Execution

- **When:** Plan is ready
- **Skill:** `subagent-driven-development` or `executing-plans` (Superpowers)
- **Output:** Working changes + git commits
- **Next:** Phase 5

### Phase 5: Completion

- **When:** All tasks are done and all reviews pass
- **Skill:** `documenting-completion`
- **Output:** `completion/YYYY-MM-DD-<feature-name>-summary.md`
- **Next:** Phase 6

### Phase 6: Branch Finish

- **When:** Completion summary is written
- **Skill:** `finishing-a-development-branch` (Superpowers)
- **Output:** Merged branch or PR

## Resuming Mid-Project

1. If `docs/superpowers/status.md` exists and is actively maintained, read it first.
2. This repository currently resumes without `status.md`, so start with the newest relevant file in `completion/`.
3. Continue with the newest relevant file in `execution-log/`.
4. If execution evidence for the current work lives in `verification-log/`, read the newest relevant file there before moving on.
5. Then read the newest relevant file in `review-log/`.
6. Finally read the newest relevant file in `debugging-log/`.
7. Resume from the latest unfinished or follow-up-worthy record.

Do NOT restart the workflow from scratch.
