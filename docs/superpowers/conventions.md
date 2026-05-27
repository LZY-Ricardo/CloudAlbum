# Enhanced Workflow — Conventions Reference

## File Purpose

| Path | What's Inside | When Created | Who Reads It |
|------|---------------|--------------|--------------|
| `decomposition/` | Sub-project breakdown, dependencies, and priority order | Before brainstorming, for large features | Anyone scoping the project |
| `specs/` | Design doc: architecture, components, data flow, and decisions | After brainstorming | Developers, reviewers |
| `plans/` | Bite-sized implementation tasks with exact files, code, and verification commands | After spec approval | AI agents, developers tracking progress |
| `execution-log/` | Per-task status, deviations, verification results, and commit SHAs | During implementation | Anyone tracking progress vs. plan |
| `verification-log/` *(fallback)* | Standalone verification records when no execution log is available yet | During verification-only work | Anyone auditing evidence for a change |
| `debugging-log/` | Symptom → root cause → fix → verification → lessons | During debugging | Anyone hitting the same issue |
| `review-log/` | Findings, resolutions, deferred items, and re-review decisions | During code review | Anyone auditing quality decisions |
| `completion/` | Spec vs. reality, known issues, deferred items, and next steps | Before merge or PR | Anyone picking up the work |
| `status.md` *(optional)* | A current recovery entrypoint for active work | Only in repos that choose to maintain it | Anyone resuming a project |

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

Exception: `debugging-log/` uses per-issue names, so one feature can have multiple bug records:

```
debugging-log/2026-05-22-login-null-pointer.md
debugging-log/2026-05-22-token-expiry-race.md
```

## Strict Rules

1. **No code before brainstorming.** Design must be approved before implementation.
2. **No completion claims without evidence.** Run verification, record results, then claim done.
3. **Every review finding has a resolution.** Use FIXED, DEFERRED (with reason + prerequisite), or REJECTED (with evidence). Never silently drop findings.
4. **Fixes must be re-verified before re-review.** Prefer the original reviewer for re-check; use a fresh reviewer only if the original reviewer is unavailable or still lacks context.
5. **No merge before completion summary.** Completion summary is the final gate.
6. **Documentation records are append-only.** Never rewrite historical entries in execution, review, debugging, or completion records.
7. **Commit documentation updates.** Do not leave workflow documentation changes uncommitted at session end.
8. **Recovery links must be real.** Only link to `status.md` or any other artifact when that file actually exists in the repository.
9. **If `status.md` is absent, resume in repository order.** Read `completion/`, then `execution-log/`, then `verification-log/` when present for the current work, then `review-log/`, then `debugging-log/`.
10. **Documentation is mandatory** when using this workflow. Ad-hoc changes under 30 minutes are exempt.

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
  ├── links to → debugging-log/*.md (per task, if applicable)
  └── includes verification results inline

review-log/log.md
  └── links to → debugging-log/*.md (if a finding required debugging)

debugging-log/issue.md
  └── links back to → execution-log or review-log (trigger source)
```
