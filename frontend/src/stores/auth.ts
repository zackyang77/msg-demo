import { defineStore } from 'pinia'
import { authApi, setAuthToken, type AuthRequest, type AuthResponse, type User } from '@/api/client'

const TOKEN_KEY = 'msgdemo_token'
const USER_KEY = 'msgdemo_user'

interface AuthState {
  user: User | null
  token: string
  loading: boolean
  error: string | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    user: null,
    token: '',
    loading: false,
    error: null
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token && state.user)
  },
  actions: {
    initFromStorage() {
      const token = localStorage.getItem(TOKEN_KEY)
      const userRaw = localStorage.getItem(USER_KEY)
      if (token && userRaw) {
        try {
          this.user = JSON.parse(userRaw) as User
          this.token = token
          setAuthToken(token)
        } catch (error) {
          console.warn('无法解析本地用户信息', error)
          this.clearSession()
        }
      }
    },
    async register(payload: AuthRequest) {
      return this.authenticate(() => authApi.register(payload))
    },
    async login(payload: AuthRequest) {
      return this.authenticate(() => authApi.login(payload))
    },
    logout() {
      this.clearSession()
    },
    async authenticate(fn: () => Promise<AuthResponse>) {
      this.loading = true
      this.error = null
      try {
        const result = await fn()
        this.user = result.user
        this.token = result.token
        setAuthToken(result.token)
        localStorage.setItem(TOKEN_KEY, result.token)
        localStorage.setItem(USER_KEY, JSON.stringify(result.user))
        return result
      } catch (error) {
        this.error = error instanceof Error ? error.message : '认证失败'
        throw error
      } finally {
        this.loading = false
      }
    },
    clearSession() {
      this.user = null
      this.token = ''
      setAuthToken(null)
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    }
  }
})

