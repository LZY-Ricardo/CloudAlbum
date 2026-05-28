import { Navigate, Route, Routes } from 'react-router-dom'
import { useAuthStore } from './stores/auth'
import Login from './pages/Login'
import Upload from './pages/Upload'
import Images from './pages/Images'
import Albums from './pages/Albums'
import Dashboard from './pages/Dashboard'
import Tokens from './pages/Tokens'
import Trash from './pages/Trash'
import Settings from './pages/Settings'
import Account from './pages/Account'
import Layout from './components/Layout'

function PrivateRoute({ children }: { children: React.ReactNode }) {
  const loggedIn = useAuthStore((state) => state.loggedIn)
  return loggedIn ? <>{children}</> : <Navigate to="/login" replace />
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
        <Route index element={<Dashboard />} />
        <Route path="upload" element={<Upload />} />
        <Route path="images" element={<Images />} />
        <Route path="albums" element={<Albums />} />
        <Route path="tokens" element={<Tokens />} />
        <Route path="trash" element={<Trash />} />
        <Route path="account" element={<Account />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  )
}
