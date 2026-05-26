import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Login from './pages/Login'
import Upload from './pages/Upload'
import Images from './pages/Images'
import Layout from './components/Layout'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const loggedIn = useAuthStore((state) => state.loggedIn)
  return loggedIn ? <>{children}</> : <Navigate to="/login" replace />
}

function DashboardPlaceholder() {
  return (
    <div className="glass-panel dashboard-placeholder">
      <div className="eyebrow">Overview</div>
      <h2 className="section-title">后台壳层已经就位。</h2>
      <p className="section-copy">上传页和图片管理页已经接入，仪表盘与更多后台视图会在后续任务中继续补齐。</p>
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
            <Layout />
          </PrivateRoute>
        }
      >
        <Route index element={<DashboardPlaceholder />} />
        <Route path="upload" element={<Upload />} />
        <Route path="images" element={<Images />} />
      </Route>
    </Routes>
  )
}
