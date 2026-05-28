# Enhanced Workflow — Conventions Reference

## File Purpose

| Directory / File | What's Inside | When Created | Who Reads It |
|-----------|--------------|--------------|-------------|
| `decomposition/` | Sub-project breakdown, dependencies, priority order | Before brainstorming, for large features | Anyone scoping the project |
| `specs/` | Design doc: architecture, interfaces, boundaries, acceptance criteria | After brainstorming | Developers, reviewers |
| `plans/` | Executable implementation plans for work that needs sequencing | After spec approval | AI agents, developers |
| `plans/<feature>.review-config.md` | Execution and review policy for the feature | Before task execution starts | Executors, reviewers, recovery sessions |
| `execution-log/` | One merged block per completed task: execution, verification, self-review, applied config, debugging | During implementation | Anyone tracking progress vs. plan |
| `debugging-log/` | Symptom → root cause → fix for reusable investigations only | During debugging | Anyone hitting the same issue |
| `review-log/` | External/deferred/cross-task review findings | During external review | Anyone auditing quality decisions |
| `completion/` | Final dashboard summary: status, coverage, artifact links, known issues | Before merge/PR | Anyone picking up the work |

## Naming Convention

All docs for the same feature share the same `<feature-name>` slug:

```
decomposition/2026-05-22-user-auth.md
specs/2026-05-22-user-auth-design.md
plans/2026-05-22-user-auth.md
plans/2026-05-22-user-auth.review-config.md
execution-log/2026-05-22-user-auth.md
review-log/2026-05-22-user-auth.md
completion/2026-05-22-user-auth-summary.md
```

Exception: `debugging-log/` uses per-issue names (one feature can have multiple bugs):

```
debugging-log/2026-05-22-login-null-pointer.md
debugging-log/2026-05-22-token-expiry-race.md
```

## Strict Rules

1. **No code before brainstorming.** Design must be approved before implementation.
2. **No completion claims without evidence.** Run verification, record results, then claim done.
3. **Every external review finding has a resolution.** FIXED, DEFERRED (with reason + prerequisite), or REJECTED (with evidence). Never silently drop.
4. **Fixes must be re-verified before re-review.** Prefer the original reviewer for re-check when continuity matters.
5. **No merge before completion summary.** Completion summary is the final gate.
6. **Documentation is append-only** for execution-log, review-log, and debugging-log.
7. **Commit documentation.** Do not leave workflow docs uncommitted at session end.
8. **Documentation is mandatory** when using the full workflow. Truly small ad-hoc changes are exempt.
9. **Self-checklist always runs.** It stays inside the execution-log block unless an external review cycle produces review-log data.
10. **Review-config governs execution.** Planned work must read `review-config.md` at task start and record how it was applied.

## Review Config Format

`plans/<feature>.review-config.md` should define:
- execution mode
- task-level review setting
- feature-level review setting
- review executor
- hard rules
- task-time usage instructions

Typical usage:
- every task starts by loading it
- every task block records `Review (applied config)`
- deviation from the plan can force escalation even if task-level external review is off

## Merged Execution-Log Task Block

This workflow's default uses a **merged execution-log block** per completed task.

Default task block shape:

```markdown
## Task N: [Task Name]

**Execution**
- what changed
- commits
- deviations (if any)

**Verification**
- exact commands
- PASS/FAIL outcomes
- uncovered areas or intentional skips

**Review (self-checklist)**
- Spec mapping
- Interface consistency
- Tests verify behavior
- Smell scan
- Spec-stated boundaries covered
- Plan deviation check

**Review (applied config)**
- what review-config required
- what review actually ran
- whether deviation escalation triggered

**Debugging**
- `N/A` or link to reusable debugging-log entry

---
```

## Review / Debugging Boundaries

- **Execution log:** default home for routine task-closeout data
- **Review log:** only for external review cycles, deferred findings, and cross-task issues
- **Debugging log:** only for investigations with reuse value
- **Completion summary:** dashboard only; link instead of retelling

## Completion Summary Shape

The default completion summary is a dashboard:

```markdown
# [Feature Name] — Completion Summary

**Date:** YYYY-MM-DD
**Branch:** [actual branch name]
**Status:** ✅ Complete / ⚠️ Deferred items exist

## Scope
One sentence describing what shipped.

## Spec coverage
| Req | Status | Notes |
|---|---|---|
| [Requirement] | ✅ | — |
| [Requirement] | ⚠️ | Deferred — see review-log |

## Artifacts
| Document | Link |
|---|---|
| Spec | [link] |
| Plan | [link] |
| Execution log | [link] |
| Review log | [link if relevant] |
| Debugging log | [link if relevant] |

## Known issues / deferred
- Omit the section if none exist.

## Summary
One line on what shipped.
One line on the next step or deferred follow-up.
```

Do **not** use the completion summary to retell the spec, replay tasks, or duplicate fixed review/debug history.

## Document Cross-References

Documents link to each other for traceability:

```
completion/summary.md
  ├── links to → specs/design.md
  ├── links to → plans/plan.md
  ├── links to → plans/review-config.md
  ├── links to → execution-log/log.md
  └── links to → review-log/log.md (if relevant)

execution-log/log.md
  ├── links to → plans/plan.md (header)
  ├── applies → plans/review-config.md
  └── links to → debugging-log/*.md (if applicable)

review-log/log.md
  └── links to → debugging-log/*.md (if a finding required debugging)

debugging-log/issue.md
  └── links back to → execution-log or review-log (trigger source)
```
