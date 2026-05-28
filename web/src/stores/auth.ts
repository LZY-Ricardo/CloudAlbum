import { create } from 'zustand'
import client from '../api/client'

export type MeInfo = {
  user_id: number
  username?: string
  auth_type?: string
  token_scope?: string
  created_at?: string
  password_changed_at?: string | null
  uses_default_password?: boolean
}

type AuthState = {
  token: string | null
  loggedIn: boolean
  me: MeInfo | null
  bannerDismissed: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  reset: () => void
  init: () => Promise<void>
  refreshMe: () => Promise<void>
  applyNewToken: (token: string) => Promise<void>
  dismissBanner: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  loggedIn: Boolean(localStorage.getItem('token')),
  me: null,
  bannerDismissed: false,

  async login(username: string, password: string) {
    const response = await client.post('/auth/login', { username, password })
    const token = response.data.token as string
    localStorage.setItem('token', token)
    set({ token, loggedIn: true, bannerDismissed: false })
    await get().refreshMe()
  },

  logout() {
    localStorage.removeItem('token')
    set({ token: null, loggedIn: false, me: null, bannerDismissed: false })
  },

  reset() {
    set({ token: null, loggedIn: false, me: null, bannerDismissed: false })
  },

  async init() {
    const token = localStorage.getItem('token')
    set({ token, loggedIn: Boolean(token) })
    if (token) {
      await get().refreshMe()
    }
  },

  async refreshMe() {
    try {
      const { data } = await client.get('/auth/me')
      set({ me: data as MeInfo })
    } catch {
      // 401 走全局拦截器；其他错误忽略
    }
  },

  async applyNewToken(token: string) {
    localStorage.setItem('token', token)
    set({ token })
    await get().refreshMe()
  },

  dismissBanner() {
    set({ bannerDismissed: true })
  },
}))
