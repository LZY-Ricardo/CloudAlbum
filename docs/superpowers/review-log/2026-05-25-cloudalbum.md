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

### Review Cycle 7 — 2026-05-25 20:14 CST

**Cycle ID:** RC-7
**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent
**Scope:** Task 7 Image + Album API + Router
**Preceded by:** Task 7 implementation and verification

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `internal/service/image.go` remote URL upload path has duplicate-content handling and storage-key reuse semantics that need explicit consistency with local upload behavior. | FIXED | OPEN | pending | Also affects Task 10 |
| 2 | IMPORTANT | `internal/service/image.go` and `internal/handler/image.go` need to distinguish omitted `album_id` from explicit `null` during image update, otherwise album assignment can be cleared unexpectedly. | FIXED | OPEN | pending | Also affects Task 10 |
| 3 | IMPORTANT | `internal/service/album.go` and `internal/handler/album.go` overwrite album fields when `name` or `cover_image_id` are omitted, causing partial updates to clear existing values. | FIXED | OPEN | pending | Also affects Task 10 |
| 4 | MINOR | `internal/router/router.go` includes a redundant group-level `RequireScope` on image routes, increasing fragility for future route additions and middleware ordering changes. | FIXED | OPEN | pending | Also affects Task 10 |
| 5 | IMPORTANT | The Task 7 behavior changes need minimal regression tests to lock in update semantics and avoid future regressions. | FIXED | OPEN | pending | Also affects Task 10 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 8 — 2026-05-26 00:07 CST

**Cycle ID:** RC-8
**Reviewer type:** CODE_QUALITY
**Reviewer:** self-review with prior reviewer findings recap
**Scope:** Task 7 Image + Album API + Router
**Preceded by:** Review Cycle 7
**Re-check of:** Review Cycle 7
**Original reviewer:** subagent
**Re-check reviewer:** implementer with explicit checklist against prior findings

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `internal/service/image.go` remote URL upload path has duplicate-content handling and storage-key reuse semantics that need explicit consistency with local upload behavior. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 2 | IMPORTANT | `internal/service/image.go` and `internal/handler/image.go` need to distinguish omitted `album_id` from explicit `null` during image update, otherwise album assignment can be cleared unexpectedly. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 3 | IMPORTANT | `internal/service/album.go` and `internal/handler/album.go` overwrite album fields when `name` or `cover_image_id` are omitted, causing partial updates to clear existing values. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 4 | MINOR | `internal/router/router.go` includes a redundant group-level `RequireScope` on image routes, increasing fragility for future route additions and middleware ordering changes. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 5 | IMPORTANT | The Task 7 behavior changes need minimal regression tests to lock in update semantics and avoid future regressions. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 6 | MINOR | `cmd/server/main.go` still called the old `router.Setup` signature after router cleanup, and `internal/router/router.go` retained dead `_ = imageSvc` / `_ = albumSvc` lines until follow-up cleanup. | FIXED | NEW_FINDING | pending | Also affects Task 10 |

#### Re-check Summary

- **Finding #1:** Verified fixed by keeping duplicate-content handling centralized in `storeProcessedImage()` for both multipart and remote URL uploads.
- **Finding #2:** Verified fixed by changing image update handling to use raw JSON maps so omitted `album_id` and explicit `null` are distinguishable.
- **Finding #3:** Verified fixed by changing album update handling to preserve omitted fields instead of clearing them.
- **Finding #4:** Verified fixed by removing the redundant image group-level `RequireScope` gate.
- **Finding #5:** Verified fixed by adding regression tests for image/album update semantics in service tests.
- **Verification evidence reviewed:** `go test ./internal/service ./internal/handler` PASS, `go build ./...` PASS.

#### New Findings During Re-check

**Finding #6:** Router cleanup changed `Setup` signature, but `cmd/server/main.go` and `internal/router/router.go` needed a final cleanup pass to remove the old call/unused placeholders.
- **Status of prior finding:** prior findings were fixed, but the cleanup introduced a small compile-time follow-up.
- **Action:** fixed immediately and re-verified with tests and full build.

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 9 — 2026-05-26 14:04 CST

**Cycle ID:** RC-9
**Reviewer type:** SPEC_COMPLIANCE
**Reviewer:** subagent
**Scope:** Task 8 React Setup + Login Page
**Preceded by:** Task 8 implementation and verification

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `web/src/api/client.ts` clears `localStorage` on 401 but does not update Zustand auth state, so if no redirect occurs the app state can still think the user is logged in. | FIXED | OPEN | pending | Also affects Task 9, 10 |

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 10 — 2026-05-26 14:07 CST

**Cycle ID:** RC-10
**Reviewer type:** SPEC_COMPLIANCE
**Reviewer:** self-review with prior reviewer finding recap
**Scope:** Task 8 React Setup + Login Page
**Preceded by:** Review Cycle 9
**Re-check of:** Review Cycle 9
**Original reviewer:** subagent
**Re-check reviewer:** implementer with explicit checklist against prior finding

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `web/src/api/client.ts` clears `localStorage` on 401 but does not update Zustand auth state, so if no redirect occurs the app state can still think the user is logged in. | FIXED | VERIFIED_FIXED | pending | Also affects Task 9, 10 |

#### Re-check Summary

- **Finding #1:** Verified fixed by adding a `reset()` action to the auth store and invoking it from the axios 401 response interceptor before any redirect logic.
- **Verification evidence reviewed:** `cd web && npm run build` PASS, `cd web && npm run dev -- --host 127.0.0.1` startup PASS.

#### Deferred Items

None.

#### Rejected Items

None.

#### Related Debugging

None.

---

### Review Cycle 11 — 2026-05-26 14:18 CST

**Cycle ID:** RC-11
**Reviewer type:** CODE_QUALITY
**Reviewer:** self-review with prior reviewer finding recap
**Scope:** Task 9 Layout + Upload Page
**Preceded by:** Task 9 code review (external reviewer report)
**Re-check of:** Task 9 code review on commit `b886988`
**Original reviewer:** review-only subagent
**Re-check reviewer:** implementer with explicit checklist against prior findings

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | IMPORTANT | `Upload.tsx` used `useMemo` to fetch album data, which is an invalid side-effect hook pattern. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 2 | IMPORTANT | Multi-file upload silently dropped per-file backend failures and could make partial or total failure look like success. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 3 | IMPORTANT | Clipboard paste support was advertised in UI copy but only worked when paste reached a focused inner element. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 4 | IMPORTANT | Dropzone exposed `role="button"`/`tabIndex` without keyboard activation support. | FIXED | VERIFIED_FIXED | pending | Also affects Task 10 |
| 5 | MINOR | Drag enter/leave state can still flicker when moving across child nodes inside the dropzone. | DEFERRED | DEFERRED | — | Target Task 10 |

#### Re-check Summary

- **Finding #1:** Verified fixed by replacing `useMemo` with `useEffect` for album fetching.
- **Finding #2:** Verified fixed by preserving and rendering per-file failure items from the backend `results` array.
- **Finding #3:** Verified fixed by adding a page-level `paste` listener via `useEffect`.
- **Finding #4:** Verified fixed by adding keyboard activation for Enter/Space on the dropzone.
- **Finding #5:** Deferred intentionally because it is a UX polish issue rather than a current correctness break, and Task 10 will revisit upload-area interactions while expanding the image-management surface.
- **Verification evidence reviewed:** `cd web && npm run build` PASS, `cd web && npm run dev -- --host 127.0.0.1` startup PASS.

#### Deferred Items

**Finding #5:** Drag enter/leave state can still flicker when moving across child nodes inside the dropzone.
- **Reason:** Current upload behavior is functional; this is interaction polish rather than a blocking correctness issue.
- **Impact:** Hover highlight may flicker while dragging across nested content, but upload itself still works.
- **Prerequisite:** Address together with Task 10 upload/image-management interaction polish to avoid churn.

#### Rejected Items

None.

#### Related Debugging

None.

---
