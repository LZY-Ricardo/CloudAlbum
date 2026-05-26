import { useEffect, useState } from 'react'
import client from '../api/client'

type Stats = {
  count: number
  total_size: number
}

type ImageItem = {
  image: {
    id: number
    original_name: string
    size: number
    storage_key: string
  }
  urls: Record<string, string>
}

export default function Dashboard() {
  const [stats, setStats] = useState<Stats>({ count: 0, total_size: 0 })
  const [recent, setRecent] = useState<ImageItem[]>([])
  const [error, setError] = useState('')

  useEffect(() => {
    client.get('/images/stats').then((res) => setStats(res.data)).catch(() => setError('统计信息加载失败。'))
    client.get('/images', { params: { page: 1, page_size: 5 } }).then((res) => setRecent(res.data.images ?? [])).catch(() => setError('最近上传加载失败。'))
  }, [])

  return (
    <div className="dashboard-page">
      <section className="dashboard-metrics-grid">
        <article className="glass-panel metric-card">
          <div className="eyebrow">Library</div>
          <div className="metric-value">{stats.count}</div>
          <div className="metric-label">总图片数</div>
        </article>
        <article className="glass-panel metric-card">
          <div className="eyebrow">Storage</div>
          <div className="metric-value">{(stats.total_size / 1024 / 1024).toFixed(2)}</div>
          <div className="metric-label">已使用 MB</div>
        </article>
      </section>

      <section className="glass-panel dashboard-feed-panel">
        <div className="eyebrow">Recent uploads</div>
        <h2 className="section-title">最近上传</h2>
        {error ? <div className="inline-error">{error}</div> : null}
        <div className="dashboard-feed-list">
          {recent.length === 0 ? (
            <div className="empty-state glass-subpanel">还没有最近上传记录。</div>
          ) : (
            recent.map((item) => (
              <article key={item.image.id} className="dashboard-feed-item glass-subpanel">
                <img src={item.urls.url} alt={item.image.original_name} className="dashboard-feed-thumb" />
                <div>
                  <div className="result-name">{item.image.original_name}</div>
                  <div className="result-meta">ID #{item.image.id} · {(item.image.size / 1024).toFixed(1)} KB</div>
                </div>
              </article>
            ))
          )}
        </div>
      </section>
    </div>
  )
}
