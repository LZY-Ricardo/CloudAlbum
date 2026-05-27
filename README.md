# CloudAlbum

私有图床管理后台。Go + Gin + GORM 后端，React + Vite + Arco Design 前端。

## 配置管理

- `configs/config.yaml` 仅作为**首次启动的种子**。第一次启动时如果 `settings` 表为空，
  CloudAlbum 会插入一行空 overrides；运行时配置最终生效值是
  `YAML 基线 ⊕ DB settings overrides`，以 DB 为权威源。
- 之后**直接修改 `config.yaml` 不再生效**。需要修改可热更新的配置（站点 base_url
  与图片处理策略），请通过后台 **Settings** 页编辑；保存后立即生效。
- 如果你确实希望让 yaml 的新值覆盖运行时：删除 `settings` 表中 `id=1` 的行
  （例如 `sqlite3 data/cloudalbum.db 'DELETE FROM settings WHERE id=1;'`），重启进程即可。
- 启动时若 yaml mtime 比 `settings.updated_at` 晚，日志会打一条 INFO 提醒，但不会
  自动覆盖运行时。
- 数据库、对象存储、`server.port`、`auth.jwt_secret` 等不在可热更新范围内，仍以
  yaml / 环境变量为准。

### UI 可编辑字段（本期）

| 字段 | 位置 | 备注 |
|---|---|---|
| `server.base_url` | Settings → 站点 | 影响图片公开链接前缀，立即生效 |
| `image.max_size` | Settings → 图片处理 | 单位 MB，范围 (0, 1024] |
| `image.allowed_types` | Settings → 图片处理 | 至少一项；仅允许 `jpg/jpeg/png/gif/webp/bmp/svg` |
| `image.auto_convert` | Settings → 图片处理 | `""` / `webp` / `jpeg` |
| `image.quality` | Settings → 图片处理 | 范围 [1, 100] |
| `image.strip_exif` | Settings → 图片处理 | 是否移除 EXIF / 隐私元数据 |

不在 UI 暴露的字段（`thumbnails`、`token_expire`、`database`、`storage`、
`server.port`、`auth.jwt_secret`）仍以 YAML 为准。

## 账户与改密

- 默认登录账号：`admin / admin123`。
- 登录后顶部会出现「默认密码 Banner」，提示尽快修改密码（不强制阻塞）。
- 在 **Account** 页可以查看账号信息、修改密码。密码至少 8 位且不能与旧密码相同。
- 修改密码后所有旧 JWT 立刻失效；当前会话续签新 JWT。API Token 不受改密影响。
