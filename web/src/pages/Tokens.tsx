import { useEffect, useState } from 'react'
import client from '../api/client'

type TokenItem = {
  id: number
  name: string
  scope: string
  last_used_at?: string | null
}

export default function Tokens() {
  const [tokens, setTokens] = useState<TokenItem[]>([])
  const [name, setName] = useState('')
  const [scope, setScope] = useState('upload')
  const [rawToken, setRawToken] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const fetchTokens = async () => {
    setLoading(true)
    setError('')
    try {
      const response = await client.get('/tokens')
      setTokens(response.data.tokens ?? [])
    } catch {
      setError('Token 列表加载失败。')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void fetchTokens()
  }, [])

  const createToken = async () => {
    if (!name.trim()) return
    setLoading(true)
    setError('')
    try {
      const response = await client.post('/tokens', { name, scope })
      setRawToken(response.data.raw_token ?? '')
      setName('')
      await fetchTokens()
    } catch {
      setError('创建 Token 失败。')
    } finally {
      setLoading(false)
    }
  }

  const deleteToken = async (id: number) => {
    if (!window.confirm('确认删除这个 Token 吗？')) return
    setLoading(true)
    setError('')
    try {
      await client.delete(`/tokens/${id}`)
      await fetchTokens()
    } catch {
      setError('删除 Token 失败。')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">API tokens</div>
        <h2 className="section-title">管理上传令牌</h2>
        <div className="management-form-grid tokens-grid">
          <input className="url-input" value={name} onChange={(e) => setName(e.target.value)} placeholder="Token 名称" />
          <select value={scope} onChange={(e) => setScope(e.target.value)}>
            <option value="read">只读</option>
            <option value="upload">上传</option>
            <option value="full">完整权限</option>
          </select>
        </div>
        <div className="management-actions">
          <button type="button" className="gradient-button" onClick={() => void createToken()} disabled={loading}>
            创建 Token
          </button>
        </div>
        {rawToken ? (
          <div className="glass-subpanel token-raw-panel">
            <div className="result-meta">仅显示一次，请立即保存：</div>
            <pre className="result-link">{rawToken}</pre>
          </div>
        ) : null}
        {error ? <div className="inline-error">{error}</div> : null}
      </section>

      <section className="management-list">
        {tokens.length === 0 ? (
          <div className="glass-panel empty-state">还没有创建任何 Token。</div>
        ) : (
          tokens.map((token) => (
            <article key={token.id} className="glass-panel management-card">
              <div>
                <div className="result-name">{token.name}</div>
                <div className="result-meta">Scope: {token.scope} · Last used: {token.last_used_at || 'Never'}</div>
              </div>
              <button type="button" className="ghost-button danger-button" onClick={() => void deleteToken(token.id)}>
                删除
              </button>
            </article>
          ))
        )}
      </section>
    </div>
  )
}
