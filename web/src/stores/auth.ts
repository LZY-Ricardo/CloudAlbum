import { create } from 'zustand'
import client from '../api/client'

type AuthState = {
  token: string | null
  loggedIn: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  reset: () => void
  init: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  loggedIn: Boolean(localStorage.getItem('token')),
  async login(username: string, password: string) {
    const response = await client.post('/auth/login', { username, password })
    const token = response.data.token as string
    localStorage.setItem('token', token)
    set({ token, loggedIn: true })
  },
  logout() {
    localStorage.removeItem('token')
    set({ token: null, loggedIn: false })
  },
  reset() {
    set({ token: null, loggedIn: false })
  },
  init() {
    const token = localStorage.getItem('token')
    set({ token, loggedIn: Boolean(token) })
  },
}))
