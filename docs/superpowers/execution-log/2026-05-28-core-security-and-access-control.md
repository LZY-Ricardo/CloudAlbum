# Core Security & Access Control — Execution Log

**Date:** 2026-05-28
**Plan:** `docs/superpowers/plans/2026-05-28-core-security-and-access-control.md`
**Review Config:** `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`

## Task 1: Review config and execution-log scaffold

**Execution**
- Created the feature review-config and execution-log scaffold.

**Verification**
- `ls docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md` → PASS

**Review (self-checklist)**
- Spec mapping: workflow prerequisite only.
- Interface consistency: N/A.
- Tests verify behavior: file existence check only.
- Smell scan: no placeholders in review-config.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Task-level review required only self-checklist; no escalation triggered.

**Debugging**
- `N/A`

## Task 2: Security config defaults

**Execution**
- Added `token`, `upload_rate_limit`, and `public_access` config sections to `internal/config/config.go`.
- Added default loading behavior and example values in `configs/config.yaml`.
- Added `internal/config/config_test.go` to lock default semantics.

**Verification**
- `go test ./internal/config -run TestLoadAppliesSecurityDefaults -v` → PASS

**Review (self-checklist)**
- Spec mapping: config model for token/upload/public access covered.
- Interface consistency: config struct names match the spec and later tasks.
- Tests verify behavior: default values asserted through `Load()`.
- Smell scan: no extra config surface added beyond the spec.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 3: Token expiry enforcement

**Execution**
- Extended `TokenService` to accept `config.Provider` and enforce expiry in `Validate()`.
- Added support for explicit `expires_in` and default expiry behavior.
- Updated `main.go` to build the new service signature.

**Verification**
- `go test ./internal/service -run 'TestTokenService(ValidateRejectsExpiredToken|CreateUsesDefaultExpiry|CreateAndValidate)' -v` → PASS

**Review (self-checklist)**
- Spec mapping: API Token expiry now actually blocks expired tokens.
- Interface consistency: `NewTokenService` / `Create` signatures align with handler task.
- Tests verify behavior: default expiry, explicit expiry, and no `last_used_at` update after expiry all covered.
- Smell scan: no extra token policy beyond the plan.
- Spec-stated boundaries covered: yes.
- Plan deviation check: one test fixture DSN was isolated by test name; no plan change required.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 4: Token handler and Tokens page

**Execution**
- Added `expires_in` request support to `internal/handler/token.go`.
- Added handler regression test for `expires_at` in the response.
- Updated `web/src/pages/Tokens.tsx` to accept optional expiry hours and render expiry status.

**Verification**
- `go test ./internal/handler -run TestTokenHandlerCreateAcceptsExpiresIn -v && cd web && npm run build` → PASS (Vite chunk-size warning only)

**Review (self-checklist)**
- Spec mapping: token creation/listing UI and API now surface expiry.
- Interface consistency: `expires_in` and `expires_at` names consistent with service/config tasks.
- Tests verify behavior: handler test covers API response path; build confirms TS surface stays valid.
- Smell scan: kept UI input simple (hours only), no extra date-picker complexity.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 5: In-memory upload limiter

**Execution**
- Added `internal/ratelimit/limiter.go` with fixed-window in-memory limiter.
- Added focused tests for blocking and window reset behavior.

**Verification**
- `go test ./internal/ratelimit -v` → PASS

**Review (self-checklist)**
- Spec mapping: single-process limiter implemented as planned.
- Interface consistency: `NewLimiter` / `Allow` used directly by handler task.
- Tests verify behavior: fixed-window block and reset covered.
- Smell scan: no persistence or distributed complexity added.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 6: Upload rate limit integration

**Execution**
- Wired limiter into `ImageHandler.Upload` and `UploadURL`.
- Added `token_id` to API-token auth context in `internal/middleware/auth.go`.
- Added `uploadLimitKey()` selection logic for JWT/token/user fallback.
- Injected limiter from `main.go`.

**Verification**
- `go test ./internal/handler -run TestImageUploadReturns429WhenRateLimited -v && go test ./internal/middleware ./internal/handler ./internal/service -count=1` → PASS

**Review (self-checklist)**
- Spec mapping: upload endpoints now return `429 rate_limit_exceeded` when over limit.
- Interface consistency: context keys and limiter API match config/service tasks.
- Tests verify behavior: handler regression covers 429 short-circuit before upload logic.
- Smell scan: avoided broader handler refactor not required for the task.
- Spec-stated boundaries covered: yes.
- Plan deviation check: test was adjusted to preload the limiter window instead of using `max=0`; runtime behavior unchanged.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 7: Public access helper

**Execution**
- Added `internal/security/public_access.go` with mode parsing, referer host extraction, and whitelist checks.
- Added helper tests covering all three modes.

**Verification**
- `go test ./internal/security -v` → PASS

**Review (self-checklist)**
- Spec mapping: `off` / `referer_whitelist` / `allow_empty_or_whitelist` all implemented.
- Interface consistency: helper API matches `PublicHandler` integration task.
- Tests verify behavior: empty referer, allowed host, and denied host covered.
- Smell scan: kept matching to exact hostname only, as planned.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 8: PublicHandler integration

**Execution**
- Injected `config.Provider` into `PublicHandler`.
- Enforced public access policy before reading image objects.
- Added a regression test for denied referer and updated existing tests to use the new constructor.

**Verification**
- `go test ./internal/handler -run 'TestPublicHandler(ImageServesStoredFileWithHeaders|ImageReturnsNotFoundWhenMissing|RejectsDisallowedReferer)' -v` → PASS

**Review (self-checklist)**
- Spec mapping: public image access now returns `403 public_access_forbidden` when policy denies the request.
- Interface consistency: provider wiring matches `main.go` and helper task.
- Tests verify behavior: allowed and denied paths both covered.
- Smell scan: existing content-type/cache headers preserved.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 9: strip_exif behavior

**Execution**
- Extended `ProcessResult` with `OriginalData`.
- Re-encode JPEG originals when `strip_exif=true`, keeping original bytes when false.
- Updated `ImageService.storeProcessedImage()` to persist `OriginalData` when provided.
- Added processor regression test and kept image-service dedup test green.

**Verification**
- `go test ./internal/image -run TestProcessorRespectsStripExifForOriginalJPEG -v && go test ./internal/service -run TestImageServiceUploadDeduplicatesByHash -v` → PASS

**Review (self-checklist)**
- Spec mapping: `image.strip_exif` now changes stored-original behavior in the upload path.
- Interface consistency: `OriginalData` is consumed only by image storage path, no unrelated surface added.
- Tests verify behavior: rewritten bytes vs original bytes covered.
- Smell scan: limited best-effort preserve behavior to JPEG only, matching current scope.
- Spec-stated boundaries covered: yes.
- Plan deviation check: narrowed best-effort preserve behavior to JPEG path rather than inventing cross-format rewriting.

**Review (applied config)**
- Self-checklist only; no external review at task level.

**Debugging**
- `N/A`

## Task 10: Full verification and closeout

**Execution**
- Ran full backend tests and frontend build.
- Updated execution log, completion summary, and `status.md` for sub-project 1.
- Deferred one external subagent review to feature closeout, per review-config.

**Verification**
- `go test ./... -count=1 && cd web && npm run build` → PASS (Vite chunk-size warning only)

**Review (self-checklist)**
- Spec mapping: all four major security gaps are covered by completed tasks.
- Interface consistency: config/service/handler/UI naming stayed aligned across the feature.
- Tests verify behavior: targeted regressions plus full suite/build both green.
- Smell scan: no extra security UI/editor scope added beyond plan.
- Spec-stated boundaries covered: yes.
- Plan deviation check: none so far.

**Review (applied config)**
- Feature-level external code review still pending and will run once for the whole feature.

**Debugging**
- `N/A`
