import { useMemo, useRef, useState } from 'react'
import client from '../api/client'

type UploadItem = {
  image: {
    id: number
    original_name: string
    size: number
    storage_key: string
  }
  urls: Record<string, string>
}

type Album = {
  id: number
  name: string
}

export default function Upload() {
  const [uploading, setUploading] = useState(false)
  const [progress, setProgress] = useState(0)
  const [error, setError] = useState('')
  const [results, setResults] = useState<UploadItem[]>([])
  const [urlValue, setUrlValue] = useState('')
  const [albumId, setAlbumId] = useState('')
  const [albums, setAlbums] = useState<Album[]>([])
  const [linkMode, setLinkMode] = useState<'url' | 'markdown' | 'html' | 'bbcode'>('markdown')
  const [dragActive, setDragActive] = useState(false)
  const inputRef = useRef<HTMLInputElement | null>(null)

  useMemo(() => {
    client.get('/albums').then((res) => setAlbums(res.data.albums ?? [])).catch(() => {})
  }, [])

  const uploadFiles = async (files: FileList | File[]) => {
    if (!files.length) return
    setUploading(true)
    setProgress(0)
    setError('')

    const formData = new FormData()
    Array.from(files).forEach((file) => formData.append('files', file))
    if (albumId) formData.append('album_id', albumId)

    try {
      const response = await client.post('/images', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (event) => {
          if (event.total) {
            setProgress(Math.round((event.loaded / event.total) * 100))
          }
        },
      })
      const next = (response.data.results ?? [])
        .filter((item: UploadItem | { error?: string }) => 'image' in item)
      setResults((prev) => [...next, ...prev])
    } catch {
      setError('上传失败，请稍后重试。')
    } finally {
      setUploading(false)
      setProgress(0)
    }
  }

  const handleDrop: React.DragEventHandler<HTMLDivElement> = (event) => {
    event.preventDefault()
    setDragActive(false)
    if (event.dataTransfer.files?.length) {
      void uploadFiles(event.dataTransfer.files)
    }
  }

  const handlePaste: React.ClipboardEventHandler<HTMLDivElement> = (event) => {
    const images = Array.from(event.clipboardData.items)
      .filter((item) => item.type.startsWith('image/'))
      .map((item) => item.getAsFile())
      .filter((file): file is File => Boolean(file))
    if (images.length > 0) {
      void uploadFiles(images)
    }
  }

  const handleUrlUpload = async () => {
    if (!urlValue.trim()) return
    setUploading(true)
    setError('')
    try {
      const response = await client.post('/images/upload-url', {
        url: urlValue.trim(),
        album_id: albumId ? Number(albumId) : undefined,
      })
      setResults((prev) => [response.data, ...prev])
      setUrlValue('')
    } catch {
      setError('远程 URL 上传失败。')
    } finally {
      setUploading(false)
    }
  }

  const copyAll = async () => {
    const text = results.map((item) => item.urls[linkMode]).join('\n')
    if (!text) return
    await navigator.clipboard.writeText(text)
  }

  return (
    <div className="upload-page" onPaste={handlePaste}>
      <section className="upload-panel glass-panel">
        <div className="upload-panel-header">
          <div>
            <div className="eyebrow">Upload center</div>
            <h2 className="section-title">将图片拖进来，或让链接自己落地。</h2>
            <p className="section-copy">支持拖拽、多文件、剪贴板粘贴与远程 URL 拉取，并在完成后立即生成多格式外链。</p>
          </div>
          <div className="upload-toolbar">
            <label className="field-inline">
              <span>目标相册</span>
              <select value={albumId} onChange={(e) => setAlbumId(e.target.value)}>
                <option value="">未分组</option>
                {albums.map((album) => (
                  <option key={album.id} value={album.id}>{album.name}</option>
                ))}
              </select>
            </label>
          </div>
        </div>

        <div
          className={`upload-dropzone${dragActive ? ' is-drag-active' : ''}`}
          onDragEnter={(event) => {
            event.preventDefault()
            setDragActive(true)
          }}
          onDragOver={(event) => event.preventDefault()}
          onDragLeave={(event) => {
            event.preventDefault()
            setDragActive(false)
          }}
          onDrop={handleDrop}
          onClick={() => inputRef.current?.click()}
          role="button"
          tabIndex={0}
        >
          <div className="upload-dropzone-icon">⬆</div>
          <h3>Drop images, paste screenshots, or click to choose files</h3>
          <p>支持多图批量上传，拖拽会自动保持当前相册上下文。</p>
          <input
            ref={inputRef}
            className="sr-only"
            type="file"
            multiple
            accept="image/*"
            onChange={(event) => {
              if (event.target.files?.length) {
                void uploadFiles(event.target.files)
              }
            }}
          />
        </div>

        <div className="url-upload-row">
          <input
            className="url-input"
            value={urlValue}
            onChange={(event) => setUrlValue(event.target.value)}
            placeholder="粘贴远程图片 URL，例如 https://example.com/cover.jpg"
          />
          <button type="button" className="gradient-button" onClick={() => void handleUrlUpload()} disabled={uploading}>
            拉取 URL
          </button>
        </div>

        {uploading ? (
          <div className="upload-progress glass-subpanel">
            <div className="upload-progress-bar">
              <span style={{ width: `${progress}%` }} />
            </div>
            <div className="upload-progress-meta">上传中… {progress}%</div>
          </div>
        ) : null}

        {error ? <div className="inline-error">{error}</div> : null}
      </section>

      <section className="results-panel glass-panel">
        <div className="results-header">
          <div>
            <div className="eyebrow">Generated links</div>
            <h2 className="section-title">上传结果与外链格式</h2>
          </div>
          <div className="results-actions">
            <select value={linkMode} onChange={(e) => setLinkMode(e.target.value as typeof linkMode)}>
              <option value="markdown">Markdown</option>
              <option value="url">URL</option>
              <option value="html">HTML</option>
              <option value="bbcode">BBCode</option>
            </select>
            <button type="button" className="ghost-button" onClick={() => void copyAll()}>
              复制全部
            </button>
          </div>
        </div>

        <div className="results-list">
          {results.length === 0 ? (
            <div className="empty-state glass-subpanel">
              上传完成后，这里会显示图片信息和不同格式的可复制链接。
            </div>
          ) : (
            results.map((item) => (
              <article key={item.image.id} className="result-card glass-subpanel">
                <div className="result-card-head">
                  <div>
                    <div className="result-name">{item.image.original_name}</div>
                    <div className="result-meta">ID #{item.image.id} · {(item.image.size / 1024).toFixed(1)} KB</div>
                  </div>
                  <button
                    type="button"
                    className="ghost-button"
                    onClick={() => void navigator.clipboard.writeText(item.urls[linkMode])}
                  >
                    复制当前格式
                  </button>
                </div>
                <pre className="result-link">{item.urls[linkMode]}</pre>
              </article>
            ))
          )}
        </div>
      </section>
    </div>
  )
}
