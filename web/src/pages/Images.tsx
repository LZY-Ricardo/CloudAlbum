import { useEffect, useMemo, useState } from 'react'
import client from '../api/client'

type ImageItem = {
  image: {
    id: number
    original_name: string
    size: number
    storage_key: string
    created_at?: string
  }
  urls: Record<string, string>
}

type Album = {
  id: number
  name: string
}

export default function Images() {
  const [images, setImages] = useState<ImageItem[]>([])
  const [albums, setAlbums] = useState<Album[]>([])
  const [selected, setSelected] = useState<number[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [keyword, setKeyword] = useState('')
  const [albumFilter, setAlbumFilter] = useState('')
  const [moveTarget, setMoveTarget] = useState('')
  const [preview, setPreview] = useState<ImageItem | null>(null)

  const selectedSet = useMemo(() => new Set(selected), [selected])

  const fetchImages = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await client.get('/images', {
        params: {
          page: 1,
          page_size: 100,
          keyword: keyword || undefined,
          album_id: albumFilter || undefined,
        },
      })
      setImages(response.data.images ?? [])
    } catch {
      setError('加载图片列表失败。')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    setSelected([])
    void fetchImages()
  }, [keyword, albumFilter])

  useEffect(() => {
    client.get('/albums').then((res) => setAlbums(res.data.albums ?? [])).catch(() => {})
  }, [])

  const toggleSelection = (id: number) => {
    setSelected((prev) => (prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]))
  }

  const clearSelection = () => setSelected([])

  const handleBatchDelete = async () => {
    if (selected.length === 0) return
    if (!window.confirm(`确认删除这 ${selected.length} 张图片吗？你之后仍可在回收站恢复它们。`)) {
      return
    }
    setLoading(true)
    setError('')
    try {
      await client.post('/images/batch', { ids: selected, action: 'delete' })
      clearSelection()
      await fetchImages()
    } catch {
      setError('批量删除失败。')
    } finally {
      setLoading(false)
    }
  }

  const handleBatchMove = async () => {
    if (selected.length === 0 || !moveTarget) return
    setLoading(true)
    setError('')
    try {
      await client.post('/images/batch', {
        ids: selected,
        action: 'move',
        album_id: moveTarget === '__none__' ? null : Number(moveTarget),
      })
      clearSelection()
      setMoveTarget('')
      await fetchImages()
    } catch {
      setError('批量移动失败。')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="images-page">
      <section className="glass-panel images-toolbar-panel">
        <div>
          <div className="eyebrow">Image library</div>
          <h2 className="section-title">沉浸式管理你的图片资产</h2>
          <p className="section-copy">支持搜索、按相册筛选、批量移动与批量删除，并可随时预览图片原始链接。</p>
        </div>

        <div className="images-toolbar-controls">
          <input
            className="url-input"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="按文件名搜索"
          />
          <select value={albumFilter} onChange={(event) => setAlbumFilter(event.target.value)}>
            <option value="">全部相册</option>
            {albums.map((album) => (
              <option key={album.id} value={album.id}>{album.name}</option>
            ))}
          </select>
        </div>
      </section>

      {selected.length > 0 ? (
        <section className="glass-panel selection-bar">
          <div>已选择 {selected.length} 张图片</div>
          <div className="selection-actions">
            <select value={moveTarget} onChange={(event) => setMoveTarget(event.target.value)}>
              <option value="">移动到相册</option>
              <option value="__none__">移出相册</option>
              {albums.map((album) => (
                <option key={album.id} value={album.id}>{album.name}</option>
              ))}
            </select>
            <button type="button" className="ghost-button" onClick={() => void handleBatchMove()} disabled={!moveTarget || loading}>
              批量移动
            </button>
            <button type="button" className="ghost-button danger-button" onClick={() => void handleBatchDelete()} disabled={loading}>
              批量删除
            </button>
            <button type="button" className="ghost-button" onClick={clearSelection}>
              清空选择
            </button>
          </div>
        </section>
      ) : null}

      {error ? <div className="inline-error">{error}</div> : null}

      <section className="images-grid">
        {loading && images.length === 0 ? (
          <div className="glass-panel empty-state">图片列表加载中…</div>
        ) : images.length === 0 ? (
          <div className="glass-panel empty-state">当前没有图片，先去上传中心添加一些内容。</div>
        ) : (
          images.map((item) => {
            const imageUrl = item.urls.url
            const active = selectedSet.has(item.image.id)
            return (
              <article key={item.image.id} className={`image-card glass-panel${active ? ' is-selected' : ''}`}>
                <button
                  type="button"
                  className={`selection-toggle${active ? ' is-selected' : ''}`}
                  onClick={() => toggleSelection(item.image.id)}
                  aria-pressed={active}
                >
                  {active ? '✓' : ''}
                </button>
                <button type="button" className="image-preview-trigger" onClick={() => setPreview(item)}>
                  <img className="image-card-cover" src={imageUrl} alt={item.image.original_name} loading="lazy" />
                </button>
                <div className="image-card-body">
                  <div>
                    <div className="result-name">{item.image.original_name}</div>
                    <div className="result-meta">ID #{item.image.id} · {(item.image.size / 1024).toFixed(1)} KB</div>
                  </div>
                  <div className="image-card-actions">
                    <button type="button" className="ghost-button" onClick={() => void navigator.clipboard.writeText(item.urls.markdown)}>
                      复制 Markdown
                    </button>
                    <button type="button" className="ghost-button" onClick={() => setPreview(item)}>
                      预览
                    </button>
                  </div>
                </div>
              </article>
            )
          })
        )}
      </section>

      {preview ? (
        <div className="preview-overlay" onClick={() => setPreview(null)}>
          <div className="preview-dialog glass-panel" onClick={(event) => event.stopPropagation()}>
            <button type="button" className="preview-close" onClick={() => setPreview(null)}>×</button>
            <img className="preview-image" src={preview.urls.url} alt={preview.image.original_name} />
            <div className="preview-meta">
              <div className="result-name">{preview.image.original_name}</div>
              <div className="result-meta">ID #{preview.image.id} · {(preview.image.size / 1024).toFixed(1)} KB</div>
              <pre className="result-link">{preview.urls.url}</pre>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  )
}
