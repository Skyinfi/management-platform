import { useEffect, useMemo, useState } from 'react'
import { authApi } from '../api/auth'
import { apiClient } from '../api/client'
import { clearAuth, getStoredToken, getStoredUser, persistAuth, attachAuthToken } from './authStore'
import type { LoginResponse } from '../api/auth'

export type AuthState = {
  ready: boolean
  authenticated: boolean
  user: LoginResponse['user'] | null
  token: string | null
  error: string | null
  login: (username: string, password: string, remember?: boolean) => Promise<void>
  logout: () => void
}

export function useAuth(): AuthState {
  const [ready, setReady] = useState(false)
  const [authenticated, setAuthenticated] = useState(false)
  const [user, setUser] = useState<LoginResponse['user'] | null>(null)
  const [token, setToken] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    apiClient.setUnauthorizedHandler(() => {
      clearAuth()
      attachAuthToken(null)
      setAuthenticated(false)
      setUser(null)
      setToken(null)
      setError('登录已过期，请重新登录')
    })

    const storedToken = getStoredToken()
    const storedUser = getStoredUser()

    if (storedToken) {
      attachAuthToken(storedToken)
      setToken(storedToken)
    }

    if (storedUser && storedToken) {
      setUser(storedUser)
      setAuthenticated(true)
    }

    async function bootstrap() {
      try {
        if (!storedToken) {
          setReady(true)
          return
        }

        const response = await authApi.me()
        setUser(response.data.user)
        setAuthenticated(true)
        persistAuth(response.data)
        attachAuthToken(response.data.token)
        setToken(response.data.token)
        setError(null)
      } catch (err) {
        clearAuth()
        attachAuthToken(null)
        setAuthenticated(false)
        setUser(null)
        setToken(null)
        setError(err instanceof Error ? err.message : '登录态已失效')
      } finally {
        setReady(true)
      }
    }

    void bootstrap()
  }, [])

  async function login(username: string, password: string, remember = true) {
    const response = await authApi.login({ username, password })
    persistAuth(response.data)
    attachAuthToken(response.data.token)
    setToken(response.data.token)
    setUser(response.data.user)
    setAuthenticated(true)
    setError(null)

    if (!remember) {
      localStorage.removeItem('app-manager-last-username')
      localStorage.setItem('app-manager-remember-me', 'false')
    }
  }

  function logout() {
    clearAuth()
    attachAuthToken(null)
    setAuthenticated(false)
    setUser(null)
    setToken(null)
  }

  return useMemo(
    () => ({ ready, authenticated, user, token, error, login, logout }),
    [ready, authenticated, user, token, error],
  )
}
