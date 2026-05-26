import { useEffect, useState } from 'react'
import client from '../api/client'

type ImageItem = {
  image: {
    id: number
    original_name: string
    size: number
  }
  urls: Record<string, string>
}

export default function Trash() {
  const [items, setItems] = useState<ImageItem[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const fetchTrash = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await client.get('/images', { params: { page: 1, page_size: 100, deleted: true } })
      setItems(response.data.images ?? [])
    } catch {
      setError('回收站加载失败。')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchTrash()
  }, [])

  const restore = async (id: number) => {
    setLoading(true)
    setError('')
    try {
      await client.post(`/images/${id}/restore`)
      await fetchTrash()
    } catch {
      setError('恢复图片失败。')
    } finally {
      setLoading(false)
    }
  }

  const destroy = async (id: number) => {
    if (!window.confirm('确认永久删除这张图片吗？此操作不可恢复。')) return
    setLoading(true)
    setError('')
    try {
      await client.delete(`/images/${id}/permanent`)
      await fetchTrash()
    } catch {
      setError('永久删除失败。')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Trash</div>
        <h2 className="section-title">回收站</h2>
        <p className="section-copy">这里保留软删除后的图片，可恢复或彻底删除。</p>
        {error ? <div className="inline-error">{error}</div> : null}
      </section>

      <section className="management-list">
        {loading && items.length === 0 ? (
          <div className="glass-panel empty-state">回收站加载中…</div>
        ) : items.length === 0 ? (
          <div className="glass-panel empty-state">回收站是空的。</div>
        ) : (
          items.map((item) => (
            <article key={item.image.id} className="glass-panel management-card">
              <div>
                <div className="result-name">{item.image.original_name}</div>
                <div className="result-meta">ID #{item.image.id} · {(item.image.size / 1024).toFixed(1)} KB</div>
              </div>
              <div className="management-actions">
                <button type="button" className="ghost-button" onClick={() => void restore(item.image.id)}>
                  恢复
                </button>
                <button type="button" className="ghost-button danger-button" onClick={() => void destroy(item.image.id)}>
                  彻底删除
                </button>
              </div>
            </article>
          ))
        )}
      </section>
    </div>
  )
}
