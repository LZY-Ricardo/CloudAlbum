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
