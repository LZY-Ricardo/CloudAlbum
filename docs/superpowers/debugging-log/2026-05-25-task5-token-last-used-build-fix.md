# Task 5 token last-used build fix

**Date:** 2026-05-25
**Triggered by:** VERIFICATION
**Context:** Task 5 from `docs/superpowers/plans/2026-05-25-cloudalbum.md`
**Status:** FIXED

## Symptom
After changing `internal/repository/token.go` to avoid sqlite-incompatible `NOW()`, `go build ./...` failed with an unused import error for `time`.

## Root Cause
The repository implementation was changed in two steps: first the `time` package import was added, but the `UpdateLastUsed` implementation still used the old expression path. That left `time` unused and broke compilation.

## Investigation Path
1. Re-ran `gofmt -w ... && go build ./... && go run cmd/server/main.go` after the repository change.
2. Read the compiler output showing `internal/repository/token.go:4:2: "time" imported and not used`.
3. Checked the updated `UpdateLastUsed` method and confirmed the implementation had not yet switched to `time.Now()`.
4. Concluded the build failure came from an incomplete two-step sqlite compatibility change.

## Fix
Updated `internal/repository/token.go` so `UpdateLastUsed` now calls `UpdateColumn("last_used_at", time.Now())`, which both uses the import and avoids relying on `NOW()` for sqlite compatibility.

**Commit:** `pending`

## Verification
- `gofmt -w /Users/zyb/workspace/person/CloudAlbum/internal/repository/token.go && go build ./... && go run /Users/zyb/workspace/person/CloudAlbum/cmd/server/main.go`
- Result: build passed and startup printed the expected sqlite and local storage initialization output.

## Lessons *(optional)*
When changing database portability logic, update the implementation and imports in the same edit so compile failures do not interrupt verification.

## Cross-References
- Triggered by: `docs/superpowers/execution-log/2026-05-25-cloudalbum.md` (Task 5 verification entry)
- Related debugging: `docs/superpowers/debugging-log/2026-05-25-storage-image-review-findings.md`
