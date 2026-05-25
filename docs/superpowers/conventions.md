# Enhanced Workflow — Conventions Reference

## File Purpose

| Directory | What's Inside | When Created | Who Reads It |
|-----------|--------------|--------------|-------------|
| `decomposition/` | Sub-project breakdown, dependencies, priority order | Before brainstorming, for large features | Anyone scoping the project |
| `specs/` | Design doc: architecture, components, data flow, decisions | After brainstorming | Developers, reviewers |
| `plans/` | Bite-sized tasks (2-5 min) with code, commands, expected output | After spec approved | AI agents, developers tracking progress |
| `execution-log/` | Per-task status, deviations, verification results, commit SHAs | During implementation | Anyone tracking progress vs. plan |
| `debugging-log/` | Symptom → root cause → fix → verification → lessons | During debugging | Anyone hitting the same issue |
| `review-log/` | Findings (CRITICAL/IMPORTANT/MINOR), resolutions, deferred items | During code review | Anyone auditing quality decisions |
| `completion/` | Spec vs. reality, known issues, deferred items, next steps | Before merge/PR | Anyone picking up the work |

## Naming Convention

All docs for the same feature share the same `<feature-name>` slug:

```
decomposition/2026-05-22-user-auth.md
specs/2026-05-22-user-auth-design.md
plans/2026-05-22-user-auth.md
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
3. **Every review finding has a resolution.** FIXED, DEFERRED (with reason + prerequisite), or REJECTED (with evidence). Never silently drop.
4. **No merge before completion summary.** Completion summary is the final gate.
5. **Documentation is append-only.** Never edit past entries in execution-log, review-log, or debugging-log.
6. **Commit documentation.** Do not leave documentation updates uncommitted at session end.
7. **Documentation is mandatory** when using this workflow. Ad-hoc changes under 30 minutes are exempt.

## Document Cross-References

Documents link to each other for traceability:

```
completion/summary.md
  ├── links to → specs/design.md
  ├── links to → plans/plan.md
  ├── links to → execution-log/log.md
  └── links to → review-log/log.md

execution-log/log.md
  ├── links to → plans/plan.md (header)
  ├── links to → debugging-log/*.md (per-task, if applicable)
  └── includes verification results inline

review-log/log.md
  └── links to → debugging-log/*.md (if finding required debugging)

debugging-log/issue.md
  └── links back to → execution-log or review-log (trigger source)
```
