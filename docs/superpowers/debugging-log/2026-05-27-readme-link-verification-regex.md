# README Link Verification Regex Mismatch

**Date:** 2026-05-27
**Triggered by:** VERIFICATION
**Context:** Execution of `docs/superpowers/plans/2026-05-27-enhanced-workflow-upgrade.md`
**Status:** FIXED

## Symptom
The README verification step failed with `AssertionError: missing README target: large features only` while checking `docs/superpowers/README.md`.

## Root Cause
The verification script extracted every parenthesized substring with `\(([^)]+)\)` instead of extracting only Markdown link targets. As a result, explanatory text inside the directory overview, such as `(large features only)`, was treated as a file path.

## Investigation Path
1. Re-ran the extraction logic against `docs/superpowers/README.md` and printed all parenthesized matches.
2. Compared the broad parenthesis matches with actual Markdown link targets.
3. Confirmed that `large features only` appeared only in the broad match set, not in the Markdown link target set.
4. Concluded that the verification script was over-matching plain text rather than validating broken links in the README.

## Fix
Updated the plan's README verification snippet to use `\[[^\]]+\]\(([^)]+)\)`, which only captures Markdown link targets.

**Commit:** `UNCOMMITTED`

## Verification
Re-ran the README/metadata verification using the narrowed regex and got:

- `version.json OK`
- `README links OK`

Also re-ran the workflow/conventions and recovery-order validations successfully.

## Lessons
When validating Markdown links, parse Markdown link syntax rather than generic parentheses to avoid false positives from prose.

## Cross-References
- Triggered by: `docs/superpowers/verification-log/2026-05-27-enhanced-workflow-upgrade.md`
- Related debugging: none
