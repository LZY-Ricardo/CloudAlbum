import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/auth'

export default function DefaultPasswordBanner() {
  const navigate = useNavigate()
  const usesDefault = useAuthStore((s) => s.me?.uses_default_password === true)
  const dismissed = useAuthStore((s) => s.bannerDismissed)
  const dismiss = useAuthStore((s) => s.dismissBanner)

  if (!usesDefault || dismissed) return null

  return (
    <div className="default-password-banner" role="alert">
      <span className="default-password-banner-icon">⚠️</span>
      <span className="default-password-banner-text">
        你正在使用默认密码 admin123，建议尽快修改。
      </span>
      <div className="default-password-banner-actions">
        <button
          type="button"
          className="default-password-banner-primary"
          onClick={() => navigate('/account', { state: { focusCurrentPassword: true } })}
        >
          立刻修改 →
        </button>
        <button
          type="button"
          className="default-password-banner-secondary"
          onClick={dismiss}
        >
          稍后
        </button>
      </div>
    </div>
  )
}
