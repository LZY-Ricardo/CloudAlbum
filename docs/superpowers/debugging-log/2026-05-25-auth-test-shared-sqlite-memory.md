# Auth service tests shared SQLite memory DB

**Date:** 2026-05-25
**Triggered by:** VERIFICATION
**Context:** Task 6 from `docs/superpowers/plans/2026-05-25-cloudalbum.md` / re-verification after code-review fixes
**Status:** FIXED

## Symptom
`go test ./internal/service` failed with a uniqueness error while running the new auth service tests:

```text
UNIQUE constraint failed: users.username
```

The failure happened in `auth_test.go` when inserting the same `users.username` value across multiple tests.

## Root Cause
The new auth tests used the shared in-memory SQLite DSN `file::memory:?cache=shared`, which caused separate test cases to attach to the same in-memory database. Because both tests inserted the same seeded user (`auth-user`), the second test hit the unique index on `users.username`.

## Investigation Path
1. Re-ran `go test ./internal/service` after fixing the handler error-mapping issues and confirmed the failure was isolated to the auth service test package.
2. Read `internal/service/auth_test.go` and verified both tests seeded the same username.
3. Checked the test database DSN and found it used `file::memory:?cache=shared`, which intentionally shares state across connections and test runs using the same DSN.
4. Concluded that the failure was not in auth logic itself, but in test isolation: the test DSN was shared across test cases.

## Fix
Changed the auth test database setup from a global shared in-memory DSN to a per-test named shared-memory DSN:

```go
 dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
```

This keeps SQLite behavior realistic for GORM while giving each test case its own isolated in-memory database.

**Commit:** `pending`

## Verification
- `go test ./internal/service` — PASS
- `go build ./...` — PASS

## Lessons
When using SQLite shared in-memory mode in tests, scope the DSN by test name or another unique identifier. Global `file::memory:?cache=shared` is convenient but can leak state between tests and create misleading failures.

## Cross-References
- Triggered by: `docs/superpowers/execution-log/2026-05-25-cloudalbum.md` (Task 6 verification entry)
- Related debugging: `docs/superpowers/debugging-log/2026-05-25-task5-token-last-used-build-fix.md`
