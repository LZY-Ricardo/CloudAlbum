# Enhanced Workflow — Process Reference

## Scope Tiering (Choose the Lightest Path That Fits)

| Tier | Typical Shape | Default Path |
|------|----------------|--------------|
| **Small** | ~≤50 LOC, one file, no new public interface, under ~30 minutes | Work directly. Skip the enhanced workflow artifacts. |
| **Medium** | Multi-step but one coherent concern | Brainstorming is still required. Use a plan only if sequencing is non-trivial. |
| **Large** | Cross-file, cross-subsystem, external API, or complex sequencing | Run the full enhanced workflow. |

## Workflow Phases (Mandatory Order When the Full Workflow Applies)

```
Phase 1: decomposing-requirements  (only for truly multi-concern requirements)
    ↓
Phase 2: brainstorming              → specs/
    ↓
Phase 3: writing-plans              → plans/ + review-config.md (only when needed)
    ↓
Phase 4: execution (default = inline)
    ├── load review-config
    ├── implement + verify
    ├── optional debugging-log (only for reusable investigations)
    └── merged execution-log task block
    ↓
Phase 5: documenting-completion     → completion/
    ↓
Phase 6: finishing-a-development-branch → merge / PR
```

**Each phase MUST complete before the next begins.** Skipping phases is not allowed when the full workflow is in use.

## Per-Task Execution Loop

During Phase 4, every planned task follows this default sub-loop:

```
Step 0: Load review-config
    ↓
Implement → commit
    ↓
Verify (tests, build, lint, type-check, manual checks)
    ↓ (if meaningful investigation needed)
Debug → optional debugging-log → re-verify
    ↓
Apply configured review strategy
    ↓
Write one merged execution-log block
    ↓
Next task
```

The merged task block contains:
- Execution
- Verification
- Review (self-checklist)
- Review (applied config)
- Debugging

## Review Strategy Lifecycle

For planned work, create:

`plans/YYYY-MM-DD-<feature-name>.review-config.md`

This file defines:
- execution mode
- task-level review setting
- feature-level review setting
- review executor
- hard rules (self-checklist always on, deviation escalation, etc.)

Execution reads this file at task start and records how it was applied inside each merged task block.

## Phase Details

### Phase 1: Requirement Decomposition (conditional)

- **When:** Requirement involves multiple independent concerns, is vague, or clearly needs ordering/dependency mapping
- **Skill:** `decomposing-requirements`
- **Output:** `decomposition/YYYY-MM-DD-<topic>.md`
- **Next:** User picks first sub-project → Phase 2
- **Skip if:** Single, well-scoped concern

### Phase 2: Brainstorming

- **When:** Always before code for medium/large work
- **Skill:** `brainstorming` (Superpowers)
- **Output:** `specs/YYYY-MM-DD-<feature-name>-design.md`
- **Next:** Planning only if sequencing warrants it; otherwise execution

### Phase 3: Planning

- **When:** Large work, or medium work with non-trivial sequencing
- **Skill:** `writing-plans` (Superpowers)
- **Output:** `plans/YYYY-MM-DD-<feature-name>.md` + `plans/YYYY-MM-DD-<feature-name>.review-config.md`
- **Next:** Phase 4

### Phase 4: Execution

- **When:** Spec is approved and implementation is ready
- **Skill:** `executing-plans` by default; `subagent-driven-development` only when isolation is worth the cost
- **Output:** Working code + git commits + execution-log entries
- **Next:** Phase 5

### Phase 5: Completion

- **When:** All tasks done, all configured reviews passed
- **Skill:** `documenting-completion`
- **Output:** `completion/YYYY-MM-DD-<feature-name>-summary.md`
- **Next:** Phase 6

The completion summary is a dashboard/hub, not a narrative recap.

### Phase 6: Branch Finish

- **When:** Completion summary is written
- **Skill:** `finishing-a-development-branch` (Superpowers)
- **Output:** Merged branch or PR

## Resuming Mid-Project

When joining a project that already has docs:

1. Check `status.md` first
2. Check the active `review-config.md`
3. Check `execution-log/`
4. Check `review-log/` for deferred/cross-task review items
5. Check `debugging-log/` for reusable investigations
6. Resume from the current phase and next recommended action

Do NOT restart the workflow from scratch.
