const DEFAULT_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api'

let authToken: string | null = null
let onUnauthorized: (() => void) | null = null

export class ApiError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

export type RequestOptions = Omit<RequestInit, 'body'> & {
  body?: unknown
}

export function setAuthToken(token: string | null) {
  authToken = token
}

export function setUnauthorizedHandler(handler: (() => void) | null) {
  onUnauthorized = handler
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const headers = new Headers(options.headers)

  if (authToken) {
    headers.set('Authorization', `Bearer ${authToken}`)
  }

  if (options.body !== undefined && !(options.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(`${DEFAULT_BASE_URL}${path}`, {
    ...options,
    headers,
    body: options.body === undefined || options.body instanceof FormData ? options.body : JSON.stringify(options.body),
  })

  const text = await response.text()
  let payload: any = null

  if (text) {
    try {
      payload = JSON.parse(text)
    } catch {
      payload = { message: text }
    }
  }

  if (!response.ok) {
    if (response.status === 401) {
      onUnauthorized?.()
    }
    throw new ApiError(payload?.message ?? '请求失败', response.status)
  }

  return payload as T
}

export const apiClient = {
  setAuthToken,
  setUnauthorizedHandler,
  get: <T>(path: string, options?: Omit<RequestOptions, 'body'>) => request<T>(path, { ...options, method: 'GET' }),
  post: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'body'>) =>
    request<T>(path, { ...options, method: 'POST', body }),
  put: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'body'>) =>
    request<T>(path, { ...options, method: 'PUT', body }),
  patch: <T>(path: string, body?: unknown, options?: Omit<RequestOptions, 'body'>) =>
    request<T>(path, { ...options, method: 'PATCH', body }),
  delete: <T>(path: string, options?: Omit<RequestOptions, 'body'>) =>
    request<T>(path, { ...options, method: 'DELETE' }),
}
