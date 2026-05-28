# 核心安全与访问控制补全 — 设计 Spec

**Date:** 2026-05-28
**Sub-project:** decomposition 子项目 1 — 核心安全与访问控制补全
**Source decomposition:** `docs/superpowers/decomposition/2026-05-27-feature-gap-roadmap.md`
**Status:** Draft（等待 user review）

## 1. 目标

补齐当前 CloudAlbum 已经在数据模型、配置项或产品表面上暴露出来、但还没有真正生效的核心安全能力，让项目在不推翻现有分享/上传模型的前提下达到“默认更安全、行为可配置、回归可验证”的状态。

本期聚焦 4 个缺口：

- API Token 过期校验真正生效
- 上传限流真正生效
- 公开图片链路具备可配置访问策略
- EXIF / 隐私元数据处理与配置项保持一致

本期采用“安全补齐，但不改产品基本分享模型”策略：

- 保留现有 `/i/*key`、`/t/*key` 永久直链作为默认公开访问模型
- 在现有链路上增加安全策略层，而不是重做分享体系
- 复用上个子项目已落地的 `config.Provider`，让安全能力同样走 `YAML → DB overrides → Provider → consumer`

## 2. 非目标

本期不做：

- 临时分享链接 / 签名链接（属于 decomposition 子项目 6）
- 多用户权限体系 / RBAC
- 企业级审计系统
- 多副本一致性限流 / Redis 限流
- 所有写接口统一限流（本期仅覆盖上传）
- 安全策略后台可视化全量编辑页
- 图片水印、内容审核、病毒扫描
- “保留 EXIF” 的逐字段强保证

## 3. 决策汇总

| 维度 | 结论 |
|---|---|
| 本期策略 | 安全补齐，但不改产品基本分享模型 |
| Token 过期 | 复用现有 `APIToken.ExpiresAt`；校验链路真正拒绝过期 token |
| Token 创建 | 支持可选过期时间；推荐 API 使用 `expires_in`，避免绝对时间时区歧义 |
| 上传限流范围 | 仅上传相关入口，覆盖 JWT 后台上传 + API Token 上传 |
| 限流粒度 | JWT 按 `user_id`；API Token 优先按 `token_id` |
| 限流实现 | 单进程内存限流，满足当前项目定位 |
| 公开访问默认模型 | 保持永久直链默认可用 |
| 防盗链模式 | `off` / `referer_whitelist` / `allow_empty_or_whitelist` |
| 防盗链白名单匹配 | 以 host 为主，不做完整 URL 前缀匹配 |
| EXIF 默认策略 | 继续默认 `strip_exif=true` |
| EXIF 关闭时语义 | best effort 保留，不承诺逐字段原样保真 |
| Settings UI 范围 | 本期仅继续暴露 `image.strip_exif`；限流/防盗链/token 默认过期策略先走 YAML |

## 4. 架构总览

### 4.1 设计思路

本期不是引入全新安全子系统，而是在现有 4 条关键链路上增加策略判断：

1. **认证链路**：在 `TokenService.Validate()` 中真正执行 API Token 过期校验
2. **上传链路**：在上传入口增加限流判定层
3. **公开访问链路**：在 `PublicHandler` 中增加公开访问策略判定层
4. **图片处理链路**：在 `Processor` 中让 `image.strip_exif` 的行为真实生效

### 4.2 运行时配置边界

安全策略继续复用 `config.Provider`：

```text
YAML / settings overrides
          │
          ▼
  config.Provider
          │
   ┌──────┼──────────┬────────────┐
   ▼      ▼          ▼            ▼
Token  Upload     Public       Image
policy limiter    access       processor
```

关键原则：

- **不推翻现有 consumer 结构**，而是在现有 service/handler 中读取 Provider snapshot
- **不把安全策略编辑 UI 一起扩大到本期范围**，避免和子项目 4 的后台设置再次强耦合
- **默认行为尽量兼容现有部署**，强化配置能力而不是强制改模型

### 4.3 新增/明确的配置维度

#### Token 策略

- 是否允许创建永不过期 token
- 默认过期时长
- 最大允许过期时长（可选，若实现成本低可纳入）
- validate 时严格拒绝过期 token

#### Upload 限流策略

- 是否启用
- 窗口长度
- 窗口内最大次数
- 作用对象：JWT 上传 / API Token 上传
- 超限错误语义

#### Public 访问策略

- `off`
- `referer_whitelist`
- `allow_empty_or_whitelist`
- host 白名单列表

#### Image 元数据策略

- 继续沿用 `image.strip_exif`
- 在 spec 中明确开启/关闭时的真实行为语义

## 5. 数据模型与配置模型

### 5.1 API Token

继续复用现有 `internal/model/token.go`：

```go
type APIToken struct {
    ID         uint       `gorm:"primaryKey" json:"id"`
    UserID     uint       `json:"user_id"`
    Name       string     `gorm:"size:100;not null" json:"name"`
    TokenHash  string     `gorm:"uniqueIndex;size:64;not null" json:"-"`
    Scope      string     `gorm:"size:20;not null" json:"scope"`
    ExpiresAt  *time.Time `json:"expires_at"`
    LastUsedAt *time.Time `json:"last_used_at"`
    CreatedAt  time.Time  `json:"created_at"`
}
```

本期不新增表；重点是让 `ExpiresAt` 从“存在字段”变成“真实约束”。

### 5.2 限流状态

本期不落库，使用进程内内存状态。

原因：

- 当前项目定位是单进程自托管
- decomposition 和现有文档里没有要求多副本一致性
- 先把行为闭环，比过早引入 Redis/DB 限流更符合 YAGNI

代价与边界：

- 服务重启后计数重置
- 多副本部署下无法共享窗口状态
- 这些限制在本期是接受的；若未来进入多实例部署，再升级为外部存储型限流器

### 5.3 Public 访问配置

本期建议在配置模型中新增一个公开访问策略段，例如：

```yaml
public_access:
  mode: off
  allowed_referer_hosts: []
```

也可以挂到 `server` 段或 `image` 段，但推荐独立为 `public_access`，因为它描述的是“公开访问策略”，不是服务监听参数，也不是图像编码参数。

### 5.4 EXIF 配置

继续沿用当前已有字段：

```yaml
image:
  strip_exif: true
```

本期不新增更细粒度的元数据保留白名单，避免把范围扩展成完整媒体策略系统。

## 6. 具体行为定义

### 6.1 API Token 过期校验

#### 创建行为

- `POST /api/v1/tokens` 支持可选过期配置
- 推荐请求字段使用 `expires_in`（秒）而不是 `expires_at`
- 若未传：
  - 若配置允许永不过期，可创建永不过期 token
  - 否则使用系统默认过期时长

#### 校验行为

`TokenService.Validate()` 调整为：

1. 规范化 raw token
2. 通过 hash 查出 token
3. **若 `ExpiresAt != nil && now > ExpiresAt`，直接拒绝**
4. 拒绝时对外统一视为无效 token
5. 仅在校验通过时更新 `last_used_at`

#### 对外语义

- 过期 token 不单独暴露“已过期”细节
- API 表现继续统一为 `401` + invalid token 语义，减少探测信息泄露

### 6.2 上传限流

#### 保护范围

仅保护上传相关入口，包括：

- JWT 登录后台执行的上传
- API Token 执行的上传

不纳入：

- 删除
- 改名
- 相册管理
- Settings 修改
- Change password

#### 识别粒度

- **JWT 上传**：按 `user_id`
- **API Token 上传**：优先按 `token_id`；若上下文里没有 token id，则退化为 `user_id`

这样可以避免一个用户下多个自动化脚本共享同一个上传窗口而互相干扰。

#### 算法选择

本期可接受固定窗口或滑动窗口，只要满足：

- 行为稳定
- 测试可预测
- 错误语义明确

推荐优先实现**固定窗口**，因为简单、可测、足够满足当前目标。

#### 超限行为

- 返回 `429 Too Many Requests`
- 稳定错误码：`rate_limit_exceeded`
- 人类可读信息：`upload rate limit exceeded`
- Web 前端展示友好提示：`上传过于频繁，请稍后再试。`

### 6.3 公开图片访问策略

#### 默认模型

继续保留现有：

- `/i/*key`
- `/t/*key`

默认仍可作为永久直链使用。

#### 模式定义

##### `off`

- 不做防盗链限制
- 现有公开图片行为完全不变

##### `referer_whitelist`

- 请求必须带 Referer
- 且 Referer host 必须命中白名单
- 否则返回 403

##### `allow_empty_or_whitelist`

- 若 Referer 为空，允许访问
- 若 Referer 不为空，则必须命中白名单
- 这是更适合当前产品默认推荐的安全模式，因为可兼容：
  - 用户直接在地址栏打开图片
  - 部分客户端 / WebView / IM 场景不携带 Referer

#### 匹配规则

- 以 **host** 为主匹配，不做完整 URL 前缀匹配
- 不在本期实现复杂通配、路径级策略、协议级细分

#### 拒绝语义

- 访问被策略拒绝时返回 **403**
- 不伪装为 404
- 推荐错误码：`public_access_forbidden`

### 6.4 EXIF / 隐私元数据处理

#### `image.strip_exif=true`

- 上传处理时先正确做方向修正
- 再以编码输出的方式移除 EXIF / 隐私元数据
- 缩略图天然不保留原始 EXIF
- 最终效果：显示方向正确，同时不保留拍摄设备、GPS 等元数据

#### `image.strip_exif=false`

- 尽量保留原图元数据
- 但由于当前链路存在 decode/encode / auto-convert，保留语义只能定义为 **best effort**
- spec 中明确：本期不承诺“逐字段原样保真”

这样可以避免对底层图像库能力做超出实际实现的承诺。

## 7. 接口设计

### 7.1 `POST /api/v1/tokens`

请求新增可选字段：

```json
{
  "name": "cli",
  "scope": "upload",
  "expires_in": 86400
}
```

说明：

- `expires_in` 单位为秒
- 不传则按系统默认策略推导
- 若传 0 或负数，视为无效请求

成功响应继续包含：

```json
{
  "token": {
    "id": 1,
    "name": "cli",
    "scope": "upload",
    "expires_at": "2026-05-29T12:00:00Z",
    "last_used_at": null,
    "created_at": "2026-05-28T12:00:00Z"
  },
  "raw_token": "ca_xxx"
}
```

### 7.2 `GET /api/v1/tokens`

返回列表中明确包含 `expires_at`，前端据此展示：

- 永不过期
- 具体过期时间
- 已过期

### 7.3 上传接口

现有上传接口无需换路径，但在处理逻辑前增加限流判定：

- 未超限：保持现有行为
- 超限：返回 `429` + `rate_limit_exceeded`

### 7.4 `/i/*key` 与 `/t/*key`

现有路径不变，但在读取对象前增加访问策略判定：

- 允许：继续读取并返回原图 / 缩略图
- 拒绝：返回 `403` + `public_access_forbidden`

## 8. 前端与后台 UI 范围

### 8.1 Token 管理页

在 `web/src/pages/Tokens.tsx` 基础上扩展：

- 创建表单增加可选过期时间输入
- Token 列表展示 `expires_at`
- 状态展示：永不过期 / 将于某时过期 / 已过期

本期不做：

- 批量吊销
- 倒计时实时刷新
- 复杂日期时间选择器增强

### 8.2 Settings 页

继续保留并允许编辑 `image.strip_exif`。

本期**不**把以下安全策略接入 Settings UI：

- upload rate limit
- public access mode / whitelist
- token 默认过期策略

它们先通过 YAML 生效。原因：避免把本期从“安全闭环”扩大成“后台设置二次扩容”。

## 9. 受影响模块

预期影响范围：

- `internal/service/token.go`
- `internal/handler/token.go`
- `internal/middleware/auth.go`
- 上传相关 handler / service
- `internal/handler/public.go`
- `internal/image/processor.go`
- `internal/config/config.go`
- `internal/config/provider.go`
- `web/src/pages/Tokens.tsx`
- 相关测试文件

## 10. 错误语义

建议统一使用稳定错误码：

| 场景 | HTTP | error |
|---|---:|---|
| token 无效 / 过期 | 401 | `invalid token` 或现有统一语义 |
| 创建 token 过期参数非法 | 400 | `invalid_request` |
| 上传超限 | 429 | `rate_limit_exceeded` |
| 公开访问被策略拒绝 | 403 | `public_access_forbidden` |

原则：

- 过期 token 不对外暴露更多细节
- 限流和公开访问拒绝要有稳定、可测试的错误码

## 11. 测试与验收标准

### 11.1 API Token

至少覆盖：

- 创建永不过期 token / 默认过期 token / 显式短期 token
- 未过期 token 可正常通过 `Validate()`
- 过期 token 返回无效 token 语义
- 过期 token 不更新 `last_used_at`
- `GET /tokens` 与 `POST /tokens` 返回 `expires_at`

### 11.2 上传限流

至少覆盖：

- JWT 上传在窗口内正常
- JWT 上传超限后 429
- API Token 上传同样受限
- 窗口重置后恢复
- 非上传接口不受影响

### 11.3 Public 访问策略

至少覆盖：

- `off` 模式下现有行为不变
- `referer_whitelist` 命中白名单可访问
- 非白名单被 403
- `allow_empty_or_whitelist` 模式下空 Referer 可访问
- `/i/*key` 与 `/t/*key` 都覆盖

### 11.4 EXIF

至少覆盖：

- `strip_exif=true` 时，输出结果不再携带原始 EXIF 元数据
- 图片方向显示正确
- `strip_exif=false` 时，行为符合 best effort 保留语义

### 11.5 回归

- 现有 JWT 登录 / Token 管理 / 上传 happy path 不被破坏
- 现有公开图片访问在 `public_access.mode=off` 下不回归
- `go test ./... -count=1` 通过
- 前端 `npm run build` 通过

## 12. 风险与取舍

### 12.1 公开直链仍是默认模型

这是有意取舍。

原因：

- 当前系统多个位置直接产出 `/i/...` URL
- decomposition 已把临时分享链接明确放到子项目 6
- 本期若切成签名链接优先，会显著扩大范围并破坏现有分享语义

### 12.2 限流只做单进程内存实现

这是与当前项目定位一致的取舍。

若未来进入多副本部署，再在子项目 2/7 或独立后续工作中升级。

### 12.3 EXIF 保留不是强保证

这是尊重当前处理链路实际能力的取舍。

本期目标是“剥离真正生效”，而不是建设完整媒体元数据保真系统。

## 13. 完成定义

本期视为完成，需要同时满足：

1. API Token 过期校验真实生效
2. 上传限流真实生效，并覆盖 JWT + API Token 上传
3. 公开图片链路具备可配置访问策略
4. `image.strip_exif` 与实际上传处理行为一致
5. 默认行为不破坏现有永久直链分享模型
6. 新增安全行为具备稳定错误语义与回归测试

## 14. 后续衔接

本期完成后：

- 子项目 2 可在更安全的默认基础上继续做部署/运行路径闭环
- 子项目 6 再引入签名链接 / 临时分享模型时，不会和本期 scope 混在一起
- 若未来需要，子项目 4 的 settings 能力可再扩展到安全策略 UI，但不作为本期前置条件
