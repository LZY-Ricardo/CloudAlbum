# Core Security & Access Control — Review Log

### Review Cycle 1 — 2026-05-28T16:16:00+08:00

**Cycle ID:** RC-1
**Reviewer type:** CODE_QUALITY
**Reviewer:** subagent `ae55bb637677211c9`
**Scope:** Full implementation
**Re-check of:** —

#### Findings

| # | Severity | Description | Resolution | Re-check status | Commit | Cross-task? |
|---|----------|-------------|------------|-----------------|--------|-------------|
| 1 | HIGH | `public.go` 公开图片按 `Referer` 放行/拒绝，但成功响应仍返回 `Cache-Control: public, max-age=31536000` 且没有 `Vary: Referer`，共享缓存可绕过访问策略。 | FIXED | VERIFIED_FIXED | — | Task 8 |
| 2 | MEDIUM | `processor.go` 在 `strip_exif=true` 时重编码 JPEG 原图，但 `Hash` / `Size` 仍基于重编码前字节。 | FIXED | VERIFIED_FIXED | — | Task 9 |
| 3 | MEDIUM | `Tokens.tsx` 的过期输入没有做数值校验，非法输入会退化成 `null`，后端按“未传 expires_in”处理。 | FIXED | VERIFIED_FIXED | — | Task 4 |
| 4 | MEDIUM | `allow_no_expiry` 策略没有真实生效：`Create()` 省略 `expires_in` 时只看 `DefaultExpiresIn`，而 `Load()` 还强制把 `AllowNoExpiry` 设成 `true`。 | FIXED | VERIFIED_FIXED | — | Task 2 / Task 3 |
| 5 | LOW | `ratelimit` state map 不清理过期 key，长期运行下内存只增不减。 | FIXED | VERIFIED_FIXED | — | Task 5 |

#### Deferred / Rejected Notes
- None.

#### Related Debugging
- None.

---
