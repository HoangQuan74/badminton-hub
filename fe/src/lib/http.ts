import axios, { AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { tokenStorage } from '@/features/auth/auth.storage'
import { useAuthStore } from '@/features/auth/auth.store'
import type { RefreshResponse } from '@/features/auth/auth.types'

const baseURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

export const publicHttp = axios.create({ baseURL, timeout: 15_000 })
export const http = axios.create({ baseURL, timeout: 15_000 })

http.interceptors.request.use((config) => {
  const token = tokenStorage.access()
  config.headers['Accept-Language'] = 'vi'
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

type RetryConfig = InternalAxiosRequestConfig & { _retry?: boolean }
let refreshPromise: Promise<string> | null = null

async function refreshAccessToken() {
  const refreshToken = tokenStorage.refresh()
  if (!refreshToken) throw new Error('Missing refresh token')
  const { data } = await publicHttp.post<RefreshResponse>('/api/v1/auth/refresh', { refresh_token: refreshToken })
  tokenStorage.save(data.tokens)
  return data.tokens.access_token
}

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const request = error.config as RetryConfig | undefined
    if (error.response?.status !== 401 || !request || request._retry) return Promise.reject(error)

    request._retry = true
    try {
      refreshPromise ??= refreshAccessToken().finally(() => { refreshPromise = null })
      const accessToken = await refreshPromise
      request.headers.Authorization = `Bearer ${accessToken}`
      return http(request)
    } catch {
      tokenStorage.clear()
      useAuthStore.getState().reset()
      return Promise.reject(error)
    }
  }
)

export function getApiErrorMessage(error: unknown, fallback: string) {
  if (!axios.isAxiosError(error)) return fallback
  const data = error.response?.data as { error?: { message?: string } } | undefined
  return data?.error?.message || fallback
}
