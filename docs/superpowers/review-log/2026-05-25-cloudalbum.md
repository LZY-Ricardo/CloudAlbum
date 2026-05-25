### Review Cycle 1 — 2026-05-25 16:43:16 CST

**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent
**Scope:** Task 2-4 (Data Models / Storage Backend / Image Processing)
**Preceded by:** Spec compliance passed for Task 2 and Task 4; Task 3 implementation aligned functionally and was accepted before quality review

#### Findings

| # | Severity | Description | Resolution | Commit | Cross-task? |
|---|----------|-------------|------------|--------|-------------|
| 1 | IMPORTANT | `internal/storage/local.go` allows path traversal because `filepath.Join(basePath, key)` is used without verifying that the resolved path stays under the storage root. | FIXED | — | Also affects Task 5, 7 |
| 2 | IMPORTANT | `internal/storage/local.go` `Get()` converts not-found into a plain formatted error, so callers cannot use `errors.Is(err, os.ErrNotExist)` to distinguish 404 from 500. | FIXED | — | Also affects Task 5, 7 |
| 3 | IMPORTANT | `internal/image/processor.go` thumbnail output behavior is inconsistent with configuration intent: `Quality` may be ignored for non-JPEG outputs and `AutoConvert` is not applied clearly. | FIXED | — | Also affects Task 7 |
| 4 | IMPORTANT | `internal/image/processor.go` `DetectImageType` is too permissive for SVG/XML and does not fully validate WebP signatures. | FIXED | — | Also affects Task 7 |
| 5 | IMPORTANT | There are no tests covering storage path safety, not-found semantics, image type detection, or thumbnail behavior. | FIXED | — | Also affects Task 5, 7 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 2 — 2026-05-25 16:53 CST

**Reviewer type:** CODE_QUALITY
**Reviewer:** self-review after remediation
**Scope:** Task 23 Resolve Task 3-4 review findings
**Preceded by:** Review Cycle 1 (code quality findings recorded for Task 2-4)

#### Findings

| # | Severity | Description | Resolution | Commit | Cross-task? |
|---|----------|-------------|------------|--------|-------------|
| 1 | IMPORTANT | `internal/storage/local.go` joined `basePath` and `key` without enforcing that paths remain under the storage root. | FIXED | this commit | Also affects Task 5, 7 |
| 2 | IMPORTANT | `internal/storage/local.go` `Get()` hid `os.ErrNotExist`, preventing callers from distinguishing missing files from other I/O errors. | FIXED | this commit | Also affects Task 5, 7 |
| 3 | IMPORTANT | `internal/image/processor.go` thumbnail output inherited source format inconsistently, so `Quality` could be ignored and `AutoConvert` behavior was unclear. | FIXED | this commit | Also affects Task 7 |
| 4 | IMPORTANT | `internal/image/processor.go` `DetectImageType` accepted overly broad WebP/SVG signatures. | FIXED | this commit | Also affects Task 7 |
| 5 | IMPORTANT | The storage and image packages lacked regression tests for the review findings. | FIXED | this commit | Also affects Task 5, 7 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging
- Findings #1-#5 → `docs/superpowers/debugging-log/2026-05-25-storage-image-review-findings.md`

---

### Review Cycle 3 — 2026-05-25 17:14 CST

**Reviewer type:** CODE_QUALITY
**Reviewer:** self-review
**Scope:** Task 5 Database Init + Repository
**Preceded by:** Task 5 implementation and verification

#### Findings

| # | Severity | Description | Resolution | Commit | Cross-task? |
|---|----------|-------------|------------|--------|-------------|
| 1 | IMPORTANT | `internal/repository/token.go` used `gorm.Expr("NOW()")` for `UpdateLastUsed`, which is not portable to sqlite and could fail when token usage is updated on the default database backend. | FIXED | e52b3e4 | Also affects Task 6, 7 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging
- Finding #1 → `docs/superpowers/debugging-log/2026-05-25-task5-token-last-used-build-fix.md`

---

### Review Cycle 4 — 2026-05-25 17:24 CST

**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent
**Scope:** Task 5 Database Init + Repository
**Preceded by:** Review Cycle 3 (Task 5 self-review portability fix already applied)

#### Findings

| # | Severity | Description | Resolution | Commit | Cross-task? |
|---|----------|-------------|------------|--------|-------------|
| 1 | IMPORTANT | `cmd/server/main.go` opens sqlite without enabling key pragmas (`foreign_keys`, `busy_timeout`, `journal_mode`), so model foreign keys are not enforced and concurrent handler writes may hit `SQLITE_BUSY`. | FIXED | e52b3e4 | Also affects Task 7 |
| 2 | IMPORTANT | `internal/repository/image.go` soft-delete / `OnlyDeleted` listing semantics need to be made explicit and covered by tests so trash listing behavior does not regress silently. | FIXED | e52b3e4 | Also affects Task 7 |
| 3 | MINOR | `internal/repository/image.go` pagination accepts non-positive page/pageSize values and can silently produce brittle offsets or empty limits. | FIXED | e52b3e4 | Also affects Task 7 |
| 4 | IMPORTANT | `internal/repository` layer has no tests covering deleted-image listing, restore semantics, and aggregate behavior on empty datasets. | FIXED | e52b3e4 | Also affects Task 7 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 5 — 2026-05-25 19:38 CST

**Cycle ID:** RC-5
**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent
**Scope:** Task 6 Auth System
**Preceded by:** Task 6 implementation and focused verification

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `internal/handler/auth.go` maps every `AuthService.Login()` error to 401 and returns raw error text, which can misclassify repository/database failures as credential failures and leak backend error details. | FIXED | OPEN | 2525a38 | Also affects Task 7 |
| 2 | IMPORTANT | `internal/handler/token.go` `Delete()` maps all non-not-found failures to 403, so operational/database failures are mislabeled as authorization problems. | FIXED | OPEN | 2525a38 | Also affects Task 7 |
| 3 | MINOR | `internal/service/token_test.go` only covers API token create/validate and does not exercise JWT/login or middleware branching. | FIXED | OPEN | 2525a38 | Also affects Task 7 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 6 — 2026-05-25 19:44 CST

**Cycle ID:** RC-6
**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent
**Scope:** Task 6 Auth System
**Preceded by:** Review Cycle 5
**Re-check of:** Review Cycle 5
**Original reviewer:** subagent
**Re-check reviewer:** fresh reviewer (original reviewer unavailable in-session)

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `internal/handler/auth.go` maps every `AuthService.Login()` error to 401 and returns raw error text, which can misclassify repository/database failures as credential failures and leak backend error details. | FIXED | VERIFIED_FIXED | 2525a38 | Also affects Task 7 |
| 2 | IMPORTANT | `internal/handler/token.go` `Delete()` maps all non-not-found failures to 403, so operational/database failures are mislabeled as authorization problems. | FIXED | VERIFIED_FIXED | 2525a38 | Also affects Task 7 |
| 3 | MINOR | `internal/service/token_test.go` only covers API token create/validate and does not exercise JWT/login or middleware branching. | FIXED | VERIFIED_FIXED | 2525a38 | Also affects Task 7 |
| 4 | IMPORTANT | `AuthService.Login()` still classified corrupted bcrypt hashes as invalid credentials instead of backend failure, so the original finding #1 was only partially fixed in the first patch. | FIXED | NEW_FINDING | pending | Also affects Task 7 |

#### Re-check Summary

- **Finding #1:** Verified fixed after separating invalid-credential handling from backend failures and removing raw internal error exposure from `AuthHandler.Login()`.
- **Finding #2:** Verified fixed after `TokenHandler.Delete()` now distinguishes not found, forbidden, and backend failures.
- **Finding #3:** Verified fixed after adding `auth_test.go` coverage for JWT login/parsing and invalid-credential behavior.
- **Fallback reason:** original reviewer unavailable as a reusable in-session reviewer, so a fresh reviewer was used for the re-check.
- **Verification evidence reviewed:** `go test ./internal/service` PASS, `go build ./...` PASS.

#### New Findings During Re-check

**Finding #4:** `AuthService.Login()` still treated corrupted bcrypt hashes as invalid credentials.
- **Status of prior finding:** original handler-level issue was partially fixed, but the service still collapsed one backend failure path into a credential error.
- **Action:** fixed immediately by only mapping `bcrypt.ErrMismatchedHashAndPassword` to `ErrInvalidCredentials` and adding a regression test for corrupted hashes.

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging
- Finding #4 → `docs/superpowers/debugging-log/2026-05-25-auth-test-shared-sqlite-memory.md`

---
