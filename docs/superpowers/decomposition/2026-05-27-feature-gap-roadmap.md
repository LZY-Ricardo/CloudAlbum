# CloudAlbum 功能缺口分解

**Date:** 2026-05-27
**Source:** 基于当前代码、既有设计文档与完成总结整理
**Goal:** 将“项目还缺少什么功能”拆成可独立推进的子项目，供后续逐项进入 brainstorming → spec → plan → execution 流程。

## Sub-project Overview

| Sub-project | Status | Priority | Dependencies | Next Step |
|---|---|---:|---|---|
| 核心安全与访问控制补全 | Pending | 1 | — | Start brainstorming |
| 数据与部署能力补全 | Pending | 2 | 核心安全与访问控制补全 | Start brainstorming |
| 图片管理体验完善 | Pending | 3 | — | Start brainstorming |
| 管理后台设置与账户闭环 | Pending | 4 | 核心安全与访问控制补全 | Start brainstorming |
| 相册与内容组织增强 | Deferred | 5 | 图片管理体验完善 | Re-prioritize later |
| 分享与生态集成增强 | Deferred | 6 | 数据与部署能力补全 | Re-prioritize later |
| 运营与可观测性增强 | Deferred | 7 | 数据与部署能力补全 | Re-prioritize later |

## Requirement Restatement

当前 CloudAlbum 已经具备单机可用的基础图床能力，但仍存在一批未闭环、未兑现设计承诺或会明显影响体验的缺口。本次目标不是直接实现，而是整理成一份可执行清单，明确：
- 哪些属于必须先补的正确性/安全性问题
- 哪些属于高频使用体验问题
- 哪些是后续增强项
- 它们之间有什么依赖关系，推荐先做什么

## Early Exit Check

这不是单一问题，而是至少横跨以下多个独立关注点：
- 安全与访问控制
- 数据存储与部署能力
- 图片管理前端体验
- 后台设置与账户闭环
- 相册/内容组织增强
- 分享与外部集成
- 运营与可观测性

因此不适合直接进入单一 brainstorming，先做分解是合理的。

## Dependency Graph

```text
核心安全与访问控制补全 ──→ 数据与部署能力补全 ──→ 分享与生态集成增强
            │                         └────────────→ 运营与可观测性增强
            └────────────→ 管理后台设置与账户闭环

图片管理体验完善 ──→ 相册与内容组织增强
```

**Critical path:** 核心安全与访问控制补全 → 数据与部署能力补全 → 运营/集成增强

## Sub-project Details

### 1) 核心安全与访问控制补全
- **Status:** Pending
- **Priority:** 1
- **Goal:** 补齐当前设计已承诺但尚未真正生效的安全与访问控制能力，避免项目在“可用”状态下带着明显风险上线。
- **Scope:**
  - API Token 过期校验真正生效
  - 上传频率限制 / rate limiting
  - 公开图片链路的防盗链或访问策略
  - EXIF/隐私元数据处理策略落地并与配置对齐
- **Explicitly Excluded:**
  - 多用户权限体系
  - 企业级审计 / RBAC
  - 临时分享链接
- **Dependencies:** —
- **Risks:**
  - 防盗链策略会影响现有外链兼容性
  - EXIF 处理需要平衡隐私、方向修正和图片质量
  - 限流策略需要兼顾 Web 后台与 API Token 客户端
- **Deliverable:** 设计承诺中的核心安全能力在代码与行为层面一致，默认部署更安全。
- **Next Step:** 对“过期 token、限流、防盗链、EXIF 剥离”的具体行为做一轮 brainstorming。

### 2) 数据与部署能力补全
- **Status:** Pending
- **Priority:** 2
- **Goal:** 补齐项目在数据库、对象存储和容器部署方面的“看起来支持，但未真正闭环”的能力。
- **Scope:**
  - PostgreSQL 运行路径接通
  - Docker / compose 真机验证
  - S3 / MinIO 真机连通性验证
  - 必要时补健康检查端点，方便部署验证
- **Explicitly Excluded:**
  - Kubernetes 部署
  - 自动扩缩容
  - 多区域高可用
- **Dependencies:** 核心安全与访问控制补全
- **Risks:**
  - PostgreSQL 接入可能暴露 repository / migration 兼容问题
  - 真实 S3 验证依赖外部凭据或 MinIO 环境
  - Docker 验证依赖当前机器环境
- **Deliverable:** SQLite / PostgreSQL / Local / S3 / Docker 至少达到“真实跑通并验证关键路径”的程度。
- **Next Step:** 明确是先接 PostgreSQL，还是先补 health check + Docker/S3 验证链路。

### 3) 图片管理体验完善
- **Status:** Pending
- **Priority:** 3
- **Goal:** 提升图片多起来之后的可管理性和高频操作体验。
- **Scope:**
  - 图片分页
  - 搜索 debounce
  - 复制链接成功反馈
  - 预览弹层可访问性（Escape / focus / keyboard）
  - 上传区 drag hover 稳定性
- **Explicitly Excluded:**
  - 彻底重做 UI 视觉风格
  - 标签系统
  - 高级搜索 DSL
- **Dependencies:** —
- **Risks:**
  - 分页会影响现有列表与批量操作状态管理
  - 预览可访问性改造涉及交互细节较多
- **Deliverable:** 图片管理页在图片量上来后仍然顺手，上传与复制反馈更稳定清晰。
- **Next Step:** 先决定优先做“分页+搜索”还是“交互 polish 一揽子”。

### 4) 管理后台设置与账户闭环
- **Status:** Pending
- **Priority:** 4
- **Goal:** 把当前偏展示性质的后台补成真正可管理、可维护的单用户产品。
- **Scope:**
  - 修改密码 / 首次改密能力
  - 系统设置页从只读概览升级为可编辑入口
  - 关键站点信息、存储策略、图片处理策略的管理入口
- **Explicitly Excluded:**
  - 多用户注册
  - 完整权限管理
  - 云控制台级复杂设置
- **Dependencies:** 核心安全与访问控制补全
- **Risks:**
  - 配置热更新 vs 重启生效需要明确策略
  - 涉及敏感配置时要避免把密钥直接暴露到前端
- **Deliverable:** 用户可以在后台完成基本账户安全和核心站点设置操作，不需要手改配置文件才能日常维护。
- **Next Step:** 先定义“哪些配置允许在 UI 修改，哪些仍保留在 YAML/环境变量”。

### 5) 相册与内容组织增强
- **Status:** Deferred
- **Priority:** 5
- **Goal:** 强化相册作为内容组织核心的能力，而不只是一个分类字段。
- **Scope:**
  - 相册封面设置 UI
  - 相册详情 / 进入相册看图
  - 更顺手的相册内批量管理
  - 视情况评估标签系统是否必要
- **Explicitly Excluded:**
  - 多级相册树
  - 智能分类/AI 标签
- **Dependencies:** 图片管理体验完善
- **Risks:**
  - 如果先不上分页，进入相册页后体验仍可能受限
  - 标签系统容易扩大范围
- **Deliverable:** 相册从“元数据管理表单”升级为真正的内容组织入口。
- **Next Step:** 等图片管理体验稳定后再重排优先级。

### 6) 分享与生态集成增强
- **Status:** Deferred
- **Priority:** 6
- **Goal:** 让 CloudAlbum 不只可上传，还能更好地对外分享和接入其他工具链。
- **Scope:**
  - 临时分享链接 / 带过期时间的访问链接
  - 自定义域名 / CDN 回源策略
  - webhook / 上传完成回调
  - 批量导出 / ZIP 下载
- **Explicitly Excluded:**
  - 公共图库社区功能
  - 第三方 OAuth 登录
- **Dependencies:** 数据与部署能力补全
- **Risks:**
  - 分享能力和防盗链策略可能互相牵制
  - CDN / 自定义域名需要更明确的 URL 策略
- **Deliverable:** 项目更适合作为真正长期使用的图床，而非仅本地管理工具。
- **Next Step:** 等底层存储与访问策略稳定后再设计分享模型。

### 7) 运营与可观测性增强
- **Status:** Deferred
- **Priority:** 7
- **Goal:** 为后续稳定运行、排障和使用分析补基础设施能力。
- **Scope:**
  - 健康检查 / readiness
  - 操作日志 / 审计记录
  - 基础运行指标（上传数、失败数、存储占用趋势）
- **Explicitly Excluded:**
  - 完整 APM 平台集成
  - 分布式 tracing
- **Dependencies:** 数据与部署能力补全
- **Risks:**
  - 指标口径如果先定义不清，后面容易推翻重来
  - 审计日志要兼顾隐私和排障价值
- **Deliverable:** 项目更容易部署、监控、排障和长期维护。
- **Next Step:** 在部署闭环完成后定义最小可观测性集合。

## Recommended Execution Order

1. **核心安全与访问控制补全**
   - 先补 correctness / security gap，避免后续功能建立在不稳基础上。
2. **数据与部署能力补全**
   - 确保项目不仅“代码支持”，而且“真实可部署”。
3. **图片管理体验完善**
   - 这是用户最常用链路，投入产出比高。
4. **管理后台设置与账户闭环**
   - 补足单用户产品的运维闭环。
5. **相册与内容组织增强**
   - 在基础管理体验稳定后做更合适。
6. **分享与生态集成增强**
   - 依赖访问策略和部署能力先稳定。
7. **运营与可观测性增强**
   - 适合作为项目进入长期运行阶段时补齐。

## Human Decision Needed

建议你下一步从下面三个方向里选一个作为首个子项目：

1. **安全优先**：先做“核心安全与访问控制补全”
2. **可部署优先**：先做“数据与部署能力补全”
3. **体验优先**：先做“图片管理体验完善”

一旦你选定，我下一步就进入该子项目的 brainstorming。