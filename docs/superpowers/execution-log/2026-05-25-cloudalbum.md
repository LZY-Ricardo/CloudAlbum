# Execution Log: CloudAlbum

**Plan:** `docs/superpowers/plans/2026-05-25-cloudalbum.md`
**Spec:** `docs/superpowers/specs/2026-05-25-cloudalbum-design.md`
**Started:** 2026-05-25

---

### Task 5: Database Init + Repository Layer — DONE

**Status:** DONE

**Completed at:** 2026-05-25 17:14 CST

**What was implemented:**
- Added repository layer files for users, images, albums, and API tokens under `internal/repository/`.
- Updated `cmd/server/main.go` to load config, initialize sqlite via `initDB`, run `AutoMigrate` for `User`, `Image`, `Album`, and `APIToken`, initialize local storage via `initStorage`, and print startup information including the resolved local storage path.
- Added parent-directory creation for the sqlite DSN before opening the database so default local startup succeeds from a clean checkout.
- Adjusted token last-used updates to use `time.Now()` instead of a database-specific `NOW()` expression so the default sqlite backend remains compatible.

**Plan vs. Reality:**
- Planned: Wire database initialization and local storage startup, with repository files matching the approved Task 5 structure.
- Actual: Implemented the planned files and startup wiring, plus a minimal sqlite compatibility fix in `TokenRepository.UpdateLastUsed` uncovered during self-review.
- Reason: The original repository method used a database-specific timestamp expression that would not be portable to the sqlite-only scope of this task.

**Decisions made:**
- Keep `initDB` limited to sqlite while explicitly returning unsupported-driver errors for future database backends.
- Create the sqlite parent directory in startup code rather than relying on the database driver to create nested paths implicitly.
- Print the resolved local storage root only when the configured storage backend is `local`, keeping the startup log backend-aware.

**Commits:** `c038667`

**Related debugging:**
- → `docs/superpowers/debugging-log/2026-05-25-task5-token-last-used-build-fix.md`

---

### Task 23: Resolve Task 3-4 review findings — DONE

**Status:** DONE

**Completed at:** 2026-05-25 16:53 CST

**What was implemented:**
- Hardened `LocalStorage` path handling by validating keys before save/get/exists/delete, rejecting absolute and escaping paths.
- Preserved `os.ErrNotExist` semantics in `Get()` by wrapping the underlying open error.
- Made thumbnail encoding deterministic via config-driven output selection, with conservative JPEG fallback for `AutoConvert=webp` so quality is consistently applied.
- Tightened image type detection for WebP (`RIFF` + `WEBP`) and SVG (actual `<svg>` root instead of generic XML).
- Added regression tests covering storage traversal, not-found behavior, MIME detection, and thumbnail encoding behavior.

**Decisions made:**
- Treat `AutoConvert=webp` as a request for deterministic JPEG thumbnails because the current codebase does not include a safe built-in WebP encoder path, and JPEG is the only reviewed format here with reliable quality control.
- Apply the same storage-key validation to all filesystem operations so read/delete paths cannot bypass write-time protections.

**Commits:** `this commit` `fix: harden storage path handling and thumbnail detection`

**Related debugging:**
- → `docs/superpowers/debugging-log/2026-05-25-storage-image-review-findings.md`

---

#### Verification — Task 23

**Timestamp:** 2026-05-25 16:53 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Focused tests | `go test ./internal/storage ./internal/image` | PASS | Regression tests reproduced the findings before the fix and passed after the fix. |
| Build | `go build ./...` | PASS | Full repository build succeeded after formatting. |

**Uncovered areas:**
- No end-to-end storage consumer flows exist yet in the repository, so verification was limited to targeted package tests and a full compile.

---

#### Verification — Task 5

**Timestamp:** 2026-05-25 17:00 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Build | `go build ./...` | PASS | Full repository build succeeded after adding repository layer and initialization wiring. |
| Startup smoke test | `go run cmd/server/main.go` | PASS | Printed startup info with sqlite DSN and resolved local storage path after DB/storage initialization. |

**Uncovered areas:**
- No repository behavior tests were added in this task; verification covered compilation and startup initialization only.
- HTTP handlers, services, and router wiring remain out of scope for Task 5 and were not exercised.

**Related debugging:**
- `docs/superpowers/debugging-log/2026-05-25-task5-token-last-used-build-fix.md`

---

#### Verification — Task 5 (post-fix)

**Timestamp:** 2026-05-25 17:14 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Format + build + startup smoke test | `gofmt -w /Users/zyb/workspace/person/CloudAlbum/internal/repository/token.go && go build ./... && go run /Users/zyb/workspace/person/CloudAlbum/cmd/server/main.go` | PASS | Re-verified after replacing sqlite-incompatible timestamp logic with `time.Now()` in `TokenRepository.UpdateLastUsed`. |

**Uncovered areas:**
- No dedicated repository tests were added in this task; verification remained at compile and initialization smoke-test level.
- HTTP handlers, services, and router wiring remain out of scope for Task 5 and were not exercised.

**Related debugging:**
- `docs/superpowers/debugging-log/2026-05-25-task5-token-last-used-build-fix.md`

---

#### Verification — Task 28

**Timestamp:** 2026-05-25 17:27 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Repository tests | `go test ./internal/repository` | PASS | Verified deleted-only listing semantics, restore behavior, empty aggregate handling, and default pagination guards. |
| Build | `go build ./...` | PASS | Full repository build succeeded after sqlite DSN and repository fixes. |

**Uncovered areas:**
- Task 28 did not exercise HTTP-layer consumers yet; repository behavior is covered at package level only.
- Album deletion policy and broader repository contracts remain intentionally out of scope for this remediation.

---

#### Verification — Task 6

**Timestamp:** 2026-05-25 17:40 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Focused service test | `go test ./internal/service` | PASS | Verified API token creation/validation flow, raw token prefixing, hashed persistence, and last-used timestamp updates. |
| Build | `go build ./...` | PASS | Full repository build succeeded after adding auth services, middleware, and handlers. |

**Uncovered areas:**
- No HTTP-layer handler or middleware integration tests were added in this task; verification remained at focused service coverage plus full compile.
- Router wiring and end-to-end login/token request flows are intentionally deferred to Task 7.

---

#### Verification — Task 6 (post-review fixes)

**Timestamp:** 2026-05-25 19:42 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Service tests | `go test ./internal/service` | PASS | Verified JWT login/parsing and API token create/validate after fixing handler error mapping and test isolation. |
| Build | `go build ./...` | PASS | Full repository build succeeded after Task 6 review fixes. |

**Uncovered areas:**
- No HTTP handler or middleware integration tests were added yet; verification remains at service-level tests plus full compile.
- Router wiring and end-to-end auth request flows remain intentionally deferred to Task 7.

**Related debugging:**
- `docs/superpowers/debugging-log/2026-05-25-auth-test-shared-sqlite-memory.md`

---

#### Verification — Task 7

**Timestamp:** 2026-05-25 20:05 CST

| Check | Command | Result | Notes |
|-------|---------|--------|-------|
| Task 7 package tests | `go test ./internal/service ./internal/handler` | PASS | Verified new image/album service behavior and public image serving headers/not-found behavior. |
| Build | `go build ./...` | PASS | Full repository build succeeded after adding image/album APIs, router setup, and server wiring. |
| Startup smoke test | `python3` controlled startup script running `go run ./cmd/server/main.go` on temporary port `18080` plus `POST /api/v1/auth/login` | PASS | Local port `8080` was already occupied by unrelated processes, so verification used a temporary config override (`18080`) and restored `configs/config.yaml` immediately afterward. Login returned HTTP 200 with a JWT token. |

**Uncovered areas:**
- No end-to-end multipart upload HTTP test was added; upload flows are covered at service level plus compile/startup verification only.
- Public thumbnail route behavior beyond storage lookup and headers was not exercised with generated thumbnail files in an integration test.

---
