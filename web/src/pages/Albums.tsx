import { useEffect, useState } from 'react'
import client from '../api/client'

type Album = {
  id: number
  name: string
  description: string
}

export default function Albums() {
  const [albums, setAlbums] = useState<Album[]>([])
  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [editing, setEditing] = useState<Album | null>(null)

  const fetchAlbums = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await client.get('/albums')
      setAlbums(response.data.albums ?? [])
    } catch {
      setError('相册列表加载失败。')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchAlbums()
  }, [])

  const resetForm = () => {
    setName('')
    setDescription('')
    setEditing(null)
  }

  const submit = async () => {
    if (!name.trim()) return
    setLoading(true)
    setError('')
    try {
      if (editing) {
        await client.put(`/albums/${editing.id}`, { name, description })
      } else {
        await client.post('/albums', { name, description })
      }
      resetForm()
      await fetchAlbums()
    } catch {
      setError(editing ? '更新相册失败。' : '创建相册失败。')
    } finally {
      setLoading(false)
    }
  }

  const removeAlbum = async (id: number) => {
    if (!window.confirm('确认删除这个相册吗？其中图片会被移出相册。')) return
    setLoading(true)
    setError('')
    try {
      await client.delete(`/albums/${id}`)
      await fetchAlbums()
    } catch {
      setError('删除相册失败。')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Albums</div>
        <h2 className="section-title">管理相册分组</h2>
        <div className="management-form-grid">
          <input className="url-input" value={name} onChange={(e) => setName(e.target.value)} placeholder="相册名称" />
          <input className="url-input" value={description} onChange={(e) => setDescription(e.target.value)} placeholder="相册描述" />
        </div>
        <div className="management-actions">
          <button type="button" className="gradient-button" onClick={() => void submit()} disabled={loading}>
            {editing ? '保存修改' : '创建相册'}
          </button>
          {editing ? (
            <button type="button" className="ghost-button" onClick={resetForm}>取消编辑</button>
          ) : null}
        </div>
        {error ? <div className="inline-error">{error}</div> : null}
      </section>

      <section className="management-list">
        {albums.length === 0 ? (
          <div className="glass-panel empty-state">还没有相册，先创建一个新的分组。</div>
        ) : (
          albums.map((album) => (
            <article key={album.id} className="glass-panel management-card">
              <div>
                <div className="result-name">{album.name}</div>
                <div className="result-meta">{album.description || '暂无描述'}</div>
              </div>
              <div className="management-actions">
                <button type="button" className="ghost-button" onClick={() => { setEditing(album); setName(album.name); setDescription(album.description || '') }}>
                  编辑
                </button>
                <button type="button" className="ghost-button danger-button" onClick={() => void removeAlbum(album.id)}>
                  删除
                </button>
              </div>
            </article>
          ))
        )}
      </section>
    </div>
  )
}
