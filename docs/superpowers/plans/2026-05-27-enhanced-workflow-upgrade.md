# Enhanced Workflow Guidance Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Force re-check and align this repository's Enhanced Superpowers guidance layer to plugin 5.1.0 semantics without overwriting historical workflow records.

**Architecture:** Treat the upgrade as a documentation-and-metadata migration. Update only the five guidance-layer files (`CLAUDE.md`, `docs/superpowers/README.md`, `docs/superpowers/workflow.md`, `docs/superpowers/conventions.md`, `docs/superpowers/version.json`) while leaving all historical directories untouched. In addition, the upgrade workflow is expected to create its own spec, plan, verification, and debugging records as a normal part of the Enhanced Superpowers process. Because plugin 5.1.0 only ships direct project templates for `CLAUDE.md`, `README.md`, and `status.md`, align `workflow.md` and `conventions.md` to current semantics instead of trying to mechanically copy missing template files.

**Tech Stack:** Markdown, JSON, Python 3, git

---

## File Map

- **Modify:** `CLAUDE.md`
  - Keep the project-level mandatory workflow guardrails aligned to current Enhanced Superpowers semantics.
- **Modify:** `docs/superpowers/README.md`
  - Keep repository-specific navigation and the real Active Features row, but remove nonexistent `status.md` links and make resume order explicit.
- **Modify:** `docs/superpowers/workflow.md`
  - Define the current phase order, execution loop, and the repository's recovery order when `status.md` is absent.
- **Modify:** `docs/superpowers/conventions.md`
  - Define file purposes, naming rules, cross-references, and the rule that recovery links must point only to files that actually exist.
- **Create:** `docs/superpowers/version.json`
  - Backfill legacy metadata with plugin/template version `5.1.0` and a real upgrade timestamp.
- **Preserve exactly:**
  - historical records under `completion/`, `execution-log/`, `review-log/`, `debugging-log/`, `specs/`, `plans/`, `decomposition/`
- **Expected new records created by this upgrade:**
  - design spec, implementation plan, verification records, and debugging log entries as the Enhanced Superpowers workflow proceeds

## Task 1: Add legacy metadata and align repository README

**Files:**
- Create: `docs/superpowers/version.json`
- Modify: `docs/superpowers/README.md`
- Verify: `docs/superpowers/version.json`, `docs/superpowers/README.md`

- [ ] **Step 1: Write the metadata file with the real plugin version and a live timestamp**

Run:

```bash
python3 - <<'PY'
import json
from datetime import datetime, timezone
from pathlib import Path

version_path = Path('/Users/zyb/workspace/person/CloudAlbum/docs/superpowers/version.json')
payload = {
    'pluginVersion': '5.1.0',
    'workflowTemplateVersion': '5.1.0',
    'initializedAt': 'legacy',
    'lastUpgradedAt': datetime.now(timezone.utc).replace(microsecond=0).isoformat().replace('+00:00', 'Z'),
}
version_path.write_text(json.dumps(payload, indent=2) + '\n', encoding='utf-8')
PY
```

Expected: command exits with status 0 and `docs/superpowers/version.json` now exists.

- [ ] **Step 2: Replace `docs/superpowers/README.md` with the aligned repository-specific content**

Write this exact file:

````markdown
# Superpowers Enhanced Workflow

This project uses the Enhanced Superpowers workflow.

- **Workflow phases and per-task loop:** see [workflow.md](workflow.md)
- **Naming conventions, file purposes, and rules:** see [conventions.md](conventions.md)
- **Resume order for this repository:** read the newest relevant record in `completion/`, then `execution-log/`, then `verification-log/` when present for the current work, then `review-log/`, then `debugging-log/`

## Directory Overview

```
docs/superpowers/
├── decomposition/     ← Requirement breakdown (large features only)
├── specs/             ← Design specifications
├── plans/             ← Implementation plans
├── execution-log/     ← Task progress + verification results
├── debugging-log/     ← Bug investigation records
├── review-log/        ← Code review findings and resolutions
├── verification-log/  ← Standalone verification evidence (fallback when no execution log exists yet)
└── completion/        ← Final completion summaries
```

## Quick Navigation

| I want to... | Read this |
|-------------|-----------|
| Understand the full workflow | [workflow.md](workflow.md) |
| Know what each file contains | [conventions.md](conventions.md) |
| Resume work in this repository | Latest relevant file in `completion/`, then `execution-log/`, then `verification-log/` when present for the current work, then `review-log/`, then `debugging-log/` |
| Find a feature's docs | Active Features table below |
| Check current progress | `execution-log/<feature-name>.md` |
| Find deferred issues | `review-log/<feature-name>.md` → Deferred Items section |
| Find past bugs | `debugging-log/` files |
| Find standalone verification evidence | `verification-log/<feature-name>.md` |
| See final feature summary | `completion/<feature-name>-summary.md` |

## Active Features

<!-- Update this table as features are started and completed -->

| Feature | Status | Spec | Plan | Execution | Review | Completion |
|---------|--------|------|------|-----------|--------|------------|
| CloudAlbum 图床 | Complete | [spec](specs/2026-05-25-cloudalbum-design.md) | [plan](plans/2026-05-25-cloudalbum.md) | [log](execution-log/2026-05-25-cloudalbum.md) | [review](review-log/2026-05-25-cloudalbum.md) | [summary](completion/2026-05-26-cloudalbum-summary.md) |

Status: `Planning` | `In Progress` | `In Review` | `Complete` | `On Hold`
````

- [ ] **Step 3: Verify the metadata fields and README links**

Run:

```bash
python3 - <<'PY'
import json
import re
from pathlib import Path

root = Path('/Users/zyb/workspace/person/CloudAlbum/docs/superpowers')
version_data = json.loads((root / 'version.json').read_text(encoding='utf-8'))
assert version_data['pluginVersion'] == '5.1.0'
assert version_data['workflowTemplateVersion'] == '5.1.0'
assert version_data['initializedAt'] == 'legacy'
assert version_data['lastUpgradedAt'].endswith('Z')

readme = (root / 'README.md').read_text(encoding='utf-8')
assert 'status.md' not in readme
links = re.findall(r'\[[^\]]+\]\(([^)]+)\)', readme)
for link in links:
    if '://' in link or link.startswith('#'):
        continue
    target = (root / link).resolve()
    if not target.exists():
        raise AssertionError(f'missing README target: {link}')

print('version.json OK')
print('README links OK')
PY
```

Expected:

```text
version.json OK
README links OK
```

- [ ] **Step 4: Commit the metadata and README alignment**

Run:

```bash
git add docs/superpowers/version.json docs/superpowers/README.md && git commit -m "$(cat <<'EOF'
docs: add workflow version metadata and align README
EOF
)"
```

Expected: a new commit is created containing only `docs/superpowers/version.json` and `docs/superpowers/README.md`.

## Task 2: Align workflow and conventions guidance to 5.1.0 semantics

**Files:**
- Modify: `docs/superpowers/workflow.md`
- Modify: `docs/superpowers/conventions.md`
- Verify: `docs/superpowers/workflow.md`, `docs/superpowers/conventions.md`

- [ ] **Step 1: Replace `docs/superpowers/workflow.md` with the aligned process reference**

Write this exact file:

````markdown
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
````

- [ ] **Step 2: Replace `docs/superpowers/conventions.md` with the aligned conventions reference**

Write this exact file:

````markdown
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
````

- [ ] **Step 3: Verify the updated workflow guidance text**

Run:

```bash
python3 - <<'PY'
from pathlib import Path

root = Path('/Users/zyb/workspace/person/CloudAlbum/docs/superpowers')
workflow = (root / 'workflow.md').read_text(encoding='utf-8')
conventions = (root / 'conventions.md').read_text(encoding='utf-8')
assert 'This repository currently resumes without `status.md`' in workflow
assert 'status.md' in conventions
assert 'Recovery links must be real.' in conventions
assert 'resume in repository order' in conventions
print('workflow.md OK')
print('conventions.md OK')
PY
```

Expected:

```text
workflow.md OK
conventions.md OK
```

- [ ] **Step 4: Commit the workflow and conventions alignment**

Run:

```bash
git add docs/superpowers/workflow.md docs/superpowers/conventions.md && git commit -m "$(cat <<'EOF'
docs: align workflow process guidance
EOF
)"
```

Expected: a new commit is created containing only `docs/superpowers/workflow.md` and `docs/superpowers/conventions.md`.

## Task 3: Align project CLAUDE rules and verify repository recovery behavior

**Files:**
- Modify: `CLAUDE.md`
- Verify: `CLAUDE.md`, preserved workflow directories

- [ ] **Step 1: Replace `CLAUDE.md` with the aligned project guidance**

Write this exact file:

```markdown
## Enhanced Superpowers Workflow

This project uses the Enhanced Superpowers workflow. The following rules are MANDATORY for all AI agents.

### Mandatory Rules

1. **Before writing ANY code**, invoke `brainstorming` skill (Superpowers). No exceptions.
2. **Before executing**, invoke `writing-plans` skill (Superpowers). No exceptions.
3. **After each task commit**, invoke `documenting-execution` skill.
4. **After each verification run**, invoke `documenting-verification` skill.
5. **After each code review cycle**, invoke `documenting-review` skill.
6. **If review finds issues, fix them, run verification, then prefer the original reviewer for re-check.** If the original reviewer is unavailable or still lacks context after a concise recap, fall back to a fresh reviewer.
7. **After resolving any bug**, invoke `documenting-debugging` skill.
8. **Before merge or PR**, invoke `documenting-completion` skill.
9. **For large requirements** (3+ features or subsystems), invoke `decomposing-requirements` skill BEFORE brainstorming.

### Strict Prohibitions

- Do NOT write code before brainstorming is approved by the human partner.
- Do NOT claim work is complete without running verification commands.
- Do NOT silently drop review findings. Every finding = FIXED, DEFERRED (with reason), or REJECTED (with evidence).
- Do NOT skip verification before re-review.
- Do NOT replace re-review with implementer self-assertion. Prefer the original reviewer; use a fresh reviewer only as fallback.
- Do NOT merge or create PR before the completion summary is written.
- Do NOT leave documentation updates uncommitted at session end.

### Documentation

- All workflow docs go in `docs/superpowers/` — see `docs/superpowers/README.md` for navigation and `docs/superpowers/workflow.md` for the required phase order.
- This repository currently resumes without `docs/superpowers/status.md`, so recovery starts from `completion/`, then `execution-log/`, then `verification-log/` when present for the current work, then `review-log/`, then `debugging-log/`.
- Documentation is MANDATORY when using this workflow. Ad-hoc changes under 30 minutes are exempt.
```

- [ ] **Step 2: Verify that historical workflow directories remain untouched and recovery order is inspectable**

Run:

```bash
python3 - <<'PY'
from pathlib import Path

root = Path('/Users/zyb/workspace/person/CloudAlbum/docs/superpowers')
required_dirs = [
    'completion',
    'execution-log',
    'verification-log',
    'review-log',
    'debugging-log',
    'specs',
    'plans',
    'decomposition',
]
for name in required_dirs:
    path = root / name
    if not path.is_dir():
        raise AssertionError(f'missing directory: {name}')
    entries = sorted(p.name for p in path.glob('*.md'))
    print(f'{name}: {len(entries)} markdown files')

claude_md = Path('/Users/zyb/workspace/person/CloudAlbum/CLAUDE.md').read_text(encoding='utf-8')
assert 'This repository currently resumes without `docs/superpowers/status.md`' in claude_md
print('CLAUDE.md recovery rule OK')
PY
```

Expected: one summary line per preserved directory plus `CLAUDE.md recovery rule OK`.

- [ ] **Step 3: Read the latest relevant records in recovery order to prove resume viability**

Run:

```bash
python3 - <<'PY'
from pathlib import Path

root = Path('/Users/zyb/workspace/person/CloudAlbum/docs/superpowers')
for directory in ['completion', 'execution-log', 'verification-log', 'review-log', 'debugging-log']:
    files = sorted(root.joinpath(directory).glob('*.md'))
    latest = files[-1].name if files else 'NONE'
    print(f'{directory}: {latest}')
PY
```

Expected: five lines in this order — `completion`, `execution-log`, `verification-log`, `review-log`, `debugging-log` — each showing the latest file name or `NONE`.

- [ ] **Step 4: Commit the CLAUDE alignment**

Run:

```bash
git add CLAUDE.md && git commit -m "$(cat <<'EOF'
docs: align project workflow guardrails
EOF
)"
```

Expected: a new commit is created containing only `CLAUDE.md`.

## Task 4: Final verification and review handoff

**Files:**
- Verify only: working tree diff and preserved directories

- [ ] **Step 1: Confirm the intended guidance-layer files changed, plus only the expected workflow records for this upgrade**

Run:

```bash
git diff --name-only HEAD~3..HEAD
```

Expected: the diff includes the five guidance-layer files below, and may also include the design/plan/debugging/verification records created during this upgrade flow.

```text
CLAUDE.md
docs/superpowers/README.md
docs/superpowers/conventions.md
docs/superpowers/version.json
docs/superpowers/workflow.md
```

- [ ] **Step 2: Confirm preserved historical records were not rewritten, aside from expected new workflow records for this upgrade**

Run:

```bash
git diff --name-only HEAD~3..HEAD -- docs/superpowers/completion docs/superpowers/execution-log docs/superpowers/review-log docs/superpowers/debugging-log docs/superpowers/specs docs/superpowers/plans docs/superpowers/decomposition
```

Expected: no output for pre-existing tracked history files; only expected new upgrade records are acceptable in `debugging-log/`, `verification-log/`, `plans/`, or `specs/` if they were created during this workflow.

- [ ] **Step 3: Review the final diff before requesting code review**

Run:

```bash
git diff --stat HEAD~3..HEAD && git diff HEAD~3..HEAD
```

Expected: diff shows the intended guidance-layer files plus any expected upgrade records, with no accidental edits to unrelated preserved history.

- [ ] **Step 4: Request review and stop**

Use `superpowers:requesting-code-review`, then follow the repository rule set:

1. If review is clean, continue into the normal verification/completion flow.
2. If review finds issues, fix them, run verification again, document the review result, and prefer the same reviewer for re-check.
3. Do not claim the upgrade complete until review findings are resolved and verification output is recorded.

## Self-Review Checklist

- **Spec coverage:**
  - Guidance-only scope is covered by Tasks 1–3.
  - Legacy `version.json` backfill is covered by Task 1.
  - README preservation of the real Active Features row and avoidance of `status.md` is covered by Task 1.
  - Workflow/conventions semantic alignment is covered by Task 2.
  - Historical record preservation and recovery-order proof are covered by Task 3 and Task 4.
- **Placeholder scan:** No `TODO`, `TBD`, or undefined "later" steps remain.
- **Type consistency:** The plan uses the same five target files and the same recovery order (`completion` → `execution-log` → `verification-log` when present → `review-log` → `debugging-log`) throughout.
