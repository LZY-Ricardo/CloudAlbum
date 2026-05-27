# Admin Settings & Account Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 decomposition 子项目 4「管理后台设置与账户闭环」—— 改密码 API、Settings 部分可编辑、运行时配置 hot-swap、Account 页与默认密码 Banner。

**Architecture:** YAML 仅 bootstrap；新增 `settings` 表（单行 JSON）作运行时权威源；`config.Provider` 用 `atomic.Pointer[Config]` 持快照，service / processor / handler 全部改为从 Provider 读；改密通过 `User.TokenVersion` 在 JWT claims 内嵌的 `tv` 字段实现旧 JWT 吊销。前端新增 Account 页 + 顶部默认密码 Banner，Settings 页改造为可编辑表单。

**Tech Stack:** Go 1.x、Gin、GORM (SQLite)、JWT (golang-jwt/v5)、React 18 + TypeScript + Vite + Zustand + Arco Design Web React。

**Spec:** `docs/superpowers/specs/2026-05-27-admin-settings-and-account-design.md`

---

## File Structure

### 新建后端文件

| 路径 | 责任 |
|---|---|
| `internal/config/overrides.go` | `Overrides` 结构 + JSON marshal/unmarshal |
| `internal/config/overrides_test.go` | overrides 反序列化 / 指针语义测试 |
| `internal/config/provider.go` | `Provider` + atomic.Pointer + Apply / Get |
| `internal/config/provider_test.go` | Provider 单测 + 并发 race 测试 |
| `internal/model/settings.go` | `Settings` 模型（单行 JSON） |
| `internal/repository/settings.go` | `SettingsRepository` + `LoadOrBootstrap` + `Save` |
| `internal/repository/settings_test.go` | repository 单测 |
| `internal/service/settings.go` | `SettingsService` + 白名单 / 字段级校验 / merge / sync.Mutex |
| `internal/service/settings_test.go` | service 单测 |
| `internal/handler/settings.go` | `SettingsHandler` GET / PUT |
| `internal/handler/settings_test.go` | handler + 路由 / 鉴权矩阵测试 |

### 修改后端文件

| 路径 | 改动概要 |
|---|---|
| `internal/model/user.go` | 新增 `TokenVersion` / `PasswordChangedAt` |
| `internal/repository/user.go` | 新增 `Update` / 改密事务方法 |
| `internal/service/auth.go` | 切到 `*Provider`；扩展 `Claims` (`TV` → `tv`)；新增 `ChangePassword` + errors |
| `internal/service/auth_test.go` | 扩展测试 |
| `internal/middleware/auth.go` | JWT 路径增加 `token_version` 比对 |
| `internal/handler/auth.go` | 新增 `ChangePassword` endpoint；扩展 `Me` 返回字段 |
| `internal/image/processor.go` | 切到 `*Provider`，每次操作读 `provider.Get().Image` |
| `internal/service/image.go` | 切到 `*Provider`，去掉 `cfg` / `baseURL` 字段 |
| `internal/service/image_test.go` | 适配新构造签名 |
| `internal/handler/public.go` | 若引用 base_url / image cfg，改为读 Provider |
| `internal/router/router.go` | 挂载 `/auth/change-password`、`/settings` |
| `main.go` | AutoMigrate 增加 `Settings`；构造 Provider；yaml mtime 对比 INFO；service wiring |
| `README.md` | "配置管理" 章节 |

### 新建前端文件

| 路径 | 责任 |
|---|---|
| `web/src/pages/Account.tsx` | 账号信息 + 改密表单 |
| `web/src/components/DefaultPasswordBanner.tsx` | 默认密码全局提示 |

### 修改前端文件

| 路径 | 改动概要 |
|---|---|
| `web/src/stores/auth.ts` | 缓存 `me`；改密续签；Banner 关闭状态；`init` 拉 me |
| `web/src/App.tsx` | 加 `/account` 路由 |
| `web/src/components/Layout.tsx` | 侧边栏插入 Account（Settings 之前）；顶栏上方挂 Banner |
| `web/src/pages/Settings.tsx` | 改造为可编辑表单 |

---

## Task Map

- **Phase A — 后端基础**：Task 1–6（Overrides / Provider / Settings 模型与 repo / User 模型 / main.go wiring）
- **Phase B — Provider 接入下游**：Task 7–9（Processor / ImageService / AuthService）
- **Phase C — 改密与 JWT 校验**：Task 10–12（ChangePassword service / middleware / handler & route）
- **Phase D — Settings 接口**：Task 13–14（SettingsService / SettingsHandler & route）
- **Phase E — 前端**：Task 15–20
- **Phase F — 文档与验收**：Task 21–22

---

## Phase A — 后端基础

### Task 1: 新增 `config.Overrides` 结构

**Files:**
- Create: `internal/config/overrides.go`
- Test: `internal/config/overrides_test.go`

- [ ] **Step 1: 写失败测试**

`internal/config/overrides_test.go`:

```go
package config

import (
	"encoding/json"
	"testing"
)

func TestOverridesEmptyJSON(t *testing.T) {
	var o Overrides
	if err := json.Unmarshal([]byte(`{}`), &o); err != nil {
		t.Fatalf("unmarshal empty: %v", err)
	}
	if o.Server.BaseURL != nil {
		t.Fatalf("BaseURL should be nil")
	}
	if o.Image.Quality != nil {
		t.Fatalf("Quality should be nil")
	}
}

func TestOverridesPartialJSON(t *testing.T) {
	var o Overrides
	raw := `{"server":{"base_url":"https://x"},"image":{"quality":90,"strip_exif":true}}`
	if err := json.Unmarshal([]byte(raw), &o); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if o.Server.BaseURL == nil || *o.Server.BaseURL != "https://x" {
		t.Fatalf("BaseURL: %#v", o.Server.BaseURL)
	}
	if o.Image.Quality == nil || *o.Image.Quality != 90 {
		t.Fatalf("Quality: %#v", o.Image.Quality)
	}
	if o.Image.StripExif == nil || *o.Image.StripExif != true {
		t.Fatalf("StripExif: %#v", o.Image.StripExif)
	}
	if o.Image.MaxSize != nil {
		t.Fatalf("MaxSize should remain nil")
	}
}

func TestOverridesMarshalOmitEmpty(t *testing.T) {
	var o Overrides
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(data) != `{"server":{},"image":{}}` {
		t.Fatalf("unexpected: %s", data)
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/config/... -run Overrides -v
```
Expected: FAIL，`undefined: Overrides`。

- [ ] **Step 3: 最小实现**

`internal/config/overrides.go`:

```go
package config

// Overrides 表示运行时对 YAML 基线的覆盖值。
// 所有字段使用指针：nil 表示"未设置 → 落回 YAML 默认"。
type Overrides struct {
	Server struct {
		BaseURL *string `json:"base_url,omitempty"`
	} `json:"server"`
	Image struct {
		MaxSize      *int64    `json:"max_size,omitempty"`
		AllowedTypes *[]string `json:"allowed_types,omitempty"`
		AutoConvert  *string   `json:"auto_convert,omitempty"`
		Quality      *int      `json:"quality,omitempty"`
		StripExif    *bool     `json:"strip_exif,omitempty"`
	} `json:"image"`
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/config/... -run Overrides -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/config/overrides.go internal/config/overrides_test.go
git commit -m "feat(config): add Overrides struct for runtime settings"
```

---

### Task 2: 新增 `config.Provider`

**Files:**
- Create: `internal/config/provider.go`
- Test: `internal/config/provider_test.go`

- [ ] **Step 1: 写失败测试**

`internal/config/provider_test.go`:

```go
package config

import (
	"sync"
	"testing"
)

func baseFixture() Config {
	return Config{
		Server: ServerConfig{Port: 8080, BaseURL: "http://localhost:8080"},
		Image: ImageConfig{
			MaxSize:      50 << 20,
			AllowedTypes: []string{"jpg", "png"},
			AutoConvert:  "webp",
			Quality:      85,
			StripExif:    true,
		},
		Auth: AuthConfig{JWTSecret: "s", TokenExpire: 168 * 3600 * 1e9},
	}
}

func TestProviderEmptyOverridesEqualsBase(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	got := p.Get()
	if got.Server.BaseURL != "http://localhost:8080" {
		t.Fatalf("BaseURL: %s", got.Server.BaseURL)
	}
	if got.Image.Quality != 85 {
		t.Fatalf("Quality: %d", got.Image.Quality)
	}
}

func TestProviderApplyOverrides(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	var o Overrides
	url := "https://img.example.com"
	q := 90
	o.Server.BaseURL = &url
	o.Image.Quality = &q
	p.Apply(o)

	got := p.Get()
	if got.Server.BaseURL != url {
		t.Fatalf("BaseURL not overridden: %s", got.Server.BaseURL)
	}
	if got.Image.Quality != 90 {
		t.Fatalf("Quality not overridden: %d", got.Image.Quality)
	}
	if got.Image.MaxSize != 50<<20 {
		t.Fatalf("MaxSize should keep base: %d", got.Image.MaxSize)
	}
}

func TestProviderConcurrentReadWrite(t *testing.T) {
	p := NewProvider(baseFixture(), Overrides{})
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					_ = p.Get()
				}
			}
		}()
	}
	for i := 0; i < 200; i++ {
		var o Overrides
		q := i%100 + 1
		o.Image.Quality = &q
		p.Apply(o)
	}
	close(stop)
	wg.Wait()
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test -race ./internal/config/... -run Provider -v
```
Expected: FAIL，`undefined: NewProvider`。

- [ ] **Step 3: 最小实现**

`internal/config/provider.go`:

```go
package config

import "sync/atomic"

// Provider 持有运行时配置快照。读侧零锁，写侧用 atomic.Pointer 整体替换。
type Provider struct {
	base      Config
	snapshot  atomic.Pointer[Config]
	overrides atomic.Pointer[Overrides]
}

func NewProvider(base Config, overrides Overrides) *Provider {
	p := &Provider{base: base}
	p.applyOverrides(overrides)
	return p
}

// Get 返回当前生效的只读快照。调用方按约定不可修改。
func (p *Provider) Get() *Config { return p.snapshot.Load() }

// Overrides 返回当前 overrides 副本，用于 UI 回显。
func (p *Provider) Overrides() *Overrides { return p.overrides.Load() }

// Apply 用新的 overrides 重新构建快照并原子替换。
func (p *Provider) Apply(o Overrides) { p.applyOverrides(o) }

func (p *Provider) applyOverrides(o Overrides) {
	merged := p.base
	if o.Server.BaseURL != nil {
		merged.Server.BaseURL = *o.Server.BaseURL
	}
	if o.Image.MaxSize != nil {
		merged.Image.MaxSize = *o.Image.MaxSize
	}
	if o.Image.AllowedTypes != nil {
		merged.Image.AllowedTypes = *o.Image.AllowedTypes
	}
	if o.Image.AutoConvert != nil {
		merged.Image.AutoConvert = *o.Image.AutoConvert
	}
	if o.Image.Quality != nil {
		merged.Image.Quality = *o.Image.Quality
	}
	if o.Image.StripExif != nil {
		merged.Image.StripExif = *o.Image.StripExif
	}
	p.snapshot.Store(&merged)
	p.overrides.Store(&o)
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test -race ./internal/config/... -run Provider -v
```
Expected: PASS（包括 race detector 不报警）。

- [ ] **Step 5: 提交**

```bash
git add internal/config/provider.go internal/config/provider_test.go
git commit -m "feat(config): add Provider for runtime config hot-swap"
```

---

### Task 3: 新增 `model.Settings`

**Files:**
- Create: `internal/model/settings.go`

- [ ] **Step 1: 实现模型**

`internal/model/settings.go`:

```go
package model

import "time"

// Settings 表存运行时配置 overrides，固定单行 id=1，payload 为 JSON。
type Settings struct {
	ID        uint      `gorm:"primaryKey"`
	Payload   string    `gorm:"type:text;not null"`
	UpdatedAt time.Time
	UpdatedBy uint
}
```

- [ ] **Step 2: 编译检查**

```bash
go build ./...
```
Expected: 成功，无新错误。

- [ ] **Step 3: 提交**

```bash
git add internal/model/settings.go
git commit -m "feat(model): add Settings model for runtime config row"
```

---

### Task 4: 新增 `SettingsRepository`

**Files:**
- Create: `internal/repository/settings.go`
- Test: `internal/repository/settings_test.go`

- [ ] **Step 1: 写失败测试**

`internal/repository/settings_test.go`:

```go
package repository

import (
	"encoding/json"
	"testing"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Settings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSettingsRepoLoadOrBootstrapEmpty(t *testing.T) {
	db := newTestDB(t)
	repo := NewSettingsRepository(db)
	o, err := repo.LoadOrBootstrap()
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if o.Server.BaseURL != nil || o.Image.Quality != nil {
		t.Fatalf("expected zero overrides, got %#v", o)
	}
	var row model.Settings
	if err := db.First(&row, 1).Error; err != nil {
		t.Fatalf("row not seeded: %v", err)
	}
	if row.Payload != "{}" && row.Payload != `{"server":{},"image":{}}` {
		t.Fatalf("payload not empty: %s", row.Payload)
	}
}

func TestSettingsRepoSaveAndLoad(t *testing.T) {
	db := newTestDB(t)
	repo := NewSettingsRepository(db)
	_, _ = repo.LoadOrBootstrap()

	var o config.Overrides
	q := 90
	o.Image.Quality = &q
	if err := repo.Save(o, 1); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.LoadOrBootstrap()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Image.Quality == nil || *got.Image.Quality != 90 {
		t.Fatalf("reload quality: %#v", got.Image.Quality)
	}
}

func TestSettingsRepoCorruptedPayload(t *testing.T) {
	db := newTestDB(t)
	if err := db.Create(&model.Settings{ID: 1, Payload: "not json"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewSettingsRepository(db)
	_, err := repo.LoadOrBootstrap()
	if err == nil {
		t.Fatalf("expected error for corrupted payload")
	}
	_ = json.Valid // keep import
}
```

注：测试中 `db := newTestDB(t)` 多次调用会因 SQLite shared in-memory 数据库串台导致 `TestSettingsRepoCorruptedPayload` 失败。改 DSN 为唯一名（每个测试使用 `file:<name>:?mode=memory&cache=shared`）：

修订辅助：把 `newTestDB` 中 DSN 改为按测试名生成。最小改动：

```go
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	...
}
```

写测试时直接采用该版本。

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/repository/... -run Settings -v
```
Expected: FAIL，`undefined: NewSettingsRepository`。

- [ ] **Step 3: 最小实现**

`internal/repository/settings.go`:

```go
package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"

	"gorm.io/gorm"
)

type SettingsRepository struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// LoadOrBootstrap 读取 settings 表的唯一一行；若表空则插入空 overrides 并返回零值。
// 若 payload 损坏，返回错误，由上层决定回退策略。
func (r *SettingsRepository) LoadOrBootstrap() (config.Overrides, error) {
	var row model.Settings
	err := r.db.First(&row, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		seed := model.Settings{ID: 1, Payload: "{}", UpdatedAt: time.Now()}
		if err := r.db.Create(&seed).Error; err != nil {
			return config.Overrides{}, fmt.Errorf("bootstrap settings: %w", err)
		}
		return config.Overrides{}, nil
	}
	if err != nil {
		return config.Overrides{}, fmt.Errorf("load settings: %w", err)
	}
	var o config.Overrides
	if err := json.Unmarshal([]byte(row.Payload), &o); err != nil {
		return config.Overrides{}, fmt.Errorf("decode settings payload: %w", err)
	}
	return o, nil
}

// Save 用新 overrides 整体覆盖唯一一行；调用方需保证传入的 overrides 已与旧值 merge。
func (r *SettingsRepository) Save(o config.Overrides, updatedBy uint) error {
	data, err := json.Marshal(o)
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return r.db.Model(&model.Settings{}).
		Where("id = ?", 1).
		Updates(map[string]any{
			"payload":    string(data),
			"updated_at": time.Now(),
			"updated_by": updatedBy,
		}).Error
}

// UpdatedAt 返回当前 settings 行的更新时间，用于 yaml mtime 比对。
func (r *SettingsRepository) UpdatedAt() (time.Time, error) {
	var row model.Settings
	if err := r.db.Select("updated_at").First(&row, 1).Error; err != nil {
		return time.Time{}, err
	}
	return row.UpdatedAt, nil
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/repository/... -run Settings -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/repository/settings.go internal/repository/settings_test.go
git commit -m "feat(repo): add SettingsRepository with bootstrap and save"
```

---

### Task 5: User 模型扩展 TokenVersion / PasswordChangedAt

**Files:**
- Modify: `internal/model/user.go`
- Modify: `internal/repository/user.go`

- [ ] **Step 1: 修改 User 模型**

将 `internal/model/user.go` 改为：

```go
package model

import "time"

type User struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	Username          string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	PasswordHash      string     `gorm:"not null" json:"-"`
	Role              string     `gorm:"size:20;default:admin" json:"role"`
	TokenVersion      uint       `gorm:"not null;default:1" json:"-"`
	PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}
```

- [ ] **Step 2: 扩展 UserRepository**

在 `internal/repository/user.go` 末尾追加：

```go
// UpdatePasswordAndBumpVersion 在一个事务内更新密码 hash、自增 TokenVersion、写入 PasswordChangedAt。
func (r *UserRepository) UpdatePasswordAndBumpVersion(userID uint, newHash string, changedAt time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var u model.User
		if err := tx.First(&u, userID).Error; err != nil {
			return err
		}
		u.PasswordHash = newHash
		u.TokenVersion = u.TokenVersion + 1
		u.PasswordChangedAt = &changedAt
		return tx.Save(&u).Error
	})
}
```

并在文件顶部 import 中确认包含 `"time"`。

- [ ] **Step 3: 编译**

```bash
go build ./...
```
Expected: 通过。

- [ ] **Step 4: 跑现有相关测试**

```bash
go test ./internal/... -run User -v
```
Expected: 已有测试不回归。

- [ ] **Step 5: 提交**

```bash
git add internal/model/user.go internal/repository/user.go
git commit -m "feat(user): add TokenVersion and PasswordChangedAt fields"
```

---

### Task 6: main.go 启动流程接入 Provider 与 Settings

**Files:**
- Modify: `main.go`

注意：本任务**只引入 Provider + Settings 接入启动流程**，下游 service 仍接受原 `config.Xxx`；service 改造在 Phase B 完成。本任务结束时 Provider 已构造但下游仍走原 `cfg.X` 字段；这样可独立提交不破坏现有逻辑。

- [ ] **Step 1: AutoMigrate 增加 Settings**

定位 `main.go` 中 `initDB` 内 `AutoMigrate` 调用（约第 101 行），改为：

```go
if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Album{}, &model.APIToken{}, &model.Settings{}); err != nil {
	return nil, fmt.Errorf("auto migrate: %w", err)
}
```

- [ ] **Step 2: 构造 Provider 并 bootstrap settings**

在 `main()` 中 repo 初始化后、service 初始化前插入：

```go
settingsRepo := repository.NewSettingsRepository(db)
overrides, err := settingsRepo.LoadOrBootstrap()
if err != nil {
	log.Printf("[config] WARNING: settings load failed, falling back to empty overrides: %v", err)
	overrides = config.Overrides{}
}
provider := config.NewProvider(*cfg, overrides)
```

注：本步骤暂不修改 service 构造签名，Provider 暂为未使用变量。下一步加 yaml mtime 比对会消耗它；若构建报 `provider declared but not used`，先加 `_ = provider`（Phase B 会去掉）。

- [ ] **Step 3: 加 yaml mtime 与 settings.updated_at 比对的 INFO 日志**

紧跟在上一步代码之后：

```go
if info, statErr := os.Stat("configs/config.yaml"); statErr == nil {
	if updated, err := settingsRepo.UpdatedAt(); err == nil && info.ModTime().After(updated) {
		log.Printf("[config] yaml file is newer than settings table (yaml=%s settings=%s); runtime still uses values from settings DB. To re-seed from YAML, delete the row in settings table and restart.",
			info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
			updated.Format("2006-01-02T15:04:05Z07:00"))
	}
}
```

- [ ] **Step 4: 编译并启动一次**

```bash
go build ./...
go run . &
sleep 1
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/
pkill -f "go run \." || pkill cloudalbum
```
Expected: 编译通过；服务 200 / 走 SPA fallback；日志能看到「Database / Storage」原有输出；settings 表 id=1 行已创建（首次启动）。

- [ ] **Step 5: 提交**

```bash
git add main.go
git commit -m "feat(main): wire config.Provider and bootstrap settings table"
```

---

## Phase B — Provider 接入下游

### Task 7: `imgpkg.Processor` 切到 `*Provider`

**Files:**
- Modify: `internal/image/processor.go`
- Modify: `main.go`
- Modify: `internal/handler/public.go`（仅签名传参）

- [ ] **Step 1: 改造 Processor 构造函数与字段**

`internal/image/processor.go` 顶部 import 增加 `"cloudalbum/internal/config"` 已有；定义改为：

```go
type Processor struct {
	provider *config.Provider
}

func NewProcessor(provider *config.Provider) *Processor {
	return &Processor{provider: provider}
}

func (p *Processor) imageCfg() config.ImageConfig {
	return p.provider.Get().Image
}
```

- [ ] **Step 2: 替换 Processor 内对 `p.cfg` 的引用**

将 `Process` 中 `for _, size := range p.cfg.Thumbnails` 改为：

```go
cfg := p.imageCfg()
...
for _, size := range cfg.Thumbnails {
```

将 `thumbnailEncoding` 改为：

```go
func (p *Processor) thumbnailEncoding() (imaging.Format, []imaging.EncodeOption) {
	cfg := p.imageCfg()
	switch strings.ToLower(cfg.AutoConvert) {
	case "", "jpg", "jpeg", "webp":
		return imaging.JPEG, []imaging.EncodeOption{imaging.JPEGQuality(cfg.Quality)}
	case "png":
		return imaging.PNG, nil
	case "gif":
		return imaging.GIF, nil
	default:
		return imaging.JPEG, []imaging.EncodeOption{imaging.JPEGQuality(cfg.Quality)}
	}
}
```

注意：`Process` 中 `thumbnailEncoding()` 调用需要在取 `cfg` 之后进行（已实现中是 `thumbFormat, thumbOptions := p.thumbnailEncoding()`，独立 cfg，本任务不需要传递）。

- [ ] **Step 3: 更新 main.go 调用点**

将 `processor := imgpkg.NewProcessor(cfg.Image)` 改为：

```go
processor := imgpkg.NewProcessor(provider)
```

- [ ] **Step 4: publicHandler 不受影响**

`internal/handler/public.go` 中 `handler.NewPublicHandler(store, processor)` 签名不变，processor 持 provider 即可。如果 `public.go` 内显式读 `cfg`，将其改为 `provider.Get()`。

```bash
grep -n "cfg\." internal/handler/public.go || echo "no cfg references"
```

若无 cfg 引用则本步骤无改动。

- [ ] **Step 5: 编译**

```bash
go build ./...
```
Expected: 通过。

- [ ] **Step 6: 跑现有 processor / image 测试**

```bash
go test ./internal/image/... ./internal/service/... -v
```
Expected：所有现有测试通过。如果 `image_test.go` 因 `NewProcessor` 签名变更编译失败，本任务**不修复**该测试；放到 Task 8 一并处理（Task 8 改 ImageService 时会更新测试）。如果 `processor_test.go` 不存在或不依赖 `NewProcessor(cfg)`，应通过。

如果当前 image_test.go 失败导致整体包不编译，**暂时跳过整包**：

```bash
go build ./internal/image/...
go test ./internal/image/...
```

- [ ] **Step 7: 提交**

```bash
git add internal/image/processor.go main.go
git commit -m "refactor(image): Processor reads ImageConfig from Provider"
```

---

### Task 8: `ImageService` 切到 `*Provider`

**Files:**
- Modify: `internal/service/image.go`
- Modify: `internal/service/image_test.go`
- Modify: `main.go`

- [ ] **Step 1: 修改 ImageService 结构与构造**

`internal/service/image.go`:

```go
type ImageService struct {
	imageRepo *repository.ImageRepository
	store     storage.Storage
	processor *imgpkg.Processor
	provider  *config.Provider
}

func NewImageService(imageRepo *repository.ImageRepository, store storage.Storage, processor *imgpkg.Processor, provider *config.Provider) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		store:     store,
		processor: processor,
		provider:  provider,
	}
}

func (s *ImageService) imageCfg() config.ImageConfig {
	return s.provider.Get().Image
}

func (s *ImageService) baseURL() string {
	return strings.TrimRight(s.provider.Get().Server.BaseURL, "/")
}
```

- [ ] **Step 2: 替换内部 `s.cfg` / `s.baseURL` 引用**

- `Upload` 中 `if file.Size > s.cfg.MaxSize`：先 `cfg := s.imageCfg()`，改为 `if file.Size > cfg.MaxSize`。
- `Upload` 中 `if !s.isAllowedType(ext)`：保持不变（`isAllowedType` 内部走 provider，下面会改）。
- `UploadFromURL` 中 `io.LimitReader(resp.Body, s.cfg.MaxSize+1)` 与 `if int64(len(data)) > s.cfg.MaxSize`：方法开头取 `cfg := s.imageCfg()`，用 `cfg.MaxSize`。
- `URLs`：

```go
func (s *ImageService) URLs(img *model.Image) map[string]string {
	url := s.baseURL() + "/i/" + strings.TrimLeft(img.StorageKey, "/")
	return map[string]string{
		"url":      url,
		"markdown": fmt.Sprintf("![%s](%s)", img.OriginalName, url),
		"html":     fmt.Sprintf(`<img src="%s" alt="%s">`, url, img.OriginalName),
		"bbcode":   fmt.Sprintf("[img]%s[/img]", url),
	}
}
```

- `isAllowedType`：

```go
func (s *ImageService) isAllowedType(ext string) bool {
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(ext)), ".")
	for _, allowed := range s.imageCfg().AllowedTypes {
		if normalized == strings.TrimPrefix(strings.ToLower(strings.TrimSpace(allowed)), ".") {
			return true
		}
	}
	return false
}
```

- 移除结构体中的 `cfg config.ImageConfig` 和 `baseURL string` 字段（已替换为 provider）。

- [ ] **Step 3: 更新 main.go 调用**

```go
imageSvc := service.NewImageService(imageRepo, store, processor, provider)
```

- [ ] **Step 4: 修复 image_test.go 构造调用**

`internal/service/image_test.go` 中所有 `NewImageService(...)` 调用：

替换：旧 `cfg config.ImageConfig` + `baseURL string` 参数 → 用本地构造 `provider`。模板：

```go
import "cloudalbum/internal/config"

func newTestProvider(t *testing.T, base config.Config) *config.Provider {
	t.Helper()
	return config.NewProvider(base, config.Overrides{})
}

// 调用点示例：
base := config.Config{
	Server: config.ServerConfig{BaseURL: "http://localhost:8080"},
	Image: config.ImageConfig{
		MaxSize:      10 << 20,
		AllowedTypes: []string{"jpg","png"},
		AutoConvert:  "webp",
		Quality:      85,
		StripExif:    true,
	},
}
provider := newTestProvider(t, base)
processor := imgpkg.NewProcessor(provider)
svc := service.NewImageService(imageRepo, store, processor, provider)
```

把测试中已有的 `cfg` 局部变量与 `baseURL` 字符串改写成上述形态。如果原测试用了独立的 `cfg.Image`，把同一字段值搬进 `base.Image`。

- [ ] **Step 5: 编译并跑测试**

```bash
go build ./...
go test ./internal/image/... ./internal/service/... -v
```
Expected: PASS。

- [ ] **Step 6: 跑 race 检测一次**

```bash
go test -race ./internal/service/...
```
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/service/image.go internal/service/image_test.go main.go
git commit -m "refactor(image): ImageService reads cfg/baseURL from Provider"
```

---

### Task 9: `AuthService` 切到 `*Provider`

**Files:**
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_test.go`
- Modify: `main.go`

- [ ] **Step 1: 修改 AuthService 字段与构造**

`internal/service/auth.go`:

```go
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	provider  *config.Provider
}

func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, provider *config.Provider) *AuthService {
	return &AuthService{userRepo: userRepo, tokenRepo: tokenRepo, provider: provider}
}

func (s *AuthService) authCfg() config.AuthConfig {
	return s.provider.Get().Auth
}
```

- [ ] **Step 2: 替换 `s.cfg` 引用**

- `GenerateJWT`：

```go
func (s *AuthService) GenerateJWT(user *model.User) (string, error) {
	cfg := s.authCfg()
	now := time.Now()
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.TokenExpire)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}
```

注：`Claims.TokenVersion` 字段将在 Task 10 添加，本任务不引入。

- `ParseJWT`：

```go
func (s *AuthService) ParseJWT(tokenStr string) (*Claims, error) {
	cfg := s.authCfg()
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return []byte(cfg.JWTSecret), nil
	})
	...
}
```

- [ ] **Step 3: 更新 main.go 调用**

```go
authSvc := service.NewAuthService(userRepo, tokenRepo, provider)
```

并删除前面 Phase A 末尾如有的 `_ = provider`（现在 provider 已被实际消费）。

- [ ] **Step 4: 修复 auth_test.go**

将 `auth_test.go` 中所有 `service.NewAuthService(..., cfg.Auth)` 改为通过 provider 构造：

```go
func newAuthSvc(t *testing.T, userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository) *service.AuthService {
	t.Helper()
	base := config.Config{Auth: config.AuthConfig{JWTSecret: "test-secret", TokenExpire: time.Hour}}
	provider := config.NewProvider(base, config.Overrides{})
	return service.NewAuthService(userRepo, tokenRepo, provider)
}
```

把现有调用替换为该工厂或等价构造。

- [ ] **Step 5: 编译与测试**

```bash
go build ./...
go test ./internal/service/... -v
```
Expected: PASS。

- [ ] **Step 6: 提交**

```bash
git add internal/service/auth.go internal/service/auth_test.go main.go
git commit -m "refactor(auth): AuthService reads AuthConfig from Provider"
```

---

## Phase C — 改密与 JWT 校验扩展

### Task 10: `AuthService.ChangePassword` + JWT Claims TokenVersion

**Files:**
- Modify: `internal/service/auth.go`
- Modify: `internal/service/auth_test.go`

- [ ] **Step 1: 扩展 Claims 与错误 sentinel**

`internal/service/auth.go`，错误定义区域追加：

```go
var (
	ErrPasswordTooShort   = errors.New("password too short")
	ErrPasswordSameAsOld  = errors.New("password same as old")
	ErrAPITokenForbidden  = errors.New("api token forbidden")
)
```

`Claims` 增加 `TokenVersion`：

```go
type Claims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	TokenVersion uint   `json:"tv"`
	jwt.RegisteredClaims
}
```

`GenerateJWT` 内 `Claims{...}` 字面量增加 `TokenVersion: user.TokenVersion`。

- [ ] **Step 2: 写失败测试**

`internal/service/auth_test.go` 追加：

```go
func TestChangePasswordSuccess(t *testing.T) {
	db := newTestDB(t) // 复用现有 helper；若没有，构造一个 with sqlite in-memory
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	svc := newAuthSvc(t, userRepo, tokenRepo)

	if err := svc.EnsureAdmin("admin", "admin123"); err != nil {
		t.Fatalf("ensure admin: %v", err)
	}
	user, _ := userRepo.FindByUsername("admin")
	originalVersion := user.TokenVersion

	newToken, changedAt, err := svc.ChangePassword(user.ID, "admin123", "new-password-1")
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if newToken == "" || changedAt.IsZero() {
		t.Fatalf("expected token & timestamp")
	}
	reloaded, _ := userRepo.FindByID(user.ID)
	if reloaded.TokenVersion != originalVersion+1 {
		t.Fatalf("token version not bumped: %d -> %d", originalVersion, reloaded.TokenVersion)
	}
	if reloaded.PasswordChangedAt == nil {
		t.Fatalf("PasswordChangedAt not set")
	}
}

func TestChangePasswordWrongOld(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	svc := newAuthSvc(t, userRepo, tokenRepo)
	_ = svc.EnsureAdmin("admin", "admin123")
	user, _ := userRepo.FindByUsername("admin")

	if _, _, err := svc.ChangePassword(user.ID, "wrong", "new-password-1"); !errors.Is(err, service.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestChangePasswordTooShort(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	svc := newAuthSvc(t, userRepo, tokenRepo)
	_ = svc.EnsureAdmin("admin", "admin123")
	user, _ := userRepo.FindByUsername("admin")

	if _, _, err := svc.ChangePassword(user.ID, "admin123", "short"); !errors.Is(err, service.ErrPasswordTooShort) {
		t.Fatalf("expected ErrPasswordTooShort, got %v", err)
	}
}

func TestChangePasswordSameAsOld(t *testing.T) {
	db := newTestDB(t)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	svc := newAuthSvc(t, userRepo, tokenRepo)
	_ = svc.EnsureAdmin("admin", "admin123")
	user, _ := userRepo.FindByUsername("admin")

	if _, _, err := svc.ChangePassword(user.ID, "admin123", "admin123"); !errors.Is(err, service.ErrPasswordSameAsOld) {
		t.Fatalf("expected ErrPasswordSameAsOld, got %v", err)
	}
}
```

如果 `auth_test.go` 没有 `newTestDB`，在文件顶部新增：

```go
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil { t.Fatalf("open: %v", err) }
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil { t.Fatalf("migrate: %v", err) }
	return db
}
```

并在 import 加 `"github.com/glebarez/sqlite"` `"gorm.io/gorm"` `"cloudalbum/internal/model"`。

- [ ] **Step 3: 运行测试验证失败**

```bash
go test ./internal/service/... -run ChangePassword -v
```
Expected: FAIL，`svc.ChangePassword undefined`。

- [ ] **Step 4: 实现 ChangePassword**

`internal/service/auth.go` 追加：

```go
const minPasswordLen = 8

// ChangePassword 校验旧密码并更新密码 hash；同事务内 bump TokenVersion 并写 PasswordChangedAt。
// 返回新签发的 JWT（避免改密用户被自己吊销）和变更时间。
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) (string, time.Time, error) {
	if len(newPassword) < minPasswordLen {
		return "", time.Time{}, ErrPasswordTooShort
	}
	if newPassword == oldPassword {
		return "", time.Time{}, ErrPasswordSameAsOld
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", time.Time{}, ErrInvalidCredentials
		}
		return "", time.Time{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return "", time.Time{}, ErrInvalidCredentials
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", time.Time{}, err
	}

	changedAt := time.Now()
	if err := s.userRepo.UpdatePasswordAndBumpVersion(userID, string(newHash), changedAt); err != nil {
		return "", time.Time{}, err
	}

	reloaded, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", time.Time{}, err
	}
	token, err := s.GenerateJWT(reloaded)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, changedAt, nil
}
```

- [ ] **Step 5: 运行测试验证通过**

```bash
go test ./internal/service/... -run ChangePassword -v
```
Expected: PASS。

- [ ] **Step 6: 检查全包不回归**

```bash
go test ./internal/service/... -v
```
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/service/auth.go internal/service/auth_test.go
git commit -m "feat(auth): add ChangePassword with token_version bump"
```

---

### Task 11: `AuthMiddleware` 增加 token_version 比对

**Files:**
- Modify: `internal/middleware/auth.go`
- Create: `internal/middleware/auth_test.go`（如无则新建）

- [ ] **Step 1: 写失败测试**

`internal/middleware/auth_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudalbum/internal/config"
	"cloudalbum/internal/middleware"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setup(t *testing.T) (*gin.Engine, *service.AuthService, *service.TokenService, *repository.UserRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil { t.Fatalf("db: %v", err) }
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil { t.Fatalf("migrate: %v", err) }
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	base := config.Config{Auth: config.AuthConfig{JWTSecret: "s", TokenExpire: 3600 * 1e9}}
	provider := config.NewProvider(base, config.Overrides{})
	authSvc := service.NewAuthService(userRepo, tokenRepo, provider)
	tokenSvc := service.NewTokenService(tokenRepo)

	r := gin.New()
	r.Use(middleware.AuthMiddleware(authSvc, tokenSvc))
	r.GET("/x", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r, authSvc, tokenSvc, userRepo
}

func TestJWTRejectsStaleTokenVersion(t *testing.T) {
	r, authSvc, _, userRepo := setup(t)
	if err := authSvc.EnsureAdmin("admin", "admin123"); err != nil { t.Fatalf("seed: %v", err) }
	user, _ := userRepo.FindByUsername("admin")
	staleToken, _ := authSvc.GenerateJWT(user)

	// bump TokenVersion
	if err := userRepo.UpdatePasswordAndBumpVersion(user.ID, user.PasswordHash, timeNow()); err != nil {
		t.Fatalf("bump: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+staleToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
```

辅助：在文件中加 `import "time"` 和 `func timeNow() time.Time { return time.Now() }` 或直接使用 `time.Now()`。

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/middleware/... -run JWT -v
```
Expected: FAIL（middleware 还没做 token_version 比对，旧 token 仍可通行）。

- [ ] **Step 3: 修改 middleware**

`internal/middleware/auth.go` 中 JWT 分支：

```go
claims, err := authSvc.ParseJWT(tokenStr)
if err != nil {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	return
}

user, err := authSvc.LookupUser(claims.UserID)
if err != nil || user.TokenVersion != claims.TokenVersion {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
	return
}

c.Set("user_id", claims.UserID)
c.Set("auth_type", "jwt")
c.Set("username", claims.Username)
c.Next()
```

`AuthService` 加方法：

```go
func (s *AuthService) LookupUser(id uint) (*model.User, error) {
	return s.userRepo.FindByID(id)
}
```

- [ ] **Step 4: 运行测试验证通过**

```bash
go test ./internal/middleware/... -v
```
Expected: PASS。

- [ ] **Step 5: 跑现有所有测试**

```bash
go test ./internal/... -v
```
Expected: 所有现有 JWT 用例仍 PASS（新签发 token 自动带最新 TokenVersion）。

- [ ] **Step 6: 提交**

```bash
git add internal/middleware/auth.go internal/middleware/auth_test.go internal/service/auth.go
git commit -m "feat(middleware): enforce token_version match on JWT"
```

---

### Task 12: ChangePassword Handler / Me 扩展 / 路由挂载

**Files:**
- Modify: `internal/handler/auth.go`
- Modify: `internal/router/router.go`
- Create: `internal/handler/auth_test.go`（若不存在）

- [ ] **Step 1: 修改 Handler 构造（注入 userRepo 用于 Me 扩展）**

`internal/handler/auth.go`:

```go
type AuthHandler struct {
	authSvc  *service.AuthService
	userRepo *repository.UserRepository
}

func NewAuthHandler(authSvc *service.AuthService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, userRepo: userRepo}
}
```

`main.go` 调用点同步：`authHandler := handler.NewAuthHandler(authSvc, userRepo)`。

- [ ] **Step 2: 实现 ChangePassword endpoint**

```go
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	if c.GetString("auth_type") != "jwt" {
		c.JSON(http.StatusForbidden, gin.H{"error": "api_token_forbidden"})
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	userID := c.GetUint("user_id")
	token, changedAt, err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong_old_password"})
		case errors.Is(err, service.ErrPasswordTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": "min 8 characters"})
		case errors.Is(err, service.ErrPasswordSameAsOld):
			c.JSON(http.StatusBadRequest, gin.H{"error": "same_as_old"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "change_password_failed"})
		}
		log.Printf("[settings] password change failed for user_id=%d ip=%s reason=%v", userID, c.ClientIP(), err)
		return
	}
	log.Printf("[settings] password changed for user_id=%d ip=%s ts=%s", userID, c.ClientIP(), changedAt.UTC().Format(time.RFC3339))
	c.JSON(http.StatusOK, gin.H{
		"token":               token,
		"password_changed_at": changedAt,
	})
}
```

确保 `auth.go` import 中包含 `"log"`、`"time"`、`"errors"`、`"cloudalbum/internal/repository"`。

- [ ] **Step 3: 扩展 Me**

替换现有 `Me`:

```go
const defaultAdminPassword = "admin123"

func (h *AuthHandler) Me(c *gin.Context) {
	response := gin.H{
		"user_id": c.GetUint("user_id"),
	}
	if username := c.GetString("username"); username != "" {
		response["username"] = username
	}
	if authType := c.GetString("auth_type"); authType != "" {
		response["auth_type"] = authType
	}
	if tokenScope := c.GetString("token_scope"); tokenScope != "" {
		response["token_scope"] = tokenScope
	}

	authType := c.GetString("auth_type")
	if authType == "jwt" {
		if user, err := h.userRepo.FindByID(c.GetUint("user_id")); err == nil {
			response["created_at"] = user.CreatedAt
			response["password_changed_at"] = user.PasswordChangedAt
			if user.Username == "admin" {
				if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(defaultAdminPassword)) == nil {
					response["uses_default_password"] = true
				} else {
					response["uses_default_password"] = false
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}
```

确保 import 包含 `"golang.org/x/crypto/bcrypt"`。

- [ ] **Step 4: 挂载路由**

`internal/router/router.go` 中 `api.GET("/auth/me", authHandler.Me)` 之后加：

```go
api.POST("/auth/change-password", authHandler.ChangePassword)
```

注意：endpoint 在 `auth` group（已 AuthMiddleware）下，确保需要登录态。

- [ ] **Step 5: 写 handler 测试**

`internal/handler/auth_test.go`（追加或新建）：

```go
package handler_test

// 测试 4 个场景：
// 1. JWT 改密成功 → 200，新 JWT 与旧 JWT TokenVersion 不同（手段：用旧 JWT 调任意接口应 401）
// 2. JWT 改密 old wrong → 401
// 3. JWT 改密 new 7 位 → 400
// 4. API Token 调 change-password → 403
```

落地完整代码：

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudalbum/internal/config"
	"cloudalbum/internal/handler"
	"cloudalbum/internal/middleware"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setup(t *testing.T) (*gin.Engine, *service.AuthService, *service.TokenService, *repository.UserRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil { t.Fatalf("db: %v", err) }
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil { t.Fatalf("migrate: %v", err) }
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	base := config.Config{Auth: config.AuthConfig{JWTSecret: "s", TokenExpire: 3600 * 1e9}}
	provider := config.NewProvider(base, config.Overrides{})
	authSvc := service.NewAuthService(userRepo, tokenRepo, provider)
	tokenSvc := service.NewTokenService(tokenRepo)
	authHandler := handler.NewAuthHandler(authSvc, userRepo)

	r := gin.New()
	api := r.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)
	authed := api.Group("")
	authed.Use(middleware.AuthMiddleware(authSvc, tokenSvc))
	authed.POST("/auth/change-password", authHandler.ChangePassword)
	authed.GET("/auth/me", authHandler.Me)
	return r, authSvc, tokenSvc, userRepo
}

func loginJWT(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("login: %d %s", w.Code, w.Body.String()) }
	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["token"]
}

func TestChangePasswordHappy(t *testing.T) {
	r, authSvc, _, _ := setup(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)

	body, _ := json.Marshal(map[string]string{"old_password": "admin123", "new_password": "abcd1234"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("change: %d %s", w.Code, w.Body.String()) }

	// 老 token 调 me 应 401
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	if w2.Code != 401 { t.Fatalf("expected old token to be 401, got %d", w2.Code) }
}

func TestChangePasswordWrongOld(t *testing.T) {
	r, authSvc, _, _ := setup(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)
	body, _ := json.Marshal(map[string]string{"old_password": "wrong", "new_password": "abcd1234"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 401 { t.Fatalf("expected 401, got %d", w.Code) }
}

func TestChangePasswordTooShort(t *testing.T) {
	r, authSvc, _, _ := setup(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)
	body, _ := json.Marshal(map[string]string{"old_password": "admin123", "new_password": "short1"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 { t.Fatalf("expected 400, got %d", w.Code) }
}

func TestChangePasswordViaAPIToken(t *testing.T) {
	r, authSvc, tokenSvc, _ := setup(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	// 创建 API Token
	user, _ := authSvc.LookupUser(1)
	apiToken, _, err := tokenSvc.Create(user.ID, "test", "upload")
	if err != nil { t.Fatalf("token create: %v", err) }

	body, _ := json.Marshal(map[string]string{"old_password": "admin123", "new_password": "abcd1234"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/change-password", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 403 { t.Fatalf("expected 403, got %d", w.Code) }
}
```

注：`tokenSvc.Create` 的具体签名需依实际 `service/token.go` 调整；若签名不同，按当前文件调整调用方式。本任务测试不深挖 API Token 创建逻辑。

- [ ] **Step 6: 运行测试**

```bash
go test ./internal/handler/... -run Change -v
```
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/handler/auth.go internal/handler/auth_test.go internal/router/router.go main.go
git commit -m "feat(auth): change-password endpoint and extend /auth/me"
```

---

## Phase D — Settings 接口

### Task 13: `SettingsService` 白名单 / 校验 / merge / Apply

**Files:**
- Create: `internal/service/settings.go`
- Create: `internal/service/settings_test.go`

- [ ] **Step 1: 写失败测试**

`internal/service/settings_test.go`:

```go
package service_test

import (
	"errors"
	"testing"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newSettingsSvc(t *testing.T) (*service.SettingsService, *config.Provider) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil { t.Fatalf("db: %v", err) }
	if err := db.AutoMigrate(&model.Settings{}); err != nil { t.Fatalf("migrate: %v", err) }
	repo := repository.NewSettingsRepository(db)
	_, _ = repo.LoadOrBootstrap()
	base := config.Config{
		Server: config.ServerConfig{BaseURL: "http://localhost:8080"},
		Image: config.ImageConfig{
			MaxSize: 50 << 20, AllowedTypes: []string{"jpg","png"},
			AutoConvert: "webp", Quality: 85, StripExif: true,
		},
	}
	provider := config.NewProvider(base, config.Overrides{})
	return service.NewSettingsService(repo, provider), provider
}

func TestSettingsUpdateHappy(t *testing.T) {
	svc, provider := newSettingsSvc(t)
	url := "https://img.example.com"
	q := 90
	input := config.Overrides{}
	input.Server.BaseURL = &url
	input.Image.Quality = &q

	if err := svc.Update(input, 1); err != nil {
		t.Fatalf("update: %v", err)
	}
	if provider.Get().Server.BaseURL != url {
		t.Fatalf("provider not updated")
	}
	if provider.Get().Image.Quality != 90 {
		t.Fatalf("quality not updated")
	}
}

func TestSettingsValidateBaseURL(t *testing.T) {
	svc, _ := newSettingsSvc(t)
	bad := "not-a-url"
	input := config.Overrides{}
	input.Server.BaseURL = &bad
	if err := svc.Update(input, 1); !errors.Is(err, service.ErrInvalidSetting) {
		t.Fatalf("expected ErrInvalidSetting, got %v", err)
	}
}

func TestSettingsValidateQuality(t *testing.T) {
	svc, _ := newSettingsSvc(t)
	q := 200
	input := config.Overrides{}
	input.Image.Quality = &q
	if err := svc.Update(input, 1); !errors.Is(err, service.ErrInvalidSetting) {
		t.Fatalf("expected ErrInvalidSetting, got %v", err)
	}
}

func TestSettingsValidateAllowedTypes(t *testing.T) {
	svc, _ := newSettingsSvc(t)
	bad := []string{"exe", "png"}
	input := config.Overrides{}
	input.Image.AllowedTypes = &bad
	if err := svc.Update(input, 1); !errors.Is(err, service.ErrInvalidSetting) {
		t.Fatalf("expected ErrInvalidSetting, got %v", err)
	}
}

func TestSettingsGetReflectsEffectiveAndOverrides(t *testing.T) {
	svc, _ := newSettingsSvc(t)
	q := 90
	input := config.Overrides{}
	input.Image.Quality = &q
	_ = svc.Update(input, 1)

	snap := svc.Snapshot()
	if snap.Effective.Image.Quality != 90 {
		t.Fatalf("effective: %d", snap.Effective.Image.Quality)
	}
	if snap.Overrides.Image.Quality == nil || *snap.Overrides.Image.Quality != 90 {
		t.Fatalf("overrides quality nil")
	}
	if snap.Overrides.Server.BaseURL != nil {
		t.Fatalf("base_url should not be marked as override")
	}
}
```

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/service/... -run Settings -v
```
Expected: FAIL，`undefined: NewSettingsService`。

- [ ] **Step 3: 实现 SettingsService**

`internal/service/settings.go`:

```go
package service

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"cloudalbum/internal/config"
	"cloudalbum/internal/repository"
)

var ErrInvalidSetting = errors.New("invalid setting")

var supportedTypes = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true,
	"webp": true, "bmp": true, "svg": true,
}

// SettingsSnapshot 包含 UI 渲染所需的全部信息：当前生效值、显式 overrides、白名单。
type SettingsSnapshot struct {
	Effective      EffectiveSettings `json:"effective"`
	Overrides      OverrideFlags     `json:"overrides"`
	EditableFields []string          `json:"editable_fields"`
}

type EffectiveSettings struct {
	Server struct {
		BaseURL string `json:"base_url"`
	} `json:"server"`
	Image struct {
		MaxSize      int64    `json:"max_size"`
		AllowedTypes []string `json:"allowed_types"`
		AutoConvert  string   `json:"auto_convert"`
		Quality      int      `json:"quality"`
		StripExif    bool     `json:"strip_exif"`
	} `json:"image"`
}

type OverrideFlags struct {
	Server struct {
		BaseURL bool `json:"base_url,omitempty"`
	} `json:"server"`
	Image struct {
		MaxSize      bool `json:"max_size,omitempty"`
		AllowedTypes bool `json:"allowed_types,omitempty"`
		AutoConvert  bool `json:"auto_convert,omitempty"`
		Quality      bool `json:"quality,omitempty"`
		StripExif    bool `json:"strip_exif,omitempty"`
	} `json:"image"`
}

var editableFields = []string{
	"server.base_url",
	"image.max_size",
	"image.allowed_types",
	"image.auto_convert",
	"image.quality",
	"image.strip_exif",
}

type SettingsService struct {
	repo     *repository.SettingsRepository
	provider *config.Provider
	mu       sync.Mutex
}

func NewSettingsService(repo *repository.SettingsRepository, provider *config.Provider) *SettingsService {
	return &SettingsService{repo: repo, provider: provider}
}

// Snapshot 返回当前生效 + overrides + 白名单，用于 GET。
func (s *SettingsService) Snapshot() SettingsSnapshot {
	cfg := s.provider.Get()
	o := s.provider.Overrides()

	snap := SettingsSnapshot{EditableFields: editableFields}
	snap.Effective.Server.BaseURL = cfg.Server.BaseURL
	snap.Effective.Image.MaxSize = cfg.Image.MaxSize
	snap.Effective.Image.AllowedTypes = cfg.Image.AllowedTypes
	snap.Effective.Image.AutoConvert = cfg.Image.AutoConvert
	snap.Effective.Image.Quality = cfg.Image.Quality
	snap.Effective.Image.StripExif = cfg.Image.StripExif
	if o != nil {
		snap.Overrides.Server.BaseURL = o.Server.BaseURL != nil
		snap.Overrides.Image.MaxSize = o.Image.MaxSize != nil
		snap.Overrides.Image.AllowedTypes = o.Image.AllowedTypes != nil
		snap.Overrides.Image.AutoConvert = o.Image.AutoConvert != nil
		snap.Overrides.Image.Quality = o.Image.Quality != nil
		snap.Overrides.Image.StripExif = o.Image.StripExif != nil
	}
	return snap
}

// Update 校验入参 → 与当前 overrides merge → 持久化 → 通知 Provider 应用 → 全程互斥。
func (s *SettingsService) Update(input config.Overrides, userID uint) error {
	if err := validate(input); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current, err := s.repo.LoadOrBootstrap()
	if err != nil {
		return fmt.Errorf("load current: %w", err)
	}
	merged := mergeOverrides(current, input)
	if err := s.repo.Save(merged, userID); err != nil {
		return fmt.Errorf("persist settings: %w", err)
	}
	s.provider.Apply(merged)
	return nil
}

func mergeOverrides(base, patch config.Overrides) config.Overrides {
	out := base
	if patch.Server.BaseURL != nil { out.Server.BaseURL = patch.Server.BaseURL }
	if patch.Image.MaxSize != nil { out.Image.MaxSize = patch.Image.MaxSize }
	if patch.Image.AllowedTypes != nil { out.Image.AllowedTypes = patch.Image.AllowedTypes }
	if patch.Image.AutoConvert != nil { out.Image.AutoConvert = patch.Image.AutoConvert }
	if patch.Image.Quality != nil { out.Image.Quality = patch.Image.Quality }
	if patch.Image.StripExif != nil { out.Image.StripExif = patch.Image.StripExif }
	return out
}

func validate(o config.Overrides) error {
	if o.Server.BaseURL != nil {
		u, err := url.Parse(*o.Server.BaseURL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("%w: server.base_url must be http/https URL", ErrInvalidSetting)
		}
	}
	if o.Image.MaxSize != nil {
		if *o.Image.MaxSize <= 0 || *o.Image.MaxSize > (1<<30) {
			return fmt.Errorf("%w: image.max_size must be in (0, 1 GiB]", ErrInvalidSetting)
		}
	}
	if o.Image.AllowedTypes != nil {
		if len(*o.Image.AllowedTypes) == 0 {
			return fmt.Errorf("%w: image.allowed_types must be non-empty", ErrInvalidSetting)
		}
		seen := make(map[string]bool)
		for _, raw := range *o.Image.AllowedTypes {
			t := strings.ToLower(strings.TrimSpace(raw))
			if !supportedTypes[t] {
				return fmt.Errorf("%w: image.allowed_types contains unsupported %q", ErrInvalidSetting, raw)
			}
			seen[t] = true
		}
	}
	if o.Image.AutoConvert != nil {
		v := strings.ToLower(*o.Image.AutoConvert)
		if v != "" && v != "webp" && v != "jpeg" && v != "jpg" {
			return fmt.Errorf("%w: image.auto_convert must be one of '', 'webp', 'jpeg'", ErrInvalidSetting)
		}
	}
	if o.Image.Quality != nil {
		if *o.Image.Quality < 1 || *o.Image.Quality > 100 {
			return fmt.Errorf("%w: image.quality must be in [1, 100]", ErrInvalidSetting)
		}
	}
	return nil
}
```

- [ ] **Step 4: 运行测试**

```bash
go test ./internal/service/... -run Settings -v
```
Expected: PASS。

- [ ] **Step 5: 提交**

```bash
git add internal/service/settings.go internal/service/settings_test.go
git commit -m "feat(settings): add SettingsService with validation and merge"
```

---

### Task 14: `SettingsHandler` GET / PUT + 路由 + 白名单字段拦截

**Files:**
- Create: `internal/handler/settings.go`
- Create: `internal/handler/settings_test.go`
- Modify: `internal/router/router.go`
- Modify: `main.go`

- [ ] **Step 1: 写失败测试**

`internal/handler/settings_test.go`:

```go
package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupSettings(t *testing.T) (*gin.Engine, *service.AuthService) { /* 类似 setup, 但要挂载 settings 路由 */ }

func TestSettingsGetJWT(t *testing.T) {
	r, authSvc := setupSettings(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("got %d", w.Code) }
	if !strings.Contains(w.Body.String(), "editable_fields") {
		t.Fatalf("missing editable_fields: %s", w.Body.String())
	}
}

func TestSettingsPutHappy(t *testing.T) {
	r, authSvc := setupSettings(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)
	body, _ := json.Marshal(map[string]any{
		"image": map[string]any{"quality": 90},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 { t.Fatalf("got %d: %s", w.Code, w.Body.String()) }
}

func TestSettingsPutUnknownField(t *testing.T) {
	r, authSvc := setupSettings(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	token := loginJWT(t, r)
	body := []byte(`{"database":{"driver":"postgres"}}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/settings", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 400 { t.Fatalf("expected 400, got %d", w.Code) }
	if !strings.Contains(w.Body.String(), "unknown_field") {
		t.Fatalf("expected unknown_field error: %s", w.Body.String())
	}
}

func TestSettingsPutAPITokenForbidden(t *testing.T) {
	r, authSvc := setupSettings(t)
	_ = authSvc.EnsureAdmin("admin", "admin123")
	// 模拟一个 API Token：直接通过 TokenService.Create
	// ... 调用，得到 apiToken
	// PUT /settings with Bearer <apiToken>
	// 期望 403 api_token_forbidden
}
```

测试中的 `setupSettings` 仿 Task 12 的 `setup`，额外构造 `SettingsHandler` 并挂载 `settings.GET/PUT`。需要构造 `SettingsRepository` 与 `SettingsService`。

- [ ] **Step 2: 运行测试验证失败**

```bash
go test ./internal/handler/... -run Settings -v
```
Expected: FAIL（handler 未实现）。

- [ ] **Step 3: 实现 Handler**

`internal/handler/settings.go`:

```go
package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"cloudalbum/internal/config"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	svc *service.SettingsService
}

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) Get(c *gin.Context) {
	c.JSON(http.StatusOK, h.svc.Snapshot())
}

func (h *SettingsHandler) Update(c *gin.Context) {
	if c.GetString("auth_type") != "jwt" {
		c.JSON(http.StatusForbidden, gin.H{"error": "api_token_forbidden"})
		return
	}

	raw, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	if field, ok := containsUnknownField(raw); !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_field", "field": field})
		return
	}

	var input config.Overrides
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown_field", "detail": err.Error()})
		return
	}

	if err := h.svc.Update(input, c.GetUint("user_id")); err != nil {
		if errors.Is(err, service.ErrInvalidSetting) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_value", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "settings_persist_failed"})
		return
	}
	c.JSON(http.StatusOK, h.svc.Snapshot())
}

// containsUnknownField 检查 raw JSON 顶层 keys 是否全在白名单 {"server","image"} 内，
// 且 server / image 子对象内的 keys 是否全在嵌套白名单内。
func containsUnknownField(raw []byte) (string, bool) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(raw, &top); err != nil {
		return "_root", false
	}
	allowedTop := map[string]map[string]bool{
		"server": {"base_url": true},
		"image":  {"max_size": true, "allowed_types": true, "auto_convert": true, "quality": true, "strip_exif": true},
	}
	for key, val := range top {
		nested, ok := allowedTop[key]
		if !ok {
			return key, false
		}
		var sub map[string]json.RawMessage
		if err := json.Unmarshal(val, &sub); err != nil {
			return key, false
		}
		for k := range sub {
			if !nested[k] {
				return key + "." + k, false
			}
		}
	}
	return "", true
}
```

- [ ] **Step 4: 挂载路由**

`internal/router/router.go` 改：

```go
settings := api.Group("/settings")
settings.GET("", settingsHandler.Get)
settings.PUT("", settingsHandler.Update)
```

`Setup` 函数参数列表新增 `settingsHandler *handler.SettingsHandler`，在末尾追加。

`main.go` 构造：

```go
settingsSvc := service.NewSettingsService(settingsRepo, provider)
settingsHandler := handler.NewSettingsHandler(settingsSvc)
```

并把 `settingsHandler` 传给 `router.Setup`。

- [ ] **Step 5: 运行测试**

```bash
go test ./internal/handler/... -run Settings -v
```
Expected: PASS。

- [ ] **Step 6: 全包回归**

```bash
go test ./... -v
```
Expected: PASS。

- [ ] **Step 7: 提交**

```bash
git add internal/handler/settings.go internal/handler/settings_test.go internal/router/router.go main.go
git commit -m "feat(settings): GET/PUT /api/v1/settings handlers and routes"
```

---

## Phase E — 前端

### Task 15: 扩展 `auth store`：me 缓存 / 改密续签 / Banner 关闭态

**Files:**
- Modify: `web/src/stores/auth.ts`

- [ ] **Step 1: 改写 auth store**

替换 `web/src/stores/auth.ts` 为：

```ts
import { create } from 'zustand'
import client from '../api/client'

export type MeInfo = {
  user_id: number
  username?: string
  auth_type?: string
  token_scope?: string
  created_at?: string
  password_changed_at?: string | null
  uses_default_password?: boolean
}

type AuthState = {
  token: string | null
  loggedIn: boolean
  me: MeInfo | null
  bannerDismissed: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  reset: () => void
  init: () => Promise<void>
  refreshMe: () => Promise<void>
  applyNewToken: (token: string) => Promise<void>
  dismissBanner: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  loggedIn: Boolean(localStorage.getItem('token')),
  me: null,
  bannerDismissed: false,

  async login(username: string, password: string) {
    const response = await client.post('/auth/login', { username, password })
    const token = response.data.token as string
    localStorage.setItem('token', token)
    set({ token, loggedIn: true, bannerDismissed: false })
    await get().refreshMe()
  },

  logout() {
    localStorage.removeItem('token')
    set({ token: null, loggedIn: false, me: null, bannerDismissed: false })
  },

  reset() {
    set({ token: null, loggedIn: false, me: null, bannerDismissed: false })
  },

  async init() {
    const token = localStorage.getItem('token')
    set({ token, loggedIn: Boolean(token) })
    if (token) {
      await get().refreshMe()
    }
  },

  async refreshMe() {
    try {
      const { data } = await client.get('/auth/me')
      set({ me: data as MeInfo })
    } catch {
      // 401 被全局拦截器处理；其他错误忽略，me 保持当前值
    }
  },

  async applyNewToken(token: string) {
    localStorage.setItem('token', token)
    set({ token })
    await get().refreshMe()
  },

  dismissBanner() {
    set({ bannerDismissed: true })
  },
}))
```

- [ ] **Step 2: 在 main.tsx 启动时调 init**

定位 `web/src/main.tsx`（或入口文件）；如果当前没有调用 `useAuthStore.getState().init()`，在入口处加：

```ts
import { useAuthStore } from './stores/auth'
useAuthStore.getState().init()
```

如果入口已有 init 调用（之前是同步版本），改为不 await（异步执行）。

- [ ] **Step 3: 启动前端检查编译**

```bash
cd web && npm run build && cd -
```
Expected: 构建成功。如果 TypeScript 报类型错，按报错位置补类型。

- [ ] **Step 4: 提交**

```bash
git add web/src/stores/auth.ts web/src/main.tsx
git commit -m "feat(web): cache me, support new-token apply and banner dismiss in auth store"
```

---

### Task 16: 顶部默认密码 Banner 组件

**Files:**
- Create: `web/src/components/DefaultPasswordBanner.tsx`
- Modify: `web/src/components/Layout.tsx`

- [ ] **Step 1: 实现 Banner**

`web/src/components/DefaultPasswordBanner.tsx`:

```tsx
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'

export default function DefaultPasswordBanner() {
  const navigate = useNavigate()
  const usesDefault = useAuthStore((s) => s.me?.uses_default_password === true)
  const dismissed = useAuthStore((s) => s.bannerDismissed)
  const dismiss = useAuthStore((s) => s.dismissBanner)

  if (!usesDefault || dismissed) return null

  return (
    <div className="default-password-banner" role="alert">
      <span className="default-password-banner-icon">⚠️</span>
      <span className="default-password-banner-text">
        你正在使用默认密码 admin123，建议尽快修改。
      </span>
      <div className="default-password-banner-actions">
        <button
          type="button"
          className="default-password-banner-primary"
          onClick={() => navigate('/account', { state: { focusCurrentPassword: true } })}
        >
          立刻修改 →
        </button>
        <button
          type="button"
          className="default-password-banner-secondary"
          onClick={dismiss}
        >
          稍后
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 加最小样式**

在 `web/src/index.css` 追加：

```css
.default-password-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  margin-bottom: 16px;
  background: linear-gradient(90deg, rgba(255, 196, 0, 0.18), rgba(255, 122, 0, 0.18));
  border: 1px solid rgba(255, 159, 0, 0.45);
  border-radius: 12px;
  color: #b15a00;
  font-size: 13px;
}
.default-password-banner-icon { font-size: 16px; }
.default-password-banner-text { flex: 1; }
.default-password-banner-actions { display: flex; gap: 8px; }
.default-password-banner-primary,
.default-password-banner-secondary {
  border-radius: 999px;
  padding: 4px 12px;
  font-size: 12px;
  cursor: pointer;
  border: 1px solid rgba(255, 159, 0, 0.55);
  background: rgba(255, 255, 255, 0.5);
}
.default-password-banner-primary { background: rgba(255, 159, 0, 0.9); color: white; border: none; }
```

- [ ] **Step 3: 在 Layout 顶部挂载**

`web/src/components/Layout.tsx`：

import 顶部追加：

```ts
import DefaultPasswordBanner from './DefaultPasswordBanner'
```

将 `<div className="dashboard-content">` 之前插入 Banner（topbar 下方、内容区上方）：

```tsx
<div className="dashboard-content">
  <DefaultPasswordBanner />
  <Outlet />
</div>
```

- [ ] **Step 4: 编译**

```bash
cd web && npm run build && cd -
```
Expected: 通过。

- [ ] **Step 5: 提交**

```bash
git add web/src/components/DefaultPasswordBanner.tsx web/src/components/Layout.tsx web/src/index.css
git commit -m "feat(web): add default-password banner in shared layout"
```

---

### Task 17: 新建 Account 页

**Files:**
- Create: `web/src/pages/Account.tsx`

- [ ] **Step 1: 实现 Account 页**

`web/src/pages/Account.tsx`:

```tsx
import { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Button, Form, Input, Message, Typography } from '@arco-design/web-react'
import client from '../api/client'
import { useAuthStore } from '../stores/auth'

const { Title } = Typography

export default function Account() {
  const location = useLocation()
  const me = useAuthStore((s) => s.me)
  const applyNewToken = useAuthStore((s) => s.applyNewToken)
  const refreshMe = useAuthStore((s) => s.refreshMe)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const oldInputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    if (!me) {
      refreshMe()
    }
  }, [me, refreshMe])

  useEffect(() => {
    if ((location.state as { focusCurrentPassword?: boolean } | null)?.focusCurrentPassword) {
      oldInputRef.current?.focus()
    }
  }, [location.state])

  const validate = (): string => {
    if (!oldPassword) return '请输入当前密码'
    if (newPassword.length < 8) return '新密码至少 8 位'
    if (newPassword === oldPassword) return '新密码不能与当前密码相同'
    if (newPassword !== confirmPassword) return '两次输入的新密码不一致'
    return ''
  }

  const handleSubmit = async () => {
    const v = validate()
    if (v) { setError(v); return }
    setSubmitting(true)
    setError('')
    try {
      const { data } = await client.post('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      })
      await applyNewToken(data.token as string)
      Message.success('密码已修改')
      setOldPassword(''); setNewPassword(''); setConfirmPassword('')
    } catch (err: any) {
      const code = err?.response?.data?.error
      if (code === 'wrong_old_password') setError('当前密码错误')
      else if (code === 'same_as_old') setError('新密码不能与当前密码相同')
      else if (code === 'invalid_request') setError('密码格式不符合要求（至少 8 位）')
      else setError('修改失败，请稍后再试')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Account</div>
        <Title heading={4} className="section-title">账号信息</Title>
        <dl className="account-meta-grid">
          <div><dt>用户名</dt><dd>{me?.username ?? '-'}</dd></div>
          <div><dt>角色</dt><dd>{me?.auth_type === 'jwt' ? 'admin' : '-'}</dd></div>
          <div><dt>创建时间</dt><dd>{me?.created_at ?? '-'}</dd></div>
          <div><dt>上次改密</dt><dd>{me?.password_changed_at ?? '从未'}</dd></div>
        </dl>
      </section>

      <section className="glass-panel management-form-panel">
        <Title heading={4} className="section-title">修改密码</Title>
        <Form layout="vertical" onSubmit={handleSubmit}>
          <Form.Item label="当前密码">
            <Input.Password
              ref={(el) => { oldInputRef.current = (el as unknown as HTMLInputElement) ?? null }}
              value={oldPassword}
              onChange={setOldPassword}
              placeholder="请输入当前密码"
            />
          </Form.Item>
          <Form.Item label="新密码（≥ 8 位）">
            <Input.Password value={newPassword} onChange={setNewPassword} placeholder="新密码" />
          </Form.Item>
          <Form.Item label="确认新密码">
            <Input.Password value={confirmPassword} onChange={setConfirmPassword} placeholder="再次输入新密码" />
          </Form.Item>
          {error ? <div className="form-error">{error}</div> : null}
          <Button type="primary" htmlType="submit" loading={submitting}>保存</Button>
        </Form>
      </section>
    </div>
  )
}
```

- [ ] **Step 2: 加小样式**

`web/src/index.css` 追加：

```css
.account-meta-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px 24px; margin: 12px 0 0; }
.account-meta-grid > div { display: flex; flex-direction: column; gap: 4px; }
.account-meta-grid dt { font-size: 12px; color: rgba(0,0,0,0.55); }
.account-meta-grid dd { margin: 0; font-size: 14px; font-weight: 500; }
.form-error { margin: 8px 0; padding: 8px 12px; border-radius: 8px; background: rgba(255, 77, 79, 0.12); color: #c00; font-size: 13px; }
```

- [ ] **Step 3: 编译**

```bash
cd web && npm run build && cd -
```
Expected: 通过。

- [ ] **Step 4: 提交**

```bash
git add web/src/pages/Account.tsx web/src/index.css
git commit -m "feat(web): add Account page with change-password form"
```

---

### Task 18: 注册路由 + 侧边栏插入 Account

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/Layout.tsx`

- [ ] **Step 1: 加路由**

`web/src/App.tsx` import：

```ts
import Account from './pages/Account'
```

`<Route path="settings" ... />` 之前加：

```tsx
<Route path="account" element={<Account />} />
```

- [ ] **Step 2: 侧边栏插入 Account 入口**

`web/src/components/Layout.tsx`，import 增加：

```ts
import { IconUser } from '@arco-design/web-react/icon'
```

`navItems` 数组在 `{ to: '/settings', ... }` **之前**插入：

```ts
{ to: '/account', label: '账户', icon: <IconUser /> },
```

- [ ] **Step 3: 编译**

```bash
cd web && npm run build && cd -
```
Expected: 通过。

- [ ] **Step 4: 提交**

```bash
git add web/src/App.tsx web/src/components/Layout.tsx
git commit -m "feat(web): mount /account route and sidebar entry"
```

---

### Task 19: Settings 页改造为可编辑表单

**Files:**
- Modify: `web/src/pages/Settings.tsx`

- [ ] **Step 1: 改写 Settings 页**

替换 `web/src/pages/Settings.tsx`：

```tsx
import { useEffect, useState } from 'react'
import { Button, Checkbox, Form, Input, InputNumber, Message, Radio, Slider, Tag, Typography } from '@arco-design/web-react'
import client from '../api/client'

const { Title } = Typography

const SUPPORTED_TYPES = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg']

type SettingsSnapshot = {
  effective: {
    server: { base_url: string }
    image: {
      max_size: number
      allowed_types: string[]
      auto_convert: string
      quality: number
      strip_exif: boolean
    }
  }
  overrides: {
    server: { base_url?: boolean }
    image: {
      max_size?: boolean
      allowed_types?: boolean
      auto_convert?: boolean
      quality?: boolean
      strip_exif?: boolean
    }
  }
  editable_fields: string[]
}

export default function Settings() {
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [snapshot, setSnapshot] = useState<SettingsSnapshot | null>(null)
  const [baseURL, setBaseURL] = useState('')
  const [maxSizeMB, setMaxSizeMB] = useState<number>(50)
  const [allowedTypes, setAllowedTypes] = useState<string[]>([])
  const [autoConvert, setAutoConvert] = useState<string>('')
  const [quality, setQuality] = useState<number>(85)
  const [stripExif, setStripExif] = useState<boolean>(true)

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const { data } = await client.get<SettingsSnapshot>('/settings')
      setSnapshot(data)
      setBaseURL(data.effective.server.base_url)
      setMaxSizeMB(Math.round(data.effective.image.max_size / (1024 * 1024)))
      setAllowedTypes(data.effective.image.allowed_types)
      setAutoConvert(data.effective.image.auto_convert)
      setQuality(data.effective.image.quality)
      setStripExif(data.effective.image.strip_exif)
    } catch {
      setError('加载设置失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [])

  const validate = (): string => {
    if (!/^https?:\/\/.+/.test(baseURL)) return '站点 Base URL 必须是 http/https 链接'
    if (maxSizeMB <= 0 || maxSizeMB > 1024) return '最大大小需在 1–1024 MB 之间'
    if (allowedTypes.length === 0) return '至少选择一种图片格式'
    if (quality < 1 || quality > 100) return '压缩质量需在 1–100 之间'
    return ''
  }

  const handleSubmit = async () => {
    const v = validate()
    if (v) { setError(v); return }
    setSubmitting(true)
    setError('')
    try {
      const payload = {
        server: { base_url: baseURL },
        image: {
          max_size: maxSizeMB * 1024 * 1024,
          allowed_types: allowedTypes,
          auto_convert: autoConvert,
          quality,
          strip_exif: stripExif,
        },
      }
      await client.put('/settings', payload)
      Message.success('设置已保存')
      await load()
    } catch (err: any) {
      const code = err?.response?.data?.error
      if (code === 'unknown_field') setError(`存在未识别字段：${err.response.data.field ?? ''}`)
      else if (code === 'invalid_value') setError(err.response.data.detail ?? '字段值不合法')
      else setError('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <div className="management-page">加载中…</div>

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Settings</div>
        <Title heading={4} className="section-title">站点</Title>
        <Form layout="vertical">
          <Form.Item
            label={
              <span>
                Base URL
                {snapshot?.overrides.server.base_url ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}
              </span>
            }
            help="用于生成图片公开链接。立即生效。"
          >
            <Input value={baseURL} onChange={setBaseURL} placeholder="https://img.example.com" />
          </Form.Item>
        </Form>
      </section>

      <section className="glass-panel management-form-panel">
        <Title heading={4} className="section-title">图片处理</Title>
        <Form layout="vertical">
          <Form.Item label={<span>最大大小（MB）{snapshot?.overrides.image.max_size ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <InputNumber min={1} max={1024} value={maxSizeMB} onChange={(v) => setMaxSizeMB(Number(v) || 0)} />
          </Form.Item>
          <Form.Item label={<span>允许格式{snapshot?.overrides.image.allowed_types ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Checkbox.Group value={allowedTypes} onChange={setAllowedTypes}>
              {SUPPORTED_TYPES.map((t) => <Checkbox key={t} value={t}>{t}</Checkbox>)}
            </Checkbox.Group>
          </Form.Item>
          <Form.Item label={<span>自动转换{snapshot?.overrides.image.auto_convert ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Radio.Group value={autoConvert} onChange={setAutoConvert}>
              <Radio value="">不转换</Radio>
              <Radio value="webp">WebP</Radio>
              <Radio value="jpeg">JPEG</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item label={<span>压缩质量{snapshot?.overrides.image.quality ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Slider min={1} max={100} value={quality} onChange={(v) => setQuality(Number(v))} />
          </Form.Item>
          <Form.Item label={<span>EXIF 剥离{snapshot?.overrides.image.strip_exif ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Checkbox checked={stripExif} onChange={setStripExif}>移除 EXIF / 隐私元数据</Checkbox>
          </Form.Item>
        </Form>
      </section>

      {error ? <div className="form-error">{error}</div> : null}
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <Button onClick={load} disabled={submitting}>恢复未保存的修改</Button>
        <Button type="primary" loading={submitting} onClick={handleSubmit}>保存</Button>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: 编译**

```bash
cd web && npm run build && cd -
```
Expected: 通过。

- [ ] **Step 3: 提交**

```bash
git add web/src/pages/Settings.tsx
git commit -m "feat(web): make Settings page editable"
```

---

### Task 20: 端到端联调（开发模式启动 + 浏览器验收）

**Files:**
- 无新文件，纯人工验收 + 修复

- [ ] **Step 1: 启动后端**

```bash
go run . &
```
监听 `:8080`。

- [ ] **Step 2: 启动前端开发服务器**

```bash
cd web && npm run dev &
cd -
```
监听 Vite 默认端口（通常 5173），代理 `/api` 到 8080。

- [ ] **Step 3: 手动验收 happy 1（默认密码 → 改密 → 重新登录）**

- 浏览器打开 dev server URL
- 默认 admin/admin123 登录
- 确认顶部出现「默认密码 Banner」
- 点「立刻修改」→ 进入 `/account`
- 改密 `admin123` → `abcd1234` → 保存 → Banner 消失
- 退出登录 → 用 `abcd1234` 重新登录成功

- [ ] **Step 4: 手动验收 happy 2（base_url 立即生效）**

- 在 Settings 页改 base_url 为 `http://127.0.0.1:8080`（或别的合法 URL）→ 保存
- 上传任意图片 → 在图片管理页复制链接 → 检查链接前缀已变

- [ ] **Step 5: 手动验收 happy 3（quality 立即生效）**

- 在 Settings 改 quality 为 50 → 保存
- 上传任意大图 → 在图片管理页查看新图文件大小（应明显小于 quality=85 时）

- [ ] **Step 6: 手动验收 error 路径**

- 改密时填错旧密码 → 看到红条「当前密码错误」
- 改密时新密码 7 位 → 看到「至少 8 位」前端拦截
- 用 curl 试探：

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"abcd1234"}' | jq -r .token)
curl -s -X PUT http://localhost:8080/api/v1/settings \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"database":{"driver":"postgres"}}'
```
Expected: `{"error":"unknown_field","field":"database"}` 400。

- [ ] **Step 7: 停止服务并清理**

```bash
pkill -f "go run \." || true
pkill -f "vite" || true
```

- [ ] **Step 8: 不需要单独 commit（无文件改动）；继续 Phase F**

---

## Phase F — 文档与验收

### Task 21: README 增加"配置管理"章节

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 追加章节**

定位 `README.md` 中合适位置（在「配置」/「Configuration」之后或附近）追加：

```markdown
## 配置管理

- `configs/config.yaml` 仅作为**首次启动的种子**。第一次启动时如果 `settings` 表为空，
  CloudAlbum 会插入一行空 overrides；运行时配置最终生效值是
  `YAML 基线 ⊕ DB settings overrides`，以 DB 为权威源。
- 之后**直接修改 `config.yaml` 不再生效**。需要修改可热更新的配置（站点 base_url
  与图片处理策略），请通过后台「Settings」页编辑；保存后立即生效。
- 如果你确实希望让 yaml 的新值覆盖运行时：删除 `settings` 表中 `id=1` 的行（例如
  `sqlite3 data/cloudalbum.db 'DELETE FROM settings WHERE id=1;'`），重启进程即可。
- 启动时若 yaml mtime 比 `settings.updated_at` 晚，日志会打一条 INFO 提醒，但不会
  自动覆盖运行时。
- 数据库、对象存储、`server.port`、`auth.jwt_secret` 等不在可热更新范围内，仍以
  yaml / 环境变量为准。
```

- [ ] **Step 2: 提交**

```bash
git add README.md
git commit -m "docs: explain settings DB authority and YAML bootstrap"
```

---

### Task 22: 工作流文档（execution-log / verification-log / review-log）

按 CLAUDE.md 工作流要求，每完成一个 task 后**调用对应文档化 skill**：

- 每个任务实现 & 提交后 → `documenting-execution`
- 每次 verification 跑完 → `documenting-verification`
- 每次 code review 完成 → `documenting-review`
- 整体完成、merge / PR 前 → `documenting-completion`

本任务不在 plan 内单独执行步骤，而是在执行过程中以 skill 调用自然产生。

---

## Self-Review

### 1. Spec 覆盖检查

| Spec 章节 | 实现任务 |
|---|---|
| §3 架构总览（YAML→DB→Provider→consumer） | Task 1–6 |
| §4.1 User 新字段 | Task 5 |
| §4.2 settings 表 | Task 3 |
| §4.3 Overrides 结构 | Task 1 |
| §4.4 改密事件日志 | Task 12（handler 中 log.Printf） |
| §5.1 POST /auth/change-password | Task 10 + Task 12 |
| §5.2 GET /auth/me 扩展 | Task 12 |
| §5.3 GET /settings | Task 14 |
| §5.4 PUT /settings + 字段校验 | Task 13 + Task 14 |
| §5.5 路由挂载 | Task 12, Task 14 |
| §5.6 JWT 校验 token_version | Task 11 |
| §6 Provider 改造 | Task 7–9 |
| §7 前端 UI | Task 15–19 |
| §8 范围 / 非目标 | 全部任务（明确不做 thumbnails / token_expire / DB UI 等） |
| §9 错误处理表 | 散落在各任务（settings persist failed / api_token_forbidden / 损坏 payload 回退见 main.go） |
| §10 测试策略 | 各任务的 test 步骤 + Task 20 手动验收 |
| §13 yaml mtime INFO | Task 6 |
| README 配置管理章节 | Task 21 |

### 2. Placeholder 扫描

- 无 TBD / TODO
- 所有 step 含具体代码或具体命令
- 测试代码完整可运行

### 3. 类型一致性

- `Overrides`：Task 1 定义，Task 2 / 4 / 13 / 14 引用，签名一致
- `Provider.Get()`：返回 `*Config`，所有消费者调用一致
- `Provider.Apply(Overrides)`：值参数，一致
- `SettingsRepository.Save(Overrides, userID)`：Task 4 定义，Task 13 调用一致
- `AuthService.ChangePassword(uint, string, string) (string, time.Time, error)`：Task 10 定义，Task 12 调用一致
- `Claims.TokenVersion`：Task 10 加入，Task 11 middleware 比对一致
- `MeInfo.uses_default_password`：Task 12 后端字段名 / Task 15 前端字段名一致

### 4. 已知风险与缓解

- **Task 7→Task 8 之间 image_test.go 暂时编译失败**：plan 明确指出 Task 8 修复，subagent 执行时只要按顺序执行不会卡住
- **`tokenSvc.Create` 签名假设**：Task 12 测试中假设 `Create(userID, name, scope) (token, model, error)`；若签名不同，按实际调整
- **Vite dev server 端口**：Task 20 验收命令假设默认端口；如果项目自定义，按实际端口走

---

## Execution Handoff

Plan complete and saved to `docs/superpowers/plans/2026-05-27-admin-settings-and-account.md`. Two execution options:

**1. Subagent-Driven (recommended)** — 每个 task 派发一个全新 subagent 实现，主会话保持 review 视角，节奏快、上下文不污染。

**2. Inline Execution** — 直接在当前 session 用 `executing-plans` 批量推进，到关键 checkpoint 暂停 review。

Which approach?






