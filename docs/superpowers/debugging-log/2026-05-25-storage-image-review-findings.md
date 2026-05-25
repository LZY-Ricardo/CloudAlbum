# Storage and image review findings remediation

**Date:** 2026-05-25
**Triggered by:** REVIEW
**Context:** Task 23 from `docs/superpowers/plans/2026-05-25-cloudalbum.md`
**Status:** FIXED

## Symptom
Prior code-quality review found that local storage accepted traversal keys and obscured `os.ErrNotExist`, while image processing produced inconsistent thumbnail output and over-identified WebP/SVG inputs.

## Root Cause
1. `LocalStorage` used `filepath.Join(basePath, key)` directly for every operation, so keys could escape the storage root and `Get()` replaced filesystem not-found errors with a plain formatted error.
2. Thumbnail encoding reused source MIME-derived formats, which meant JPEG quality settings were silently ignored for PNG/GIF outputs and `AutoConvert` intent was not represented explicitly.
3. `DetectImageType` accepted any payload with `WEBP` at bytes 8-11 and treated generic XML declarations as SVG without confirming a `<svg>` root element.

## Investigation Path
1. Added focused storage and image tests for traversal, missing-file semantics, MIME detection, and thumbnail output behavior.
2. Ran `go test ./internal/storage ./internal/image` and confirmed each reported review finding reproduced as a concrete failing assertion.
3. Compared the failing tests with the current implementations and traced the issues to unchecked joined paths, lossy error wrapping, source-format thumbnail encoding, and permissive signature checks.
4. Implemented minimal fixes in the affected packages only and re-ran the focused tests plus `go build ./...`.

## Fix
- Added `resolvePath()` in `internal/storage/local.go` and routed save/get/exists/delete through it.
- Wrapped the underlying `os.Open` error in `Get()` so callers can still use `errors.Is(err, os.ErrNotExist)`.
- Replaced source-format thumbnail encoding with `thumbnailEncoding()` so output format is explicit and conservative.
- Tightened WebP detection to require both `RIFF` and `WEBP`, and tightened SVG detection to require an actual `<svg>` document root.
- Added regression tests in `internal/storage/local_test.go` and `internal/image/processor_test.go`.

**Commit:** `this commit`

## Verification
- `go test ./internal/storage ./internal/image`
- `go build ./...`

## Lessons
- Filesystem path validation must be shared by all read/write operations, not only save paths.
- MIME sniffers should prefer narrow positive identification over permissive heuristics, especially for text-based formats.

## Cross-References
- Triggered by: `docs/superpowers/review-log/2026-05-25-cloudalbum.md`
- Related debugging: None
