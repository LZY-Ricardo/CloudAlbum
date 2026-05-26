export default function Settings() {
  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Settings</div>
        <h2 className="section-title">系统配置概览</h2>
        <p className="section-copy">当前阶段先展示系统默认配置方向，后续任务再逐步补可编辑设置与更细粒度的偏好管理。</p>
      </section>

      <section className="settings-grid">
        <article className="glass-panel settings-card">
          <div className="result-name">Storage</div>
          <div className="result-meta">当前默认使用本地存储，设计上已兼容后续 S3 扩展。</div>
        </article>
        <article className="glass-panel settings-card">
          <div className="result-name">Image pipeline</div>
          <div className="result-meta">上传时进行类型识别、缩略图生成与链接格式输出。</div>
        </article>
        <article className="glass-panel settings-card">
          <div className="result-name">Auth</div>
          <div className="result-meta">后台登录使用 JWT，外部上传工具使用 API Token。</div>
        </article>
      </section>
    </div>
  )
}
