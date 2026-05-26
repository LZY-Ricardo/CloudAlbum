import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Card, Form, Input, Typography } from '@arco-design/web-react'
import { IconLock, IconUser } from '@arco-design/web-react/icon'
import { useAuthStore } from '../stores/auth'

const { Title, Paragraph, Text } = Typography

export default function Login() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('admin123')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async () => {
    setLoading(true)
    setError('')
    try {
      await login(username, password)
      navigate('/')
    } catch {
      setError('用户名或密码错误，或服务暂时不可用。')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="login-shell">
      <div className="login-ambient login-ambient-left" />
      <div className="login-ambient login-ambient-right" />
      <div className="login-grid">
        <section className="login-copy">
          <div className="eyebrow">CLOUDALBUM</div>
          <Title heading={1} className="login-title">
            Your image hub, framed in glass.
          </Title>
          <Paragraph className="login-description">
            管理你的个人图床、生成外链、组织相册，并以干净优雅的方式维护整个图库工作流。
          </Paragraph>
          <div className="login-highlights">
            <div className="highlight-card">
              <span className="highlight-label">Storage</span>
              <Text>Local / S3-ready architecture</Text>
            </div>
            <div className="highlight-card">
              <span className="highlight-label">Workflow</span>
              <Text>Upload, organize, link, publish</Text>
            </div>
          </div>
        </section>

        <Card className="login-card" bordered={false}>
          <div className="login-card-header">
            <div className="login-card-badge">Secure Access</div>
            <Title heading={4} className="login-card-title">
              Sign in to CloudAlbum
            </Title>
            <Paragraph className="login-card-subtitle">
              进入管理后台，开始管理图片与相册。
            </Paragraph>
          </div>

          <Form layout="vertical" onSubmit={handleSubmit}>
            <Form.Item label="用户名">
              <Input
                size="large"
                prefix={<IconUser />}
                value={username}
                onChange={setUsername}
                placeholder="请输入用户名"
              />
            </Form.Item>
            <Form.Item label="密码">
              <Input.Password
                size="large"
                prefix={<IconLock />}
                value={password}
                onChange={setPassword}
                placeholder="请输入密码"
              />
            </Form.Item>

            {error ? <div className="login-error">{error}</div> : null}

            <Button
              type="primary"
              htmlType="submit"
              long
              size="large"
              loading={loading}
              className="login-submit"
            >
              进入后台
            </Button>
          </Form>
        </Card>
      </div>
    </div>
  )
}
