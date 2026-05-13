import { create } from 'zustand'
import * as authApi from '../api/auth'
import * as profileApi from '../api/profile'
import type { User } from '../types'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  initialized: boolean
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, email: string) => Promise<void>
  logout: () => void
  fetchProfile: () => Promise<void>
  loadFromStorage: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: null,
  user: null,
  isAuthenticated: false,
  initialized: false,
  loading: false,

  login: async (username, password) => {
    set({ loading: true })
    try {
      const { data } = await authApi.login({ username, password })
      localStorage.setItem('token', data.access_token)
      set({ token: data.access_token, isAuthenticated: true })
      await get().fetchProfile()
    } finally {
      set({ loading: false })
    }
  },

  register: async (username, password, email) => {
    set({ loading: true })
    try {
      await authApi.register({ username, password, email })
    } finally {
      set({ loading: false })
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null, isAuthenticated: false })
  },

  fetchProfile: async () => {
    try {
      const { data } = await profileApi.getProfile()
      set({ user: data })
    } catch {
      get().logout()
    }
  },

  loadFromStorage: async () => {
    const token = localStorage.getItem('token')
    if (token) {
      set({ token, isAuthenticated: true })
      await get().fetchProfile()
    }
    set({ initialized: true })
  },
}))
