import { apiClient } from './client'
import type { ApiResponse } from './types'

export type LoginRequest = {
  username: string
  password: string
}

export type LoginResponse = {
  token: string
  expiresAt: string
  user: {
    id: string
    name: string
    role: string
  }
}

export const authApi = {
  login: (body: LoginRequest) => apiClient.post<ApiResponse<LoginResponse>>('/auth/login', body),
  me: () => apiClient.get<ApiResponse<LoginResponse>>('/auth/me'),
}
