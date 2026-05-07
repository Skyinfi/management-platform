import { apiClient } from '../api'
import type { LoginResponse } from '../api/auth'
import type { ApiResponse } from '../api/types'

const STORAGE_KEY = 'app-manager-token'
const USER_KEY = 'app-manager-user'

export type AuthUser = LoginResponse['user']

export function getStoredToken() {
  return localStorage.getItem(STORAGE_KEY)
}

export function getStoredUser(): AuthUser | null {
  const raw = localStorage.getItem(USER_KEY)
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as AuthUser
  } catch {
    return null
  }
}

export function persistAuth(payload: LoginResponse) {
  localStorage.setItem(STORAGE_KEY, payload.token)
  localStorage.setItem(USER_KEY, JSON.stringify(payload.user))
}

export function clearAuth() {
  localStorage.removeItem(STORAGE_KEY)
  localStorage.removeItem(USER_KEY)
}

export function attachAuthToken(token: string | null) {
  apiClient.setAuthToken(token)
}

export async function fetchCurrentUser() {
  return apiClient.get<ApiResponse<LoginResponse>>('/auth/me')
}
