# 管理后台设置与账户闭环 — 设计 Spec

**Date:** 2026-05-27
**Sub-project:** decomposition 子项目 4 — 管理后台设置与账户闭环
**Source decomposition:** `docs/superpowers/decomposition/2026-05-27-feature-gap-roadmap.md`
**Status:** Draft（等待 user review）

## 1. 目标

把当前偏展示性质的后台补成真正可管理、可维护的单用户产品。本期采用「渐进闭环」策略：

- 让 admin 能在 UI 内完成 **改密** —— 摆脱写死的 `admin/admin123`
- 让 Settings 页从只读升级为 **部分字段可编辑** —— 不再需要改 YAML 才能调整高频运行参数
- 为后续扩展（更多可编辑字段、可观测性、子项目 1 的限流）建立 **运行时配置 hot-swap** 底座

不做的：多用户、权限分层、操作审计表、应用层限流、字段级"恢复默认"按钮、密码复杂度规则、国际化、前端测试框架。详见 §8。

## 2. 决策汇总

| 维度 | 结论 |
|---|---|
| 本期范围 | 改密码 + Settings 部分可编辑（渐进闭环） |
| 首次改密策略 | 强提醒不强制（默认密码时全局 Banner） |
| 改密后会话处理 | bump `token_version` 吊销所有旧 JWT；API Token 不受影响 |
| 密码策略 | 长度 ≥ 8；新 ≠ 旧；改密事件写 stdout 日志；本期不做接口限流 |
| Settings 可编辑项 | `server.base_url` + `image.{max_size, allowed_types, auto_convert, quality, strip_exif}` |
| 配置存储 | DB 为权威源（新增 `settings` 表，单行 JSON）；YAML 仅首次 bootstrap |
| 生效策略 | 本期字段全部"立即生效"，通过 `atomic.Pointer[Config]` 整包替换 |
| 前端结构 | 侧边栏新增 Account 入口（位于 Settings 之前）；Settings 改造为可编辑 |
| Account 页元素 | 改密表单 + 账号只读信息 + 默认密码 Banner |

## 3. 架构总览

```
              ┌────────────────────┐
启动期        │ YAML (configs)     │
              └─────────┬──────────┘
                        │ Load + bootstrap (仅在 settings 表为空时)
                        ▼
              ┌────────────────────┐
权威源        │ DB: settings 表    │  ──── 单行 JSON, id=1, payload, updated_at, updated_by
              └─────────┬──────────┘
                        │ 启动加载 + 每次写入后 reload
                        ▼
              ┌────────────────────┐
运行时        │ config.Provider    │  ──── atomic.Pointer[Config]，持有最新 snapshot
              └─────────┬──────────┘
                        │ Get() snapshot
                        ▼
              ┌────────────────────┐
消费方        │ service / handler  │  ──── 不再持值拷贝，改为持有 Provider
              └────────────────────┘
```

关键点：

- **YAML 不再是运行时权威源**。后续人为修改 `configs/config.yaml` 不会反映到运行时（除非 settings 表被清空）。这点会在 README + 启动 INFO 日志中说明。
- **写入流程**：handler → service 校验 → repo 写 `settings` 表（一次事务）→ 通知 Provider 重新构建 snapshot → atomic 替换。
- **Provider 只暴露**：`Get() *Config`、`Overrides() *Overrides`、`Apply(Overrides)`。
- **受影响消费方**：`AuthService`（读 `token_expire` / `jwt_secret`）、`ImageService`（读 image 段）、公开图片链接拼接处（读 `server.base_url`）。

非目标：

- settings 字段级历史 / 审计表
- 多实例同步（项目为单进程）

## 4. 数据模型

### 4.1 `User` 模型变更（`internal/model/user.go`）

```go
type User struct {
    ID                uint       `gorm:"primaryKey" json:"id"`
    Username          string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
    PasswordHash      string     `gorm:"not null" json:"-"`
    Role              string     `gorm:"size:20;default:admin" json:"role"`
    TokenVersion      uint       `gorm:"not null;default:1" json:"-"`         // 新增
    PasswordChangedAt *time.Time `json:"password_changed_at,omitempty"`        // 新增
    CreatedAt         time.Time  `json:"created_at"`
}
```

- `TokenVersion`：每次改密 +1。JWT claims 新增 `tv` 字段（uint），签发时写入当前 `TokenVersion`，校验时与 DB 当前值比对，不一致即拒绝。
- `PasswordChangedAt`：Account 页展示。`nil` 表示"从未改过"——是默认密码 Banner 的辅助判据，但**默认密码检测的主信号是 bcrypt 比对 `admin123`**（详见 §5.2）。

迁移：GORM `AutoMigrate` 自动加列；旧记录 `TokenVersion` 默认 1、`PasswordChangedAt` 为 NULL。

### 4.2 新增 `settings` 表（`internal/model/settings.go`）

```go
type Settings struct {
    ID        uint      `gorm:"primaryKey"`            // 固定为 1
    Payload   string    `gorm:"type:text;not null"`    // JSON
    UpdatedAt time.Time
    UpdatedBy uint                                     // user_id
}
```

### 4.3 `Overrides` 反序列化结构（`internal/config/overrides.go`，新增）

```go
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

- 全部指针：区分"未设置 → 落回 YAML 默认"和"显式空值"。
- `Config snapshot = YAML 基线 ⊕ Overrides`，合并逻辑只在 `Provider.Apply()` 内执行。
- Schema 演化：新增可编辑字段时只加 `*T`；旧 JSON 反序列化得到 `nil` → 自动用 YAML 默认值，无需 DB migration。

### 4.4 改密事件日志

不新增表。stdout（项目当前日志方式）：

```
[settings] password changed for user_id=1 ip=127.0.0.1 ts=2026-05-27T10:30:00Z
[settings] password change failed for user_id=1 reason=wrong-old-password ip=...
```

不记录任何密码内容；不持久化失败次数（接口限流统一由子项目 1 处理）。

## 5. 后端接口

### 5.1 `POST /api/v1/auth/change-password`

请求：

```json
{
  "old_password": "...",
  "new_password": "..."
}
```

成功 200：

```json
{
  "token": "<new JWT>",
  "password_changed_at": "2026-05-27T10:30:00Z"
}
```

错误码：

- `400 invalid_request` — 字段缺失 / 长度 < 8
- `401 wrong_old_password`
- `400 same_as_old`
- `403 api_token_forbidden` — API Token 不允许调此接口

服务端流程：

1. 校验 `auth_type == "jwt"`，否则 403。
2. bcrypt 比对 `old_password`，不一致 → 401。
3. 校验 `new_password` 长度 ≥ 8 且 ≠ `old_password`，否则 400。
4. **事务内**：`PasswordHash = bcrypt(new)`；`TokenVersion += 1`；`PasswordChangedAt = now`。
5. 使用更新后的 `TokenVersion` 签发新 JWT，写入 response（避免改密用户被自己吊销）。
6. 写一条 stdout 日志（含 user_id / IP / 时间）。

对应错误 sentinel（`internal/service/auth.go`）：

- `ErrInvalidCredentials`（已存在）→ 401 `wrong_old_password`
- `ErrPasswordTooShort`（新增）→ 400 `invalid_request`
- `ErrPasswordSameAsOld`（新增）→ 400 `same_as_old`
- `ErrAPITokenForbidden`（新增）→ 403 `api_token_forbidden`

### 5.2 `GET /api/v1/auth/me`（扩展现有接口）

在现有返回（`user_id` / `username` / `auth_type` / `token_scope`）基础上追加：

```json
{
  "password_changed_at": null,
  "uses_default_password": true,
  "created_at": "2026-05-25T10:00:00Z"
}
```

- `password_changed_at`：`User.PasswordChangedAt` 透传，`nil` → 序列化为 JSON `null`。
- `uses_default_password`：直接 bcrypt 比对当前 hash 与字面量 `admin123`。**仅在** `auth_type=jwt` 且 `username == "admin"` 时计算并返回；其他情况不返回此字段。
  - 选择直接 bcrypt 比对而非"是否改过密码"作为判据，是因为用户可能改回默认值，那时仍应提示。
  - 本期 admin 用户名不可改（项目无对应接口，本期也不引入）。若未来引入改名能力，此处判据需同步升级（例如改为"hash 匹配启动时记录的 bootstrap 默认值"）。
- `created_at`：`User.CreatedAt` 透传。

API Token 调此接口时仅返回原有字段，不返回新增的 3 个字段。

### 5.3 `GET /api/v1/settings`

允许 JWT 与 API Token 都调用（脚本上传可能需要读取 `max_size` / `allowed_types` 预判）。

响应：

```json
{
  "effective": {
    "server":  { "base_url": "..." },
    "image":   {
      "max_size": 52428800,
      "allowed_types": ["jpg","jpeg","png","gif","webp","bmp","svg"],
      "auto_convert": "webp",
      "quality": 85,
      "strip_exif": true
    }
  },
  "overrides": {
    "server":  { "base_url": true },
    "image":   { "quality": true }
  },
  "editable_fields": [
    "server.base_url",
    "image.max_size",
    "image.allowed_types",
    "image.auto_convert",
    "image.quality",
    "image.strip_exif"
  ]
}
```

- 仅返回白名单字段；**永不**包含 `jwt_secret` / `database` / `storage` / `token_expire` / `server.port`。
- `overrides` 段表明哪些字段被显式覆盖，UI 据此显示"已修改"角标。
- `editable_fields` 由后端权威给出，避免前后端漂移。

### 5.4 `PUT /api/v1/settings`

仅 JWT，API Token 调用返回 `403 api_token_forbidden`。

请求（部分更新，未传入的字段保持当前值）：

```json
{
  "server": { "base_url": "https://img.example.com" },
  "image":  { "quality": 90, "strip_exif": true }
}
```

响应：与 `GET /api/v1/settings` 同格式。

服务端流程：

1. 校验 `auth_type == "jwt"`。
2. **白名单校验**：请求体 JSON 只能出现白名单字段（含嵌套），否则 `400 unknown_field`，错误体包含字段名。
3. **字段级校验**：
   - `base_url`：必须为 http/https URL（`url.Parse` + scheme 检查）
   - `max_size`：范围 `(0, 1 GiB]`
   - `allowed_types`：非空，元素属于 `{jpg, jpeg, png, gif, webp, bmp, svg}`，去重
   - `auto_convert`：∈ `{"", "webp", "jpeg"}`
   - `quality`：`[1, 100]`
   - `strip_exif`：bool
4. 读取当前 `Overrides`（DB）→ 与请求体 merge（部分字段）→ 写回 DB（`UPDATE settings WHERE id=1`）。
5. 通知 `Provider.Apply(newOverrides)` → atomic snapshot 替换。
6. 返回新的 `effective + overrides`。

**并发**：`SettingsService.Update()` 内部用 `sync.Mutex` 串行化"读旧 → merge → 写新"，避免单进程下的 lost update。

### 5.5 路由挂载（`internal/router/router.go`）

```go
auth := api.Group("/auth")
auth.POST("/change-password", authHandler.ChangePassword)
// auth.GET("/me") 已存在, 扩展返回字段即可

settings := api.Group("/settings")
settings.GET("",  settingsHandler.Get)     // JWT + API Token 都可
settings.PUT("",  settingsHandler.Update)  // 仅 JWT, handler 内显式校验 auth_type
```

不引入新 scope（项目是单用户）。区分 JWT / API Token 通过 `auth_type` 上下文字段完成。

### 5.6 JWT 校验链路扩展（`internal/middleware/auth.go`）

**仅 `auth_type == "jwt"` 的请求走 token_version 比对。API Token 路径走原有 token 表状态校验，不受本节影响。**

JWT 校验在 `AuthService.ParseJWT` 之后新增一步：

1. `claims.UserID` → `UserRepository.FindByID`
2. 比较 `claims.TokenVersion == user.TokenVersion`，不一致 → 返回 `invalid token`

每次 JWT API 调用多一次 user 主键查询；SQLite 单用户量级 < 100 µs，可接受。本期不加 in-memory 缓存。

## 6. 运行时配置 hot-swap 改造

### 6.1 新增 `config.Provider`（`internal/config/provider.go`）

```go
package config

import "sync/atomic"

type Provider struct {
    base      Config                     // YAML 基线, 启动时定型, 之后只读
    snapshot  atomic.Pointer[Config]     // 当前生效快照
    overrides atomic.Pointer[Overrides]  // 当前 overrides, 用于回显
}

func NewProvider(base Config, overrides Overrides) *Provider {
    p := &Provider{base: base}
    p.applyOverrides(overrides)
    return p
}

func (p *Provider) Get() *Config            { return p.snapshot.Load() }
func (p *Provider) Overrides() *Overrides   { return p.overrides.Load() }
func (p *Provider) Apply(o Overrides)       { p.applyOverrides(o) }

func (p *Provider) applyOverrides(o Overrides) {
    merged := p.base
    if o.Server.BaseURL    != nil { merged.Server.BaseURL    = *o.Server.BaseURL }
    if o.Image.MaxSize     != nil { merged.Image.MaxSize     = *o.Image.MaxSize }
    if o.Image.AllowedTypes != nil { merged.Image.AllowedTypes = *o.Image.AllowedTypes }
    if o.Image.AutoConvert != nil { merged.Image.AutoConvert  = *o.Image.AutoConvert }
    if o.Image.Quality     != nil { merged.Image.Quality      = *o.Image.Quality }
    if o.Image.StripExif   != nil { merged.Image.StripExif    = *o.Image.StripExif }
    p.snapshot.Store(&merged)
    p.overrides.Store(&o)
}
```

- 读侧零锁，`Get()` 是 `atomic.Pointer.Load`。
- 写侧只在 `SettingsService.Update` 调一次 `Apply`。
- `Get()` 返回的快照按约定**只读**，调用方按需取字段：`provider.Get().Image.Quality`。

### 6.2 启动流程改造（`main.go`）

```go
base, err := config.Load(cfgPath)
db := ...
settingsRepo := repository.NewSettingsRepository(db)
overrides, err := settingsRepo.LoadOrBootstrap()  // 空表 → 写入 {id:1, payload:"{}"}
provider := config.NewProvider(*base, overrides)

authSvc     := service.NewAuthService(userRepo, tokenRepo, provider)
imageSvc    := service.NewImageService(..., provider)
settingsSvc := service.NewSettingsService(settingsRepo, provider)
```

`LoadOrBootstrap`：表里有一行就读那行；没有就插入 `{id:1, payload:"{}"}` 并返回零值 `Overrides`。YAML 不主动同步到 settings —— 种子是"空 overrides"，意味着首次启动一切走 YAML 默认。

### 6.3 YAML 与 DB 时间戳对比（启动期）

如果 `configs/config.yaml` 的 `mtime` 比 `settings.updated_at` 晚，启动时打 INFO：

```
[config] yaml file is newer than settings table; runtime still uses values from settings DB.
         To re-seed from YAML, delete the row in `settings` table and restart.
```

这是软提醒，不阻塞启动。

### 6.4 service 改造范围

| 现有位置 | 现状 | 改造 |
|---|---|---|
| `AuthService.cfg config.AuthConfig` | 启动时值拷贝 `token_expire` / `jwt_secret` | 持 `*config.Provider`；签发时 `provider.Get().Auth.TokenExpire` |
| `ImageService.cfg config.ImageConfig`（或同等位置） | 启动时值拷贝整个 image 段 | 持 `*config.Provider`；上传 / 缩略图 / 校验路径每次 `provider.Get().Image.X` |
| 公开图片链接拼接 | 散落使用 `cfg.Server.BaseURL` | 全部改成 `provider.Get().Server.BaseURL` |
| `handler.PublicHandler` 涉及 base_url 处 | 同上 | 同上 |

**plan 阶段会先 grep `cfg.Image` / `cfg.Server.BaseURL` 所有引用点，落到一份明确清单**。如果发现散得很厉害（十几处以上），plan 中单列一项"重构 Provider 引用"独立任务，独立提交，避免与 settings 接口任务耦合。

### 6.5 快照实时性语义

- 一次 HTTP 请求生命周期内，第一次 `provider.Get()` 之后再发生的写入不会反映给本次请求剩余逻辑。这是合理的，避免请求中途配置漂移。
- 实现约定：每个 handler / service 方法在开头取一次 `cfg := h.provider.Get()`，后续用 `cfg.X`。

### 6.6 不可变集合（明确写入 spec）

- `database`、`storage`、`server.port` 一律不进 Provider 的"可覆盖集合"。`Provider.base` 里照常存，但 `Apply()` 无路径修改。
- `auth.jwt_secret` 不在 UI 白名单，但走 Provider 是为未来一致性；本期等同 YAML 值。

## 7. 前端 UI

### 7.1 路由 / 导航

- 侧边栏新增菜单项 **Account**，放在 Settings **之前**（账户 → 设置，自上而下）。
- 新增路由：`/account` → `web/src/pages/Account.tsx`
- 现有 `Settings.tsx` 改造为可编辑表单。

### 7.2 Account 页（`web/src/pages/Account.tsx`，新增）

```
┌─ Account ─────────────────────────────────────────────┐
│  账号信息                                             │
│  ─────────                                            │
│  用户名      admin                                    │
│  角色        admin                                    │
│  创建时间    2026-05-25 10:00:00                      │
│  上次改密    从未  /  2026-05-27 10:30:00             │
├───────────────────────────────────────────────────────┤
│  修改密码                                             │
│  ─────────                                            │
│  当前密码    [____________________]                   │
│  新密码      [____________________]   (≥ 8 位)        │
│  确认新密码  [____________________]                   │
│  [取消]   [保存]                                      │
└───────────────────────────────────────────────────────┘
```

交互：

- 表单本地校验：两次新密码一致、长度 ≥ 8、新 ≠ 旧。
- 后端返回新 JWT → `auth store` 替换；不跳转登录页。
- 成功 → toast「密码已修改」+ 重新拉 `/auth/me`（Banner 据此消失）。
- 失败 → 表单顶部红条显示后端 `error` 文案；不清空"当前密码"输入框。

### 7.3 Settings 页改造（`web/src/pages/Settings.tsx`）

替换现有 3 张只读卡片，分两个 Section：

```
┌─ Settings ────────────────────────────────────────────┐
│  站点                                                 │
│  ─────                                                │
│  Base URL     [https://img.example.com_________]      │
│               用于生成图片公开链接。立即生效。        │
│                                                       │
│  图片处理                                             │
│  ─────                                                │
│  最大大小     [____50____] MB                         │
│  允许格式     [jpg] [png] [webp] [gif] [bmp] [svg]    │
│  自动转换     ( ) 不转换  (•) WebP  ( ) JPEG          │
│  压缩质量     [────●────────] 85   (1–100)            │
│  EXIF 剥离    [✓] 移除 EXIF/隐私元数据                │
│                                                       │
│  [恢复未保存的修改]      [保存]                       │
└───────────────────────────────────────────────────────┘
```

交互：

- 进入页面 `GET /api/v1/settings` → 用 `effective` 填充控件，`overrides` 决定"已修改"角标。
- 表单态本地维护；只有"保存"才发 `PUT`。
- 字段级前端校验同后端规则（max_size 上限 1 GiB、allowed_types 非空、quality 1–100 等）。
- 成功 → toast「设置已保存」+ 重新拉 `GET` 覆盖本地态。
- 失败 → 顶部红条显示后端 `error` + 字段级错误高亮（依据后端返回的 `field`）。

### 7.4 默认密码 Banner

放在 `App.tsx` 顶层，所有页面共享：

```
┌──────────────────────────────────────────────────────────┐
│ ⚠️ 你正在使用默认密码 admin123  [立刻修改 →]  [稍后]    │
└──────────────────────────────────────────────────────────┘
```

- 渲染条件：`auth store` 中缓存的 `me.uses_default_password === true`。
- 「立刻修改」→ 跳 `/account` 并把焦点落到"当前密码"。
- 「稍后」→ Banner 组件内部 state 标记关闭（不写持久化存储）。当前页面挂载期间不再展示；用户刷新页面、切换 tab 重开、或重新登录时 Banner 会重新出现。强提醒，不强制。
- Banner 不阻塞任何页面操作。

### 7.5 auth store 变更

- 登录成功 / 应用启动时拉一次 `/auth/me`，缓存 `me` 全字段。
- 改密成功后 `setToken(newToken)` + 重新拉 `/auth/me`。
- 401 拦截器逻辑不变：清空 token + 跳登录页。

## 8. 范围与非目标

### 8.1 In scope

**后端**：

- `User` 新增 `TokenVersion` / `PasswordChangedAt`（AutoMigrate）
- 新增 `settings` 表 + `SettingsRepository` + `LoadOrBootstrap`
- 新增 `config.Provider` + `Overrides`；`AuthService` / `ImageService` / 公开链接拼接处改持 Provider
- 新增接口：`POST /auth/change-password`、`GET /settings`、`PUT /settings`
- 扩展 `GET /auth/me` 返回字段
- `middleware.AuthMiddleware` 增加 token_version 比对
- 启动时 yaml mtime vs settings.updated_at 比较，必要时打 INFO

**前端**：

- 侧边栏新增 Account 入口
- `web/src/pages/Account.tsx`：账号信息 + 改密表单
- `App.tsx` 顶层默认密码 Banner
- `web/src/pages/Settings.tsx` 改造为可编辑
- `auth store` 处理 me 缓存 + 改密续签

**文档**：

- 本 spec
- README 增加"配置管理"章节
- `docs/superpowers/` 流程文档：plan、execution-log、verification-log、review-log、completion

### 8.2 Out of scope

- 多用户 / 注册 / 邀请
- 权限分层（除 JWT vs API Token 的差异之外）
- 操作审计表 / 字段级历史
- 密码复杂度规则、强度可视化
- 改密 / settings 接口的应用层限流 —— 统一交给子项目 1
- Settings 字段级"恢复默认值"按钮
- `thumbnails` / `token_expire` / `storage` / `database` / `server.port` / `jwt_secret` 的 UI 编辑
- 配置变更事件总线 / 订阅广播
- 国际化、暗色模式适配
- 前端单测 / e2e 框架引入

## 9. 错误处理与边界

| 场景 | 行为 |
|---|---|
| settings payload 反序列化失败 | 启动 ERROR 日志，回退用空 `Overrides` 启动，不阻塞 |
| settings 表写入失败 | API 返回 `500 settings_persist_failed`；不替换 atomic snapshot |
| 并发 PUT `/settings` | `SettingsService` 内部 `sync.Mutex` 串行化 |
| 改密事务失败 | 不改 `TokenVersion`、不签新 JWT，返回 `500 change_password_failed`；旧 JWT 继续可用 |
| 改密后旧 JWT 残留在其他 tab | 下次 API 调用返回 `401 invalid token` → 已有全局拦截器跳登录页 |
| `uses_default_password` 仅 admin + JWT 计算 | 其他情况不返回此字段 |
| API Token 调改密接口 | `403 api_token_forbidden`，请求体不读取 |
| API Token 调 `PUT /settings` | `403 api_token_forbidden` |
| API Token 调 `GET /settings` | 允许 |
| YAML `auth.jwt_secret == "change-me-in-production"` | 启动 WARN（顺手补，非本期主目标） |

## 10. 测试策略

### 10.1 后端单测（go test）

- `internal/config/provider_test.go`
  - `Apply` 空 Overrides → snapshot == base
  - 部分字段覆盖 → 仅指定字段变化，其他保持 base
  - 并发 `Get` + `Apply` 不产生 race（`-race`）
- `internal/repository/settings_test.go`
  - `LoadOrBootstrap` 空表 → 写入一行，返回零值
  - 损坏 payload → 返回错误（让上层决定回退）
- `internal/service/auth_test.go`（扩展现有文件）
  - `ChangePassword` 成功 → hash 更新 + TokenVersion +1 + PasswordChangedAt 写入
  - 旧密码错 → `ErrInvalidCredentials`
  - 新密码 < 8 位 → `ErrPasswordTooShort`
  - 新 == 旧 → `ErrPasswordSameAsOld`
  - 改密后用老 JWT 调任意接口 → `ErrInvalidToken`（token_version 比对）
- `internal/service/settings_test.go`（新增）
  - 白名单外字段 → 拒
  - 字段级校验每条 case 一个子测试
  - 成功 → DB 一行更新 + Provider snapshot 反映新值
- `internal/handler/auth_test.go` / `settings_test.go`
  - 路由 + 鉴权矩阵：JWT 通过、API Token 写被拒、未鉴权 401

### 10.2 前端

本期**不引入测试框架**，沿用项目当前现状。

替代：plan 里写明"实现完成后必须人工跑通"以下场景，作为 `verification-log` 依据：

- happy: 默认密码 → Banner 出现 → 改密成功 → Banner 消失 → 退出登录用新密码登录
- happy: 修改 `base_url` + 立刻上传 → 返回链接前缀已变
- happy: 修改 `image.quality` → 立刻上传 → 新文件大小 / quality metadata 符合新规则
- error: 旧密码错（页面错误提示）
- error: 新密码 7 位（前端拦截）
- error: PUT settings 携带 `database.driver` → 400 unknown_field

### 10.3 回归覆盖

- 已有 `auth_test.go` / `image_test.go` / `public_test.go` 必须保持绿。
- `ImageService` 改持 Provider 后所有 image 测试不应回归。

## 11. 性能 / 资源

- 每个 API 请求多一次 `User` 主键查询。SQLite 量级 < 100 µs，单用户可忽略。
- `atomic.Pointer[Config]` 读零分配；写每次拷贝一个 `Config` 值（百字节量级）。
- 不引入新 goroutine / 后台任务。

## 12. 风险与缓解

| 风险 | 缓解 |
|---|---|
| `base_url` 引用点散落 | plan 阶段先 grep 所有引用，列出文件清单，单列任务处理 |
| `ImageService` 改持 Provider 的 patch 体积大 | plan 阶段单列任务，独立提交 |
| token_version 比对每请求多一次 DB 查询 | 单用户 SQLite 可忽略；多用户场景后续加缓存 |
| YAML 直改不生效的认知偏差 | README 显式说明 + 启动 INFO 日志 |
| settings payload 损坏导致启动失败 | ERROR 日志 + 回退空 Overrides，不阻塞启动 |
| 改密事务中 TokenVersion bump 与 JWT 续签的一致性 | 同一事务保证持久化；事务成功后才签发新 JWT |
| 单进程并发 PUT `/settings` 的 lost update | `SettingsService` 内部 `sync.Mutex` 串行化 |

## 13. 依赖关系

- **上游**：子项目 1「核心安全与访问控制补全」尚未完成。本期**不被它阻塞** —— 改密事件日志、API Token vs JWT 分流、白名单校验都已自成闭环；子项目 1 完成后只是顺势接上「改密接口」的全局限流。
- **下游**：本期落地的 `config.Provider` 为后续"扩展更多可编辑字段"、子项目 7（健康检查/可观测性）提供底座。

## 14. 文档与交付

- 本 spec：`docs/superpowers/specs/2026-05-27-admin-settings-and-account-design.md`
- 实施计划：下一步通过 `writing-plans` 产出 `docs/superpowers/plans/2026-05-27-admin-settings-and-account.md`
- README 增加"配置管理"章节：YAML 仅 bootstrap、运行时以 settings 表为准、如何重新种子
- 流程产出：`execution-log/2026-05-27-admin-settings-and-account.md`、`review-log/2026-05-27-admin-settings-and-account.md`、`completion/2026-05-27-admin-settings-and-account-summary.md`，按工作流逐项产出
