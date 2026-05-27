# Enhanced Workflow Guidance Upgrade Design

Date: 2026-05-27
Topic: In-place upgrade of Enhanced Superpowers guidance files for this project

## Goal

Upgrade this repository's Enhanced Superpowers guidance layer to the current plugin template semantics without resetting the project workflow and without overwriting any historical records.

## Current State

The project is already initialized for the Enhanced Superpowers workflow:

- `docs/superpowers/` exists
- project `CLAUDE.md` exists and contains the Enhanced Superpowers workflow rules
- guidance files exist:
  - `docs/superpowers/README.md`
  - `docs/superpowers/workflow.md`
  - `docs/superpowers/conventions.md`
- historical records already exist in workflow directories
- `docs/superpowers/version.json` is missing

This qualifies the project for legacy recognition and in-place guidance upgrade.

## Chosen Approach

Use a template-alignment forced re-check.

This means each guidance file is re-checked against the current plugin guidance even if it appears close to current state. The upgrade updates only the guidance/template layer, while preserving all project records and real project state.

## Scope

Files to update:

- `CLAUDE.md`
- `docs/superpowers/README.md`
- `docs/superpowers/workflow.md`
- `docs/superpowers/conventions.md`
- `docs/superpowers/version.json`

Directories and records that must not be overwritten:

- `docs/superpowers/execution-log/`
- `docs/superpowers/verification-log/`
- `docs/superpowers/review-log/`
- `docs/superpowers/debugging-log/`
- `docs/superpowers/completion/`
- `docs/superpowers/specs/`
- `docs/superpowers/plans/`
- `docs/superpowers/decomposition/`

## Non-Goals

This upgrade will not:

- restart the project workflow
- rewrite or normalize historical logs
- fabricate missing work records
- replace real project state with example template content
- create fake recovery entrypoints such as a synthetic `status.md`

## Per-File Strategy

### `CLAUDE.md`

- Align mandatory workflow rules to the current Enhanced Superpowers guidance
- Preserve this repository's explicit use of the enhanced workflow
- Preserve stricter project-local rules where they do not conflict with current guidance

### `docs/superpowers/README.md`

- Update the navigation and recovery guidance structure to current semantics
- Preserve the real Active Features table entries for this repository
- Avoid references to files that do not exist in this project unless those files are intentionally added as part of the upgrade
- If the current template references `status.md` but the project does not use it, adapt the README to the real project state instead of introducing a fake entrypoint

### `docs/superpowers/workflow.md`

- Align workflow phase order and recovery guidance to the current template semantics
- Preserve compatibility with the repository's existing documentation lifecycle and logs
- Keep resume behavior explicit so future sessions continue from existing records rather than restarting

### `docs/superpowers/conventions.md`

- Align directory purposes, naming rules, cross-reference guidance, and strict rules to the current template semantics
- Treat the file as forward-looking guidance rather than a mandate to rewrite historical records

### `docs/superpowers/version.json`

Create and backfill legacy metadata with these fields:

- `pluginVersion`
- `workflowTemplateVersion`
- `initializedAt`
- `lastUpgradedAt`

Backfill rules:

- `initializedAt`: `"legacy"`
- `lastUpgradedAt`: timestamp for this upgrade
- `pluginVersion`: current installed plugin version
- `workflowTemplateVersion`: current installed workflow template version

## Upgrade Rules

1. Guidance semantics follow the current installed plugin version.
2. Real project state takes priority over example template content.
3. Historical records are preserved exactly as they are.
4. Recovery paths must be valid for this repository after the upgrade.
5. Missing metadata is backfilled conservatively for a recognized legacy project.

## Verification Plan

After the upgrade, verify three areas.

### 1. Historical record integrity

Confirm these directories still exist and were not overwritten:

- `completion/`
- `execution-log/`
- `verification-log/`
- `review-log/`
- `debugging-log/`
- `specs/`
- `plans/`
- `decomposition/`

### 2. Guidance layer completeness

Confirm:

- the four guidance files reflect current workflow semantics
- `docs/superpowers/version.json` exists
- `version.json` contains all required fields

### 3. Resume readiness

Confirm that future work can resume from existing records in this order:

1. `completion/`
2. `execution-log/`
3. `verification-log/` when present for the current work
4. `review-log/`
5. `debugging-log/`

The expected result is that an agent can continue from project records without restarting the workflow from zero.

## Risks and Mitigations

### Risk: overwriting project-specific context

Mitigation: preserve real Active Features entries and adapt template language to the repository's actual files.

### Risk: introducing dead links

Mitigation: do not reference optional files such as `status.md` unless they truly exist or are intentionally created.

### Risk: confusing old logs with new rules

Mitigation: update only forward-looking guidance files and leave historical records untouched.

## Success Criteria

The upgrade is successful when:

- guidance files are aligned to current plugin semantics
- `docs/superpowers/version.json` exists with complete legacy-backfilled metadata
- historical workflow records remain intact
- the repository can resume from existing completion, execution, verification, review, and debugging records without ambiguity
