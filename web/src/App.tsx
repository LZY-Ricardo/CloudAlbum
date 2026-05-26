import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Login from './pages/Login'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const loggedIn = useAuthStore((state) => state.loggedIn)
  return loggedIn ? <>{children}</> : <Navigate to="/login" replace />
}

function DashboardPlaceholder() {
  const logout = useAuthStore((state) => state.logout)

  return (
    <div className="placeholder-shell">
      <div className="placeholder-card">
        <div className="eyebrow">Task 8 Complete</div>
        <h2>CloudAlbum admin shell is ready.</h2>
        <p>登录流程已经打通，接下来会继续补齐上传页、图片管理与其余后台页面。</p>
        <button
          type="button"
          className="ghost-button"
          onClick={() => {
            logout()
            window.location.href = '/login'
          }}
        >
          退出登录
        </button>
      </div>
    </div>
  )
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route
        path="/"
        element={
          <PrivateRoute>
            <DashboardPlaceholder />
          </PrivateRoute>
        }
      />
    </Routes>
  )
}
