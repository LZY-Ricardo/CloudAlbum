import { useEffect, useState } from 'react'
import { Button, Checkbox, Form, Input, InputNumber, Message, Radio, Slider, Tag, Typography } from '@arco-design/web-react'
import client from '../api/client'

const { Title } = Typography

const SUPPORTED_TYPES = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp', 'svg']

type SettingsSnapshot = {
  effective: {
    server: { base_url: string }
    image: {
      max_size: number
      allowed_types: string[]
      auto_convert: string
      quality: number
      strip_exif: boolean
    }
  }
  overrides: {
    server: { base_url?: boolean }
    image: {
      max_size?: boolean
      allowed_types?: boolean
      auto_convert?: boolean
      quality?: boolean
      strip_exif?: boolean
    }
  }
  editable_fields: string[]
}

export default function Settings() {
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [snapshot, setSnapshot] = useState<SettingsSnapshot | null>(null)
  const [baseURL, setBaseURL] = useState('')
  const [maxSizeMB, setMaxSizeMB] = useState<number>(50)
  const [allowedTypes, setAllowedTypes] = useState<string[]>([])
  const [autoConvert, setAutoConvert] = useState<string>('')
  const [quality, setQuality] = useState<number>(85)
  const [stripExif, setStripExif] = useState<boolean>(true)

  const load = async () => {
    setLoading(true)
    setError('')
    try {
      const { data } = await client.get<SettingsSnapshot>('/settings')
      setSnapshot(data)
      setBaseURL(data.effective.server.base_url)
      setMaxSizeMB(Math.round(data.effective.image.max_size / (1024 * 1024)))
      setAllowedTypes(data.effective.image.allowed_types)
      setAutoConvert(data.effective.image.auto_convert)
      setQuality(data.effective.image.quality)
      setStripExif(data.effective.image.strip_exif)
    } catch {
      setError('加载设置失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    load()
  }, [])

  const validate = (): string => {
    if (!/^https?:\/\/.+/.test(baseURL)) return '站点 Base URL 必须是 http/https 链接'
    if (maxSizeMB <= 0 || maxSizeMB > 1024) return '最大大小需在 1–1024 MB 之间'
    if (allowedTypes.length === 0) return '至少选择一种图片格式'
    if (quality < 1 || quality > 100) return '压缩质量需在 1–100 之间'
    return ''
  }

  const handleSubmit = async () => {
    const v = validate()
    if (v) {
      setError(v)
      return
    }
    setSubmitting(true)
    setError('')
    try {
      const payload = {
        server: { base_url: baseURL },
        image: {
          max_size: maxSizeMB * 1024 * 1024,
          allowed_types: allowedTypes,
          auto_convert: autoConvert,
          quality,
          strip_exif: stripExif,
        },
      }
      await client.put('/settings', payload)
      Message.success('设置已保存')
      await load()
    } catch (err: any) {
      const code = err?.response?.data?.error
      if (code === 'unknown_field') setError(`存在未识别字段：${err.response.data.field ?? ''}`)
      else if (code === 'invalid_value') setError(err.response.data.detail ?? '字段值不合法')
      else setError('保存失败')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <div className="management-page">加载中…</div>

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Settings</div>
        <Title heading={4} className="section-title">站点</Title>
        <Form layout="vertical">
          <Form.Item
            label={
              <span>
                Base URL
                {snapshot?.overrides.server.base_url ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}
              </span>
            }
            help="用于生成图片公开链接。立即生效。"
          >
            <Input value={baseURL} onChange={setBaseURL} placeholder="https://img.example.com" />
          </Form.Item>
        </Form>
      </section>

      <section className="glass-panel management-form-panel">
        <Title heading={4} className="section-title">图片处理</Title>
        <Form layout="vertical">
          <Form.Item label={<span>最大大小（MB）{snapshot?.overrides.image.max_size ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <InputNumber min={1} max={1024} value={maxSizeMB} onChange={(v) => setMaxSizeMB(Number(v) || 0)} />
          </Form.Item>
          <Form.Item label={<span>允许格式{snapshot?.overrides.image.allowed_types ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Checkbox.Group value={allowedTypes} onChange={setAllowedTypes}>
              {SUPPORTED_TYPES.map((t) => <Checkbox key={t} value={t}>{t}</Checkbox>)}
            </Checkbox.Group>
          </Form.Item>
          <Form.Item label={<span>自动转换{snapshot?.overrides.image.auto_convert ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Radio.Group value={autoConvert} onChange={setAutoConvert}>
              <Radio value="">不转换</Radio>
              <Radio value="webp">WebP</Radio>
              <Radio value="jpeg">JPEG</Radio>
            </Radio.Group>
          </Form.Item>
          <Form.Item label={<span>压缩质量{snapshot?.overrides.image.quality ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Slider min={1} max={100} value={quality} onChange={(v) => setQuality(Number(v))} />
          </Form.Item>
          <Form.Item label={<span>EXIF 剥离{snapshot?.overrides.image.strip_exif ? <Tag color="orange" style={{ marginLeft: 8 }}>已修改</Tag> : null}</span>}>
            <Checkbox checked={stripExif} onChange={setStripExif}>移除 EXIF / 隐私元数据</Checkbox>
          </Form.Item>
        </Form>
      </section>

      {error ? <div className="form-error">{error}</div> : null}
      <div style={{ display: 'flex', gap: 12, marginTop: 12 }}>
        <Button onClick={load} disabled={submitting}>恢复未保存的修改</Button>
        <Button type="primary" loading={submitting} onClick={handleSubmit}>保存</Button>
      </div>
    </div>
  )
}
