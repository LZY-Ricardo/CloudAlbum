# Core Security & Access Control Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans by default for this plan. Use superpowers:subagent-driven-development only if a task is truly independent and the isolation cost is worth it. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 decomposition 子项目 1「核心安全与访问控制补全」—— 让 API Token 过期校验、上传限流、公开图片访问策略、EXIF/隐私元数据处理在现有产品模型上真实生效。

**Architecture:** 复用现有 `config.Provider` 作为运行时配置入口，不重做上传或分享模型，而是在既有认证、上传、公开访问、图片处理链路上分别增加安全策略层。Token 过期继续使用现有 `APIToken.ExpiresAt` 字段；上传限流采用单进程内存状态；公开访问在 `PublicHandler` 内按配置模式做 Referer host 判定；EXIF 行为在 `Processor` 内与 `image.strip_exif` 一致。

**Tech Stack:** Go 1.x、Gin、GORM (SQLite)、React 18 + TypeScript + Vite、现有 `config.Provider` / `imaging` 图像处理链路。

**Spec:** `docs/superpowers/specs/2026-05-28-core-security-and-access-control-design.md`

---

## File Structure

### 新建后端文件

| 路径 | 责任 |
|---|---|
| `internal/ratelimit/limiter.go` | 进程内上传限流器；按 key 维护固定窗口状态 |
| `internal/ratelimit/limiter_test.go` | 限流器窗口 / 重置 / 并发基本行为测试 |
| `internal/security/public_access.go` | 公开访问模式、Referer host 解析与命中判断 |
| `internal/security/public_access_test.go` | `off` / `referer_whitelist` / `allow_empty_or_whitelist` 行为测试 |

### 修改后端文件

| 路径 | 改动概要 |
|---|---|
| `internal/config/config.go` | 新增 token / upload rate limit / public access 配置结构与默认值 |
| `configs/config.yaml` | 补示例默认配置 |
| `internal/service/token.go` | `Create` 支持可选 `expires_in`；`Validate` 真正拒绝过期 token |
| `internal/service/token_test.go` | token 过期 / `last_used_at` / 默认过期策略测试 |
| `internal/handler/token.go` | `POST /tokens` 入参加 `expires_in`；响应继续返回 `expires_at` |
| `internal/middleware/auth.go` | API Token 路径把 `token_id` 写入 context，供上传限流使用 |
| `internal/handler/image.go` | 上传前接入限流判定；超限统一返回 `429 rate_limit_exceeded` |
| `internal/handler/public.go` | 公开访问模式判定；拒绝时返回 `403 public_access_forbidden` |
| `internal/handler/public_test.go` | 增加公开访问策略相关测试 |
| `internal/image/processor.go` | `strip_exif=true` 时真正移除原图 EXIF；`false` 时 best effort 保留 |
| `internal/image/processor_test.go` | EXIF 行为和方向保持测试 |
| `main.go` | 构造 limiter / public access 依赖并注入 handler |

### 修改前端文件

| 路径 | 改动概要 |
|---|---|
| `web/src/pages/Tokens.tsx` | 创建表单增加可选过期时间；列表展示 `expires_at` 与状态 |

### 文档文件

| 路径 | 责任 |
|---|---|
| `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md` | 本功能执行/评审策略 |
| `docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md` | 每个任务一段 merged execution-log |
| `docs/superpowers/completion/2026-05-28-core-security-and-access-control-summary.md` | 功能完成总结 |

---

## Task Map

- **Task 1–2**：配置模型与 review-config / execution-log 骨架
- **Task 3–4**：API Token 过期策略与 Token 管理页输入输出
- **Task 5–6**：上传限流器与上传 handler 接入
- **Task 7–8**：公开图片访问策略与 `PublicHandler` 接入
- **Task 9**：EXIF 行为真实落地
- **Task 10**：全量回归、文档收口与 feature-level review 前置验证

---

### Task 1: 写 review-config 与 execution-log 骨架

**Files:**
- Create: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
- Create: `docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

This task is documentation-only. Create the files with the exact expected headings but deliberately leave them absent before creation so the “failure” is the missing file state documented in the execution log.

- [ ] **Step 2: Run the test to verify it fails**

Run: `ls docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md`
Expected: `No such file or directory` for both paths.

- [ ] **Step 3: Implement the minimal behavior**

Create `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md` with exactly:

```md
# Core Security & Access Control Review Config

- **Execution mode:** inline (`superpowers:executing-plans` default)
- **Task-level review:** spec-only self-checklist in every task; no external review by default
- **Feature-level review:** spec + code
- **Review executor:** hybrid (main session self-checklist + external reviewer at feature closeout)
- **Hard rules:**
  - self-checklist always runs
  - TDD failing-test step cannot be skipped
  - any plan deviation triggers one escalation review
  - feature closeout must include at least one external review dimension
```

Create `docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md` with exactly:

```md
# Core Security & Access Control — Execution Log

**Date:** 2026-05-28
**Plan:** `docs/superpowers/plans/2026-05-28-core-security-and-access-control.md`
**Review Config:** `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
```

- [ ] **Step 4: Run verification**

Run: `ls docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md`
Expected: both files exist.

- [ ] **Step 5: Write the merged execution-log block**

Append one task block to:
`docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md`

The block must contain:

```markdown
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
```

---

### Task 2: 新增安全配置结构与默认值

**Files:**
- Modify: `internal/config/config.go`
- Modify: `configs/config.yaml`
- Test: `internal/config/config_test.go` (create if missing)

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_test.go` with:

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAppliesSecurityDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(`server: {}`), 0644); err != nil {
		t.Fatalf("WriteFile(): %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load(): %v", err)
	}

	if cfg.Token.DefaultExpiresIn != 7*24*time.Hour {
		t.Fatalf("DefaultExpiresIn = %v", cfg.Token.DefaultExpiresIn)
	}
	if cfg.Token.AllowNoExpiry != true {
		t.Fatalf("AllowNoExpiry = %v, want true", cfg.Token.AllowNoExpiry)
	}
	if cfg.UploadRateLimit.Window != time.Minute {
		t.Fatalf("Window = %v, want 1m", cfg.UploadRateLimit.Window)
	}
	if cfg.UploadRateLimit.MaxRequests != 20 {
		t.Fatalf("MaxRequests = %d, want 20", cfg.UploadRateLimit.MaxRequests)
	}
	if cfg.PublicAccess.Mode != "off" {
		t.Fatalf("Mode = %q, want off", cfg.PublicAccess.Mode)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/config -run TestLoadAppliesSecurityDefaults -v`
Expected: FAIL with unknown config fields / missing test file.

- [ ] **Step 3: Implement the minimal behavior**

In `internal/config/config.go`:

```go
type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Database        DatabaseConfig        `yaml:"database"`
	Storage         StorageConfig         `yaml:"storage"`
	Image           ImageConfig           `yaml:"image"`
	Auth            AuthConfig            `yaml:"auth"`
	Token           TokenPolicyConfig     `yaml:"token"`
	UploadRateLimit UploadRateLimitConfig `yaml:"upload_rate_limit"`
	PublicAccess    PublicAccessConfig    `yaml:"public_access"`
}

type TokenPolicyConfig struct {
	AllowNoExpiry    bool          `yaml:"allow_no_expiry"`
	DefaultExpiresIn time.Duration `yaml:"default_expires_in"`
}

type UploadRateLimitConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Window      time.Duration `yaml:"window"`
	MaxRequests int           `yaml:"max_requests"`
}

type PublicAccessConfig struct {
	Mode               string   `yaml:"mode"`
	AllowedRefererHosts []string `yaml:"allowed_referer_hosts"`
}
```

Set defaults in `Load()`:

```go
if cfg.Token.DefaultExpiresIn == 0 {
	cfg.Token.DefaultExpiresIn = 7 * 24 * time.Hour
}
cfg.Token.AllowNoExpiry = true // keep zero-value default compatible
if cfg.UploadRateLimit.Window == 0 {
	cfg.UploadRateLimit.Window = time.Minute
}
if cfg.UploadRateLimit.MaxRequests == 0 {
	cfg.UploadRateLimit.MaxRequests = 20
}
if cfg.PublicAccess.Mode == "" {
	cfg.PublicAccess.Mode = "off"
}
```

Update `configs/config.yaml` with explicit example values:

```yaml
token:
  allow_no_expiry: true
  default_expires_in: 168h

upload_rate_limit:
  enabled: false
  window: 1m
  max_requests: 20

public_access:
  mode: off
  allowed_referer_hosts: []
```

- [ ] **Step 4: Run verification**

Run: `go test ./internal/config -run TestLoadAppliesSecurityDefaults -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 2 block mirroring actual commands and outcomes.

---

### Task 3: 让 TokenService 真正支持过期策略

**Files:**
- Modify: `internal/service/token.go`
- Modify: `internal/service/token_test.go`
- Modify: `main.go`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Extend `internal/service/token_test.go` with:

```go
func TestTokenServiceValidateRejectsExpiredToken(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	provider := testTokenProvider()
	svc := NewTokenService(tokenRepo, provider)

	created, rawToken, err := svc.Create(7, "cli", "upload", 1)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ExpiresAt == nil {
		t.Fatal("Create() should set expires_at when expires_in provided")
	}

	validated, err := svc.Validate(rawToken)
	if err != nil {
		t.Fatalf("Validate() immediate error = %v", err)
	}
	if validated.LastUsedAt == nil {
		t.Fatal("LastUsedAt should be updated before expiry")
	}

	created.ExpiresAt = ptrTime(created.CreatedAt.Add(-time.Minute))
	if err := db.Model(&model.APIToken{}).Where("id = ?", created.ID).Update("expires_at", created.ExpiresAt).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}
	before := validated.LastUsedAt
	_, err = svc.Validate(rawToken)
	if err == nil {
		t.Fatal("Validate() error = nil, want invalid token")
	}

	reloaded, err := tokenRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if reloaded.LastUsedAt == nil || !reloaded.LastUsedAt.Equal(*before) {
		t.Fatalf("LastUsedAt changed on expired token: before=%v after=%v", before, reloaded.LastUsedAt)
	}
}

func TestTokenServiceCreateUsesDefaultExpiry(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	provider := testTokenProvider()
	svc := NewTokenService(tokenRepo, provider)

	created, _, err := svc.Create(7, "cli", "upload", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ExpiresAt == nil {
		t.Fatal("default expiry should be applied")
	}
}
```

Also add helpers in the same file:

```go
func testTokenProvider() *config.Provider {
	base := config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: true, DefaultExpiresIn: 7 * 24 * time.Hour}}
	return config.NewProvider(base, config.Overrides{})
}

func ptrTime(v time.Time) *time.Time { return &v }
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/service -run 'TestTokenService(ValidateRejectsExpiredToken|CreateUsesDefaultExpiry)' -v`
Expected: FAIL because `NewTokenService` and `Create` signatures do not yet support provider / expiry.

- [ ] **Step 3: Implement the minimal behavior**

Change `internal/service/token.go` to:

```go
type TokenService struct {
	tokenRepo *repository.TokenRepository
	provider  *config.Provider
}

func NewTokenService(tokenRepo *repository.TokenRepository, provider *config.Provider) *TokenService
func (s *TokenService) Create(userID uint, name, scope string, expiresIn *int64) (*model.APIToken, string, error)
```

Implementation rules:

- `expiresIn == nil`:
  - if `provider.Get().Token.AllowNoExpiry == true`, apply `DefaultExpiresIn` when > 0; if later product wants “nil means never expire”, that would require a different API contract, so do **not** invent it here.
- `expiresIn != nil && *expiresIn <= 0` → return `errors.New("invalid expires_in")`
- when an expiry is chosen, set `token.ExpiresAt = now.Add(time.Duration(*expiresIn) * time.Second)` or `DefaultExpiresIn`
- in `Validate(rawToken string)`:
  - trim prefix / hash lookup stays the same
  - before `UpdateLastUsed`, add:

```go
if token.ExpiresAt != nil && time.Now().After(*token.ExpiresAt) {
	return nil, errors.New("invalid token")
}
```

Update `main.go` to construct `tokenSvc := service.NewTokenService(tokenRepo, provider)`.

- [ ] **Step 4: Run verification**

Run: `go test ./internal/service -run 'TestTokenService(ValidateRejectsExpiredToken|CreateUsesDefaultExpiry|CreateAndValidate)' -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 3 block with exact command outcomes.

---

### Task 4: 扩展 Token handler 与 Token 管理页

**Files:**
- Modify: `internal/handler/token.go`
- Create: `internal/handler/token_test.go`
- Modify: `web/src/pages/Tokens.tsx`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Create `internal/handler/token_test.go` with:

```go
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenHandlerCreateAcceptsExpiresIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, userID := newTokenHandlerTestDB(t)
	repo := repository.NewTokenRepository(db)
	provider := config.NewProvider(config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: true, DefaultExpiresIn: 7 * 24 * time.Hour}}, config.Overrides{})
	h := NewTokenHandler(service.NewTokenService(repo, provider))

	rec := httptest.NewRecorder()
	ctx, engine := gin.CreateTestContext(rec)
	ctx.Set("user_id", userID)
	engine.POST("/tokens", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.Create(c)
	})

	body, _ := json.Marshal(gin.H{"name": "cli", "scope": "upload", "expires_in": 3600})
	req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)
	_ = ctx

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"expires_at"`)) {
		t.Fatalf("response missing expires_at: %s", rec.Body.String())
	}
}

func newTokenHandlerTestDB(t *testing.T) (*gorm.DB, uint) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil { t.Fatalf("open db: %v", err) }
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil { t.Fatalf("migrate: %v", err) }
	user := &model.User{Username: "admin", PasswordHash: "hash", Role: "admin"}
	if err := db.Create(user).Error; err != nil { t.Fatalf("create user: %v", err) }
	return db, user.ID
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/handler -run TestTokenHandlerCreateAcceptsExpiresIn -v`
Expected: FAIL because handler request struct lacks `expires_in` and TokenService signature mismatches.

- [ ] **Step 3: Implement the minimal behavior**

In `internal/handler/token.go`:

```go
type CreateTokenRequest struct {
	Name      string `json:"name" binding:"required"`
	Scope     string `json:"scope" binding:"required,oneof=read upload full"`
	ExpiresIn *int64 `json:"expires_in,omitempty"`
}
```

Call:

```go
token, rawToken, err := h.tokenSvc.Create(c.GetUint("user_id"), req.Name, req.Scope, req.ExpiresIn)
```

When `err.Error() == "invalid expires_in"`, return `400` with `{"error":"invalid_request"}`.

In `web/src/pages/Tokens.tsx`:

- extend `TokenItem` with `expires_at?: string | null`
- add local state `expiresInHours`
- submit `expires_in` only when the input is non-empty
- render token status line:

```tsx
const statusLabel = !token.expires_at
  ? '永不过期'
  : new Date(token.expires_at).getTime() <= Date.now()
    ? '已过期'
    : `过期于 ${new Date(token.expires_at).toLocaleString()}`
```

Use a simple numeric input in hours to avoid pulling in new date-picker logic.

- [ ] **Step 4: Run verification**

Run: `go test ./internal/handler -run TestTokenHandlerCreateAcceptsExpiresIn -v && cd web && npm run build`
Expected: handler test PASS; web build PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 4 block with exact outcomes.

---

### Task 5: 新增进程内上传限流器

**Files:**
- Create: `internal/ratelimit/limiter.go`
- Create: `internal/ratelimit/limiter_test.go`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Create `internal/ratelimit/limiter_test.go` with:

```go
package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterBlocksAfterMaxRequests(t *testing.T) {
	l := NewLimiter(true, 50*time.Millisecond, 2)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("second allow: %v", err)
	}
	if err := l.Allow("u:1"); err == nil {
		t.Fatal("third allow should be rate limited")
	}
}

func TestLimiterResetsAfterWindow(t *testing.T) {
	l := NewLimiter(true, 20*time.Millisecond, 1)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("first allow: %v", err)
	}
	if err := l.Allow("u:1"); err == nil {
		t.Fatal("second allow should be blocked")
	}
	time.Sleep(25 * time.Millisecond)
	if err := l.Allow("u:1"); err != nil {
		t.Fatalf("allow after reset: %v", err)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/ratelimit -v`
Expected: FAIL because package does not exist.

- [ ] **Step 3: Implement the minimal behavior**

Create `internal/ratelimit/limiter.go` with:

```go
package ratelimit

import (
	"errors"
	"sync"
	"time"
)

var ErrRateLimited = errors.New("rate_limit_exceeded")

type bucket struct {
	windowStart time.Time
	count       int
}

type Limiter struct {
	enabled bool
	window  time.Duration
	max     int
	mu      sync.Mutex
	state   map[string]bucket
}

func NewLimiter(enabled bool, window time.Duration, max int) *Limiter
func (l *Limiter) Allow(key string) error
```

Implementation rules:

- disabled limiter always returns nil
- empty key returns nil (handler will already scope correctly)
- if current time is after `windowStart + window`, reset count to 1 and move windowStart
- else increment and compare against `max`
- when exceeded, return `ErrRateLimited`

- [ ] **Step 4: Run verification**

Run: `go test ./internal/ratelimit -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 5 block with actual command output.

---

### Task 6: 在上传 handler 接入限流

**Files:**
- Modify: `internal/handler/image.go`
- Modify: `internal/middleware/auth.go`
- Modify: `main.go`
- Test: `internal/handler/image_test.go` (create if missing)

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Create `internal/handler/image_test.go` with:

```go
package handler

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudalbum/internal/ratelimit"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

type stubImageService struct{}
func (s *stubImageService) Upload(userID uint, file *multipart.FileHeader, albumID *uint) (interface{}, error) { return gin.H{"id": 1}, nil }

func TestImageUploadReturns429WhenRateLimited(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	limiter := ratelimit.NewLimiter(true, time.Minute, 0)
	h := &ImageHandler{imageSvc: service.NewImageService(nil, nil, nil, nil), uploadLimiter: limiter}
	engine.POST("/images", func(c *gin.Context) {
		c.Set("user_id", uint(1))
		c.Set("auth_type", "jwt")
		h.Upload(c)
	})

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("files", "a.jpg")
	part.Write([]byte("abc"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/images", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
}
```

Before implementing, simplify the handler dependency if needed by introducing an uploader interface instead of using the concrete `*service.ImageService` in tests.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/handler -run TestImageUploadReturns429WhenRateLimited -v`
Expected: FAIL because `ImageHandler` has no limiter field / no 429 branch.

- [ ] **Step 3: Implement the minimal behavior**

Refactor `internal/handler/image.go` just enough to make upload testable:

```go
type imageUploader interface {
	Upload(userID uint, file *multipart.FileHeader, albumID *uint) (*model.Image, error)
	UploadFromURL(userID uint, rawURL string, albumID *uint) (*model.Image, error)
	// keep the existing methods used by the handler
}

type ImageHandler struct {
	imageSvc       imageUploader
	uploadLimiter  *ratelimit.Limiter
}
```

Add helper methods:

```go
func (h *ImageHandler) limitUpload(c *gin.Context) bool
func uploadLimitKey(c *gin.Context) string
```

Rules:

- JWT: key = `jwt:user:<user_id>`
- API Token with context token id: key = `token:<token_id>`
- else fallback `user:<user_id>`
- if `Allow` returns `ratelimit.ErrRateLimited`, respond `429` with `{"error":"rate_limit_exceeded"}` and stop
- call `limitUpload()` at the start of both `Upload()` and `UploadURL()`

In `internal/middleware/auth.go`, after successful API Token validation add:

```go
c.Set("token_id", apiToken.ID)
```

In `main.go`, build:

```go
uploadLimiter := ratelimit.NewLimiter(
	provider.Get().UploadRateLimit.Enabled,
	provider.Get().UploadRateLimit.Window,
	provider.Get().UploadRateLimit.MaxRequests,
)
imageHandler := handler.NewImageHandler(imageSvc, uploadLimiter)
```

Update constructor signature accordingly.

- [ ] **Step 4: Run verification**

Run: `go test ./internal/handler -run TestImageUploadReturns429WhenRateLimited -v && go test ./internal/middleware ./internal/handler ./internal/service -count=1`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 6 block.

---

### Task 7: 新增公开访问策略判断模块

**Files:**
- Create: `internal/security/public_access.go`
- Create: `internal/security/public_access_test.go`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Create `internal/security/public_access_test.go` with:

```go
package security

import "testing"

func TestAllowPublicAccessModes(t *testing.T) {
	if !AllowPublicAccess("off", nil, "") {
		t.Fatal("off mode should allow")
	}
	if AllowPublicAccess("referer_whitelist", []string{"example.com"}, "") {
		t.Fatal("whitelist mode should reject empty referer")
	}
	if !AllowPublicAccess("allow_empty_or_whitelist", []string{"example.com"}, "") {
		t.Fatal("allow_empty_or_whitelist should allow empty referer")
	}
	if !AllowPublicAccess("referer_whitelist", []string{"example.com"}, "https://example.com/a.jpg") {
		t.Fatal("expected allowed host")
	}
	if AllowPublicAccess("referer_whitelist", []string{"example.com"}, "https://evil.com/a.jpg") {
		t.Fatal("unexpected allowed host")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/security -v`
Expected: FAIL because package/functions do not exist.

- [ ] **Step 3: Implement the minimal behavior**

Create `internal/security/public_access.go` with:

```go
package security

import (
	"net/url"
	"strings"
)

func AllowPublicAccess(mode string, allowedHosts []string, referer string) bool
```

Implementation rules:

- `off` → always true
- `referer_whitelist` → require non-empty referer and host hit
- `allow_empty_or_whitelist` → empty referer true; non-empty referer must hit
- normalize hosts with `strings.ToLower(strings.TrimSpace(...))`
- parse referer with `url.Parse`; parse failure = reject
- compare exact host name, but strip port via `u.Hostname()`

- [ ] **Step 4: Run verification**

Run: `go test ./internal/security -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 7 block.

---

### Task 8: 在 PublicHandler 接入公开访问策略

**Files:**
- Modify: `internal/handler/public.go`
- Modify: `internal/handler/public_test.go`
- Modify: `main.go`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Extend `internal/handler/public_test.go` with:

```go
func TestPublicHandlerRejectsDisallowedReferer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	provider := config.NewProvider(config.Config{
		Image: testPublicImageConfig(),
		PublicAccess: config.PublicAccessConfig{Mode: "referer_whitelist", AllowedRefererHosts: []string{"good.example"}},
	}, config.Overrides{})
	h := NewPublicHandler(&stubPublicStorage{files: map[string]string{"demo/test.jpg": "image-bytes"}}, imgpkg.NewProcessor(provider), provider)
	engine.GET("/i/*key", h.Image)

	req := httptest.NewRequest(http.MethodGet, "/i/demo/test.jpg", nil)
	req.Header.Set("Referer", "https://evil.example/path")
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "public_access_forbidden") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/handler -run TestPublicHandlerRejectsDisallowedReferer -v`
Expected: FAIL because handler constructor does not accept provider and no 403 branch exists.

- [ ] **Step 3: Implement the minimal behavior**

Refactor `internal/handler/public.go`:

```go
type PublicHandler struct {
	store     storage.Storage
	processor *imgpkg.Processor
	provider  *config.Provider
}

func NewPublicHandler(store storage.Storage, processor *imgpkg.Processor, provider *config.Provider) *PublicHandler
```

In `serve()` before storage lookup:

```go
cfg := h.provider.Get().PublicAccess
if !security.AllowPublicAccess(cfg.Mode, cfg.AllowedRefererHosts, c.GetHeader("Referer")) {
	c.JSON(http.StatusForbidden, gin.H{"error": "public_access_forbidden"})
	return
}
```

Update `main.go` to pass `provider` into `NewPublicHandler(...)`.

Keep existing cache/content-type behavior unchanged when access is allowed.

- [ ] **Step 4: Run verification**

Run: `go test ./internal/handler -run 'TestPublicHandler(ImageServesStoredFileWithHeaders|ImageReturnsNotFoundWhenMissing|RejectsDisallowedReferer)' -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 8 block.

---

### Task 9: 让 `strip_exif` 在 Processor 中真实生效

**Files:**
- Modify: `internal/image/processor.go`
- Modify: `internal/image/processor_test.go`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

Extend `internal/image/processor_test.go` with a round-trip assertion that the original bytes differ when `strip_exif=true` and stay byte-identical when `strip_exif=false` for a JPEG fixture carrying EXIF-like APP1 payload:

```go
func TestProcessorRespectsStripExifForOriginalJPEG(t *testing.T) {
	original := jpegWithAPP1Fixture(t)
	base := config.Config{Image: config.ImageConfig{AutoConvert: "jpeg", Quality: 85, StripExif: true, Thumbnails: []config.ThumbnailSize{{Name: "thumb", Width: 32, Height: 32}}}}
	stripProvider := config.NewProvider(base, config.Overrides{})
	keepCfg := base
	keepCfg.Image.StripExif = false
	keepProvider := config.NewProvider(keepCfg, config.Overrides{})

	stripped, err := NewProcessor(stripProvider).Process(original, "image/jpeg")
	if err != nil { t.Fatalf("Process(strip): %v", err) }
	kept, err := NewProcessor(keepProvider).Process(original, "image/jpeg")
	if err != nil { t.Fatalf("Process(keep): %v", err) }

	if bytes.Equal(stripped.OriginalData, original) {
		t.Fatal("strip_exif=true should rewrite original bytes")
	}
	if !bytes.Equal(kept.OriginalData, original) {
		t.Fatal("strip_exif=false should preserve original bytes for jpeg best-effort path")
	}
}
```

This test requires extending `ProcessResult` to expose `OriginalData []byte`.

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/image -run TestProcessorRespectsStripExifForOriginalJPEG -v`
Expected: FAIL because `OriginalData` does not exist and Process currently never rewrites original bytes.

- [ ] **Step 3: Implement the minimal behavior**

In `internal/image/processor.go`:

- extend `ProcessResult` with `OriginalData []byte`
- when `mimeType == "image/jpeg"`:
  - decode with `AutoOrientation(true)` as today
  - if `cfg.StripExif == true`, encode the oriented image back to JPEG using `cfg.Quality`, assign bytes to `OriginalData`
  - else assign `OriginalData = data`
- for non-JPEG inputs, keep `OriginalData = data` in this task; do **not** invent cross-format rewriting logic not required by the spec

Then update the image upload path (same file or `internal/service/image.go` depending on current boundary) so original object storage uses `result.OriginalData` instead of raw upload bytes when it is non-nil.

- [ ] **Step 4: Run verification**

Run: `go test ./internal/image -run TestProcessorRespectsStripExifForOriginalJPEG -v && go test ./internal/service -run TestImageServiceUploadDeduplicatesByHash -v`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 9 block.

---

### Task 10: 全量回归、feature closeout 验证与收口文档

**Files:**
- Modify: `docs/superpowers/execution-log/2026-05-28-core-security-and-access-control.md`
- Create: `docs/superpowers/completion/2026-05-28-core-security-and-access-control-summary.md`
- Modify: `docs/superpowers/status.md`

- [ ] **Step 0: Load review config**

Read: `docs/superpowers/plans/2026-05-28-core-security-and-access-control.review-config.md`
Apply the configured task-level / feature-level review strategy before implementing this task.

- [ ] **Step 1: Write the failing test**

This task is verification/documentation-only. The “failing test” is the absence of a feature closeout run and missing completion summary before final verification.

- [ ] **Step 2: Run the test to verify it fails**

Run: `ls docs/superpowers/completion/2026-05-28-core-security-and-access-control-summary.md`
Expected: `No such file or directory`.

- [ ] **Step 3: Implement the minimal behavior**

Run full verification exactly:

```bash
go test ./... -count=1
cd web && npm run build
```

Then write `docs/superpowers/completion/2026-05-28-core-security-and-access-control-summary.md` as a dashboard with:

- scope sentence
- spec coverage table for token expiry / upload rate limit / public access / EXIF
- links to spec / plan / review-config / execution-log
- known issues or explicit “none”
- summary + next step

Update `docs/superpowers/status.md`:

- set active sub-project to decomposition 子项目 1
- set current phase to `completion` or `in review` depending on whether external review has run
- point Current spec / plan / completion summary to the new feature
- set next recommended action to feature-level review and PR preparation

- [ ] **Step 4: Run verification**

Run: `go test ./... -count=1 && cd web && npm run build`
Expected: PASS.

- [ ] **Step 5: Write the merged execution-log block**

Append a Task 10 block including the exact verification commands and actual outcomes.

---

## Self-Review Against Spec

- **Spec coverage:**
  - Token 过期校验 → Task 2–4
  - 上传限流 → Task 5–6
  - 公开访问策略 → Task 7–8
  - EXIF 行为落地 → Task 9
  - 文档 / completion / status 收口 → Task 10
- **Placeholder scan:** no TBD / TODO / “similar to task N” placeholders remain.
- **Type consistency:** `TokenPolicyConfig` / `UploadRateLimitConfig` / `PublicAccessConfig` names are consistent across config, service, handler tasks.
- **Lightweight plan check:** stripping code blocks still leaves a coherent sequence: config → token → UI → limiter → public access → EXIF → full verification.
