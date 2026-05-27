import { useEffect, useRef, useState } from 'react'
import { useLocation } from 'react-router-dom'
import { Button, Form, Input, Message, Typography } from '@arco-design/web-react'
import client from '../api/client'
import { useAuthStore } from '../stores/auth'

const { Title } = Typography

export default function Account() {
  const location = useLocation()
  const me = useAuthStore((s) => s.me)
  const applyNewToken = useAuthStore((s) => s.applyNewToken)
  const refreshMe = useAuthStore((s) => s.refreshMe)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const oldInputRef = useRef<HTMLInputElement | null>(null)

  useEffect(() => {
    if (!me) {
      refreshMe()
    }
  }, [me, refreshMe])

  useEffect(() => {
    if ((location.state as { focusCurrentPassword?: boolean } | null)?.focusCurrentPassword) {
      oldInputRef.current?.focus()
    }
  }, [location.state])

  const validate = (): string => {
    if (!oldPassword) return '请输入当前密码'
    if (newPassword.length < 8) return '新密码至少 8 位'
    if (newPassword === oldPassword) return '新密码不能与当前密码相同'
    if (newPassword !== confirmPassword) return '两次输入的新密码不一致'
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
      const { data } = await client.post('/auth/change-password', {
        old_password: oldPassword,
        new_password: newPassword,
      })
      await applyNewToken(data.token as string)
      Message.success('密码已修改')
      setOldPassword('')
      setNewPassword('')
      setConfirmPassword('')
    } catch (err: any) {
      const code = err?.response?.data?.error
      if (code === 'wrong_old_password') setError('当前密码错误')
      else if (code === 'same_as_old') setError('新密码不能与当前密码相同')
      else if (code === 'invalid_request') setError('密码格式不符合要求（至少 8 位）')
      else setError('修改失败，请稍后再试')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="management-page">
      <section className="glass-panel management-form-panel">
        <div className="eyebrow">Account</div>
        <Title heading={4} className="section-title">账号信息</Title>
        <dl className="account-meta-grid">
          <div><dt>用户名</dt><dd>{me?.username ?? '-'}</dd></div>
          <div><dt>角色</dt><dd>{me?.auth_type === 'jwt' ? 'admin' : '-'}</dd></div>
          <div><dt>创建时间</dt><dd>{me?.created_at ? new Date(me.created_at).toLocaleString() : '-'}</dd></div>
          <div><dt>上次改密</dt><dd>{me?.password_changed_at ? new Date(me.password_changed_at).toLocaleString() : '从未'}</dd></div>
        </dl>
      </section>

      <section className="glass-panel management-form-panel">
        <Title heading={4} className="section-title">修改密码</Title>
        <Form layout="vertical" onSubmit={handleSubmit}>
          <Form.Item label="当前密码">
            <Input.Password
              ref={(el) => { oldInputRef.current = (el as unknown as HTMLInputElement) ?? null }}
              value={oldPassword}
              onChange={setOldPassword}
              placeholder="请输入当前密码"
            />
          </Form.Item>
          <Form.Item label="新密码（≥ 8 位）">
            <Input.Password value={newPassword} onChange={setNewPassword} placeholder="新密码" />
          </Form.Item>
          <Form.Item label="确认新密码">
            <Input.Password value={confirmPassword} onChange={setConfirmPassword} placeholder="再次输入新密码" />
          </Form.Item>
          {error ? <div className="form-error">{error}</div> : null}
          <Button type="primary" htmlType="submit" loading={submitting}>保存</Button>
        </Form>
      </section>
    </div>
  )
}
