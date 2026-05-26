import { Link, Outlet, useLocation } from 'react-router-dom'
import { IconUpload, IconImage, IconHome, IconPoweroff } from '@arco-design/web-react/icon'
import { useAuthStore } from '../stores/auth'

const navItems = [
  { to: '/', label: '概览', icon: <IconHome /> },
  { to: '/upload', label: '上传中心', icon: <IconUpload /> },
  { to: '/images', label: '图片管理', icon: <IconImage /> },
]

export default function Layout() {
  const location = useLocation()
  const logout = useAuthStore((state) => state.logout)

  return (
    <div className="dashboard-shell">
      <aside className="dashboard-sidebar glass-panel">
        <div>
          <div className="sidebar-brand">
            <div className="sidebar-brand-badge">CA</div>
            <div>
              <div className="sidebar-brand-title">CloudAlbum</div>
              <div className="sidebar-brand-subtitle">Private image console</div>
            </div>
          </div>

          <nav className="sidebar-nav">
            {navItems.map((item) => {
              const active = item.to === '/'
                ? location.pathname === '/'
                : location.pathname.startsWith(item.to)
              return (
                <Link
                  key={item.to}
                  to={item.to}
                  className={`sidebar-link${active ? ' is-active' : ''}`}
                >
                  <span className="sidebar-link-icon">{item.icon}</span>
                  <span>{item.label}</span>
                </Link>
              )
            })}
          </nav>
        </div>

        <button
          type="button"
          className="sidebar-logout"
          onClick={() => {
            logout()
            window.location.href = '/login'
          }}
        >
          <IconPoweroff />
          <span>退出登录</span>
        </button>
      </aside>

      <main className="dashboard-main">
        <div className="dashboard-topbar glass-panel">
          <div>
            <div className="eyebrow">CloudAlbum Admin</div>
            <h1 className="dashboard-title">上传与管理你的图片资产</h1>
          </div>
          <div className="dashboard-status">
            <span className="status-dot" />
            <span>Backend connected</span>
          </div>
        </div>

        <div className="dashboard-content">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
