# CloudAlbum - 个人图床设计文档

> 日期：2026-05-25
> 状态：已批准

## 概述

CloudAlbum 是一个自托管的个人图床服务，采用 Go 后端 + React 前端（内嵌 SPA）的单体架构，Docker 一键部署。

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go，Gin/Echo 框架 |
| 数据库 | GORM ORM，默认 SQLite，可切换 PostgreSQL |
| 前端 | React + TypeScript + Arco Design |
| 图片处理 | Go 原生 image 库 + libvips（可选） |
| 部署 | Docker 单容器，前端构建产物通过 go:embed 嵌入二进制 |

## 项目结构

```
CloudAlbum/
├── cmd/
│   └── server/           # 入口 main.go
├── internal/
│   ├── config/           # 配置加载 (YAML)
│   ├── handler/          # HTTP handlers (API)
│   ├── middleware/        # 鉴权、日志、CORS 等
│   ├── model/            # 数据模型
│   ├── repository/       # 数据库操作
│   ├── service/          # 业务逻辑
│   ├── storage/          # 存储后端接口 + 实现
│   └── image/            # 图片处理（压缩、转码、缩略图）
├── web/                  # React 前端项目
│   ├── src/
│   └── dist/             # 构建产物
├── embed.go              # embed web/dist 到二进制
├── configs/
│   └── config.yaml       # 默认配置
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

## 架构方案：Go API + React SPA 内嵌

Go 后端提供 RESTful API，React 前端构建后将静态文件通过 `go:embed` 嵌入到 Go 二进制中。部署时只需一个 Docker 容器。

- 开发时前后端独立（前端 dev server 代理到后端）
- 生产环境前端静态文件由 Go 直接服务
- 单容器部署，无额外依赖

## 数据模型

### User（用户）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| username | string | 用户名 |
| password_hash | string | 密码哈希 |
| role | string | 角色 (admin) |
| created_at | time | 创建时间 |

### Image（图片）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| user_id | uint | 上传者 |
| storage_key | string | 存储路径 |
| filename | string | 存储文件名 |
| original_name | string | 原始文件名 |
| size | int64 | 文件大小 |
| mime_type | string | MIME 类型 |
| width | int | 图片宽度 |
| height | int | 图片高度 |
| hash | string | SHA256，用于去重 |
| album_id | uint | 所属相册（nullable） |
| created_at | time | 创建时间 |
| deleted_at | time | 软删除时间（gorm.DeletedAt） |

去重策略：相同 hash 的文件不重复存储，只在数据库新增一条记录指向同一个存储文件。

### Album（相册/分组）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| user_id | uint | 所属用户 |
| name | string | 相册名 |
| description | string | 描述 |
| cover_image_id | uint | 封面图片（nullable） |
| sort_order | int | 排序 |
| created_at | time | 创建时间 |

单层结构，不支持嵌套子相册。

### Token（API Token）

| 字段 | 类型 | 说明 |
|---|---|---|
| id | uint | 主键 |
| user_id | uint | 所属用户 |
| name | string | Token 名称 |
| token_hash | string | Token 哈希 |
| scope | string | 权限范围 (read/upload/full) |
| last_used_at | time | 最后使用时间 |
| expires_at | time | 过期时间（nullable） |
| created_at | time | 创建时间 |

## API 设计

### 认证

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/v1/auth/login | 登录，返回 JWT |
| POST | /api/v1/auth/logout | 登出 |

### Token 管理

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/v1/tokens | 列出 Token |
| POST | /api/v1/tokens | 创建 Token |
| DELETE | /api/v1/tokens/:id | 删除 Token |

### 图片

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /api/v1/images | 上传（单张/批量，multipart） |
| POST | /api/v1/images/upload-url | 远程 URL 上传 |
| GET | /api/v1/images | 图片列表（分页、按相册/日期筛选） |
| GET | /api/v1/images/:id | 图片详情（含多格式链接） |
| PUT | /api/v1/images/:id | 更新（重命名、移动相册） |
| DELETE | /api/v1/images/:id | 删除图片 |
| POST | /api/v1/images/batch | 批量操作（移动、删除） |

### 相册

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /api/v1/albums | 相册列表 |
| POST | /api/v1/albums | 创建相册 |
| PUT | /api/v1/albums/:id | 更新相册 |
| DELETE | /api/v1/albums/:id | 删除相册 |

### 公共访问（无需鉴权）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /i/:key | 直接访问原图 |
| GET | /t/:key | 访问缩略图 |

## 上传方式

1. **点击选择** — 按钮/区域触发浏览器原生文件选择器，支持多选
2. **拖拽上传** — 拖文件到页面指定区域
3. **剪贴板粘贴** — Ctrl+V 粘贴截图
4. **远程 URL** — 粘贴图片地址拉取到本地存储
5. **API Token** — 供 PicGo 等第三方客户端调用

## 图片处理流水线

上传时同步处理，步骤可配置开关：

```
原始图片
  → EXIF 剥脱（隐私保护）
  → 格式转码（可选 WebP/AVIF）
  → 压缩（质量可配，默认 85）
  → 生成缩略图（多种尺寸）
  → 计算 SHA256（去重）
  → 存储到后端
```

缩略图尺寸：`thumb` (200x200)、`medium` (800x600)、`large` (1200x900)。

## 管理后台

基于 React + Arco Design（Arco Dark 主题 + 自定义样式覆盖）。

### 视觉风格：毛玻璃 + 青绿渐变（暗色主题）

整体调性：深色底、毛玻璃卡片、青绿渐变强调色，现代感强，图片内容突出。

**设计令牌 (Design Tokens):**

```css
:root {
  /* 背景层 */
  --bg-primary: #0a0a0f;
  --bg-secondary: #111118;
  --bg-card: rgba(255, 255, 255, 0.04);
  --bg-card-hover: rgba(255, 255, 255, 0.07);
  --bg-glass: rgba(255, 255, 255, 0.05);
  --bg-sidebar: rgba(17, 17, 24, 0.8);

  /* 强调色 */
  --accent-from: #06b6d4;
  --accent-to: #10b981;
  --accent-solid: #06b6d4;
  --accent-glow: rgba(6, 182, 212, 0.15);

  /* 文字 */
  --text-primary: #f0f0f5;
  --text-secondary: #9ca3af;
  --text-muted: #6b7280;

  /* 边框 */
  --border: rgba(255, 255, 255, 0.08);
  --border-hover: rgba(255, 255, 255, 0.15);

  /* 状态色 */
  --success: #10b981;
  --warning: #f59e0b;
  --danger: #ef4444;

  /* 圆角 */
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 16px;

  /* 阴影 */
  --shadow-card: 0 4px 24px rgba(0, 0, 0, 0.3);
  --shadow-glow: 0 0 20px rgba(6, 182, 212, 0.2);
}
```

**字体:**
- 标题: `Outfit` (Google Fonts) — 几何无衬线，现代感
- 正文: `DM Sans` (Google Fonts) — 清晰易读
- 中文回退: `system-ui, "PingFang SC", "Microsoft YaHei"`

**核心视觉效果:**
1. **毛玻璃卡片** — `backdrop-filter: blur(20px)` + 半透明背景 + 微边框
2. **渐变强调** — 按钮、选中态、进度条使用 `linear-gradient(135deg, #06b6d4, #10b981)`
3. **悬浮微动效** — 卡片 hover 上浮 2px + 边框渐亮 + 轻微阴影增强
4. **图片卡片** — hover 时底部渐变遮罩渐显，操作按钮淡入
5. **页面加载** — stagger fade-in（子元素依次渐现，间隔 50ms）
6. **侧边栏** — 毛玻璃背景 + 选中项渐变高亮条
7. **上传区域** — 虚线边框呼吸动画 + 拖入时边框变实线 + 渐变边框
8. **灯箱** — 深色 overlay + 模糊背景 + 图片 scale-in 动画

**按钮样式:**
- 主要按钮: 渐变背景 (#06b6d4 → #10b981)，hover 时渐变位移，无圆角膨胀
- 次要按钮: 透明底 + 渐变边框
- 危险按钮: #ef4444 实色

**表格/列表:**
- 斑马纹用 rgba(255,255,255,0.02) 交替
- 行 hover 用 rgba(255,255,255,0.05)

### 页面

| 页面 | 功能 |
|---|---|
| 仪表盘 | 存储用量、图片数量、最近上传、流量统计 |
| 图片管理 | 瀑布流/网格视图、多选批量操作、搜索、按相册/日期筛选 |
| 相册管理 | 创建/编辑/删除相册、设置封面、排序 |
| 上传页 | 五种上传方式、上传进度、完成后一键复制链接 |
| 回收站 | 软删除恢复、彻底删除 |
| Token 管理 | 创建/吊销 API Token |
| 系统设置 | 存储后端配置、图片处理规则、站点信息 |

### 交互

- 图片卡片悬浮显示快捷操作（复制链接、移动、删除）
- 右键上下文菜单
- 批量选择后底部浮动操作栏
- 上传完成后自动复制链接（可配置格式）
- 图片预览灯箱（缩放、左右切换）

### 链接输出格式

- 原始 URL：`{base_url}/i/{key}`
- Markdown：`![filename](url)`
- HTML：`<img src="url" alt="filename">`
- BBCode：`[img]url[/img]`

## 存储后端（可插拔）

```go
type Storage interface {
    Save(ctx context.Context, key string, data io.Reader) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)
}
```

内置实现：
- `LocalStorage` — 本地文件系统
- `S3Storage` — S3 兼容对象存储（AWS S3、MinIO、Cloudflare R2 等）

## 安全

- JWT 鉴权，Token 过期自动刷新
- API Token 支持 Scope 权限（只读/上传/全权限）
- 上传文件类型白名单
- 文件大小限制（可配）
- 防盗链（Referer 白名单，可配置开关）
- Rate Limiting（上传频率限制）

## 配置

```yaml
server:
  port: 8080
  base_url: "http://localhost:8080"

database:
  driver: sqlite  # sqlite | postgres
  dsn: "./data/cloudalbum.db"

storage:
  driver: local   # local | s3
  local:
    path: "./data/images"
  s3:
    bucket: ""
    region: ""
    endpoint: ""

image:
  max_size: 50MB
  allowed_types: ["jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"]
  auto_convert: webp
  quality: 85
  strip_exif: true
  thumbnails:
    - name: thumb
      width: 200
      height: 200
    - name: medium
      width: 800
      height: 600
    - name: large
      width: 1200
      height: 900

auth:
  jwt_secret: ""
  token_expire: 7d
```

## 部署

Docker 单容器部署，提供 Dockerfile 和 docker-compose.yml。

- Go 多阶段构建：编译 → 前端构建 → 最终镜像
- 数据持久化：挂载 `/data` 目录（数据库 + 本地图片）
- 支持 ARM64 / AMD64
